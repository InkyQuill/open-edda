package fileproject

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDraftRoundTripAndDelete(t *testing.T) {
	root := t.TempDir()
	draft, err := WriteDraft(root, WriteDraftInput{
		FileID:       "story-1",
		BasePath:     "story/chapter-01.md",
		BaseSHA256:   "base-hash",
		BodyMarkdown: "# Draft\n\nNew text.\n",
	})
	if err != nil {
		t.Fatalf("WriteDraft error = %v", err)
	}
	if draft.Metadata.FileID != "story-1" {
		t.Fatalf("draft file id = %q", draft.Metadata.FileID)
	}

	read, err := ReadDraft(root, "story-1")
	if err != nil {
		t.Fatalf("ReadDraft error = %v", err)
	}
	if read.Metadata.BasePath != "story/chapter-01.md" {
		t.Fatalf("draft base path = %q", read.Metadata.BasePath)
	}
	if read.BodyMarkdown != "# Draft\n\nNew text.\n" {
		t.Fatalf("draft body = %q", read.BodyMarkdown)
	}

	if err := DeleteDraft(root, "story-1"); err != nil {
		t.Fatalf("DeleteDraft error = %v", err)
	}
	if _, err := ReadDraft(root, "story-1"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("ReadDraft after delete error = %v, want fs.ErrNotExist", err)
	}
}

func TestDraftRejectsPathTraversalFileID(t *testing.T) {
	root := t.TempDir()
	for _, fileID := range []string{"../story-1", "story/1", ""} {
		_, err := WriteDraft(root, WriteDraftInput{FileID: fileID, BodyMarkdown: "text"})
		if !errors.Is(err, ErrInvalidFileID) {
			t.Fatalf("WriteDraft(%q) error = %v, want ErrInvalidFileID", fileID, err)
		}
	}
}

func TestSaveCanonicalFileWritesResolvedStableFile(t *testing.T) {
	root := copyFileProjectFixture(t, filepath.Join("testdata", "partial"))
	mustWriteIDMap(t, root, map[string]string{"story/chapter-01.md": "story-1"})
	before := mustLayoutFile(t, root, "story/chapter-01.md")

	saved, err := SaveCanonicalFile(root, SaveCanonicalInput{
		FileID:         "story-1",
		BodyMarkdown:   "# Chapter 1\n\nSaved text.\n",
		ExpectedSHA256: before.SHA256,
	})
	if err != nil {
		t.Fatalf("SaveCanonicalFile error = %v", err)
	}
	if saved.ID != "story-1" || saved.Path != "story/chapter-01.md" {
		t.Fatalf("saved file = %#v", saved)
	}
	if saved.SHA256 != HashMarkdown("# Chapter 1\n\nSaved text.\n") {
		t.Fatalf("saved hash = %q", saved.SHA256)
	}
	data, err := os.ReadFile(filepath.Join(root, "story", "chapter-01.md"))
	if err != nil {
		t.Fatalf("read saved file: %v", err)
	}
	if string(data) != "# Chapter 1\n\nSaved text.\n" {
		t.Fatalf("saved body = %q", string(data))
	}
}

func TestSaveCanonicalFileRejectsStaleHash(t *testing.T) {
	root := copyFileProjectFixture(t, filepath.Join("testdata", "partial"))
	mustWriteIDMap(t, root, map[string]string{"story/chapter-01.md": "story-1"})

	_, err := SaveCanonicalFile(root, SaveCanonicalInput{
		FileID:         "story-1",
		BodyMarkdown:   "stale overwrite",
		ExpectedSHA256: strings.Repeat("0", 64),
	})
	if !errors.Is(err, ErrFileConflict) {
		t.Fatalf("SaveCanonicalFile error = %v, want ErrFileConflict", err)
	}
	data, err := os.ReadFile(filepath.Join(root, "story", "chapter-01.md"))
	if err != nil {
		t.Fatalf("read original file: %v", err)
	}
	if strings.Contains(string(data), "stale overwrite") {
		t.Fatalf("stale save changed file:\n%s", string(data))
	}
}

func TestPromoteDraftSavesAndDeletesDraft(t *testing.T) {
	root := copyFileProjectFixture(t, filepath.Join("testdata", "partial"))
	mustWriteIDMap(t, root, map[string]string{"story/chapter-01.md": "story-1"})
	before := mustLayoutFile(t, root, "story/chapter-01.md")
	if _, err := WriteDraft(root, WriteDraftInput{
		FileID:       "story-1",
		BasePath:     before.Path,
		BaseSHA256:   before.SHA256,
		BodyMarkdown: "# Chapter 1\n\nPromoted draft.\n",
	}); err != nil {
		t.Fatalf("WriteDraft error = %v", err)
	}

	saved, err := PromoteDraft(root, SaveDraftInput{FileID: "story-1"})
	if err != nil {
		t.Fatalf("PromoteDraft error = %v", err)
	}
	if saved.SHA256 != HashMarkdown("# Chapter 1\n\nPromoted draft.\n") {
		t.Fatalf("promoted hash = %q", saved.SHA256)
	}
	if _, err := ReadDraft(root, "story-1"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("ReadDraft after promote error = %v, want fs.ErrNotExist", err)
	}
}

func mustWriteIDMap(t *testing.T, root string, items map[string]string) {
	t.Helper()
	if err := WriteIDMap(root, IDMap{SchemaVersion: CurrentSchemaVersion, Items: items}); err != nil {
		t.Fatalf("WriteIDMap error = %v", err)
	}
}

func mustLayoutFile(t *testing.T, root string, rel string) LayoutFile {
	t.Helper()
	layout, err := Scan(root)
	if err != nil {
		t.Fatalf("Scan error = %v", err)
	}
	for _, file := range layout.Files {
		if file.Path == rel {
			return file
		}
	}
	t.Fatalf("layout file %q not found", rel)
	return LayoutFile{}
}

func copyFileProjectFixture(t *testing.T, source string) string {
	t.Helper()
	root := t.TempDir()
	if err := filepath.WalkDir(source, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		target := filepath.Join(root, rel)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	}); err != nil {
		t.Fatalf("copy fixture: %v", err)
	}
	return root
}
