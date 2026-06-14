package auth

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

func newID(prefix string) string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("auth: generate id: " + err.Error())
	}
	return prefix + "_" + hex.EncodeToString(b[:])
}

func nowString() string {
	return time.Now().UTC().Format(time.RFC3339)
}
