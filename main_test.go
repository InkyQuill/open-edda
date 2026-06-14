package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"git.inkyquill.net/inky/writer/app"
)

func TestBuildDependenciesRequiresAuthForProjectRoutes(t *testing.T) {
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

	// Unauthenticated request is rejected.
	req := httptest.NewRequest(http.MethodPost, "/api/projects", bytes.NewBufferString(`{
		"title": "Elysium",
		"language": "en"
	}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status = %d, want %d; body = %s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}

	// Register to get a token.
	regBody := bytes.NewBufferString(`{"email":"test@writersmoke.invalid","password":"test1234"}`)
	regReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", regBody)
	regRec := httptest.NewRecorder()
	handler.ServeHTTP(regRec, regReq)
	if regRec.Code != http.StatusCreated {
		t.Fatalf("register status = %d, want %d; body = %s", regRec.Code, http.StatusCreated, regRec.Body.String())
	}
	var regResp struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(regRec.Body).Decode(&regResp); err != nil {
		t.Fatalf("decode register response: %v", err)
	}

	// Authenticated request succeeds.
	req = httptest.NewRequest(http.MethodPost, "/api/projects", bytes.NewBufferString(`{
		"title": "Elysium",
		"language": "en"
	}`))
	req.Header.Set("Authorization", "Bearer "+regResp.Token)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("authenticated status = %d, want %d; body = %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
}

func TestBuildDependenciesRejectsWrongPassword(t *testing.T) {
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

	// Register a user.
	regBody := bytes.NewBufferString(`{"email":"test2@writersmoke.invalid","password":"test1234"}`)
	regReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", regBody)
	regRec := httptest.NewRecorder()
	handler.ServeHTTP(regRec, regReq)
	if regRec.Code != http.StatusCreated {
		t.Fatalf("register status = %d, want %d; body = %s", regRec.Code, http.StatusCreated, regRec.Body.String())
	}

	// Wrong password.
	loginBody := bytes.NewBufferString(`{"email":"test2@writersmoke.invalid","password":"wrong"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", loginBody)
	loginRec := httptest.NewRecorder()
	handler.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusUnauthorized {
		t.Fatalf("wrong password status = %d, want %d; body = %s", loginRec.Code, http.StatusUnauthorized, loginRec.Body.String())
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
