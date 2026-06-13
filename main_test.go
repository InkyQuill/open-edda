package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"git.inkyquill.net/inky/writer/app"
)

func TestBuildDependenciesMountsProjectRoutes(t *testing.T) {
	t.Setenv("WRITER_DB_PATH", filepath.Join(t.TempDir(), "writer.db"))
	t.Setenv("WRITER_MIGRATIONS_PATH", "migrations")

	deps, cleanup, err := buildDependencies()
	if err != nil {
		t.Fatalf("build dependencies: %v", err)
	}
	defer cleanup()

	handler := app.New(deps)
	req := httptest.NewRequest(http.MethodPost, "/api/projects", bytes.NewBufferString(`{
		"title": "Elysium",
		"language": "en"
	}`))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
}
