package main

import (
	"bytes"
	"context"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"testing"
	"time"
)

func TestSavedRefreshRenewsAccessAndRotatesNearExpiry(t *testing.T) {
	server, id, _, db := importTestServer(t, nil)
	hash, err := bcrypt.GenerateFromPassword([]byte("local-test-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("UPDATE authors SET password_hash=? WHERE id='author'", string(hash)); err != nil {
		t.Fatal(err)
	}
	t.Setenv("OPEN_EDDA_TOKEN", "")
	t.Setenv("OPEN_EDDA_URL", "")
	if err := runLogin([]string{"--server", server, "--email", "import@example.invalid", "--password-stdin"}, strings.NewReader("local-test-password\n"), &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	saved, err := readConnection()
	if err != nil {
		t.Fatal(err)
	}
	if saved.RefreshToken == "" {
		t.Fatal("refresh not saved")
	}
	path, err := connectionPath()
	if err != nil {
		t.Fatal(err)
	}
	saved.Token = "expired-access"
	saved.RefreshExpiresAt = time.Now().Add(24 * time.Hour).Unix()
	if _, err = db.Exec("UPDATE refresh_sessions SET expires_at=?", saved.RefreshExpiresAt); err != nil {
		t.Fatal(err)
	}
	if err = writePrivateJSON(path, saved); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(runSyncTest(t, "projects"), id) {
		t.Fatal("refresh did not authorize request")
	}
	renewed, err := readConnection()
	if err != nil {
		t.Fatal(err)
	}
	if renewed.Token == saved.Token || renewed.RefreshToken == saved.RefreshToken || renewed.SessionID != saved.SessionID {
		t.Fatal("session rotation failed")
	}
	if renewed.RefreshExpiresAt-time.Now().Unix() < 29*86400 {
		t.Fatal("refresh not extended")
	}
	t.Setenv("OPEN_EDDA_TOKEN", "explicit-env-token")
	value, err := refreshSavedAccess(context.Background(), server, "explicit-env-token")
	if err != nil || value != "explicit-env-token" {
		t.Fatal("environment token replaced", err)
	}
}
