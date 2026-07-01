package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStatusReportsUninitializedLayout(t *testing.T) {
	root := copyFixture(t, filepath.Join("..", "..", "fileproject", "testdata", "alchemist-lite"))

	var stdout bytes.Buffer
	if err := run([]string{"status", root}, &stdout, &bytes.Buffer{}); err != nil {
		t.Fatalf("status error = %v", err)
	}

	output := stdout.String()
	for _, want := range []string{
		"Project: uninitialized Edda folder",
		"story: 2",
		"character: 2",
		"worldbuilding: 3",
		"skill: 1",
		"missing_metadata",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("status output missing %q:\n%s", want, output)
		}
	}
}

func TestInitCreatesMetadataAndStatusReadsIt(t *testing.T) {
	root := copyFixture(t, filepath.Join("..", "..", "fileproject", "testdata", "partial"))

	var initOut bytes.Buffer
	if err := run([]string{"init", root, "--title", "Alchemy Draft", "--id", "project-1"}, &initOut, &bytes.Buffer{}); err != nil {
		t.Fatalf("init error = %v", err)
	}
	if !strings.Contains(initOut.String(), "Initialized Edda project: Alchemy Draft (project-1)") {
		t.Fatalf("init output = %s", initOut.String())
	}

	var statusOut bytes.Buffer
	if err := run([]string{"status", root}, &statusOut, &bytes.Buffer{}); err != nil {
		t.Fatalf("status after init error = %v", err)
	}
	if !strings.Contains(statusOut.String(), "Project: Alchemy Draft (project-1)") {
		t.Fatalf("status output = %s", statusOut.String())
	}
	if strings.Contains(statusOut.String(), "missing_metadata") {
		t.Fatalf("status still reports missing metadata:\n%s", statusOut.String())
	}
}

func TestInitRequiresTitle(t *testing.T) {
	var stderr bytes.Buffer
	err := run([]string{"init", t.TempDir()}, &bytes.Buffer{}, &stderr)
	if err == nil || !strings.Contains(err.Error(), "project title is required") {
		t.Fatalf("init without title error = %v", err)
	}
}

func copyFixture(t *testing.T, source string) string {
	t.Helper()
	root := t.TempDir()
	if err := filepath.WalkDir(source, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		target := filepath.Join(root, rel)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	}); err != nil {
		t.Fatalf("copy fixture: %v", err)
	}
	return root
}
