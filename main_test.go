package main

import (
	"bytes"
	"context"
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
	t.Setenv("OPEN_EDDA_API_KEY_ENCRYPTION_SECRET", "test-api-key-encryption-secret-32")
	t.Setenv("OPEN_EDDA_BOOTSTRAP_EMAIL", "test@writersmoke.invalid")
	t.Setenv("OPEN_EDDA_BOOTSTRAP_PASSWORD", "test1234")
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

	// Login with the bootstrapped single user to get a token.
	loginBody := bytes.NewBufferString(`{"email":"test@writersmoke.invalid","password":"test1234"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", loginBody)
	loginRec := httptest.NewRecorder()
	handler.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("login status = %d, want %d; body = %s", loginRec.Code, http.StatusOK, loginRec.Body.String())
	}
	var loginResp struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(loginRec.Body).Decode(&loginResp); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	if loginResp.Token == "" {
		t.Fatal("login response token is empty")
	}

	// Authenticated request succeeds.
	req = httptest.NewRequest(http.MethodPost, "/api/projects", bytes.NewBufferString(`{
		"title": "Elysium",
		"language": "en"
	}`))
	req.Header.Set("Authorization", "Bearer "+loginResp.Token)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("authenticated status = %d, want %d; body = %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
}

func TestBuildDependenciesBootstrapsSingleUserFromEnv(t *testing.T) {
	t.Setenv("OPEN_EDDA_DB_PATH", filepath.Join(t.TempDir(), "edda.db"))
	t.Setenv("OPEN_EDDA_MIGRATIONS_PATH", "migrations")
	t.Setenv("OPEN_EDDA_JWT_SECRET", "test-secret-32-bytes-minimum-value")
	t.Setenv("OPEN_EDDA_API_KEY_ENCRYPTION_SECRET", "test-api-key-encryption-secret-32")
	t.Setenv("OPEN_EDDA_BOOTSTRAP_EMAIL", "solo@writersmoke.invalid")
	t.Setenv("OPEN_EDDA_BOOTSTRAP_PASSWORD", "solo-password")
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
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(`{
		"email": "solo@writersmoke.invalid",
		"password": "solo-password"
	}`))
	loginRec := httptest.NewRecorder()
	handler.ServeHTTP(loginRec, loginReq)

	if loginRec.Code != http.StatusOK {
		t.Fatalf("login status = %d, want %d; body = %s", loginRec.Code, http.StatusOK, loginRec.Body.String())
	}
	var loginResp struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(loginRec.Body).Decode(&loginResp); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	if loginResp.Token == "" {
		t.Fatal("login response token is empty")
	}
}

func TestPublicRegistrationRouteIsDisabled(t *testing.T) {
	t.Setenv("OPEN_EDDA_DB_PATH", filepath.Join(t.TempDir(), "edda.db"))
	t.Setenv("OPEN_EDDA_MIGRATIONS_PATH", "migrations")
	t.Setenv("OPEN_EDDA_JWT_SECRET", "test-secret-32-bytes-minimum-value")
	t.Setenv("OPEN_EDDA_API_KEY_ENCRYPTION_SECRET", "test-api-key-encryption-secret-32")
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
	regReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(`{
		"email": "new@writersmoke.invalid",
		"password": "test1234"
	}`))
	regRec := httptest.NewRecorder()
	handler.ServeHTTP(regRec, regReq)

	if regRec.Code != http.StatusNotFound {
		t.Fatalf("register status = %d, want %d; body = %s", regRec.Code, http.StatusNotFound, regRec.Body.String())
	}
}

func TestBuildDependenciesRejectsCrossAuthorProjectAccess(t *testing.T) {
	t.Setenv("OPEN_EDDA_DB_PATH", filepath.Join(t.TempDir(), "edda.db"))
	t.Setenv("OPEN_EDDA_MIGRATIONS_PATH", "migrations")
	t.Setenv("OPEN_EDDA_JWT_SECRET", "test-secret-32-bytes-minimum-value")
	t.Setenv("OPEN_EDDA_API_KEY_ENCRYPTION_SECRET", "test-api-key-encryption-secret-32")
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
	if _, err := deps.AuthService.Register(context.Background(), "author-a@writersmoke.invalid", "test1234"); err != nil {
		t.Fatalf("seed author A: %v", err)
	}
	if _, err := deps.AuthService.Register(context.Background(), "author-b@writersmoke.invalid", "test1234"); err != nil {
		t.Fatalf("seed author B: %v", err)
	}
	authorAToken := loginTestAuthor(t, handler, "author-a@writersmoke.invalid", "test1234")
	authorBToken := loginTestAuthor(t, handler, "author-b@writersmoke.invalid", "test1234")

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
	t.Setenv("OPEN_EDDA_API_KEY_ENCRYPTION_SECRET", "test-api-key-encryption-secret-32")
	t.Setenv("OPEN_EDDA_BOOTSTRAP_EMAIL", "test2@writersmoke.invalid")
	t.Setenv("OPEN_EDDA_BOOTSTRAP_PASSWORD", "test1234")
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

	// Wrong password.
	loginBody := bytes.NewBufferString(`{"email":"test2@writersmoke.invalid","password":"wrong"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", loginBody)
	loginRec := httptest.NewRecorder()
	handler.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusUnauthorized {
		t.Fatalf("wrong password status = %d, want %d; body = %s", loginRec.Code, http.StatusUnauthorized, loginRec.Body.String())
	}
}

func loginTestAuthor(t *testing.T, handler http.Handler, email string, password string) string {
	t.Helper()
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(`{
		"email": "`+email+`",
		"password": "`+password+`"
	}`))
	loginRec := httptest.NewRecorder()
	handler.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("login %s status = %d, want %d; body = %s", email, loginRec.Code, http.StatusOK, loginRec.Body.String())
	}
	var loginResp struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(loginRec.Body).Decode(&loginResp); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	if loginResp.Token == "" {
		t.Fatalf("login %s returned empty token", email)
	}
	return loginResp.Token
}

func TestBuildDependenciesRequiresFrontendBuild(t *testing.T) {
	t.Setenv("OPEN_EDDA_DB_PATH", filepath.Join(t.TempDir(), "edda.db"))
	t.Setenv("OPEN_EDDA_MIGRATIONS_PATH", "migrations")
	t.Setenv("OPEN_EDDA_JWT_SECRET", "test-secret-32-bytes-minimum-value")
	t.Setenv("OPEN_EDDA_API_KEY_ENCRYPTION_SECRET", "test-api-key-encryption-secret-32")
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

func TestAPIKeyEncryptionSecretRequiresDedicatedEnv(t *testing.T) {
	t.Setenv("OPEN_EDDA_API_KEY_ENCRYPTION_SECRET", "")
	t.Setenv("WRITER_API_KEY_ENCRYPTION_SECRET", "")

	secret, err := apiKeyEncryptionSecret()
	if err == nil {
		t.Fatalf("apiKeyEncryptionSecret() error = nil, secret = %q", secret)
	}
}

func TestAPIKeyEncryptionSecretAcceptsConfiguredSecret(t *testing.T) {
	t.Setenv("OPEN_EDDA_API_KEY_ENCRYPTION_SECRET", "test-api-key-encryption-secret-32")

	secret, err := apiKeyEncryptionSecret()
	if err != nil {
		t.Fatalf("apiKeyEncryptionSecret() error = %v", err)
	}
	if secret != "test-api-key-encryption-secret-32" {
		t.Fatalf("secret = %q, want configured secret", secret)
	}
}
