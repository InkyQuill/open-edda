package auth

import (
	"strings"
	"time"
)

func newID(prefix string) string {
	return prefix + "_" + strings.ReplaceAll(time.Now().UTC().Format("20060102150405.000000000"), ".", "")
}

func nowString() string {
	return time.Now().UTC().Format(time.RFC3339)
}
