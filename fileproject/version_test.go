package fileproject

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateFileVersionTargetAcceptsCurrentHash(t *testing.T) {
	root := copyFileProjectFixture(t, filepath.Join("testdata", "partial"))
	_, files, err := SyncStableIDs(root)
	if err != nil {
		t.Fatalf("ListStableFiles error = %v", err)
	}
	file := files[0]

	validated, err := ValidateFileVersionTarget(root, FileVersionTarget{
		FileID:         file.ID,
		Path:           file.Path,
		Kind:           file.Kind,
		ExpectedSHA256: file.SHA256,
	})
	if err != nil {
		t.Fatalf("ValidateFileVersionTarget error = %v", err)
	}
	if validated.ID != file.ID {
		t.Fatalf("validated id = %q", validated.ID)
	}
}

func TestValidateFileVersionTargetRejectsStaleHash(t *testing.T) {
	root := copyFileProjectFixture(t, filepath.Join("testdata", "partial"))
	_, files, err := SyncStableIDs(root)
	if err != nil {
		t.Fatalf("ListStableFiles error = %v", err)
	}

	_, err = ValidateFileVersionTarget(root, FileVersionTarget{
		FileID:         files[0].ID,
		ExpectedSHA256: "stale",
	})
	if !errors.Is(err, ErrFileConflict) {
		t.Fatalf("ValidateFileVersionTarget error = %v, want ErrFileConflict", err)
	}
}

func TestValidateFileVersionTargetRejectsInvalidByteRanges(t *testing.T) {
	root := copyFileProjectFixture(t, filepath.Join("testdata", "partial"))
	_, files, err := SyncStableIDs(root)
	if err != nil {
		t.Fatalf("ListStableFiles error = %v", err)
	}
	negative := int64(-1)
	tooLarge := files[0].Size + 1

	for name, target := range map[string]FileVersionTarget{
		"negative start": {FileID: files[0].ID, StartByte: &negative},
		"negative end":   {FileID: files[0].ID, EndByte: &negative},
		"oversize start": {FileID: files[0].ID, StartByte: &tooLarge},
		"oversize end":   {FileID: files[0].ID, EndByte: &tooLarge},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := ValidateFileVersionTarget(root, target); err == nil {
				t.Fatalf("ValidateFileVersionTarget accepted %#v", target)
			}
		})
	}
}

func TestListFileCheckpointHistoryFiltersByFileID(t *testing.T) {
	root := copyFileProjectFixture(t, filepath.Join("testdata", "partial"))
	_, files, err := SyncStableIDs(root)
	if err != nil {
		t.Fatalf("ListStableFiles error = %v", err)
	}
	fileID := files[0].ID
	first, err := CreateCheckpoint(root, CreateCheckpointInput{Message: "first"})
	if err != nil {
		t.Fatalf("CreateCheckpoint first error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "story", "chapter-01.md"), []byte("# Chapter 1\n\nChanged.\n"), 0o644); err != nil {
		t.Fatalf("write changed file: %v", err)
	}
	second, err := CreateCheckpoint(root, CreateCheckpointInput{Message: "second"})
	if err != nil {
		t.Fatalf("CreateCheckpoint second error = %v", err)
	}

	history, err := ListFileCheckpointHistory(root, fileID)
	if err != nil {
		t.Fatalf("ListFileCheckpointHistory error = %v", err)
	}
	if len(history) != 2 {
		t.Fatalf("history count = %d", len(history))
	}
	if history[0].CheckpointID != first.ID || history[1].CheckpointID != second.ID {
		t.Fatalf("history order = %#v", history)
	}
	if history[0].SHA256 == history[1].SHA256 {
		t.Fatalf("file history hashes did not change: %#v", history)
	}
}

func TestListFileCheckpointHistoryReturnsEmptySlice(t *testing.T) {
	root := copyFileProjectFixture(t, filepath.Join("testdata", "partial"))
	history, err := ListFileCheckpointHistory(root, "story-missing")
	if err != nil {
		t.Fatalf("ListFileCheckpointHistory error = %v", err)
	}
	if history == nil {
		t.Fatalf("history is nil")
	}
	if len(history) != 0 {
		t.Fatalf("history = %#v", history)
	}
}
