package project

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"path"
	"strings"
	"unicode/utf8"
)

// PocketReview is a projection of one pinned chapter and its adjacent sidecar.
// Book roots (not just chapter UUIDs) scope identity within a series project.
type PocketReview struct {
	VersionID string         `json:"versionId"`
	BookRoot  string         `json:"bookRoot"`
	BookID    string         `json:"bookId"`
	Chapter   TreeEntry      `json:"chapter"`
	Sidecar   TreeEntry      `json:"sidecar"`
	Note      string         `json:"note"`
	Source    string         `json:"source"`
	Edits     []PocketRecord `json:"edits"`
	Signals   []PocketRecord `json:"signals"`
	document  pocketDocument
	version   ProjectVersion
}

func (s *VersionStore) pocketBytes(ctx context.Context, author, project string, v ProjectVersion, e TreeEntry) ([]byte, error) {
	if e.Kind != "file" || e.Bytes > 1<<20 {
		return nil, pocketInvalid("review files must be UTF-8 files up to 1 MiB; use raw files for larger documents")
	}
	f, err := s.OpenVersionFile(ctx, author, project, v.ID, e.ID)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, (1<<20)+1))
	if err != nil {
		return nil, err
	}
	if len(data) > 1<<20 || !utf8.Valid(data) {
		return nil, pocketInvalid("invalid or oversized UTF-8 file")
	}
	return data, nil
}

// ReadPocketReview never modifies either file or guesses identity without a manifest.
func (s *VersionStore) ReadPocketReview(ctx context.Context, author, project, version, entryID string) (PocketReview, error) {
	result := PocketReview{}
	v, err := s.Version(ctx, author, project, version)
	if err != nil {
		return result, err
	}
	byPath := map[string]TreeEntry{}
	for _, e := range v.Entries {
		byPath[e.Path] = e
		if e.ID == entryID {
			result.Chapter = e
		}
	}
	chapter := result.Chapter
	if chapter.ID == "" || chapter.Kind != "file" || !strings.HasSuffix(chapter.Path, ".md") {
		return result, sql.ErrNoRows
	}
	root := path.Dir(chapter.Path)
	manifest, ok := byPath[path.Join(root, ".pocket-editor.json")]
	if !ok {
		return result, sql.ErrNoRows
	}
	candidates := []string{strings.TrimSuffix(chapter.Path, ".md") + ".review.json", chapter.Path + ".review.json"}
	for _, name := range candidates {
		if e, ok := byPath[name]; ok {
			if result.Sidecar.ID != "" {
				return result, pocketInvalid("both sidecar filenames exist; reconcile them explicitly")
			}
			result.Sidecar = e
		}
	}
	data, err := s.pocketBytes(ctx, author, project, v, manifest)
	if err != nil {
		return result, err
	}
	bookID, chapterID, err := pocketIdentity(data, path.Base(chapter.Path))
	if err != nil {
		return result, err
	}
	if result.Sidecar.ID == "" {
		result.Sidecar = TreeEntry{ID: "review-" + pocketHash([]byte(chapter.ID)), Path: candidates[0], Kind: "file"}
		for _, e := range v.Entries {
			if e.ID == result.Sidecar.ID {
				return result, pocketInvalid("review identity already occupied")
			}
		}
		data, err = json.Marshal(map[string]any{"schema_version": 1, "chapter_id": chapterID, "source_path": path.Base(chapter.Path), "chapter_note": "", "edits": []any{}, "signals": []any{}})
	} else {
		data, err = s.pocketBytes(ctx, author, project, v, result.Sidecar)
	}
	if err != nil {
		return result, err
	}
	d, err := decodePocket(data)
	if err != nil {
		return result, err
	}
	if d.ChapterID != chapterID || d.SourcePath != path.Base(chapter.Path) {
		return result, pocketInvalid("sidecar identity does not match the manifest chapter")
	}
	source, err := s.pocketBytes(ctx, author, project, v, chapter)
	if err != nil {
		return result, err
	}
	d.classify(source)
	result.VersionID = v.ID
	result.BookRoot = root
	result.BookID = bookID
	result.Source = string(source)
	result.Note = d.Note
	result.Edits = d.Edits
	result.Signals = d.Signals
	result.document = d
	result.version = v
	return result, nil
}
func pocketIdentity(data []byte, name string) (string, string, error) {
	raw, err := pocketObject(data)
	if err != nil {
		return "", "", err
	}
	var version int
	var bookID, title string
	var chapters []json.RawMessage
	var ignored []string
	for _, f := range []struct {
		key      string
		target   any
		required bool
	}{{"schema_version", &version, true}, {"book_id", &bookID, true}, {"chapters", &chapters, false}, {"ignored_files", &ignored, false}} {
		if err = pocketField(raw, f.key, f.target, f.required); err != nil {
			return "", "", err
		}
	}
	if err = pocketField(raw, "title", &title, version == 1); err != nil {
		return "", "", err
	}
	if (version != 1 && version != 2) || !pocketUUID.MatchString(bookID) {
		return "", "", pocketInvalid("invalid book manifest")
	}
	paths, ids := map[string]bool{}, map[string]bool{}
	for _, p := range ignored {
		if !pocketChild(p) || paths[p] {
			return "", "", pocketInvalid("invalid ignored files")
		}
		paths[p] = true
	}
	chapterID := ""
	for _, data := range chapters {
		c, err := pocketObject(data)
		if err != nil {
			return "", "", err
		}
		var id, p, title string
		for _, f := range []struct {
			key      string
			target   any
			required bool
		}{{"id", &id, true}, {"path", &p, true}, {"title", &title, version == 1}} {
			if err = pocketField(c, f.key, f.target, f.required); err != nil {
				return "", "", err
			}
		}
		if !pocketUUID.MatchString(id) || !pocketChild(p) || ids[id] || paths[p] {
			return "", "", pocketInvalid("invalid or duplicate manifest chapter")
		}
		ids[id] = true
		paths[p] = true
		if p == name {
			chapterID = id
		}
	}
	if chapterID == "" {
		return "", "", sql.ErrNoRows
	}
	return bookID, chapterID, nil
}

