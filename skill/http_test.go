package skill

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestHTTPSkillRoutesImportListGetAndSelectSessionSkills(t *testing.T) {
	db := openMigratedTestDB(t)
	handler := newTestSkillHTTP(NewService(db))

	imported := postZip[Skill](t, handler, "/api/projects/project-1/skills/import", testSkillArchive(t), http.StatusCreated)
	if imported.ProjectID != "project-1" {
		t.Fatalf("imported project ID = %q, want project-1", imported.ProjectID)
	}
	if imported.Name != "style-pass" {
		t.Fatalf("imported name = %q, want style-pass", imported.Name)
	}
	if !imported.ScriptsDisabled || imported.ScriptCount != 1 {
		t.Fatalf("imported script status = (%d, %v), want disabled script", imported.ScriptCount, imported.ScriptsDisabled)
	}
	if !strings.Contains(imported.InstructionsMarkdown, "Always tighten prose.") {
		t.Fatalf("imported instructions = %q, want full instructions", imported.InstructionsMarkdown)
	}

	skills := getJSON[[]Skill](t, handler, "/api/projects/project-1/skills", http.StatusOK)
	if len(skills) != 1 {
		t.Fatalf("skills count = %d, want 1", len(skills))
	}
	summary := skills[0]
	if summary.ID != imported.ID {
		t.Fatalf("summary ID = %q, want %q", summary.ID, imported.ID)
	}
	if summary.InstructionsMarkdown != "" {
		t.Fatalf("summary instructions = %q, want omitted", summary.InstructionsMarkdown)
	}
	if len(summary.Files) == 0 {
		t.Fatal("summary files = 0, want file metadata")
	}
	for _, file := range summary.Files {
		if file.BodyText != "" {
			t.Fatalf("summary file %s body = %q, want omitted", file.RelativePath, file.BodyText)
		}
	}

	full := getJSON[Skill](t, handler, "/api/projects/project-1/skills/"+imported.ID, http.StatusOK)
	if full.ID != imported.ID {
		t.Fatalf("full ID = %q, want %q", full.ID, imported.ID)
	}
	if !strings.Contains(full.InstructionsMarkdown, "Always tighten prose.") {
		t.Fatalf("full instructions = %q, want instructions", full.InstructionsMarkdown)
	}
	if len(full.Files) < 2 {
		t.Fatalf("full files = %d, want at least 2", len(full.Files))
	}
	var foundTemplateBody, foundScriptBody bool
	for _, file := range full.Files {
		if file.RelativePath == "templates/rewrite.md" && file.BodyText == "Rewrite template" {
			foundTemplateBody = true
		}
		if file.RelativePath == "scripts/analyze.sh" && file.BodyText == "echo should-not-execute" {
			foundScriptBody = true
		}
	}
	if !foundTemplateBody || !foundScriptBody {
		t.Fatalf("full files = %#v, want template and script bodies", full.Files)
	}

	selected := putJSON[[]Skill](t, handler, "/api/projects/project-1/agent/sessions/session-1/skills", `{
		"skillIds": ["`+imported.ID+`"]
	}`, http.StatusOK)
	if len(selected) != 1 || selected[0].ID != imported.ID {
		t.Fatalf("selected = %#v, want imported skill", selected)
	}
	if selected[0].InstructionsMarkdown != "" {
		t.Fatalf("selected summary instructions = %q, want omitted", selected[0].InstructionsMarkdown)
	}

	sessionSkills := getJSON[[]Skill](t, handler, "/api/projects/project-1/agent/sessions/session-1/skills", http.StatusOK)
	if len(sessionSkills) != 1 || sessionSkills[0].ID != imported.ID {
		t.Fatalf("session skills = %#v, want imported skill", sessionSkills)
	}
}

