package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func accessNeedsRefresh(token string, now time.Time) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return true
	}
	data, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return true
	}
	var claims struct {
		Expires int64 `json:"exp"`
	}
	if json.Unmarshal(data, &claims) != nil {
		return true
	}
	return claims.Expires-now.Unix() < 60
}

func refreshSavedAccess(ctx context.Context, server, token string) (string, error) {
	if os.Getenv("OPEN_EDDA_TOKEN") != "" {
		return token, nil
	}
	saved, err := readConnection()
	if err != nil {
		return "", err
	}
	if saved.Server != server || saved.Token != token || saved.RefreshToken == "" {
		return token, nil
	}
	now := time.Now()
	if !accessNeedsRefresh(token, now) && saved.RefreshExpiresAt-now.Unix() >= 7*86400 {
		return token, nil
	}
	path, err := connectionPath()
	if err != nil {
		return "", err
	}
	unlock, err := lockSession()
	if err != nil {
		return "", err
	}
	defer unlock()
	latest, err := readConnection()
	if err != nil {
		return "", err
	}
	if latest.Server != saved.Server {
		return token, nil
	}
	if latest.Token != saved.Token {
		// Another command refreshed while we waited; do not overwrite its session.
		if latest.RefreshToken == saved.RefreshToken || (saved.SessionID != "" && latest.SessionID == saved.SessionID) {
			return latest.Token, nil
		}
		return token, nil
	}
	client, err := newImportClient(server, "_", saved.Token)
	if err != nil {
		return "", err
	}
	client.root = server + "/api/"
	body, err := json.Marshal(map[string]string{"refreshToken": saved.RefreshToken})
	if err != nil {
		return "", err
	}
	response, err := client.request(ctx, "POST", "auth/refresh", bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	var result connection
	if err := json.NewDecoder(io.LimitReader(response.Body, 64<<10)).Decode(&result); err != nil {
		return "", err
	}
	if result.Token == "" || result.RefreshToken == "" {
		return "", io.ErrUnexpectedEOF
	}
	result.SessionID = saved.SessionID
	result.Server = server
	if err := writePrivateJSON(path, result); err != nil {
		return "", err
	}
	return result.Token, nil
}

func lockSession() (func(), error) {
	path, err := connectionPath()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}
	lock, err := os.OpenFile(filepath.Join(filepath.Dir(path), "session.lock"), os.O_CREATE|os.O_RDWR|syscall.O_NOFOLLOW, 0600)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); err != nil {
		lock.Close()
		return nil, err
	}
	return func() { syscall.Flock(int(lock.Fd()), syscall.LOCK_UN); lock.Close() }, nil
}
