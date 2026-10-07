package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestMoveRejectsUntrackedDestinationOverlapWithoutMutation(t *testing.T) {
	for _, tc := range []struct{ name, untrack, remove, from, to string }{
		{"exact", "chapter.md", "chapter.md", "kb/person.md", "chapter.md"},
		{"under directory", "kb", "kb/person.md", "chapter.md", "kb/new.md"},
		{"contains child", "kb/person.md", "kb", "chapter.md", "kb"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a, b, _, _ := updatePair(t)
			t.Chdir(a)
			runSyncTest(t, "rm", tc.untrack)
			if err := os.RemoveAll(filepath.Join(a, tc.remove)); err != nil {
				t.Fatal(err)
			}
			before, err := os.ReadFile(checkoutPath(a))
			if err != nil {
				t.Fatal(err)
			}
			source, err := os.ReadFile(filepath.Join(a, tc.from))
			if err != nil {
				t.Fatal(err)
			}
			var out bytes.Buffer
			if err := run([]string{"move", a, "--from", tc.from, "--to", tc.to}, &out, &out); err == nil {
				t.Fatal("overlapping move allowed")
			}
			after, err := os.ReadFile(checkoutPath(a))
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(before, after) {
				t.Fatal("rejected move changed checkout state")
			}
			assertUpdateFile(t, a, tc.from, string(source))
			runSyncTest(t, "send", a)
			runSyncTest(t, "take", b)
			assertUpdateFile(t, b, tc.from, string(source))
		})
	}
}

func TestMoveCarriesUntrackedChildrenWithSourceDirectory(t *testing.T) {
	a, b, _, _ := updatePair(t)
	t.Chdir(a)
	writeUpdateFile(t, a, "kb/notes.md", "keep me")
	runSyncTest(t, "send", a)
	runSyncTest(t, "rm", "kb/person.md")
	runSyncTest(t, "move", a, "--from", "kb", "--to", "notes")
	state, err := readCheckout(a)
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Untracked) != 1 || state.Untracked[0] != "notes/person.md" {
		t.Fatal("exclusion not moved")
	}
	assertUpdateFile(t, a, "notes/person.md", "base person")
	runSyncTest(t, "send", a)
	runSyncTest(t, "take", b)
	assertUpdateFile(t, b, "notes/notes.md", "keep me")
	if _, err := os.Stat(filepath.Join(b, "notes/person.md")); !os.IsNotExist(err) {
		t.Fatal("excluded child published")
	}
}
