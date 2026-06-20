package auth

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"git.inkyquill.net/inky/writer/internal/httputil"
	"git.inkyquill.net/inky/writer/store"
	"github.com/go-chi/chi/v5"
	"github.com/pressly/goose/v3"
)

func TestLoginRejectsOversizedJSON(t *testing.T) {
	r := chi.NewRouter()
	RegisterRoutes(r, &Service{})

	body := `{"email":"` + strings.Repeat("a", int(httputil.DefaultJSONBodyLimit)) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusRequestEntityTooLarge, rec.Body.String())
	}
	if got, want := strings.TrimSpace(rec.Body.String()), `{"error":"request body too large"}`; got != want {
		t.Fatalf("body = %s, want %s", got, want)
	}
}

func TestLoginRateLimiterReturnsTooManyRequests(t *testing.T) {
	previousLimiter := defaultLoginLimiter
	defaultLoginLimiter = newLoginLimiter(5, time.Hour, 1000)
	t.Cleanup(func() { defaultLoginLimiter = previousLimiter })

	db := openAuthHTTPTestDB(t)
	service := NewService(db, strings.Repeat("s", MinSecretBytes))
	if _, err := service.Register(context.Background(), "limit@example.invalid", "correct-password"); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	r := chi.NewRouter()
	RegisterRoutes(r, service)

	for range 5 {
		rec := postLogin(t, r, "192.0.2.44:12345", "limit@example.invalid", "wrong-password")
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusUnauthorized, rec.Body.String())
		}
	}

	rec := postLogin(t, r, "192.0.2.44:54321", "limit@example.invalid", "wrong-password")
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusTooManyRequests, rec.Body.String())
	}
	if got, want := strings.TrimSpace(rec.Body.String()), `{"error":"too many login attempts"}`; got != want {
		t.Fatalf("body = %s, want %s", got, want)
	}
}

func postLogin(t *testing.T, r http.Handler, remoteAddr, email, password string) *httptest.ResponseRecorder {
	t.Helper()

	body := strings.NewReader(`{"email":"` + email + `","password":"` + password + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", body)
	req.RemoteAddr = remoteAddr
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	return rec
}

func openAuthHTTPTestDB(t *testing.T) *sql.DB {
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

	return db
}
