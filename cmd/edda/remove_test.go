package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRemoveKeepsLocalDataPublishesDeletionAndCanUndo(t *testing.T) {
	a, b, _, _ := updatePair(t)
	original, err := readCheckout(a)
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(filepath.Join(a, "kb"))
	runSyncTest(t, "rm", "person.md")
	assertUpdateFile(t, a, "kb/person.md", "base person")
	// Rejected batches must not persist their earlier, valid removals.
	var out bytes.Buffer
	if err := run([]string{"rm", "../chapter.md", "../../outside"}, &out, &out); err == nil {
		t.Fatal("outside path accepted")
	}
	state, err := readCheckout(a)
	if err != nil || len(state.Untracked) != 1 {
		t.Fatalf("partial batch persisted: %+v %v", state.Untracked, err)
	}
	runSyncTest(t, "send", a)
	runSyncTest(t, "take", b)
	assertUpdateFile(t, a, "kb/person.md", "base person")
	if _, err := os.Stat(filepath.Join(b, "kb/person.md")); !os.IsNotExist(err) {
		t.Fatal("remote deletion not received")
	}
	history := filepath.Join(t.TempDir(), "old")
	runSyncTest(t, "get", history, "--server", original.Server, "--project", original.Base.ProjectID, "--version", original.Base.ID)
	assertUpdateFile(t, history, "kb/person.md", "base person")
	runSyncTest(t, "send", a) // must not re-add the locally retained file
	runSyncTest(t, "rm", "--undo", "person.md")
	runSyncTest(t, "send", a)
	runSyncTest(t, "take", b)
	assertUpdateFile(t, b, "kb/person.md", "base person")
}

func TestRemoveDirectorySurvivesDisjointRemoteUpdate(t *testing.T) {
	a, b, _, _ := updatePair(t)
	t.Chdir(a)
	runSyncTest(t, "rm", "kb")
	writeUpdateFile(t, b, "notes.md", "remote notes")
	runSyncTest(t, "send", b)
	runSyncTest(t, "take", a)
	runSyncTest(t, "send", a)
	assertUpdateFile(t, a, "kb/person.md", "base person")
	assertUpdateFile(t, a, "notes.md", "remote notes")
	runSyncTest(t, "take", b)
	if _, err := os.Stat(filepath.Join(b, "kb")); !os.IsNotExist(err) {
		t.Fatal("remote directory retained")
	}
}

func TestRemoveRefusesToOverwriteConcurrentChanges(t *testing.T) {
	a, b, _, _ := updatePair(t)
	t.Chdir(a)
	runSyncTest(t, "rm", "kb")
	writeUpdateFile(t, b, "kb/person.md", "remote edit")
	runSyncTest(t, "send", b)
	var out bytes.Buffer
	if err := run([]string{"take", a}, &out, &out); err == nil || !strings.Contains(err.Error(), "rm --undo") {
		t.Fatalf("%v %s", err, out.String())
	}
	assertUpdateFile(t, a, "kb/person.md", "base person")
	runSyncTest(t, "rm", "--undo", "kb")
	runSyncTest(t, "take", a)
	assertUpdateFile(t, a, "kb/person.md", "remote edit")
}
