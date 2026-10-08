package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
)

func TestRefreshLifetimesRotationAndRevocation(t *testing.T) {
	db := openAuthHTTPTestDB(t)
	service := NewService(db, "refresh-test-secret-at-least-32-bytes")
	ctx := context.Background()
	initial, err := service.Register(ctx, "refresh@example.invalid", "password123")
	if err != nil {
		t.Fatal(err)
	}
	claims, err := service.ValidateToken(initial.Token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.ExpiresAt.Sub(claims.IssuedAt.Time) != 24*time.Hour {
		t.Fatal("access lifetime")
	}
	now := time.Now()
	if delta := initial.RefreshExpiresAt - now.Unix(); delta < 30*86400-2 || delta > 30*86400 {
		t.Fatal("refresh lifetime", delta)
	}
	boundary := time.Unix(initial.RefreshExpiresAt, 0).Add(-refreshRenewWindow)
	same, err := service.refreshAt(ctx, initial.RefreshToken, boundary)
	if err != nil || same.RefreshToken != initial.RefreshToken || same.RefreshExpiresAt != initial.RefreshExpiresAt {
		t.Fatalf("early renewal: %+v %v", same, err)
	}
	renewed, err := service.refreshAt(ctx, initial.RefreshToken, boundary.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if renewed.RefreshToken == initial.RefreshToken || renewed.RefreshExpiresAt != boundary.Add(time.Second+refreshLifetime).Unix() {
		t.Fatal("refresh not rotated for 30 days")
	}
	if _, err := service.refreshAt(ctx, initial.RefreshToken, boundary.Add(time.Second)); !errors.Is(err, ErrInvalidRefresh) {
		t.Fatal("old token accepted", err)
	}
	if _, err := service.refreshAt(ctx, renewed.RefreshToken, time.Unix(renewed.RefreshExpiresAt, 0)); !errors.Is(err, ErrInvalidRefresh) {
		t.Fatal("expired token accepted", err)
	}
	if err := service.RevokeRefresh(ctx, renewed.RefreshToken); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Refresh(ctx, renewed.RefreshToken); !errors.Is(err, ErrInvalidRefresh) {
		t.Fatal("revoked token accepted", err)
	}
}

func TestConcurrentRefreshRotationHasOneWinner(t *testing.T) {
	db := openAuthHTTPTestDB(t)
	service := NewService(db, "refresh-test-secret-at-least-32-bytes")
	session, err := service.Register(context.Background(), "parallel@example.invalid", "password123")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(session.RefreshExpiresAt, 0).Add(-24 * time.Hour)
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for range 2 {
		wg.Go(func() { _, err := service.refreshAt(context.Background(), session.RefreshToken, now); errs <- err })
	}
	wg.Wait()
	close(errs)
	winners := 0
	for err := range errs {
		if err == nil {
			winners++
		} else if !errors.Is(err, ErrInvalidRefresh) {
			t.Fatal(err)
		}
	}
	if winners != 1 {
		t.Fatalf("winners: %d", winners)
	}
}

func TestBrowserRefreshCookieAndLogout(t *testing.T) {
	db := openAuthHTTPTestDB(t)
	service := NewService(db, "refresh-test-secret-at-least-32-bytes")
	if _, err := service.Register(context.Background(), "cookie@example.invalid", "password123"); err != nil {
		t.Fatal(err)
	}
	r := chi.NewRouter()
	RegisterRoutes(r, service)
	req := httptest.NewRequest("POST", "https://edda.example/auth/login", strings.NewReader(`{"email":"cookie@example.invalid","password":"password123"}`))
	req.Header.Set("X-Edda-Session", "cookie")
	out := httptest.NewRecorder()
	r.ServeHTTP(out, req)
	if out.Code != 200 {
		t.Fatal(out.Body.String())
	}
	var body AuthResponse
	if err := json.Unmarshal(out.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.RefreshToken != "" {
		t.Fatal("refresh exposed to browser JS")
	}
	cookie := out.Result().Cookies()[0]
	if !cookie.HttpOnly || !cookie.Secure || cookie.SameSite != http.SameSiteStrictMode || cookie.MaxAge < 29*86400 {
		t.Fatalf("cookie flags %+v", cookie)
	}
	for step, endpoint := range []string{"refresh", "logout", "refresh"} {
		req := httptest.NewRequest("POST", "https://edda.example/auth/"+endpoint, nil)
		req.Header.Set("X-Edda-Session", "cookie")
		req.AddCookie(cookie)
		out := httptest.NewRecorder()
		r.ServeHTTP(out, req)
		switch endpoint {
		case "logout":
			if out.Code != 204 {
				t.Fatal(out.Code)
			}
		case "refresh":
			want := 200
			if step == 2 {
				want = 401
			}
			if out.Code != want {
				t.Fatal(out.Code)
			}
		}
	}
	if _, err := service.Refresh(context.Background(), cookie.Value); !errors.Is(err, ErrInvalidRefresh) {
		t.Fatal("logout failed to revoke", err)
	}
	req = httptest.NewRequest("POST", "https://edda.example/auth/refresh", strings.NewReader(`{}`))
	req.AddCookie(cookie)
	out = httptest.NewRecorder()
	r.ServeHTTP(out, req)
	if out.Code != 401 {
		t.Fatal("cookie accepted without custom header")
	}
}

func TestSecureCookiesBehindProxyAndOnLoopback(t *testing.T) {
	for _, tc := range []struct {
		url    string
		secure bool
	}{
		{"http://edda.example/api/auth/login", true},
		{"http://127.0.0.1:4187/api/auth/login", false},
		{"http://localhost:4187/api/auth/login", false},
		{"http://[::1]:4187/api/auth/login", false},
		{"https://localhost/api/auth/login", true},
	} {
		req := httptest.NewRequest("POST", tc.url, nil)
		if got := secureSessionCookie(req); got != tc.secure {
			t.Errorf("%s: secure=%v", tc.url, got)
		}
	}
}