func TestImportLocalSkillRequiresAllowedRoot(t *testing.T) {
	db := openMigratedTestDB(t)
	handler := newTestSkillHTTP(NewService(db))

	req := httptest.NewRequest(http.MethodPost, "/api/projects/project-1/skills/import-local", bytes.NewBufferString(`{
		"directory": "/tmp/some-skill"
	}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "/tmp/some-skill") {
		t.Fatalf("response leaked requested path: %s", rec.Body.String())
	}
}

func newTestSkillHTTP(service *Service) http.Handler {
	r := chi.NewRouter()
	r.Route("/api", func(r chi.Router) {
		RegisterRoutes(r, service)
	})
	return r
}

func testSkillArchive(t *testing.T) []byte {
	t.Helper()

	var body bytes.Buffer
	zw := zip.NewWriter(&body)
	writeZipEntry(t, zw, "style-pass/SKILL.md", `---
name: style-pass
description: Tightens style
routing:
  actions: [rewrite]
  contentKinds: [chapter]
  priority: 10
---
Always tighten prose.
`)
	writeZipEntry(t, zw, "style-pass/templates/rewrite.md", "Rewrite template")
	writeZipEntry(t, zw, "style-pass/scripts/analyze.sh", "echo should-not-execute")
	if err := zw.Close(); err != nil {
		t.Fatalf("close skill archive: %v", err)
	}
	return body.Bytes()
}

func writeZipEntry(t *testing.T, zw *zip.Writer, name string, body string) {
	t.Helper()
	writer, err := zw.Create(name)
	if err != nil {
		t.Fatalf("create zip entry %s: %v", name, err)
	}
	if _, err := writer.Write([]byte(body)); err != nil {
		t.Fatalf("write zip entry %s: %v", name, err)
	}
}

func getJSON[T any](t *testing.T, handler http.Handler, path string, wantStatus int) T {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return decodeTestResponse[T](t, rec, wantStatus)
}

func postZip[T any](t *testing.T, handler http.Handler, path string, body []byte, wantStatus int) T {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/zip")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return decodeTestResponse[T](t, rec, wantStatus)
}

func putJSON[T any](t *testing.T, handler http.Handler, path string, body string, wantStatus int) T {
	t.Helper()
	req := httptest.NewRequest(http.MethodPut, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return decodeTestResponse[T](t, rec, wantStatus)
}

func decodeTestResponse[T any](t *testing.T, rec *httptest.ResponseRecorder, wantStatus int) T {
	t.Helper()
	if rec.Code != wantStatus {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, wantStatus, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("content type = %q, want application/json", got)
	}
	var value T
	if err := json.NewDecoder(rec.Body).Decode(&value); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return value
}

func TestImportBuiltinDefaultSkillParsesFrontmatter(t *testing.T) {
	db := openMigratedTestDB(t)
	handler := newTestSkillHTTP(NewService(db))

	archive := zipDirFromBuiltin(t, "default", "story-sense")
	skill := postZip[Skill](t, handler, "/api/projects/project-1/skills/import", archive, http.StatusCreated)

	if skill.Name != "story-sense" {
		t.Fatalf("Name = %q, want story-sense", skill.Name)
	}
	if skill.Description == "" {
		t.Fatal("description is empty")
	}
	if len(skill.RoutingHints) == 0 {
		t.Fatal("no routing hints imported")
	}
	if skill.ScriptCount > 0 && !skill.ScriptsDisabled {
		t.Fatal("script-bearing skill should have scriptsDisabled true")
	}
}

func TestImportBuiltinOptionalSkillWithScripts(t *testing.T) {
	db := openMigratedTestDB(t)
	handler := newTestSkillHTTP(NewService(db))

	archive := zipDirFromBuiltin(t, "optional", "character-names")
	skill := postZip[Skill](t, handler, "/api/projects/project-1/skills/import", archive, http.StatusCreated)

	if skill.Name != "character-names" {
		t.Fatalf("Name = %q, want character-names", skill.Name)
	}
	if len(skill.Files) < 2 {
		t.Fatalf("character-names should have bundled data and template files, got %d", len(skill.Files))
	}
	if !skill.ScriptsDisabled {
		t.Fatal("scriptsDisabled should be true")
	}

	detail := getJSON[Skill](t, handler, "/api/projects/project-1/skills/"+skill.ID, http.StatusOK)
	foundData := false
	for _, file := range detail.Files {
		if file.Purpose == "data" {
			foundData = true
			break
		}
	}
	if !foundData {
		t.Fatal("expected data file in character-names skill")
	}
}

func TestImportBuiltinMergedSkill(t *testing.T) {
	db := openMigratedTestDB(t)
	handler := newTestSkillHTTP(NewService(db))

	archive := zipDirFromBuiltin(t, "default", "avoid-cliches")
	skill := postZip[Skill](t, handler, "/api/projects/project-1/skills/import", archive, http.StatusCreated)

	if skill.Name != "avoid-cliches" {
		t.Fatalf("Name = %q, want avoid-cliches", skill.Name)
	}
}

func zipDirFromBuiltin(t *testing.T, category, name string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	dir := "../docs/skills/builtin/" + category + "/" + name
	if _, err := os.ReadDir(dir); err != nil {
		t.Fatalf("read builtin dir %s: %v", dir, err)
	}
	var walk func(base string) error
	walk = func(base string) error {
		entries, err := os.ReadDir(base)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			fullPath := base + "/" + entry.Name()
			relPath := strings.TrimPrefix(fullPath, dir+"/")
			if entry.IsDir() {
				if err := walk(fullPath); err != nil {
					return err
				}
				continue
			}
			w, err := zw.Create(relPath)
			if err != nil {
				return err
			}
			content, err := os.ReadFile(fullPath)
			if err != nil {
				return err
			}
			if _, err := w.Write(content); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk(dir); err != nil {
		t.Fatalf("walk builtin dir: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return buf.Bytes()
}
