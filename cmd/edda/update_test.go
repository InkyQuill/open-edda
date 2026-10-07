package main

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
)

func updatePair(t *testing.T) (string, string, string, string) {
	t.Helper()
	server, id, _, _ := importTestServer(t, nil)
	parent := t.TempDir()
	a, b := filepath.Join(parent, "a"), filepath.Join(parent, "b")
	runSyncTest(t, "get", a, "--server", server, "--project", id)
	for name, body := range map[string]string{"chapter.md": "base chapter", "notes.md": "base notes", "kb/person.md": "base person"} {
		writeUpdateFile(t, a, name, body)
	}
	runSyncTest(t, "send", a)
	runSyncTest(t, "get", b, "--server", server, "--project", id)
	return a, b, server, id
}
func writeUpdateFile(t *testing.T, root, name, body string) {
	t.Helper()
	target := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
}
func assertUpdateFile(t *testing.T, root, name, want string) {
	t.Helper()
	body, err := os.ReadFile(filepath.Join(root, name))
	if err != nil || string(body) != want {
		t.Fatalf("%s=%q, want %q: %v", name, body, want, err)
	}
}

func TestTakeMergesDifferentFilesAndPreservesExcludedData(t *testing.T) {
	a, b, _, _ := updatePair(t)
	writeUpdateFile(t, a, "chapter.md", "local chapter")
	writeUpdateFile(t, a, ".env", "local secret")
	writeUpdateFile(t, a, "kb/.env", "nested secret")
	if err := os.Mkdir(filepath.Join(a, ".git"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("../outside", filepath.Join(a, ".git", "link")); err != nil {
		t.Fatal(err)
	}
	writeUpdateFile(t, b, "notes.md", "remote notes")
	writeUpdateFile(t, b, "kb/person.md", "remote person")
	runSyncTest(t, "send", b)
	inode, err := os.Stat(a)
	if err != nil {
		t.Fatal(err)
	}
	output := runSyncTest(t, "take", a)
	assertUpdateFile(t, a, "chapter.md", "local chapter")
	assertUpdateFile(t, a, "notes.md", "remote notes")
	assertUpdateFile(t, a, "kb/person.md", "remote person")
	assertUpdateFile(t, a, ".env", "local secret")
	assertUpdateFile(t, a, "kb/.env", "nested secret")
	after, err := os.Stat(a)
	if err != nil || !os.SameFile(inode, after) {
		t.Fatal("checkout root was replaced")
	}
	if link, err := os.Readlink(filepath.Join(a, ".git", "link")); err != nil || link != "../outside" {
		t.Fatal("excluded link changed", err)
	}
	if !strings.Contains(output, "Backups:") {
		t.Fatal(output)
	}
	if !strings.Contains(runSyncTest(t, "status", a), "Local changes") {
		t.Fatal("local changes acknowledged without send")
	}
	runSyncTest(t, "send", a)
	runSyncTest(t, "take", b)
	assertUpdateFile(t, b, "chapter.md", "local chapter")
}

func TestTakeConflictChoicesAndDeleteVersusEdit(t *testing.T) {
	for _, choice := range []string{"local", "remote"} {
		t.Run(choice, func(t *testing.T) {
			a, b, _, _ := updatePair(t)
			writeUpdateFile(t, a, "chapter.md", "local")
			writeUpdateFile(t, b, "chapter.md", "remote")
			runSyncTest(t, "send", b)
			if err := run([]string{"take", a}, &bytes.Buffer{}, &bytes.Buffer{}); err == nil {
				t.Fatal("conflict silently chosen")
			}
			assertUpdateFile(t, a, "chapter.md", "local")
			state, err := readCheckout(a)
			if err != nil || state.Update == "" {
				t.Fatal("plan not retained", err)
			}
			assertUpdateFile(t, filepath.Join(a, ".edda", state.Update, "base"), "chapter.md", "base chapter")
			if err := run([]string{"send", a}, &bytes.Buffer{}, &bytes.Buffer{}); err == nil {
				t.Fatal("send during conflict accepted")
			}
			runSyncTest(t, "resolve", a, "--path", "chapter.md", "--use", choice)
			runSyncTest(t, "take", a)
			assertUpdateFile(t, a, "chapter.md", choice)
		})
	}
	a, b, _, _ := updatePair(t)
	writeUpdateFile(t, a, "kb/person.md", "local changed person")
	if err := os.RemoveAll(filepath.Join(b, "kb")); err != nil {
		t.Fatal(err)
	}
	runSyncTest(t, "send", b)
	if err := run([]string{"take", a}, &bytes.Buffer{}, &bytes.Buffer{}); err == nil {
		t.Fatal("delete/edit conflict not detected")
	}
	runSyncTest(t, "resolve", a, "--path", "kb/person.md", "--use", "local")
	runSyncTest(t, "take", a)
	assertUpdateFile(t, a, "kb/person.md", "local changed person")
}

func TestTakeRefusesStalePlanAndRestartKeepsEdits(t *testing.T) {
	a, b, _, _ := updatePair(t)
	writeUpdateFile(t, a, "chapter.md", "local")
	writeUpdateFile(t, b, "chapter.md", "remote")
	runSyncTest(t, "send", b)
	if err := run([]string{"take", a}, &bytes.Buffer{}, &bytes.Buffer{}); err == nil {
		t.Fatal("expected conflict")
	}
	runSyncTest(t, "resolve", a, "--path", "chapter.md", "--use", "local")
	writeUpdateFile(t, a, "chapter.md", "edited again")
	if err := run([]string{"take", a}, &bytes.Buffer{}, &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "changed since") {
		t.Fatalf("stale plan applied: %v", err)
	}
	assertUpdateFile(t, a, "chapter.md", "edited again")
	if err := run([]string{"take", a, "--restart"}, &bytes.Buffer{}, &bytes.Buffer{}); err == nil {
		t.Fatal("expected renewed conflict")
	}
	runSyncTest(t, "resolve", a, "--path", "chapter.md", "--use", "local")
	runSyncTest(t, "take", a)
	assertUpdateFile(t, a, "chapter.md", "edited again")
}

func TestRecoverPartialApplyPreservesPostCrashEdits(t *testing.T) {
	a, b, _, _ := updatePair(t)
	writeUpdateFile(t, a, "chapter.md", "local")
	writeUpdateFile(t, b, "chapter.md", "remote")
	runSyncTest(t, "send", b)
	if err := run([]string{"take", a}, &bytes.Buffer{}, &bytes.Buffer{}); err == nil {
		t.Fatal("expected conflict")
	}
	state, err := readCheckout(a)
	if err != nil {
		t.Fatal(err)
	}
	p, err := loadUpdate(a, state)
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(a, ".edda", state.Update)
	if err := os.Mkdir(filepath.Join(dir, "backup"), 0700); err != nil {
		t.Fatal(err)
	}
	p.Phase = "applying"
	p.Units = []updateUnit{{Name: "chapter.md", Old: true, New: true, Phase: "installed"}}
	if err := saveUpdate(a, state, p); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(filepath.Join(a, "chapter.md"), filepath.Join(dir, "backup", "chapter.md")); err != nil {
		t.Fatal(err)
	}
	writeUpdateFile(t, a, "chapter.md", "edited after crash")
	runSyncTest(t, "take", a)
	assertUpdateFile(t, a, "chapter.md", "local")
	displaced, err := os.ReadDir(filepath.Join(dir, "displaced"))
	if err != nil || len(displaced) != 1 {
		t.Fatal("post-crash edits missing", err)
	}
	assertUpdateFile(t, filepath.Join(dir, "displaced"), displaced[0].Name(), "edited after crash")
	recovered, err := readCheckout(a)
	if err != nil || !reflect.DeepEqual(recovered, p.Before) {
		t.Fatal("base not rolled back", err)
	}
}

func TestSnapshotSubsetForConcurrentEditCheck(t *testing.T) {
	root := t.TempDir()
	writeUpdateFile(t, root, "one/file", "one")
	writeUpdateFile(t, root, "two", "two")
	all, err := localSnapshot(context.Background(), root, "")
	if err != nil {
		t.Fatal(err)
	}
	one, err := localSnapshot(context.Background(), root, "", "one")
	if err != nil || !reflect.DeepEqual(one, nodesUnder(all, "one")) {
		t.Fatal("wrong snapshot subset", err)
	}
}

// A separate test process exits without deferred cleanup, exercising the files
// and journal left by the actual application algorithm at durable boundaries.
func TestUpdateCrashHelper(t *testing.T) {
	root := os.Getenv("EDDA_TEST_UPDATE_ROOT")
	if root == "" {
		return
	}
	state, err := readCheckout(root)
	if err != nil {
		t.Fatal(err)
	}
	p, err := loadUpdate(root, state)
	if err != nil {
		t.Fatal(err)
	}
	err = applyUpdateObserved(context.Background(), root, state, p, func(phase string) {
		if phase == os.Getenv("EDDA_TEST_UPDATE_PHASE") {
			os.Exit(23)
		}
	})
	t.Fatalf("crash point not reached: %v", err)
}

func TestTakeRecoversProcessExitAtDurableBoundaries(t *testing.T) {
	for _, phase := range []string{"backup", "install", "before-commit", "after-commit"} {
		t.Run(phase, func(t *testing.T) {
			a, b, _, _ := updatePair(t)
			writeUpdateFile(t, a, "chapter.md", "local")
			writeUpdateFile(t, b, "chapter.md", "remote")
			writeUpdateFile(t, b, "notes.md", "remote notes")
			runSyncTest(t, "send", b)
			if err := run([]string{"take", a}, &bytes.Buffer{}, &bytes.Buffer{}); err == nil {
				t.Fatal("expected conflict")
			}
			runSyncTest(t, "resolve", a, "--path", "chapter.md", "--use", "remote")
			before, err := readCheckout(a)
			if err != nil {
				t.Fatal(err)
			}
			p, err := loadUpdate(a, before)
			if err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(os.Args[0], "-test.run=^TestUpdateCrashHelper$")
			cmd.Env = append(os.Environ(), "EDDA_TEST_UPDATE_ROOT="+a, "EDDA_TEST_UPDATE_PHASE="+phase)
			output, err := cmd.CombinedOutput()
			var exited *exec.ExitError
			if !errors.As(err, &exited) || exited.ExitCode() != 23 {
				t.Fatalf("%v: %s", err, output)
			}
			runSyncTest(t, "take", a)
			after, err := readCheckout(a)
			if err != nil || after.Update != "" {
				t.Fatalf("recovery unfinished: %v", err)
			}
			if phase == "after-commit" {
				assertUpdateFile(t, a, "chapter.md", "remote")
				assertUpdateFile(t, a, "notes.md", "remote notes")
				if after.Base.ID != p.Remote.ID {
					t.Fatal("committed base lost")
				}
			} else {
				assertUpdateFile(t, a, "chapter.md", "local")
				assertUpdateFile(t, a, "notes.md", "base notes")
				if after.Base.ID != before.Base.ID {
					t.Fatal("uncommitted base advanced")
				}
			}
		})
	}
}

func TestTakeCannotDeleteDirectoryWithExcludedFiles(t *testing.T) {
	a, b, _, _ := updatePair(t)
	writeUpdateFile(t, a, "kb/.env", "secret")
	if err := os.RemoveAll(filepath.Join(b, "kb")); err != nil {
		t.Fatal(err)
	}
	runSyncTest(t, "send", b)
	if err := run([]string{"take", a}, &bytes.Buffer{}, &bytes.Buffer{}); err == nil {
		t.Fatal("deleted excluded files")
	}
	assertUpdateFile(t, a, "kb/.env", "secret")
	assertUpdateFile(t, a, "kb/person.md", "base person")
}

func TestTakeReconcilesUnpublishedPendingSnapshot(t *testing.T) {
	var reject atomic.Bool
	server, id, _, _ := importTestServer(t, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/files/versions") && reject.Load() {
				http.Error(w, "offline", http.StatusServiceUnavailable)
				return
			}
			next.ServeHTTP(w, r)
		})
	})
	parent := t.TempDir()
	a, b := filepath.Join(parent, "a"), filepath.Join(parent, "b")
	runSyncTest(t, "get", a, "--server", server, "--project", id)
	runSyncTest(t, "get", b, "--server", server, "--project", id)
	writeUpdateFile(t, a, "local.md", "staged")
	reject.Store(true)
	if err := run([]string{"send", a}, &bytes.Buffer{}, &bytes.Buffer{}); err == nil {
		t.Fatal("expected failed publication")
	}
	pending, err := readCheckout(a)
	if err != nil || pending.Pending == nil {
		t.Fatal("missing pending", err)
	}
	if err := run([]string{"take", a}, &bytes.Buffer{}, &bytes.Buffer{}); err == nil {
		t.Fatal("abandoned retryable send")
	}
	reject.Store(false)
	writeUpdateFile(t, b, "remote.md", "remote")
	runSyncTest(t, "send", b)
	writeUpdateFile(t, a, "local.md", "later edit")
	runSyncTest(t, "take", a)
	assertUpdateFile(t, a, "local.md", "later edit")
	assertUpdateFile(t, a, "remote.md", "remote")
	assertUpdateFile(t, filepath.Join(a, ".edda", pending.Pending.Directory), pending.Pending.Inventory.Entries[0].ID, "staged")
	state, err := readCheckout(a)
	if err != nil || state.Pending != nil {
		t.Fatal("obsolete pending remains active", err)
	}
	runSyncTest(t, "send", a)
	runSyncTest(t, "take", b)
	assertUpdateFile(t, b, "local.md", "later edit")
}

func TestTakeStructuralConflictCanChooseWholeTree(t *testing.T) {
	a, b, _, _ := updatePair(t)
	// One side replaces a directory with a file while the other edits its child.
	if err := os.RemoveAll(filepath.Join(a, "kb")); err != nil {
		t.Fatal(err)
	}
	writeUpdateFile(t, a, "kb", "local file")
	writeUpdateFile(t, b, "kb/person.md", "remote person")
	runSyncTest(t, "send", b)
	if err := run([]string{"take", a}, &bytes.Buffer{}, &bytes.Buffer{}); err == nil {
		t.Fatal("expected conflict")
	}
	runSyncTest(t, "resolve", a, "--path", "kb/person.md", "--use", "remote")
	if err := run([]string{"take", a}, &bytes.Buffer{}, &bytes.Buffer{}); err == nil {
		t.Fatal("expected structural conflict")
	}
	runSyncTest(t, "resolve", a, "--path", ".", "--use", "local")
	runSyncTest(t, "take", a)
	assertUpdateFile(t, a, "kb", "local file")
}
