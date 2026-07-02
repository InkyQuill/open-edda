package fileproject

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPreserveConflictStoresVersionsAndLayoutIgnoresThem(t *testing.T) {
	root := copyFileProjectFixture(t, filepath.Join("testdata", "partial"))
	record, err := PreserveConflict(root, PreserveConflictInput{
		FileID:         "story-1",
		Path:           "story/chapter-01.md",
		BaseMarkdown:   "base",
		LocalMarkdown:  "local",
		ServerMarkdown: "server",
	})
	if err != nil {
		t.Fatalf("PreserveConflict error = %v", err)
	}
	if record.LocalSHA256 != HashMarkdown("local") {
		t.Fatalf("local hash = %q", record.LocalSHA256)
	}
	for _, name := range []string{"base.md", "local.md", "server.md", "metadata.json"} {
		if _, err := os.Stat(filepath.Join(root, ".edda", "conflicts", "story-1", name)); err != nil {
			t.Fatalf("conflict file %s missing: %v", name, err)
		}
	}
	layout, err := Scan(root)
	if err != nil {
		t.Fatalf("Scan error = %v", err)
	}
	if len(layout.Files) != 1 {
		t.Fatalf("layout file count = %d, want 1", len(layout.Files))
	}
}

func TestListConflictsExcludesResolvedRecords(t *testing.T) {
	root := copyFileProjectFixture(t, filepath.Join("testdata", "partial"))
	mustWriteIDMap(t, root, map[string]string{"story/chapter-01.md": "story-1"})
	if _, err := PreserveConflict(root, PreserveConflictInput{
		FileID:         "story-1",
		Path:           "story/chapter-01.md",
		BaseMarkdown:   "base",
		LocalMarkdown:  "# Chapter 1\n\nLocal.\n",
		ServerMarkdown: "# Chapter 1\n\nServer.\n",
	}); err != nil {
		t.Fatalf("PreserveConflict error = %v", err)
	}
	conflicts, err := ListConflicts(root)
	if err != nil {
		t.Fatalf("ListConflicts error = %v", err)
	}
	if len(conflicts) != 1 {
		t.Fatalf("conflict count = %d", len(conflicts))
	}
	if _, _, err := ResolveConflict(root, ResolveConflictInput{FileID: "story-1", Use: ConflictVersionLocal}); err != nil {
		t.Fatalf("ResolveConflict error = %v", err)
	}
	conflicts, err = ListConflicts(root)
	if err != nil {
		t.Fatalf("ListConflicts after resolve error = %v", err)
	}
	if len(conflicts) != 0 {
		t.Fatalf("resolved conflicts still listed: %#v", conflicts)
	}
}

func TestResolveConflictWritesCanonicalFile(t *testing.T) {
	root := copyFileProjectFixture(t, filepath.Join("testdata", "partial"))
	mustWriteIDMap(t, root, map[string]string{"story/chapter-01.md": "story-1"})
	if _, err := PreserveConflict(root, PreserveConflictInput{
		FileID:         "story-1",
		Path:           "story/chapter-01.md",
		BaseMarkdown:   "base",
		LocalMarkdown:  "# Chapter 1\n\nLocal.\n",
		ServerMarkdown: "# Chapter 1\n\nServer.\n",
	}); err != nil {
		t.Fatalf("PreserveConflict error = %v", err)
	}

	record, saved, err := ResolveConflict(root, ResolveConflictInput{FileID: "story-1", Use: ConflictVersionServer})
	if err != nil {
		t.Fatalf("ResolveConflict error = %v", err)
	}
	if record.ResolvedAt == nil || record.Resolution != "server" {
		t.Fatalf("resolved record = %#v", record)
	}
	if saved.SHA256 != HashMarkdown("# Chapter 1\n\nServer.\n") {
		t.Fatalf("saved hash = %q", saved.SHA256)
	}
	body, err := os.ReadFile(filepath.Join(root, "story", "chapter-01.md"))
	if err != nil {
		t.Fatalf("read canonical file: %v", err)
	}
	if string(body) != "# Chapter 1\n\nServer.\n" {
		t.Fatalf("canonical body = %q", string(body))
	}
}
