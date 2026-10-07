package main

import (
	"bytes"
	"context"
	"database/sql"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/InkyQuill/open-edda/app"
	"github.com/InkyQuill/open-edda/auth"
	"github.com/InkyQuill/open-edda/project"
	"github.com/InkyQuill/open-edda/store"
	"github.com/pressly/goose/v3"
)

func importTestServer(t *testing.T, wrap func(http.Handler) http.Handler) (string, string, *project.VersionStore, *sql.DB) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	db, err := store.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatal(err)
	}
	if err := goose.Up(db, "../../migrations"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("INSERT INTO authors(id,email,password_hash,created_at) VALUES('author','import@example.invalid','unused','now')"); err != nil {
		t.Fatal(err)
	}
	service := project.NewService(db)
	p, err := service.CreateProject(context.Background(), project.CreateProjectInput{AuthorID: "author", Title: "Imported book", StorageMode: "files"})
	if err != nil {
		t.Fatal(err)
	}
	versions, err := project.NewVersionStore(db, t.TempDir(), project.VersionLimits{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { versions.Close() })
	secret := "import-test-secret-at-least-32-bytes"
	token, err := auth.GenerateToken("author", "import@example.invalid", secret)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("OPEN_EDDA_TOKEN", token)
	var handler http.Handler = app.New(&app.Dependencies{AuthService: auth.NewService(db, secret), ProjectService: service, VersionStore: versions})
	if wrap != nil {
		handler = wrap(handler)
	}
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server.URL, p.ID, versions, db
}

func TestImportRoundTripLostResponseAndRetryAfterHeadAdvance(t *testing.T) {
	var drop atomic.Bool
	drop.Store(true)
	server, id, versions, db := importTestServer(t, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/files/versions") && drop.CompareAndSwap(true, false) {
				result := httptest.NewRecorder()
				next.ServeHTTP(result, r)
				if result.Code != 201 {
					t.Errorf("publication: %d %s", result.Code, result.Body.String())
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
	root := t.TempDir()
	files := map[string][]byte{"project.md": []byte("---\r\ncustom: value\r\n---\r\n"), "kb/一.md": []byte("Персонаж"), "work/empty.txt": {}, "assets/data.bin": {0, 255, 6}, ".pocket-editor.json": []byte("{}"), "timeline.yaml": []byte("events: []\n")}
	for path, body := range files {
		full := filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(full), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, body, 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(root, "empty-folder"), 0700); err != nil {
		t.Fatal(err)
	}
	args := []string{"import", root, "--server", server, "--project", id}
	var output bytes.Buffer
	if err := run(args, &output, &bytes.Buffer{}); err == nil {
		t.Fatal("lost response reported success")
	} else {
		t.Logf("first import failure: %v", err)
	}
	head, err := versions.Version(context.Background(), "author", id, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(head.Entries) == 0 {
		t.Fatal("publication did not commit")
	}
	for _, entry := range head.Entries {
		if entry.Kind != "file" {
			continue
		}
		reader, err := versions.OpenVersionFile(context.Background(), "author", id, head.ID, entry.ID)
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(reader)
		reader.Close()
		if err != nil || !bytes.Equal(body, files[entry.Path]) {
			t.Fatalf("bytes differ for %s: %v", entry.Path, err)
		}
	}
	advanced, err := versions.Publish(context.Background(), project.PublishVersionInput{AuthorID: "author", ProjectID: id, ExpectedVersion: head.ID, OperationID: "web-save", Entries: head.Entries})
	if err != nil {
		t.Fatal(err)
	}
	output.Reset()
	if err := run(args, &output, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "Already imported as version "+head.ID) {
		t.Fatal(output.String())
	}
	now, err := versions.Version(context.Background(), "author", id, "")
	if err != nil || now.ID != advanced.ID {
		t.Fatal("retry moved head", err)
	}
	var count int
	if err := db.QueryRow("SELECT count(*) FROM project_versions WHERE project_id=?", id).Scan(&count); err != nil || count != 3 {
		t.Fatalf("duplicate version: %d %v", count, err)
	}
	state, err := readCheckout(root)
	if err != nil || state.Base.ID != head.ID || state.Base.ProjectID != id || state.Server != server {
		t.Fatalf("import binding must use original receipt: %+v %v", state, err)
	}
	if err := os.WriteFile(filepath.Join(root, "new.md"), []byte("new local work"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := run(args, &bytes.Buffer{}, &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "use edda send") {
		t.Fatalf("nonempty target accepted: %v", err)
	}
}

func TestImportFailuresDoNotPublishAndDryRunNeedsNoServer(t *testing.T) {
	var uploads atomic.Int32
	server, id, versions, _ := importTestServer(t, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "PUT" && uploads.Add(1) == 2 {
				http.Error(w, "unavailable", 503)
				return
			}
			next.ServeHTTP(w, r)
		})
	})
	root := t.TempDir()
	for _, name := range []string{"a.md", "b.md"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(name), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := run([]string{"import", root, "--dry-run"}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if uploads.Load() != 0 {
		t.Fatal("dry run contacted server")
	}
	args := []string{"import", root, "--server", server, "--project", id}
	if err := run(args, &bytes.Buffer{}, &bytes.Buffer{}); err == nil {
		t.Fatal("failed upload reported success")
	} else {
		t.Logf("first upload failure: %v", err)
	}
	head, err := versions.Version(context.Background(), "author", id, "")
	if err != nil || len(head.Entries) != 0 {
		t.Fatal("partial tree became visible", err)
	}
	if err := run(args, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	t.Setenv("OPEN_EDDA_TOKEN", "expired")
	args[1] = t.TempDir()
	if err := run(args, &bytes.Buffer{}, &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "authentication") {
		t.Fatalf("invalid auth: %v", err)
	}
}

func TestImportDoesNotForwardCredentialsOnRedirect(t *testing.T) {
	var reached atomic.Bool
	destination := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { reached.Store(true) }))
	defer destination.Close()
	redirect := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, destination.URL, 307) }))
	defer redirect.Close()
	client, err := newImportClient(redirect.URL, "project", "token")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.version(context.Background(), "versions/current"); err == nil || reached.Load() {
		t.Fatal("redirect followed with credentials")
	}
}

func TestImportRejectsAmbiguousFolderArguments(t *testing.T) {
	if err := run([]string{"import", "first", "second", "--dry-run"}, &bytes.Buffer{}, &bytes.Buffer{}); err == nil || !strings.Contains(err.Error(), "one folder") {
		t.Fatalf("ambiguous folders: %v", err)
	}
}
