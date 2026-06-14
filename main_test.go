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
	t.Setenv("OPEN_EDDA_DB_PATH", filepath.Join(t.TempDir(), "edda.db"))
	t.Setenv("OPEN_EDDA_MIGRATIONS_PATH", "migrations")
	t.Setenv("OPEN_EDDA_JWT_SECRET", "test-secret-32-bytes-minimum-value")
	staticPath := t.TempDir()
	if err := os.WriteFile(filepath.Join(staticPath, "index.html"), []byte("<!doctype html><html></html>"), 0o644); err != nil {
		t.Fatalf("write static index: %v", err)
	}
	t.Setenv("OPEN_EDDA_STATIC_PATH", staticPath)

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
	if regResp.Token == "" {
		t.Fatal("register response token is empty")
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

func TestBuildDependenciesRejectsCrossAuthorProjectAccess(t *testing.T) {
	t.Setenv("OPEN_EDDA_DB_PATH", filepath.Join(t.TempDir(), "edda.db"))
	t.Setenv("OPEN_EDDA_MIGRATIONS_PATH", "migrations")
	t.Setenv("OPEN_EDDA_JWT_SECRET", "test-secret-32-bytes-minimum-value")
	staticPath := t.TempDir()
	if err := os.WriteFile(filepath.Join(staticPath, "index.html"), []byte("<!doctype html><html></html>"), 0o644); err != nil {
		t.Fatalf("write static index: %v", err)
	}
	t.Setenv("OPEN_EDDA_STATIC_PATH", staticPath)

	deps, cleanup, err := buildDependencies()
	if err != nil {
		t.Fatalf("build dependencies: %v", err)
	}
	defer cleanup()

	handler := app.New(deps)
	authorAToken := registerTestAuthor(t, handler, "author-a@writersmoke.invalid")
	authorBToken := registerTestAuthor(t, handler, "author-b@writersmoke.invalid")

	createReq := httptest.NewRequest(http.MethodPost, "/api/projects", bytes.NewBufferString(`{
		"title": "A Private Story",
		"language": "en"
	}`))
	createReq.Header.Set("Authorization", "Bearer "+authorAToken)
	createRec := httptest.NewRecorder()
	handler.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d; body = %s", createRec.Code, http.StatusCreated, createRec.Body.String())
	}
	var created struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createRec.Body).Decode(&created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/projects/"+created.ID+"/content?kind=chapter", nil)
	listReq.Header.Set("Authorization", "Bearer "+authorBToken)
	listRec := httptest.NewRecorder()
	handler.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusNotFound {
		t.Fatalf("cross-author status = %d, want %d; body = %s", listRec.Code, http.StatusNotFound, listRec.Body.String())
	}
}

func TestBuildDependenciesRejectsWrongPassword(t *testing.T) {
	t.Setenv("OPEN_EDDA_DB_PATH", filepath.Join(t.TempDir(), "edda.db"))
	t.Setenv("OPEN_EDDA_MIGRATIONS_PATH", "migrations")
	t.Setenv("OPEN_EDDA_JWT_SECRET", "test-secret-32-bytes-minimum-value")
	staticPath := t.TempDir()
	if err := os.WriteFile(filepath.Join(staticPath, "index.html"), []byte("<!doctype html><html></html>"), 0o644); err != nil {
		t.Fatalf("write static index: %v", err)
	}
	t.Setenv("OPEN_EDDA_STATIC_PATH", staticPath)

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

func registerTestAuthor(t *testing.T, handler http.Handler, email string) string {
	t.Helper()
	regReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(`{
		"email": "`+email+`",
		"password": "test1234"
	}`))
	regRec := httptest.NewRecorder()
	handler.ServeHTTP(regRec, regReq)
	if regRec.Code != http.StatusCreated {
		t.Fatalf("register %s status = %d, want %d; body = %s", email, regRec.Code, http.StatusCreated, regRec.Body.String())
	}
	var regResp struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(regRec.Body).Decode(&regResp); err != nil {
		t.Fatalf("decode register response: %v", err)
	}
	if regResp.Token == "" {
		t.Fatalf("register %s returned empty token", email)
	}
	return regResp.Token
}

func TestBuildDependenciesRequiresFrontendBuild(t *testing.T) {
	t.Setenv("OPEN_EDDA_DB_PATH", filepath.Join(t.TempDir(), "edda.db"))
	t.Setenv("OPEN_EDDA_MIGRATIONS_PATH", "migrations")
	t.Setenv("OPEN_EDDA_JWT_SECRET", "test-secret-32-bytes-minimum-value")
	t.Setenv("OPEN_EDDA_STATIC_PATH", t.TempDir())

	deps, cleanup, err := buildDependencies()
	if err == nil {
		cleanup()
		t.Fatalf("buildDependencies() error = nil, deps = %#v", deps)
	}
}

func TestBuildDependenciesRequiresJWTSecret(t *testing.T) {
	t.Setenv("OPEN_EDDA_DB_PATH", filepath.Join(t.TempDir(), "edda.db"))
	t.Setenv("OPEN_EDDA_MIGRATIONS_PATH", "migrations")
	staticPath := t.TempDir()
	if err := os.WriteFile(filepath.Join(staticPath, "index.html"), []byte("<!doctype html><html></html>"), 0o644); err != nil {
		t.Fatalf("write static index: %v", err)
	}
	t.Setenv("OPEN_EDDA_STATIC_PATH", staticPath)

	deps, cleanup, err := buildDependencies()
	if err == nil {
		cleanup()
		t.Fatalf("buildDependencies() error = nil, deps = %#v", deps)
	}
}
