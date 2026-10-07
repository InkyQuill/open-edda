package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIgnoreTrackedSendAndTakePreservesLocalFiles(t *testing.T) {
	a, b, _, _ := updatePair(t)
	writeUpdateFile(t, a, ".eddaignore", "chapter.md\n*.log\n")
	writeUpdateFile(t, a, "private.log", "local only")
	writeUpdateFile(t, a, "chapter.md", "updated tracked text")
	runSyncTest(t, "send", a)
	runSyncTest(t, "take", b)
	assertUpdateFile(t, b, "chapter.md", "updated tracked text")
	if _, err := os.Stat(filepath.Join(b, "private.log")); !os.IsNotExist(err) {
		t.Fatal("uploaded ignored log")
	}
	writeUpdateFile(t, b, "chapter.md", "remote change")
	runSyncTest(t, "send", b)
	runSyncTest(t, "take", a)
	assertUpdateFile(t, a, "private.log", "local only")
	assertUpdateFile(t, a, "chapter.md", "remote change")
}

func TestTakeRejectsIncomingIgnoredPath(t *testing.T) {
	a, b, _, _ := updatePair(t)
	writeUpdateFile(t, a, ".eddaignore", "*.log\n")
	writeUpdateFile(t, a, "private.log", "local")
	writeUpdateFile(t, b, "private.log", "remote")
	runSyncTest(t, "send", b)
	var out bytes.Buffer
	if err := run([]string{"take", a}, &out, &out); err == nil || !strings.Contains(err.Error(), "overlap local ignore") {
		t.Fatalf("%v %s", err, out.String())
	}
	assertUpdateFile(t, a, "private.log", "local")
}

func TestInventorySummaryAndAttachProblems(t *testing.T) {
	root := t.TempDir()
	writeUpdateFile(t, root, "draft.md", "text")
	if err := os.Symlink("missing-inside", filepath.Join(root, "bad-link")); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := runImport([]string{root, "--dry-run"}, &out, &out); err == nil {
		t.Fatal("accepted link")
	}
	if !strings.Contains(out.String(), "bad-link") || strings.Contains(out.String(), `file "draft.md"`) {
		t.Fatal(out.String())
	}
	out.Reset()
	_ = runImport([]string{root, "--dry-run", "--verbose"}, &out, &out)
	if !strings.Contains(out.String(), `file "draft.md"`) {
		t.Fatal(out.String())
	}
	server, id, _, _ := importTestServer(t, nil)
	out.Reset()
	err := runAttach([]string{root, "--server", server, "--project", id}, &out)
	if err == nil {
		t.Fatal("accepted link")
	}
	if !strings.Contains(out.String(), "bad-link") {
		t.Fatal(out.String())
	}
}
