package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/InkyQuill/open-edda/project"
	"golang.org/x/crypto/bcrypt"
)

func runSyncTest(t *testing.T, args ...string) string {
	t.Helper()
	var out bytes.Buffer
	if err := run(args, &out, &bytes.Buffer{}); err != nil {
		t.Fatalf("%s: %v", args[0], err)
	}
	return out.String()
}
func TestGetSendTwoCopiesAndWebConflict(t *testing.T) {
	server, id, versions, _ := importTestServer(t, nil)
	parent := t.TempDir()
	a := filepath.Join(parent, "a")
	b := filepath.Join(parent, "b")
	for _, root := range []string{a, b} {
		runSyncTest(t, "get", root, "--server", server, "--project", id)
	}
	if err := os.Mkdir(filepath.Join(a, "kb"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(a, "kb", "一.md"), []byte("Первая версия.\r\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(a, "binary"), []byte{0, 255}, 0600); err != nil {
		t.Fatal(err)
	}
	runSyncTest(t, "send", a)
	first, err := versions.Version(context.Background(), "author", id, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Entries) != 3 {
		t.Fatal(first)
	}
	// A fresh second checkout receives exactly the bytes acknowledged by send.
	c := filepath.Join(parent, "c")
	runSyncTest(t, "get", c, "--server", server, "--project", id)
	body, err := os.ReadFile(filepath.Join(c, "kb", "一.md"))
	if err != nil || string(body) != "Первая версия.\r\n" {
		t.Fatal(string(body), err)
	}
	if err := os.WriteFile(filepath.Join(c, "kb", "一.md"), []byte("Из второй копии."), 0600); err != nil {
		t.Fatal(err)
	}
	runSyncTest(t, "send", c)
	second, err := versions.Version(context.Background(), "author", id, "")
	if err != nil {
		t.Fatal(err)
	}
	if second.Entries[2].ID != first.Entries[2].ID {
		t.Fatal("same path lost identity")
	}
	// The older independent copy cannot replace changes it never downloaded.
	if err := os.WriteFile(filepath.Join(b, "draft.md"), []byte("Моя работа"), 0600); err != nil {
		t.Fatal(err)
	}
	before := snapshotFiles(t, b)
	if err := run([]string{"send", b}, &bytes.Buffer{}, &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "advanced") {
		t.Fatalf("stale send: %v", err)
	}
	after := snapshotFiles(t, b)
	delete(after, ".edda/sync.lock")
	delete(before, ".edda/sync.lock")
	if !reflect.DeepEqual(before, after) {
		t.Fatal("conflict changed local work")
	}
	if err := run([]string{"get", b, "--server", server, "--project", id}, &bytes.Buffer{}, &bytes.Buffer{}); err == nil {
		t.Fatal("get overwrote existing directory")
	}
	if !strings.Contains(runSyncTest(t, "status", b), "Local changes") {
		t.Fatal("status lost dirty state")
	}
	// Deletion and empty directories are represented in the complete manifest.
	if err := os.Remove(filepath.Join(c, "binary")); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(c, "empty"), 0700); err != nil {
		t.Fatal(err)
	}
	runSyncTest(t, "send", c)
	d := filepath.Join(parent, "d")
	runSyncTest(t, "get", d, "--server", server, "--project", id)
	if _, err := os.Stat(filepath.Join(d, "binary")); !os.IsNotExist(err) {
		t.Fatal("deleted file returned")
	}
	if info, err := os.Stat(filepath.Join(d, "empty")); err != nil || !info.IsDir() {
		t.Fatal("empty directory lost", err)
	}
}

func TestSendLostResponseKeepsSnapshotAndLaterEdits(t *testing.T) {
	var drop atomic.Bool
	server, id, versions, db := importTestServer(t, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/files/versions") && drop.CompareAndSwap(true, false) {
				result := httptest.NewRecorder()
				next.ServeHTTP(result, r)
				if result.Code != 201 {
					t.Error(result.Body.String())
				}
				connection, _, err := w.(http.Hijacker).Hijack()
				if err != nil {
					t.Error(err)
					return
				}
				connection.Close()
				return
			}
			next.ServeHTTP(w, r)
		})
	})
	root := filepath.Join(t.TempDir(), "book")
	runSyncTest(t, "get", root, "--server", server, "--project", id)
	file := filepath.Join(root, "chapter.md")
	if err := os.WriteFile(file, []byte("staged"), 0600); err != nil {
		t.Fatal(err)
	}
	drop.Store(true)
	if err := run([]string{"send", root}, &bytes.Buffer{}, &bytes.Buffer{}); err == nil {
		t.Fatal("lost response reported success")
	}
	state, err := readCheckout(root)
	if err != nil || state.Pending == nil {
		t.Fatal("pending receipt not durable", err)
	}
	oldBase := state.Base.ID
	if err := os.WriteFile(file, []byte("later local edit"), 0600); err != nil {
		t.Fatal(err)
	}
	runSyncTest(t, "send", root)
	state, err = readCheckout(root)
	if err != nil || state.Pending != nil || state.Base.ID == oldBase {
		t.Fatal("ack did not advance base", err)
	}
	body, err := os.ReadFile(file)
	if err != nil || string(body) != "later local edit" {
		t.Fatal("later edit lost", err)
	}
	if !strings.Contains(runSyncTest(t, "status", root), "Local changes") {
		t.Fatal("later edit incorrectly clean")
	}
	var count int
	if err := db.QueryRow("SELECT count(*) FROM project_versions WHERE project_id=?", id).Scan(&count); err != nil || count != 2 {
		t.Fatal("duplicate publication", count, err)
	}
	runSyncTest(t, "send", root)
	head, err := versions.Version(context.Background(), "author", id, "")
	if err != nil {
		t.Fatal(err)
	}
	if head.ID == state.Base.ID {
		t.Fatal("later edits not published")
	}
}

