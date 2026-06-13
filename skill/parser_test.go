package skill

import (
	"archive/zip"
	"bytes"
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
