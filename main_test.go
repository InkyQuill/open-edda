package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"git.inkyquill.net/inky/writer/app"
)

func TestBuildDependenciesMountsProjectRoutes(t *testing.T) {
	t.Setenv("WRITER_DB_PATH", filepath.Join(t.TempDir(), "writer.db"))
	t.Setenv("WRITER_MIGRATIONS_PATH", "migrations")
	staticPath := t.TempDir()
	if err := os.WriteFile(filepath.Join(staticPath, "index.html"), []byte("<!doctype html><html></html>"), 0o644); err != nil {
		t.Fatalf("write static index: %v", err)
	}
	t.Setenv("WRITER_STATIC_PATH", staticPath)

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

func TestBuildDependenciesRequiresFrontendBuild(t *testing.T) {
	t.Setenv("WRITER_DB_PATH", filepath.Join(t.TempDir(), "writer.db"))
	t.Setenv("WRITER_MIGRATIONS_PATH", "migrations")
	t.Setenv("WRITER_STATIC_PATH", t.TempDir())

	deps, cleanup, err := buildDependencies()
	if err == nil {
		cleanup()
		t.Fatalf("buildDependencies() error = nil, deps = %#v", deps)
	}
}
