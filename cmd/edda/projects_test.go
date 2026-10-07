package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProjectsOutput(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		args       []string
		want       string
	}{
		{"readable", `[{"id":"project-1","title":"Проверка PVC · 2026-10-06","storageMode":"files","slug":"test","language":""}]`, nil, "Проверка PVC · 2026-10-06"},
		{"empty", `[]`, nil, "No projects yet."},
		{"null", `null`, nil, "No projects yet."},
		{"json", `[{"id":"project-1","title":"Книга","extra":"preserved"}]`, []string{"--json"}, ""},
		{"empty json", `[]`, []string{"--json"}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/projects" || r.Header.Get("Authorization") != "Bearer test-token" {
					t.Errorf("unexpected request: %s", r.URL.Path)
				}
				w.Write([]byte(tc.body))
			}))
			defer server.Close()
			t.Setenv("XDG_CONFIG_HOME", t.TempDir())
			t.Setenv("OPEN_EDDA_URL", server.URL)
			t.Setenv("OPEN_EDDA_TOKEN", "test-token")
			var output bytes.Buffer
			if err := run(append([]string{"projects"}, tc.args...), &output, &output); err != nil {
				t.Fatal(err)
			}
			if len(tc.args) > 0 {
				var compact bytes.Buffer
				if err := json.Compact(&compact, output.Bytes()); err != nil {
					t.Fatal(err)
				}
				if compact.String() != tc.body {
					t.Fatalf("got %s", output.String())
				}
			} else {
				if !strings.Contains(output.String(), tc.want) || json.Valid(output.Bytes()) {
					t.Fatalf("got %s", output.String())
				}
				if tc.name == "readable" && !strings.Contains(output.String(), "project-1") {
					t.Fatal("missing project ID")
				}
			}
		})
	}
}

func TestProjectsHelpAndArguments(t *testing.T) {
	var output bytes.Buffer
	if err := run([]string{"projects", "--help"}, &output, &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "--json") {
		t.Fatal(output.String())
	}
	if err := run([]string{"projects", "extra"}, &output, &output); err == nil {
		t.Fatal("accepted positional argument")
	}
}
