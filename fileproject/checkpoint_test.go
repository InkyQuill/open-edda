package fileproject

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateCheckpointSnapshotsStableFilesAndIgnoresOperationalDrafts(t *testing.T) {
	root := copyFileProjectFixture(t, filepath.Join("testdata", "partial"))
	if _, err := WriteDraft(root, WriteDraftInput{FileID: "story-1", BodyMarkdown: "operational draft"}); err != nil {
		t.Fatalf("WriteDraft error = %v", err)
	}

	checkpoint, err := CreateCheckpoint(root, CreateCheckpointInput{Message: "first save"})
	if err != nil {
		t.Fatalf("CreateCheckpoint error = %v", err)
	}
	if checkpoint.Message != "first save" {
		t.Fatalf("message = %q", checkpoint.Message)
	}
	if len(checkpoint.Files) != 1 {
		t.Fatalf("checkpoint files = %d, want 1", len(checkpoint.Files))
	}
	if checkpoint.Files[0].Path != "story/chapter-01.md" {
		t.Fatalf("checkpoint path = %q", checkpoint.Files[0].Path)
	}
	if _, err := os.Stat(filepath.Join(root, ".edda", "checkpoints", checkpoint.ID, "files", "story", "chapter-01.md")); err != nil {
		t.Fatalf("snapshot file missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".edda", "checkpoints", checkpoint.ID, "files", ".edda", "drafts", "story-1.md")); !os.IsNotExist(err) {
		t.Fatalf("operational draft was checkpointed, stat error = %v", err)
	}
}

func TestDiffCheckpointReportsWorkingTreeChanges(t *testing.T) {
	root := copyFileProjectFixture(t, filepath.Join("testdata", "partial"))
	checkpoint, err := CreateCheckpoint(root, CreateCheckpointInput{Message: "base"})
	if err != nil {
		t.Fatalf("CreateCheckpoint error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "story", "chapter-01.md"), []byte("# Chapter 1\n\nChanged.\n"), 0o644); err != nil {
		t.Fatalf("write changed file: %v", err)
	}

	diff, err := DiffCheckpoint(root, checkpoint.ID, "")
	if err != nil {
		t.Fatalf("DiffCheckpoint error = %v", err)
	}
	if len(diff) != 1 {
		t.Fatalf("diff length = %d, want 1: %#v", len(diff), diff)
	}
	if diff[0].Status != CheckpointDiffModified || diff[0].Path != "story/chapter-01.md" {
		t.Fatalf("diff entry = %#v", diff[0])
	}
}

func TestRestoreCheckpointRollsBackFilesAndIDs(t *testing.T) {
	root := copyFileProjectFixture(t, filepath.Join("testdata", "partial"))
	checkpoint, err := CreateCheckpoint(root, CreateCheckpointInput{Message: "base"})
	if err != nil {
		t.Fatalf("CreateCheckpoint error = %v", err)
	}
	originalID := checkpoint.Files[0].ID
	if err := os.WriteFile(filepath.Join(root, "story", "chapter-01.md"), []byte("# Chapter 1\n\nChanged.\n"), 0o644); err != nil {
		t.Fatalf("write changed file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "story", "extra.md"), []byte("# Extra\n"), 0o644); err != nil {
		t.Fatalf("write extra file: %v", err)
	}

	if _, err := RestoreCheckpoint(root, checkpoint.ID); err != nil {
		t.Fatalf("RestoreCheckpoint error = %v", err)
	}
	body, err := os.ReadFile(filepath.Join(root, "story", "chapter-01.md"))
	if err != nil {
		t.Fatalf("read restored file: %v", err)
	}
	if strings.Contains(string(body), "Changed") {
		t.Fatalf("file was not restored:\n%s", string(body))
	}
	if _, err := os.Stat(filepath.Join(root, "story", "extra.md")); !os.IsNotExist(err) {
		t.Fatalf("extra file still exists, stat error = %v", err)
	}
	stable, err := ResolveStableFile(root, originalID)
	if err != nil {
		t.Fatalf("ResolveStableFile restored id error = %v", err)
	}
	if stable.Path != "story/chapter-01.md" {
		t.Fatalf("restored id path = %q", stable.Path)
	}
}

func TestListCheckpointsSortsByCreation(t *testing.T) {
	root := copyFileProjectFixture(t, filepath.Join("testdata", "partial"))
	first, err := CreateCheckpoint(root, CreateCheckpointInput{Message: "first"})
	if err != nil {
		t.Fatalf("first checkpoint error = %v", err)
	}
	second, err := CreateCheckpoint(root, CreateCheckpointInput{Message: "second"})
	if err != nil {
		t.Fatalf("second checkpoint error = %v", err)
	}

	checkpoints, err := ListCheckpoints(root)
	if err != nil {
		t.Fatalf("ListCheckpoints error = %v", err)
	}
	if len(checkpoints) != 2 {
		t.Fatalf("checkpoint count = %d", len(checkpoints))
	}
	if checkpoints[0].ID != first.ID || checkpoints[1].ID != second.ID {
		t.Fatalf("checkpoint order = %#v", checkpoints)
	}
}
