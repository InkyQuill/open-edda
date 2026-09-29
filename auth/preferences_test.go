package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestProfileThemePersistenceAndIsolation(t *testing.T) {
	db := openAuthHTTPTestDB(t)
	service := NewService(db, strings.Repeat("s", MinSecretBytes))
	first, err := service.Register(context.Background(), "first@example.invalid", "correct-password")
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.Register(context.Background(), "second@example.invalid", "correct-password")
	if err != nil {
		t.Fatal(err)
	}
	router := chi.NewRouter()
	RegisterProfileRoutes(router, service)
	request := func(method, token, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, "/auth/preferences", strings.NewReader(body))
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		return rec
	}
	if rec := request(http.MethodGet, "", ""); rec.Code != 401 {
		t.Fatalf("unauthenticated: %d", rec.Code)
	}
	if rec := request(http.MethodGet, first.Token, ""); rec.Code != 200 || !strings.Contains(rec.Body.String(), `"themeId":"thoth-light"`) {
		t.Fatalf("default: %d %s", rec.Code, rec.Body.String())
	}
	if rec := request(http.MethodPut, first.Token, `{"themeId":"thoth-dark"}`); rec.Code != 200 {
		t.Fatalf("save: %d %s", rec.Code, rec.Body.String())
	}
	// Re-create the service to demonstrate database persistence, not in-memory state.
	router = chi.NewRouter()
	RegisterProfileRoutes(router, NewService(db, strings.Repeat("s", MinSecretBytes)))
	if rec := request(http.MethodGet, first.Token, ""); !strings.Contains(rec.Body.String(), `"themeId":"thoth-dark"`) {
		t.Fatalf("persisted: %s", rec.Body.String())
	}
	if rec := request(http.MethodGet, second.Token, ""); !strings.Contains(rec.Body.String(), `"themeId":"thoth-light"`) {
		t.Fatalf("isolated: %s", rec.Body.String())
	}
	for _, body := range []string{`{"themeId":"unknown"}`, `{"themeId":""}`, `{"themeId":5}`, `{`} {
		if rec := request(http.MethodPut, first.Token, body); rec.Code != 400 {
			t.Fatalf("invalid %s: %d", body, rec.Code)
		}
	}
	if rec := request(http.MethodGet, first.Token, ""); !strings.Contains(rec.Body.String(), `"themeId":"thoth-dark"`) {
		t.Fatalf("invalid update changed preference: %s", rec.Body.String())
	}
}