type PocketAction struct {
	ExpectedVersion string `json:"expectedVersion"`
	OperationID     string `json:"operationId"`
	EntryID         string `json:"entryId"`
	Action          string `json:"action"`
	RecordID        string `json:"recordId"`
	Note            string `json:"note"`
	From            int    `json:"from"`
	To              int    `json:"to"`
	Selected        string `json:"selected"`
	After           string `json:"after"`
	SignalType      string `json:"signalType"`
	Comment         string `json:"comment"`
	ActionVersion   string `json:"actionVersion"`
	Metadata        string `json:"metadata"`
}

// ChangePocketReview publishes the chapter and sidecar together. Rebuilding from
// the pinned base makes a lost-response replay use the same normal Publish receipt.
func (s *VersionStore) ChangePocketReview(ctx context.Context, author, project string, input PocketAction) (ProjectVersion, error) {
	if input.ExpectedVersion == "" || input.OperationID == "" {
		return ProjectVersion{}, pocketInvalid("version and operation ID are required")
	}
	if input.Action == "start" {
		return s.startPocket(ctx, author, project, input)
	}
	if input.Action == "undo" {
		return s.undoPocket(ctx, author, project, input)
	}
	review, err := s.ReadPocketReview(ctx, author, project, input.ExpectedVersion, input.EntryID)
	if err == sql.ErrNoRows && input.Action == "metadata" {
		return s.plainMetadata(ctx, author, project, input)
	}
	if err != nil {
		return ProjectVersion{}, err
	}
	source, data, err := changePocket(review, input)
	if err != nil {
		return ProjectVersion{}, err
	}
	replacements := map[string][]byte{review.Sidecar.ID: data}
	if input.Action == "apply" || input.Action == "metadata" {
		replacements[review.Chapter.ID] = source
	}
	entries := append([]TreeEntry(nil), review.version.Entries...)
	if review.Sidecar.SHA256 == "" && input.Action != "metadata" {
		entries = append(entries, review.Sidecar)
	}
	for i, e := range entries {
		body, ok := replacements[e.ID]
		if !ok {
			continue
		}
		if len(body) > 1<<20 {
			return ProjectVersion{}, pocketInvalid("result exceeds browser review limit")
		}
		hash := pocketHash(body)
		if err = s.UploadObject(ctx, author, project, hash, int64(len(body)), bytes.NewReader(body)); err != nil {
			return ProjectVersion{}, err
		}
		entries[i].SHA256 = hash
		entries[i].Bytes = int64(len(body))
	}
	return s.Publish(ctx, PublishVersionInput{AuthorID: author, ProjectID: project, ExpectedVersion: input.ExpectedVersion, OperationID: input.OperationID, Entries: entries})
}
