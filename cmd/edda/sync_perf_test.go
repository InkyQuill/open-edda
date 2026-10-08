package main

import (
	"bytes"
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/InkyQuill/open-edda/fileproject"
)

func TestSyncTransferVolume(t *testing.T) {
	var uploads, downloads atomic.Int64
	server, id, _, _ := importTestServer(t, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "PUT" && strings.Contains(r.URL.Path, "/objects/") {
				uploads.Add(r.ContentLength)
			}
			if r.Method == "GET" && strings.Contains(r.URL.Path, "/entries/") {
				downloads.Add(1)
			}
			next.ServeHTTP(w, r)
		})
	})
	parent := t.TempDir()
	a, b := filepath.Join(parent, "a"), filepath.Join(parent, "b")
	runSyncTest(t, "get", a, "--server", server, "--project", id)
	for i, name := range []string{"one.bin", "two.bin", "three.bin", "four.bin"} {
		writeUpdateFile(t, a, name, strings.Repeat(string(rune('a'+i)), 1<<20))
	}
	writeUpdateFile(t, a, "chapter.md", "old")
	runSyncTest(t, "send", a)
	runSyncTest(t, "get", b, "--server", server, "--project", id)
	uploads.Store(0)
	downloads.Store(0)
	writeUpdateFile(t, a, "chapter.md", "new chapter")
	start := time.Now()
	runSyncTest(t, "send", a)
	sendElapsed := time.Since(start)
	start = time.Now()
	runSyncTest(t, "take", b)
	t.Logf("4 MiB unchanged + 11-byte edit: send=%s upload=%d bytes; take=%s download=%d files", sendElapsed, uploads.Load(), time.Since(start), downloads.Load())
	assertUpdateFile(t, b, "chapter.md", "new chapter")
	if got := uploads.Load(); got != 11 {
		t.Fatalf("sent unchanged contents: %d bytes", got)
	}
	if got := downloads.Load(); got != 1 {
		t.Fatalf("downloaded unchanged/base contents: %d files", got)
	}

	// New paths referring to known content require only a manifest change.
	uploads.Store(0)
	downloads.Store(0)
	writeUpdateFile(t, a, "copy.md", "new chapter")
	if err := os.Rename(filepath.Join(a, "one.bin"), filepath.Join(a, "renamed.bin")); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(a, "two.bin")); err != nil {
		t.Fatal(err)
	}
	runSyncTest(t, "send", a)
	runSyncTest(t, "take", b)
	if uploads.Load() != 0 || downloads.Load() != 0 {
		t.Fatalf("metadata-only change transferred contents: upload=%d download=%d", uploads.Load(), downloads.Load())
	}
	assertUpdateFile(t, b, "copy.md", "new chapter")
	assertUpdateFile(t, b, "renamed.bin", strings.Repeat("a", 1<<20))

	// Identical new contents at two paths are transferred once, with a complete count.
	uploads.Store(0)
	downloads.Store(0)
	writeUpdateFile(t, a, "new-one.md", "duplicate")
	writeUpdateFile(t, a, "new-two.md", "duplicate")
	var output bytes.Buffer
	if err := runNetworkSend([]string{a}, &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "Uploading changed files: 1/1 (100%)") {
		t.Fatalf("incorrect progress: %s", &output)
	}
	runSyncTest(t, "take", b)
	if uploads.Load() != 9 || downloads.Load() != 1 {
		t.Fatalf("duplicate contents transferred repeatedly: upload=%d download=%d", uploads.Load(), downloads.Load())
	}
	assertUpdateFile(t, b, "new-two.md", "duplicate")
}

