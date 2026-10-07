package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/InkyQuill/open-edda/project"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateAttachMoveRoundTrip(t *testing.T) {
	server, _, versions, _ := importTestServer(t, nil)
	created := runSyncTest(t, "create", "--json", "--title", "Flexible", "--server", server)
	var p project.StoryProject
	if err := json.Unmarshal([]byte(created), &p); err != nil {
		t.Fatal(err)
	}
	a := t.TempDir()
	writeUpdateFile(t, a, "draft.md", "local")
	runSyncTest(t, "attach", a, "--server", server, "--project", p.ID)
	runSyncTest(t, "send", a)
	first, err := versions.Version(context.Background(), "author", p.ID, "")
	if err != nil {
		t.Fatal(err)
	}
	b := filepath.Join(t.TempDir(), "b")
	runSyncTest(t, "get", b, "--server", server, "--project", p.ID)
	runSyncTest(t, "move", a, "--from", "draft.md", "--to", "chapter.md")
	runSyncTest(t, "send", a)
	runSyncTest(t, "take", b)
	second, err := versions.Version(context.Background(), "author", p.ID, "")
	if err != nil {
		t.Fatal(err)
	}
	if first.Entries[0].ID != second.Entries[0].ID {
		t.Fatal("rename lost identity")
	}
	assertUpdateFile(t, b, "chapter.md", "local")
	if _, err := os.Stat(filepath.Join(b, "draft.md")); !os.IsNotExist(err) {
		t.Fatal("old path retained")
	}
	c := t.TempDir()
	writeUpdateFile(t, c, "other.md", "other")
	if err := run([]string{"attach", c, "--server", server, "--project", p.ID}, &bytes.Buffer{}, &bytes.Buffer{}); err == nil {
		t.Fatal("attached unrelated divergent folder")
	}
	if _, err := os.Stat(filepath.Join(c, ".edda")); !os.IsNotExist(err) {
		t.Fatal("failed attach created metadata")
	}
}
func TestLocalRenameRemoteEditPreservesIdentityAfterResolution(t *testing.T) {
	a, b, _, _ := updatePair(t)
	before, err := readCheckout(a)
	if err != nil {
		t.Fatal(err)
	}
	runSyncTest(t, "move", a, "--from", "chapter.md", "--to", "renamed.md")
	writeUpdateFile(t, b, "chapter.md", "remote edit")
	runSyncTest(t, "send", b)
	if err := run([]string{"take", a}, &bytes.Buffer{}, &bytes.Buffer{}); err == nil {
		t.Fatal("rename/edit guessed")
	}
	runSyncTest(t, "resolve", a, "--path", ".", "--use", "local")
	runSyncTest(t, "take", a)
	runSyncTest(t, "send", a)
	after, err := readCheckout(a)
	if err != nil {
		t.Fatal(err)
	}
	if entryMap(before.Base.Entries)["chapter.md"].ID != entryMap(after.Base.Entries)["renamed.md"].ID {
		t.Fatal("identity lost during take")
	}
}
func TestMoveRecovery(t *testing.T) {
	a, _, _, _ := updatePair(t)
	state, err := readCheckout(a)
	if err != nil {
		t.Fatal(err)
	}
	entries := append([]project.TreeEntry{}, state.Base.Entries...)
	for i := range entries {
		if entries[i].Path == "chapter.md" {
			entries[i].Path = "renamed.md"
		}
	}
	state.Move = &pendingMove{From: "chapter.md", To: "renamed.md", Entries: entries}
	if err := writePrivateJSON(checkoutPath(a), state); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(filepath.Join(a, "chapter.md"), filepath.Join(a, "renamed.md")); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"send", a}, &bytes.Buffer{}, &bytes.Buffer{}); err == nil {
		t.Fatal("sent interrupted move")
	}
	runSyncTest(t, "move", a)
	runSyncTest(t, "send", a)
	assertUpdateFile(t, a, "renamed.md", "base chapter")
}

func TestHistoryRestoreRetainsDirtyWorkAndRetries(t *testing.T) {
	a, b, _, _ := updatePair(t)
	before, err := readCheckout(a)
	if err != nil {
		t.Fatal(err)
	}
	writeUpdateFile(t, a, "chapter.md", "new server version")
	runSyncTest(t, "send", a)
	runSyncTest(t, "take", b)
	writeUpdateFile(t, a, "notes.md", "unsent local notes")
	var page project.VersionPage
	if err := json.Unmarshal([]byte(runSyncTest(t, "history", a, "--json")), &page); err != nil || len(page.Versions) < 3 {
		t.Fatal("history missing", err)
	}
	first := runSyncTest(t, "restore", a, "--version", before.Base.ID)
	second := runSyncTest(t, "restore", a, "--version", before.Base.ID)
	if first != second {
		t.Fatal("restore retry created duplicate version")
	}
	assertUpdateFile(t, a, "notes.md", "unsent local notes")
	runSyncTest(t, "take", a)
	assertUpdateFile(t, a, "chapter.md", "base chapter")
	assertUpdateFile(t, a, "notes.md", "unsent local notes")
	runSyncTest(t, "send", a)
	runSyncTest(t, "take", b)
	assertUpdateFile(t, b, "notes.md", "unsent local notes")
}

func TestAttachedExclusionsFollowDirectoryMove(t *testing.T) {
	server, id, _, _ := importTestServer(t, nil)
	a := t.TempDir()
	writeUpdateFile(t, a, "kb/person.md", "person")
	if err := os.Symlink("/outside", filepath.Join(a, "kb", "local-skill")); err != nil {
		t.Fatal(err)
	}
	runSyncTest(t, "attach", a, "--server", server, "--project", id, "--exclude", "kb/local-skill")
	runSyncTest(t, "send", a)
	runSyncTest(t, "move", a, "--from", "kb", "--to", "knowledge")
	runSyncTest(t, "send", a)
	state, err := readCheckout(a)
	if err != nil || len(state.Excludes) != 1 || state.Excludes[0] != "knowledge/local-skill" {
		t.Fatal("exclusion not moved", err)
	}
	b := filepath.Join(t.TempDir(), "b")
	runSyncTest(t, "get", b, "--server", server, "--project", id)
	assertUpdateFile(t, b, "knowledge/person.md", "person")
	writeUpdateFile(t, b, "other.md", "remote")
	runSyncTest(t, "send", b)
	runSyncTest(t, "take", a)
	if target, err := os.Readlink(filepath.Join(a, "knowledge", "local-skill")); err != nil || target != "/outside" {
		t.Fatal("excluded symlink lost", err)
	}
}
