package fileproject

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScanClassifiesAlchemistStyleLayout(t *testing.T) {
	layout, err := Scan(filepath.Join("testdata", "alchemist-lite"))
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	counts := CountByKind(layout.Files)
	if counts[LayoutKindStory] != 2 {
		t.Fatalf("story count = %d, want 2", counts[LayoutKindStory])
	}
	if counts[LayoutKindCharacter] != 2 {
		t.Fatalf("character count = %d, want 2", counts[LayoutKindCharacter])
	}
	if counts[LayoutKindWorldbuilding] != 3 {
		t.Fatalf("worldbuilding count = %d, want 3", counts[LayoutKindWorldbuilding])
	}
	if counts[LayoutKindStoryline] != 2 {
		t.Fatalf("storyline count = %d, want 2", counts[LayoutKindStoryline])
	}
	if counts[LayoutKindDraft] != 2 {
		t.Fatalf("draft count = %d, want 2", counts[LayoutKindDraft])
	}
	if counts[LayoutKindGuidance] != 2 {
		t.Fatalf("guidance count = %d, want 2", counts[LayoutKindGuidance])
	}
	if counts[LayoutKindSkill] != 1 {
		t.Fatalf("skill count = %d, want 1", counts[LayoutKindSkill])
	}

	assertHasFile(t, layout, "story/chapter-01.md", LayoutKindStory)
	assertHasFile(t, layout, "worldbuilding/magic/Alchemy.md", LayoutKindWorldbuilding)
	assertHasFile(t, layout, ".agents/skills/story-coach/SKILL.md", LayoutKindSkill)
	if !hasWarning(layout, "missing_metadata", ".edda/project.json") {
		t.Fatalf("warnings missing metadata warning: %#v", layout.Warnings)
	}
	if hasWarningCode(layout, "missing_root") || hasWarningCode(layout, "missing_index") {
		t.Fatalf("complete fixture should not have root/index warnings: %#v", layout.Warnings)
	}
}

func TestScanWarnsForPartialLayout(t *testing.T) {
	layout, err := Scan(filepath.Join("testdata", "partial"))
	if err != nil {
		t.Fatalf("Scan(partial) error = %v", err)
	}

	if CountByKind(layout.Files)[LayoutKindStory] != 1 {
		t.Fatalf("partial story count = %d, want 1", CountByKind(layout.Files)[LayoutKindStory])
	}
	if !hasWarning(layout, "missing_root", "characters") {
		t.Fatalf("warnings missing characters root: %#v", layout.Warnings)
	}
	if !hasWarning(layout, "missing_index", "story/_index.md") {
		t.Fatalf("warnings missing story index: %#v", layout.Warnings)
	}
}

func TestScanIgnoresOperationalNoise(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, root, "story/_index.md", "# Story\n")
	mustWrite(t, root, "story/chapter-01.md", "# Chapter\n")
	mustWrite(t, root, ".DS_Store", "noise")
	mustWrite(t, root, ".edda/state.local.json", "{}")
	mustWrite(t, root, ".edda/drafts/chapter-01.md", "draft")
	mustWrite(t, root, ".edda/conflicts/chapter-01.local.md", "conflict")

	layout, err := Scan(root)
	if err != nil {
		t.Fatalf("Scan(noise) error = %v", err)
	}

	for _, file := range layout.Files {
		if strings.Contains(file.Path, ".DS_Store") || strings.Contains(file.Path, ".edda/drafts") || strings.Contains(file.Path, ".edda/conflicts") {
			t.Fatalf("operational noise was scanned: %#v", file)
		}
	}
}

func TestInitAndReadMetadata(t *testing.T) {
	root := t.TempDir()
	metadata, err := InitMetadata(root, InitMetadataInput{
		ID:    "project-1",
		Title: "Alchemy Draft",
	})
	if err != nil {
		t.Fatalf("InitMetadata() error = %v", err)
	}
	if metadata.SchemaVersion != CurrentSchemaVersion || metadata.LayoutVersion != CurrentLayoutVersion {
		t.Fatalf("metadata versions = %#v", metadata)
	}

	read, err := ReadMetadata(root)
	if err != nil {
		t.Fatalf("ReadMetadata() error = %v", err)
	}
	if read.ID != "project-1" || read.Title != "Alchemy Draft" {
		t.Fatalf("read metadata = %#v", read)
	}

	layout, err := Scan(root)
	if err != nil {
		t.Fatalf("Scan(metadata root) error = %v", err)
	}
	if layout.Metadata == nil || layout.Metadata.ID != "project-1" {
		t.Fatalf("layout metadata = %#v", layout.Metadata)
	}
	assertHasFile(t, layout, ".edda/project.json", LayoutKindMetadata)
}

func TestReadMetadataMissingAndMalformed(t *testing.T) {
	_, err := ReadMetadata(t.TempDir())
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("ReadMetadata(missing) error = %v, want fs.ErrNotExist", err)
	}

	_, err = ReadMetadata(filepath.Join("testdata", "invalid"))
	if err == nil || !strings.Contains(err.Error(), "parse project metadata") {
		t.Fatalf("ReadMetadata(invalid) error = %v, want parse error", err)
	}
	_, err = Scan(filepath.Join("testdata", "invalid"))
	if err == nil || !strings.Contains(err.Error(), "parse project metadata") {
		t.Fatalf("Scan(invalid) error = %v, want parse error", err)
	}
}

func assertHasFile(t *testing.T, layout ProjectLayout, path string, kind LayoutKind) {
	t.Helper()
	for _, file := range layout.Files {
		if file.Path == path && file.Kind == kind {
			return
		}
	}
	t.Fatalf("file %s with kind %s not found in %#v", path, kind, layout.Files)
}

func hasWarning(layout ProjectLayout, code string, path string) bool {
	for _, warning := range layout.Warnings {
		if warning.Code == code && warning.Path == path {
			return true
		}
	}
	return false
}

func hasWarningCode(layout ProjectLayout, code string) bool {
	for _, warning := range layout.Warnings {
		if warning.Code == code {
			return true
		}
	}
	return false
}

func mustWrite(t *testing.T, root string, rel string, body string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", rel, err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
}
