package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestLoginNonInteractiveValidation(t *testing.T) {
	for _, tc := range []struct {
		name        string
		args        []string
		input, want string
	}{
		{"requires terminal", nil, "", "interactive login needs a terminal"},
		{"requires email", []string{"--password-stdin"}, "password", "email is required"},
		{"empty password", []string{"--server", "https://example.invalid", "--email", "a@example.invalid", "--password-stdin"}, "\n", "password must not be empty"},
		{"long password", []string{"--server", "https://example.invalid", "--email", "a@example.invalid", "--password-stdin"}, strings.Repeat("x", 4097), "password input too long"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("XDG_CONFIG_HOME", t.TempDir())
			t.Setenv("OPEN_EDDA_URL", "")
			var output bytes.Buffer
			err := runLogin(tc.args, strings.NewReader(tc.input), &output)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("got %v; want %s", err, tc.want)
			}
		})
	}
}

func TestLoginHelp(t *testing.T) {
	var output bytes.Buffer
	if err := runLogin([]string{"--help"}, strings.NewReader(""), &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "hidden password") {
		t.Fatal(output.String())
	}
}
