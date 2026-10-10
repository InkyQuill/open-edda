package project

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"path"
	"strings"
)

func pocketStableUUID(seed string) string {
	h := pocketHash([]byte(seed))
	return h[:8] + "-" + h[8:12] + "-5" + h[13:16] + "-a" + h[17:20] + "-" + h[20:32]
}

// Starting a review is explicit. Existing manifests retain their schema, IDs,
// ordering and unknown metadata; no author content or folder is created/moved.
func (s *VersionStore) startPocket(ctx context.Context, author, project string, input PocketAction) (ProjectVersion, error) {
	v, err := s.Version(ctx, author, project, input.ExpectedVersion)
	if err != nil {
		return ProjectVersion{}, err
	}
	var chapter, manifest TreeEntry
	for _, e := range v.Entries {
		if e.ID == input.EntryID {
			chapter = e
		}
	}
	if chapter.Kind != "file" || !strings.HasSuffix(chapter.Path, ".md") {
		return ProjectVersion{}, pocketInvalid("select a Markdown file")
	}
	if _, err = s.pocketBytes(ctx, author, project, v, chapter); err != nil {
		return ProjectVersion{}, err
	}
	manifestPath := path.Join(path.Dir(chapter.Path), ".pocket-editor.json")
	for _, e := range v.Entries {
		if e.Path == manifestPath {
			manifest = e
		}
	}
	var raw pocketJSON
	name := path.Base(chapter.Path)
	schema := 2
	if manifest.ID != "" {
		data, e := s.pocketBytes(ctx, author, project, v, manifest)
		if e != nil {
			return ProjectVersion{}, e
		}
		if _, _, e = pocketIdentity(data, name); e != nil && e != sql.ErrNoRows {
			return ProjectVersion{}, e
		}
		raw, err = pocketObject(data)
		if err != nil {
			return ProjectVersion{}, err
		}
		if err = pocketField(raw, "schema_version", &schema, true); err != nil {
			return ProjectVersion{}, err
		}
	} else {
		raw = pocketJSON{}
		_ = pocketSet(raw, "schema_version", 2)
		_ = pocketSet(raw, "book_id", pocketStableUUID(project+":"+manifestPath))
		manifest = TreeEntry{ID: "manifest-" + pocketHash([]byte(manifestPath)), Path: manifestPath, Kind: "file"}
		for _, e := range v.Entries {
			if e.ID == manifest.ID {
				return ProjectVersion{}, pocketInvalid("manifest identity already occupied")
			}
		}
	}
	var chapters []pocketJSON
	if err = pocketField(raw, "chapters", &chapters, false); err != nil {
		return ProjectVersion{}, err
	}
	found := false
	for _, c := range chapters {
		var p string
		_ = pocketField(c, "path", &p, true)
		if p == name {
			found = true
		}
	}
	if !found {
		// Do not adopt an unbound existing review by guessing its chapter identity.
		for _, e := range v.Entries {
			if e.Path == strings.TrimSuffix(chapter.Path, ".md")+".review.json" || e.Path == chapter.Path+".review.json" {
				return ProjectVersion{}, pocketInvalid("existing sidecar needs an explicit matching chapter identity")
			}
		}
		c := pocketJSON{}
		_ = pocketSet(c, "id", pocketStableUUID(project+":"+chapter.ID))
		_ = pocketSet(c, "path", name)
		if schema == 1 {
			_ = pocketSet(c, "title", strings.TrimSuffix(name, ".md"))
		}
		chapters = append(chapters, c)
		_ = pocketSet(raw, "chapters", chapters)
		var ignored []string
		if err = pocketField(raw, "ignored_files", &ignored, false); err != nil {
			return ProjectVersion{}, err
		}
		if _, exists := raw["ignored_files"]; exists {
			next := []string{}
			for _, p := range ignored {
				if p != name {
					next = append(next, p)
				}
			}
			_ = pocketSet(raw, "ignored_files", next)
		}
	}
	data, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return ProjectVersion{}, err
	}
	data = append(data, '\n')
	if len(data) > 1<<20 {
		return ProjectVersion{}, pocketInvalid("manifest exceeds browser limit")
	}
	if _, _, err = pocketIdentity(data, name); err != nil {
		return ProjectVersion{}, err
	}
	manifest.SHA256 = pocketHash(data)
	manifest.Bytes = int64(len(data))
	if err = s.UploadObject(ctx, author, project, manifest.SHA256, manifest.Bytes, bytes.NewReader(data)); err != nil {
		return ProjectVersion{}, err
	}
	entries := []TreeEntry{}
	for _, e := range v.Entries {
		if e.ID != manifest.ID {
			entries = append(entries, e)
		}
	}
	entries = append(entries, manifest)
	return s.Publish(ctx, PublishVersionInput{AuthorID: author, ProjectID: project, ExpectedVersion: input.ExpectedVersion, OperationID: input.OperationID, Entries: entries})
}