func TestSnapshotLeavesReservedTopLevelDataUntouched(t *testing.T) {
	root, target := t.TempDir(), t.TempDir()
	writeUpdateFile(t, root, ".git/objects/large", "repository data")
	writeUpdateFile(t, root, "node_modules/package/index.js", "cache")
	writeUpdateFile(t, root, ".env", "secret")
	writeUpdateFile(t, root, "book/.env", "nested secret")
	writeUpdateFile(t, root, "book/chapter.md", "text")
	var paths []string
	ctx := fileproject.WithProgress(context.Background(), func(p fileproject.Progress) { paths = append(paths, p.Path) })
	nodes, err := localSnapshot(ctx, root, target)
	if err != nil {
		t.Fatal(err)
	}
	for _, node := range nodes {
		if !strings.HasPrefix(node.Path, "book") {
			t.Fatalf("snapshotted reserved path: %s", node.Path)
		}
	}
	for _, name := range paths {
		if name != "" && !strings.HasPrefix(name, "book") {
			t.Fatalf("read reserved path: %s", name)
		}
	}
	assertUpdateFile(t, target, "book/.env", "nested secret")
	if _, err := os.Lstat(filepath.Join(target, ".git")); !os.IsNotExist(err) {
		t.Fatalf("copied repository: %v", err)
	}
}

func TestMissingObjectSendRetryKeepsFrozenChanges(t *testing.T) {
	var fail atomic.Bool
	server, id, _, _ := importTestServer(t, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "PUT" && strings.Contains(r.URL.Path, "/objects/") && fail.CompareAndSwap(true, false) {
				http.Error(w, "temporary upload failure", http.StatusServiceUnavailable)
				return
			}
			next.ServeHTTP(w, r)
		})
	})
	root := filepath.Join(t.TempDir(), "a")
	runSyncTest(t, "get", root, "--server", server, "--project", id)
	writeUpdateFile(t, root, "unchanged.md", "known content")
	runSyncTest(t, "send", root)
	writeUpdateFile(t, root, "new.md", "frozen")
	fail.Store(true)
	if err := runNetworkSend([]string{root}, &bytes.Buffer{}); err == nil {
		t.Fatal("failed upload reported success")
	}
	state, err := readCheckout(root)
	if err != nil || state.Pending == nil {
		t.Fatalf("missing pending state: %v", err)
	}
	for _, entry := range state.Pending.Inventory.Entries {
		if entry.Path == "unchanged.md" {
			if _, err := os.Lstat(filepath.Join(root, ".edda", state.Pending.Directory, entry.ID)); !os.IsNotExist(err) {
				t.Fatalf("unchanged file unnecessarily staged: %v", err)
			}
		}
	}
	writeUpdateFile(t, root, "new.md", "later edit")
	runSyncTest(t, "send", root)
	other := filepath.Join(t.TempDir(), "b")
	runSyncTest(t, "get", other, "--server", server, "--project", id)
	assertUpdateFile(t, other, "new.md", "frozen")
	assertUpdateFile(t, other, "unchanged.md", "known content")
	assertUpdateFile(t, root, "new.md", "later edit")
	if !strings.Contains(runSyncTest(t, "status", root), "Local changes") {
		t.Fatal("later edits marked sent")
	}

	// A reused object must still pass the same digest check as a downloaded one.
	state, err = readCheckout(other)
	if err != nil {
		t.Fatal(err)
	}
	cache, err := os.OpenRoot(other)
	if err != nil {
		t.Fatal(err)
	}
	defer cache.Close()
	reuse := objectReuse{Root: cache, Paths: map[string]string{}}
	for _, entry := range state.Base.Entries {
		reuse.Paths[entry.SHA256] = entry.Path
	}
	writeUpdateFile(t, other, "new.md", "broken") // Same length as "frozen".
	c, err := resolveConnection(server)
	if err != nil {
		t.Fatal(err)
	}
	client, err := newImportClient(server, id, c.Token)
	if err != nil {
		t.Fatal(err)
	}
	err = downloadCheckout(context.Background(), client, state.Base, t.TempDir(), server, reuse)
	if err == nil || !strings.Contains(err.Error(), "checksum") {
		t.Fatalf("accepted corrupted cached bytes: %v", err)
	}
}
