package project

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"git.inkyquill.net/inky/writer/auth"
	"git.inkyquill.net/inky/writer/internal/httputil"
	"github.com/go-chi/chi/v5"
)

func TestHTTPCreateProjectRejectsOversizedJSON(t *testing.T) {
	r := chi.NewRouter()
	RegisterRoutes(r, &Service{})

	body := `{"title":"` + strings.Repeat("a", int(httputil.DefaultJSONBodyLimit)) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/projects", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), auth.AuthorIDKey, "author-1"))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusRequestEntityTooLarge, rec.Body.String())
	}
	if got, want := strings.TrimSpace(rec.Body.String()), `{"error":"request body too large"}`; got != want {
		t.Fatalf("body = %s, want %s", got, want)
	}
}
