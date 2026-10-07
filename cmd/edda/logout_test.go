package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLogoutClearsLocalCredentialsBeforeServerRevocation(t *testing.T) {
	for _, status := range []int{204, 401, 503} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			t.Setenv("XDG_CONFIG_HOME", t.TempDir())
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				saved, err := readConnection()
				if err != nil || saved.Token != "" || saved.RefreshToken != "" || saved.SessionID != "" {
					t.Error("credentials still saved during network request", err)
				}
				w.WriteHeader(status)
			}))
			defer server.Close()
			checkLogout(t, server.URL, status != 204)
		})
	}
}

func TestLogoutClearsLocalCredentialsWithUnavailableOrInvalidServer(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	address := server.URL
	server.Close()
	for _, address := range []string{address, ":invalid"} {
		t.Run(address, func(t *testing.T) {
			t.Setenv("XDG_CONFIG_HOME", t.TempDir())
			checkLogout(t, address, true)
		})
	}
}

func checkLogout(t *testing.T, server string, warning bool) {
	t.Helper()
	path, err := connectionPath()
	if err != nil {
		t.Fatal(err)
	}
	if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err = writePrivateJSON(path, connection{Server: server, Token: "test-access", RefreshToken: "test-refresh", RefreshExpiresAt: 123, SessionID: "session"}); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err = runLogout(&out); err != nil {
		t.Fatal(err)
	}
	saved, err := readConnection()
	if err != nil {
		t.Fatal(err)
	}
	if saved.Token != "" || saved.RefreshToken != "" || saved.SessionID != "" || saved.RefreshExpiresAt != 0 || saved.Server != server {
		t.Fatal("local logout incomplete")
	}
	if strings.Contains(out.String(), "Warning:") != warning {
		t.Fatalf("unexpected output: %s", out.String())
	}
}