func TestGetRejectsCorruptDownloadAndReservedPaths(t *testing.T) {
	var corrupt atomic.Bool
	server, id, versions, _ := importTestServer(t, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if corrupt.Load() && strings.Contains(r.URL.Path, "/entries/") {
				w.Write([]byte("bad"))
				return
			}
			next.ServeHTTP(w, r)
		})
	})
	source := filepath.Join(t.TempDir(), "source")
	runSyncTest(t, "get", source, "--server", server, "--project", id)
	if err := os.WriteFile(filepath.Join(source, "file"), []byte("good"), 0600); err != nil {
		t.Fatal(err)
	}
	runSyncTest(t, "send", source)
	parent := t.TempDir()
	destination := filepath.Join(parent, "failed")
	corrupt.Store(true)
	if err := run([]string{"get", destination, "--server", server, "--project", id}, &bytes.Buffer{}, &bytes.Buffer{}); err == nil {
		t.Fatal("bad checksum accepted")
	}
	names, err := os.ReadDir(parent)
	if err != nil || len(names) != 0 {
		t.Fatal("partial checkout exposed", names, err)
	}
	corrupt.Store(false)
	head, err := versions.Version(context.Background(), "author", id, "")
	if err != nil {
		t.Fatal(err)
	}
	_, err = versions.Publish(context.Background(), project.PublishVersionInput{AuthorID: "author", ProjectID: id, ExpectedVersion: head.ID, OperationID: "reserved", Entries: []project.TreeEntry{{ID: "state", Path: ".edda", Kind: "directory"}}})
	if err == nil {
		t.Fatal("server accepted reserved checkout path")
	}
}

func TestLoginStoresCredentialsOutsideProjectAndScopesServer(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("OPEN_EDDA_URL", "")
	t.Setenv("OPEN_EDDA_TOKEN", "")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Error("credentials attached to login")
		}
		var input map[string]string
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Error(err)
			return
		}
		if input["password"] != "private-password" {
			t.Error("wrong password")
		}
		w.Write([]byte(`{"token":"private-token"}`))
	}))
	defer server.Close()
	var output bytes.Buffer
	if err := runLogin([]string{"--server", server.URL, "--email", "a@example.invalid", "--password-stdin"}, strings.NewReader("private-password\n"), &output); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(output.String(), "private-") {
		t.Fatal("credential leaked")
	}
	path, err := connectionPath()
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatal("credential permissions", err)
	}
	c, err := resolveConnection("")
	if err != nil || c.Token != "private-token" {
		t.Fatal(c, err)
	}
	if _, err := resolveConnection("https://different.invalid"); err == nil {
		t.Fatal("saved token forwarded to different server")
	}
}

func TestSavedLoginAgainstActualServer(t *testing.T) {
	server, id, _, db := importTestServer(t, nil)
	hash, err := bcrypt.GenerateFromPassword([]byte("local-test-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("UPDATE authors SET password_hash=? WHERE id='author'", string(hash)); err != nil {
		t.Fatal(err)
	}
	t.Setenv("OPEN_EDDA_TOKEN", "")
	t.Setenv("OPEN_EDDA_URL", "")
	if err = runLogin([]string{"--server", server, "--email", "import@example.invalid", "--password-stdin"}, strings.NewReader("local-test-password\n"), &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(runSyncTest(t, "projects"), id) {
		t.Fatal("project discovery failed")
	}
	root := filepath.Join(t.TempDir(), "book")
	runSyncTest(t, "get", root, "--project", id)
	runSyncTest(t, "logout")
	if err := run([]string{"send", root}, &bytes.Buffer{}, &bytes.Buffer{}); err == nil {
		t.Fatal("logged-out send authorized")
	}
}
