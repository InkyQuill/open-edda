package fileproject

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAssignStableIDsPersistsAcrossScans(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, root, "story/_index.md", "# Story\n")
	mustWrite(t, root, "story/chapter-01.md", "# Chapter\n")

	firstLayout, err := Scan(root)
	if err != nil {
		t.Fatalf("Scan(first) error = %v", err)
	}
	firstMap, firstFiles, err := AssignStableIDs(root, firstLayout)
	if err != nil {
		t.Fatalf("AssignStableIDs(first) error = %v", err)
	}
	if err := WriteIDMap(root, firstMap); err != nil {
		t.Fatalf("WriteIDMap(first) error = %v", err)
	}

	secondLayout, err := Scan(root)
	if err != nil {
		t.Fatalf("Scan(second) error = %v", err)
	}
	secondMap, secondFiles, err := AssignStableIDs(root, secondLayout)
	if err != nil {
		t.Fatalf("AssignStableIDs(second) error = %v", err)
	}

	if firstMap.Items["story/chapter-01.md"] == "" {
		t.Fatalf("missing first chapter id: %#v", firstMap.Items)
	}
	if secondMap.Items["story/chapter-01.md"] != firstMap.Items["story/chapter-01.md"] {
		t.Fatalf("chapter id changed: first=%q second=%q", firstMap.Items["story/chapter-01.md"], secondMap.Items["story/chapter-01.md"])
	}
	if len(firstFiles) != len(secondFiles) {
		t.Fatalf("stable file count changed: first=%d second=%d", len(firstFiles), len(secondFiles))
	}
}

func TestAssignStableIDsAddsAndPrunesPaths(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, root, "story/chapter-01.md", "# Chapter 1\n")
	mustWrite(t, root, "story/chapter-02.md", "# Chapter 2\n")

	layout, err := Scan(root)
	if err != nil {
		t.Fatalf("Scan(first) error = %v", err)
	}
	idMap, _, err := AssignStableIDs(root, layout)
	if err != nil {
		t.Fatalf("AssignStableIDs(first) error = %v", err)
	}
	if err := WriteIDMap(root, idMap); err != nil {
		t.Fatalf("WriteIDMap(first) error = %v", err)
	}
	chapterOneID := idMap.Items["story/chapter-01.md"]

	if err := os.Remove(filepath.Join(root, "story", "chapter-02.md")); err != nil {
		t.Fatalf("remove chapter 2: %v", err)
	}
	mustWrite(t, root, "characters/Protagonist.md", "# Protagonist\n")

	layout, err = Scan(root)
	if err != nil {
		t.Fatalf("Scan(second) error = %v", err)
	}
	idMap, _, err = AssignStableIDs(root, layout)
	if err != nil {
		t.Fatalf("AssignStableIDs(second) error = %v", err)
	}

	if idMap.Items["story/chapter-01.md"] != chapterOneID {
		t.Fatalf("chapter 1 id changed")
	}
	if _, ok := idMap.Items["story/chapter-02.md"]; ok {
		t.Fatalf("deleted chapter retained in ids: %#v", idMap.Items)
	}
	if idMap.Items["characters/Protagonist.md"] == "" {
		t.Fatalf("new character missing id: %#v", idMap.Items)
	}
}

func TestAssignStableIDsExcludesEddaOperationalFiles(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, root, "story/chapter-01.md", "# Chapter 1\n")
	if _, err := InitMetadata(root, InitMetadataInput{ID: "project-1", Title: "Draft"}); err != nil {
		t.Fatalf("InitMetadata error = %v", err)
	}
	mustWrite(t, root, ".edda/state.local.json", "{}")
	mustWrite(t, root, ".edda/drafts/story-1.md", "draft")
	mustWrite(t, root, ".edda/conflicts/story-1/local.md", "local")

	layout, err := Scan(root)
	if err != nil {
		t.Fatalf("Scan error = %v", err)
	}
	idMap, files, err := AssignStableIDs(root, layout)
	if err != nil {
		t.Fatalf("AssignStableIDs error = %v", err)
	}
	if len(files) != 1 || files[0].Path != "story/chapter-01.md" {
		t.Fatalf("stable files = %#v", files)
	}
	for path := range idMap.Items {
		if strings.HasPrefix(path, ".edda/") {
			t.Fatalf("operational path received stable id: %#v", idMap.Items)
		}
	}
}

func TestReadIDMapMissing(t *testing.T) {
	_, err := ReadIDMap(t.TempDir())
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("ReadIDMap(missing) error = %v, want fs.ErrNotExist", err)
	}
}

func TestReadIDMapRejectsUnsupportedSchemaVersion(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".edda"), 0o755); err != nil {
		t.Fatalf("create .edda: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".edda", "ids.json"), []byte(`{"schemaVersion":999,"items":{}}`), 0o644); err != nil {
		t.Fatalf("write ids map: %v", err)
	}

	if _, err := ReadIDMap(root); err == nil {
		t.Fatalf("ReadIDMap accepted unsupported schema version")
	}
}
