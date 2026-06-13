package app

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"testing/fstest"

	"git.inkyquill.net/inky/writer/project"
	"git.inkyquill.net/inky/writer/store"
	"github.com/pressly/goose/v3"
)

func TestHealth(t *testing.T) {
	t.Parallel()

	handler := New(nil)
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("content type = %q, want %q", got, "application/json")
	}
	if got := rec.Body.String(); got != `{"status":"ok"}`+"\n" {
		t.Fatalf("body = %q", got)
	}
}

func TestStaticFrontendServesIndexAndKeepsAPINotFound(t *testing.T) {
	handler := New(&Dependencies{
		StaticFS: fstest.MapFS{
			"index.html": {Data: []byte("<!doctype html><html><body>Writer</body></html>")},
		},
	})

	for _, path := range []string{"/", "/projects/project-1"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("%s status = %d, want %d", path, rec.Code, http.StatusOK)
		}
		if got := rec.Header().Get("Content-Type"); got != "text/html; charset=utf-8" {
			t.Fatalf("%s content type = %q, want text/html; charset=utf-8", path, got)
		}
	}

	for _, path := range []string{"/api", "/api/missing"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("%s status = %d, want %d", path, rec.Code, http.StatusNotFound)
		}
	}
}

func TestCreateProjectContent(t *testing.T) {
	handler := newTestApp(t)
	body := bytes.NewBufferString(`{
		"kind": "chapter",
		"title": "Chapter 1",
		"bodyMarkdown": "# Chapter 1\n\nOpening text.",
		"metadataJson": "{}",
		"sortOrder": 1
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/projects/project-1/content", body)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("content type = %q, want %q", got, "application/json")
	}

	var got project.ContentItem
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Kind != project.KindChapter {
		t.Fatalf("kind = %q, want %q", got.Kind, project.KindChapter)
	}
	if got.Title != "Chapter 1" {
		t.Fatalf("title = %q, want Chapter 1", got.Title)
	}
	if got.CurrentRevision != 1 {
		t.Fatalf("currentRevision = %d, want 1", got.CurrentRevision)
	}
}

func TestListProjectContentByKind(t *testing.T) {
	handler := newTestApp(t)
	createBody := bytes.NewBufferString(`{
		"kind": "chapter",
		"title": "Chapter 1",
		"bodyMarkdown": "# Chapter 1\n\nOpening text.",
		"metadataJson": "{}",
		"sortOrder": 1
	}`)
	createReq := httptest.NewRequest(http.MethodPost, "/api/projects/project-1/content", createBody)
	createRec := httptest.NewRecorder()
	handler.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d; body = %s", createRec.Code, http.StatusCreated, createRec.Body.String())
	}

	req := httptest.NewRequest(http.MethodGet, "/api/projects/project-1/content?kind=chapter", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("content type = %q, want %q", got, "application/json")
	}

	var got []project.ContentItem
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("content count = %d, want 1", len(got))
	}
	if got[0].Kind != project.KindChapter {
		t.Fatalf("kind = %q, want %q", got[0].Kind, project.KindChapter)
	}
	if got[0].Title != "Chapter 1" {
		t.Fatalf("title = %q, want Chapter 1", got[0].Title)
	}
	if got[0].CurrentRevision != 1 {
		t.Fatalf("currentRevision = %d, want 1", got[0].CurrentRevision)
	}
}

func TestListRevisionsDoesNotCrossProjectBoundary(t *testing.T) {
	handler := newTestApp(t)
	content := createTestContent(t, handler, "project-1", "Chapter 1")
	req := httptest.NewRequest(http.MethodGet, "/api/projects/project-2/content/"+content.ID+"/revisions", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestProjectContentRoutesMapClientErrors(t *testing.T) {
	handler := newTestApp(t)

	listMissing := httptest.NewRequest(http.MethodGet, "/api/projects/missing/content?kind=chapter", nil)
	listRec := httptest.NewRecorder()
	handler.ServeHTTP(listRec, listMissing)
	if listRec.Code != http.StatusNotFound {
		t.Fatalf("list missing project status = %d, want %d; body = %s", listRec.Code, http.StatusNotFound, listRec.Body.String())
	}

	createMissing := httptest.NewRequest(http.MethodPost, "/api/projects/missing/content", bytes.NewBufferString(`{
		"kind": "chapter",
		"title": "Chapter 1",
		"bodyMarkdown": "Opening.",
		"metadataJson": "{}"
	}`))
	createMissingRec := httptest.NewRecorder()
	handler.ServeHTTP(createMissingRec, createMissing)
	if createMissingRec.Code != http.StatusNotFound {
		t.Fatalf("create missing project status = %d, want %d; body = %s", createMissingRec.Code, http.StatusNotFound, createMissingRec.Body.String())
	}

	createTestContent(t, handler, "project-1", "Chapter 1")
	createDuplicate := httptest.NewRequest(http.MethodPost, "/api/projects/project-1/content", bytes.NewBufferString(`{
		"kind": "chapter",
		"title": "Chapter 1",
		"bodyMarkdown": "Duplicate.",
		"metadataJson": "{}"
	}`))
	createDuplicateRec := httptest.NewRecorder()
	handler.ServeHTTP(createDuplicateRec, createDuplicate)
	if createDuplicateRec.Code != http.StatusConflict {
		t.Fatalf("create duplicate status = %d, want %d; body = %s", createDuplicateRec.Code, http.StatusConflict, createDuplicateRec.Body.String())
	}
}

func TestCreateProjectRejectsTrailingJSON(t *testing.T) {
	handler := newTestApp(t)
	req := httptest.NewRequest(http.MethodPost, "/api/projects", bytes.NewBufferString(`{"title":"One","language":"en"}{"title":"Two"}`))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func createTestContent(t *testing.T, handler http.Handler, projectID string, title string) project.ContentItem {
	t.Helper()

	body := bytes.NewBufferString(`{
		"kind": "chapter",
		"title": "` + title + `",
		"bodyMarkdown": "# ` + title + `\n\nOpening text.",
		"metadataJson": "{}",
		"sortOrder": 1
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+projectID+"/content", body)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d; body = %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var content project.ContentItem
	if err := json.NewDecoder(rec.Body).Decode(&content); err != nil {
		t.Fatalf("decode content response: %v", err)
	}
	return content
}

func newTestApp(t *testing.T) http.Handler {
	t.Helper()

	db := openMigratedTestDB(t)
	return New(&Dependencies{
		ProjectService: project.NewService(db),
	})
}

func openMigratedTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := store.Open(filepath.Join(t.TempDir(), "writer.db"))
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatalf("set goose dialect: %v", err)
	}
	if err := goose.Up(db, filepath.Join("..", "migrations")); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	_, err = db.Exec(`
			INSERT INTO authors (id, email, password_hash, created_at)
			VALUES ('author-1', 'author@example.com', 'hash', '2026-06-13T00:00:00Z');
			INSERT INTO story_projects (id, author_id, title, slug, language, created_at, updated_at)
			VALUES
				('project-1', 'author-1', 'Test', 'test', 'en', '2026-06-13T00:00:00Z', '2026-06-13T00:00:00Z'),
				('project-2', 'author-1', 'Other', 'other', 'en', '2026-06-13T00:00:00Z', '2026-06-13T00:00:00Z');
		`)
	if err != nil {
		t.Fatalf("seed project: %v", err)
	}

	return db
}
