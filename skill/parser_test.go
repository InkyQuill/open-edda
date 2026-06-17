package skill

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseSkillArchiveReadsFrontmatterRoutingAndFiles(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	writeZipFile(t, zw, "style/SKILL.md", `---
name: style-pass
description: Use when rewriting prose for style and rhythm
route:
  actionKinds: [rewrite, read_check]
  contentKinds: [chapter]
  tags: [style, prose]
  priority: 70
metadata:
  useCases:
    - Rewrite prose when style and rhythm are the main problem.
    - Check a chapter for sentence-level drag before applying edits.
  doNotUse:
    - Do not use for broad plot restructuring.
---

# Style Pass

Prefer concrete verbs and keep the author's POV.
`)
	writeZipFile(t, zw, "style/templates/rewrite.md", "Rewrite template")
	writeZipFile(t, zw, "style/references/checklist.md", "Checklist")
	writeZipFile(t, zw, "style/scripts/analyze.sh", "#!/bin/sh\necho disabled\n")
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}

	parsed, err := ParseSkillArchive(bytes.NewReader(buf.Bytes()), int64(buf.Len()), "style.zip")
	if err != nil {
		t.Fatalf("ParseSkillArchive() error = %v", err)
	}
	if parsed.Name != "style-pass" {
		t.Fatalf("Name = %q, want style-pass", parsed.Name)
	}
	if parsed.Description != "Use when rewriting prose for style and rhythm" {
		t.Fatalf("Description = %q", parsed.Description)
	}
	var metadata struct {
		UseCases []string `json:"useCases"`
		DoNotUse []string `json:"doNotUse"`
	}
	if err := json.Unmarshal([]byte(parsed.MetadataJSON), &metadata); err != nil {
		t.Fatalf("unmarshal MetadataJSON: %v", err)
	}
	if len(metadata.UseCases) != 2 || metadata.UseCases[0] != "Rewrite prose when style and rhythm are the main problem." {
		t.Fatalf("metadata useCases = %#v", metadata.UseCases)
	}
	if len(metadata.DoNotUse) != 1 || metadata.DoNotUse[0] != "Do not use for broad plot restructuring." {
		t.Fatalf("metadata doNotUse = %#v", metadata.DoNotUse)
	}
	if !strings.Contains(parsed.InstructionsMarkdown, "Prefer concrete verbs") {
		t.Fatalf("InstructionsMarkdown missing body: %q", parsed.InstructionsMarkdown)
	}
	if parsed.ScriptCount != 1 || !parsed.ScriptsDisabled {
		t.Fatalf("script status = (%d, %v), want disabled script count", parsed.ScriptCount, parsed.ScriptsDisabled)
	}
	if len(parsed.RoutingHints) != 6 {
		t.Fatalf("RoutingHints = %d, want 6 action/content/tag hints", len(parsed.RoutingHints))
	}
	assertImportedFile(t, parsed.Files, "SKILL.md", FilePurposeInstruction, false)
	assertImportedFile(t, parsed.Files, "templates/rewrite.md", FilePurposeTemplate, false)
	assertImportedFile(t, parsed.Files, "references/checklist.md", FilePurposeReference, false)
	assertImportedFile(t, parsed.Files, "scripts/analyze.sh", FilePurposeScript, true)
}

func TestParseBuiltinSkillDirectories(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "docs", "skills", "builtin")
	for _, category := range []string{"default", "optional"} {
		category := category
		entries, err := os.ReadDir(filepath.Join(root, category))
		if err != nil {
			t.Fatalf("read %s builtins: %v", category, err)
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			name := entry.Name()
			t.Run(category+"/"+name, func(t *testing.T) {
				t.Parallel()

				parsed, err := ParseSkillDirectory(filepath.Join(root, category, name))
				if err != nil {
					t.Fatalf("ParseSkillDirectory() error = %v", err)
				}
				if parsed.Name == "" {
					t.Fatal("Name is empty")
				}
				if parsed.Description == "" {
					t.Fatal("Description is empty")
				}
				if strings.TrimSpace(parsed.InstructionsMarkdown) == "" {
					t.Fatal("InstructionsMarkdown is empty")
				}
				var metadata struct {
					UseCases []string `json:"useCases"`
					DoNotUse []string `json:"doNotUse"`
				}
				if err := json.Unmarshal([]byte(parsed.MetadataJSON), &metadata); err != nil {
					t.Fatalf("unmarshal MetadataJSON: %v", err)
				}
				if len(metadata.UseCases) == 0 {
					t.Fatalf("metadata useCases empty in %s", parsed.MetadataJSON)
				}
				if len(metadata.DoNotUse) == 0 {
					t.Fatalf("metadata doNotUse empty in %s", parsed.MetadataJSON)
				}
				assertImportedFile(t, parsed.Files, "SKILL.md", FilePurposeInstruction, false)
			})
		}
	}
}

func TestParseSkillArchiveRejectsUnsafePaths(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	writeZipFile(t, zw, "../SKILL.md", "---\nname: unsafe\n---\n")
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}

	_, err := ParseSkillArchive(bytes.NewReader(buf.Bytes()), int64(buf.Len()), "unsafe.zip")
	if err == nil {
		t.Fatal("ParseSkillArchive() error = nil, want unsafe path error")
	}
}

func TestParseSkillArchiveRejectsAggregateUncompressedBytes(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	writeZipFile(t, zw, "SKILL.md", "---\nname: too-large\n---\n")
	body := strings.Repeat("a", int(maxSkillFileBytes))
	for i := 0; i < 10; i++ {
		writeZipFile(t, zw, fmt.Sprintf("data/chunk-%02d.txt", i), body)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}

	_, err := ParseSkillArchive(bytes.NewReader(buf.Bytes()), int64(buf.Len()), "too-large.zip")
	if err == nil {
		t.Fatal("ParseSkillArchive() error = nil, want aggregate size error")
	}
}

func TestParseSkillArchiveRejectsTooManyFiles(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	writeZipFile(t, zw, "SKILL.md", "---\nname: too-many\n---\n")
	for i := 0; i < 257; i++ {
		writeZipFile(t, zw, fmt.Sprintf("data/file-%03d.txt", i), "")
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}

	_, err := ParseSkillArchive(bytes.NewReader(buf.Bytes()), int64(buf.Len()), "too-many.zip")
	if err == nil {
		t.Fatal("ParseSkillArchive() error = nil, want file count error")
	}
}

func writeZipFile(t *testing.T, zw *zip.Writer, name string, body string) {
	t.Helper()

	w, err := zw.Create(name)
	if err != nil {
		t.Fatalf("create zip file %q: %v", name, err)
	}
	if _, err := w.Write([]byte(body)); err != nil {
		t.Fatalf("write zip file %q: %v", name, err)
	}
}

func assertImportedFile(t *testing.T, files []ImportedSkillFile, path string, purpose FilePurpose, disabled bool) {
	t.Helper()

	for _, file := range files {
		if file.RelativePath != path {
			continue
		}
		if file.Purpose != purpose {
			t.Fatalf("%s purpose = %q, want %q", path, file.Purpose, purpose)
		}
		if file.ScriptDisabled != disabled {
			t.Fatalf("%s ScriptDisabled = %v, want %v", path, file.ScriptDisabled, disabled)
		}
		return
	}
	t.Fatalf("missing imported file %q in %#v", path, files)
}
