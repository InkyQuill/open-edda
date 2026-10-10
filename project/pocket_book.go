package project

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"path"
	"strings"
)

type PocketBookChapter struct {
	ID    string  `json:"id"`
	Path  string  `json:"path"`
	Title *string `json:"title,omitempty"`
}
type PocketBookAction struct {
	ExpectedVersion string              `json:"expectedVersion"`
	OperationID     string              `json:"operationId"`
	EntryID         string              `json:"entryId"`
	Title           string              `json:"title"`
	Chapters        []PocketBookChapter `json:"chapters"`
	CreatePaths     []string            `json:"createPaths,omitempty"`
}

// ChangePocketBook preserves raw unknown wire fields, including numbers that a
// JavaScript client cannot represent exactly. Missing chapters stay in the spine.
func (s *VersionStore) ChangePocketBook(ctx context.Context, author, project string, input PocketBookAction) (ProjectVersion, error) {
	if input.ExpectedVersion == "" || input.OperationID == "" || strings.TrimSpace(input.Title) == "" || strings.ContainsAny(input.Title, "\r\n") {
		return ProjectVersion{}, pocketInvalid("version, operation and one-line book title are required")
	}
	v, err := s.Version(ctx, author, project, input.ExpectedVersion)
	if err != nil {
		return ProjectVersion{}, err
	}
	var manifest TreeEntry
	byPath := map[string]TreeEntry{}
	for _, e := range v.Entries {
		byPath[e.Path] = e
		if e.ID == input.EntryID {
			manifest = e
		}
	}
	if manifest.Kind != "file" || path.Base(manifest.Path) != ".pocket-editor.json" {
		return ProjectVersion{}, pocketInvalid("select a book manifest")
	}
	data, err := s.pocketBytes(ctx, author, project, v, manifest)
	if err != nil {
		return ProjectVersion{}, err
	}
	if _, _, err = pocketIdentity(data, ""); err != nil && err != sql.ErrNoRows {
		return ProjectVersion{}, err
	}
	raw, err := pocketObject(data)
	if err != nil {
		return ProjectVersion{}, err
	}
	var existing []pocketJSON
	if err = pocketField(raw, "chapters", &existing, false); err != nil {
		return ProjectVersion{}, err
	}
	byID := map[string]pocketJSON{}
	for _, c := range existing {
		var id string
		_ = pocketField(c, "id", &id, true)
		byID[id] = c
	}
	create := map[string]bool{}
	for _, name := range input.CreatePaths {
		if create[name] {
			return ProjectVersion{}, pocketInvalid("duplicate creation path")
		}
		create[name] = true
	}
	created := map[string][]byte{}
	entries := append([]TreeEntry{}, v.Entries...)
	chapters := []pocketJSON{}
	sidecars := map[string]bool{}
	for _, c := range input.Chapters {
		if !pocketUUID.MatchString(c.ID) || !pocketChild(c.Path) || !strings.HasSuffix(c.Path, ".md") || (c.Title != nil && strings.ContainsAny(*c.Title, "\r\n")) {
			return ProjectVersion{}, pocketInvalid("invalid chapter")
		}
		for _, name := range []string{strings.TrimSuffix(c.Path, ".md") + ".review.json", c.Path + ".review.json"} {
			if sidecars[name] {
				return ProjectVersion{}, pocketInvalid("chapter sidecar names overlap")
			}
			sidecars[name] = true
		}
		row, known := byID[c.ID]
		if known {
			if create[c.Path] {
				return ProjectVersion{}, pocketInvalid("chapter is already in book")
			}
			var original string
			_ = pocketField(row, "path", &original, true)
			if original != c.Path {
				return ProjectVersion{}, pocketInvalid("chapter paths cannot be reassigned")
			}
		} else {
			full := path.Join(path.Dir(manifest.Path), c.Path)
			file, exists := byPath[full]
			if create[c.Path] {
				if exists || strings.HasPrefix(c.Path, ".") || c.Title == nil || strings.TrimSpace(*c.Title) == "" {
					return ProjectVersion{}, pocketInvalid("new chapter needs an unused filename and title")
				}
				for _, name := range []string{strings.TrimSuffix(full, ".md") + ".review.json", full + ".review.json"} {
					if _, occupied := byPath[name]; occupied {
						return ProjectVersion{}, pocketInvalid("new chapter sidecar name is occupied")
					}
				}
				id := "book-chapter-" + pocketHash([]byte(manifest.ID+":"+c.ID))
				for _, entry := range entries {
					if entry.ID == id {
						return ProjectVersion{}, pocketInvalid("new chapter identity is occupied")
					}
				}
				body := []byte("# " + strings.TrimSpace(*c.Title) + "\n")
				created[id] = body
				entries = append(entries, TreeEntry{ID: id, Path: full, Kind: "file", SHA256: pocketHash(body), Bytes: int64(len(body))})
				delete(create, c.Path)
			} else {
				if !exists || file.Kind != "file" {
					return ProjectVersion{}, pocketInvalid("new chapter file does not exist")
				}
				if _, err := s.pocketBytes(ctx, author, project, v, file); err != nil {
					return ProjectVersion{}, err
				}
				found := false
				for _, name := range []string{strings.TrimSuffix(full, ".md") + ".review.json", full + ".review.json"} {
					if sidecar, exists := byPath[name]; exists {
						if found {
							return ProjectVersion{}, pocketInvalid("two sidecars exist for chapter")
						}
						found = true
						data, err := s.pocketBytes(ctx, author, project, v, sidecar)
						if err != nil {
							return ProjectVersion{}, err
						}
						review, err := decodePocket(data)
						if err != nil {
							return ProjectVersion{}, err
						}
						if review.ChapterID != c.ID || review.SourcePath != c.Path {
							return ProjectVersion{}, pocketInvalid("review identity does not match added chapter")
						}
					}
				}
			}
			row = pocketJSON{}
			_ = pocketSet(row, "id", c.ID)
			_ = pocketSet(row, "path", c.Path)
		}
		if c.Title != nil {
			_ = pocketSet(row, "title", *c.Title)
		}
		chapters = append(chapters, row)
	}
	if len(create) != 0 {
		return ProjectVersion{}, pocketInvalid("creation path is not a new chapter")
	}
	_ = pocketSet(raw, "title", strings.TrimSpace(input.Title))
	_ = pocketSet(raw, "chapters", chapters)
	data, err = json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return ProjectVersion{}, err
	}
	data = append(data, '\n')
	if len(data) > 1<<20 {
		return ProjectVersion{}, pocketInvalid("manifest exceeds browser limit")
	}
	if _, _, err = pocketIdentity(data, ""); err != nil && err != sql.ErrNoRows {
		return ProjectVersion{}, err
	}
	hash := pocketHash(data)
	if err = s.UploadObject(ctx, author, project, hash, int64(len(data)), bytes.NewReader(data)); err != nil {
		return ProjectVersion{}, err
	}
	for _, body := range created {
		if err = s.UploadObject(ctx, author, project, pocketHash(body), int64(len(body)), bytes.NewReader(body)); err != nil {
			return ProjectVersion{}, err
		}
	}
	for i := range entries {
		if entries[i].ID == manifest.ID {
			entries[i].SHA256 = hash
			entries[i].Bytes = int64(len(data))
		}
	}
	return s.Publish(ctx, PublishVersionInput{AuthorID: author, ProjectID: project, ExpectedVersion: input.ExpectedVersion, OperationID: input.OperationID, Entries: entries})
}
