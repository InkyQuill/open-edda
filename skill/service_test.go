package skill

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"git.inkyquill.net/inky/writer/store"
	"github.com/pressly/goose/v3"
)

func TestServiceInstallsListsGetsReimportsAndRendersSkill(t *testing.T) {
	db := openMigratedTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	installed, err := service.Install(ctx, InstallInput{
		ProjectID:   "project-1",
		SourceType:  SourceTypeUpload,
		SourceLabel: "style-pass.zip",
		Imported:    importedStylePass("Rewrite template"),
	})
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if installed.Name != "style-pass" {
		t.Fatalf("installed skill name = %q, want style-pass", installed.Name)
	}
	if installed.ScriptCount != 1 || !installed.ScriptsDisabled {
		t.Fatalf("script status = (%d, %v), want disabled script", installed.ScriptCount, installed.ScriptsDisabled)
	}
	if len(installed.Files) != 2 {
		t.Fatalf("installed files = %d, want 2", len(installed.Files))
	}
	if len(installed.RoutingHints) != 1 {
		t.Fatalf("installed routing hints = %d, want 1", len(installed.RoutingHints))
	}

	skills, err := service.List(ctx, "project-1")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(skills) != 1 {
		t.Fatalf("List() returned %d skills, want 1", len(skills))
	}
	got := skills[0]
	if got.InstructionsMarkdown != "" {
		t.Fatal("list response should not include instructions")
	}
	if got.ScriptCount != 1 || !got.ScriptsDisabled {
		t.Fatalf("script status = (%d, %v), want disabled script", got.ScriptCount, got.ScriptsDisabled)
	}
	if len(got.Files) != 2 {
		t.Fatalf("list files = %d, want 2", len(got.Files))
	}
	if got.Files[0].BodyText != "" {
		t.Fatal("list response should not include file body text")
	}

	fetched, err := service.Get(ctx, "project-1", installed.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if fetched.InstructionsMarkdown != "Always tighten prose." {
		t.Fatalf("instructions = %q, want full instructions", fetched.InstructionsMarkdown)
	}
	if len(fetched.Files) != 2 || fetched.Files[0].BodyText == "" || fetched.Files[1].BodyText == "" {
		t.Fatalf("Get() files = %#v, want file bodies", fetched.Files)
	}
	if len(fetched.RoutingHints) != 1 || fetched.RoutingHints[0].ActionKind != "rewrite" {
		t.Fatalf("routing hints = %#v, want rewrite hint", fetched.RoutingHints)
	}

	byName, err := service.GetByName(ctx, "project-1", "style-pass")
	if err != nil {
		t.Fatalf("GetByName() error = %v", err)
	}
	if byName.ID != installed.ID {
		t.Fatalf("GetByName() ID = %q, want %q", byName.ID, installed.ID)
	}

	routable, err := service.ListRoutable(ctx, "project-1", "rewrite", "chapter")
	if err != nil {
		t.Fatalf("ListRoutable() error = %v", err)
	}
	if len(routable) != 1 || routable[0].ID != installed.ID {
		t.Fatalf("ListRoutable() = %#v, want installed skill", routable)
	}

	reinstalled, err := service.Install(ctx, InstallInput{
		ProjectID:   "project-1",
		SourceType:  SourceTypeUpload,
		SourceLabel: "style-pass-v2.zip",
		Imported: ImportedSkill{
			Name:                 "style-pass",
			DisplayName:          "Style Pass",
			Description:          "Second import",
			InstructionsMarkdown: "Prefer direct language.",
			MetadataJSON:         "{}",
			SourceLabel:          "style-pass-v2.zip",
			Files: []ImportedSkillFile{
				{
					RelativePath: "templates/rewrite.md",
					Purpose:      FilePurposeTemplate,
					MediaType:    "text/markdown; charset=utf-8",
					BodyText:     "Replacement template",
					Bytes:        int64(len("Replacement template")),
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("reinstall Install() error = %v", err)
	}
	if reinstalled.ID != installed.ID {
		t.Fatalf("reinstalled ID = %q, want existing ID %q", reinstalled.ID, installed.ID)
	}
	if len(reinstalled.Files) != 1 {
		t.Fatalf("reinstalled files = %d, want old files replaced", len(reinstalled.Files))
	}
	if reinstalled.Files[0].BodyText != "Replacement template" {
		t.Fatalf("reinstalled file body = %q, want replacement", reinstalled.Files[0].BodyText)
	}

	renderedSkill, err := service.Install(ctx, InstallInput{
		ProjectID:   "project-1",
		SourceType:  SourceTypeUpload,
		SourceLabel: "style-pass.zip",
		Imported:    importedStylePass("Rewrite template"),
	})
	if err != nil {
		t.Fatalf("restore Install() error = %v", err)
	}
	rendered, skillForModel, err := service.RenderForModel(ctx, RenderSkillInput{
		ProjectID: "project-1",
		SkillID:   renderedSkill.ID,
	})
	if err != nil {
		t.Fatalf("RenderForModel() error = %v", err)
	}
	if skillForModel.ID != renderedSkill.ID {
		t.Fatalf("RenderForModel() skill ID = %q, want %q", skillForModel.ID, renderedSkill.ID)
	}
	for _, want := range []string{
		`<skill_content name="style-pass">`,
		`<instructions>`,
		`Always tighten prose.`,
		`<script_status disabled="true" count="1">Script files are imported for reference only. Writer does not execute bundled skill scripts in v1.</script_status>`,
		`<file path="templates/rewrite.md" purpose="template" bytes="16" readable="true">Use read_skill_file with this path to load the file only if needed.</file>`,
		`<file path="scripts/analyze.sh" purpose="script" bytes="23" disabled="true" readable="false">Script file is disabled and not readable through skill file loading.</file>`,
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("RenderForModel() missing %q in:\n%s", want, rendered)
		}
	}
	if strings.Contains(rendered, "Rewrite template") {
		t.Fatalf("RenderForModel() included supporting file body:\n%s", rendered)
	}
	if strings.Contains(rendered, "echo should-not-execute") {
		t.Fatalf("RenderForModel() included executable script body:\n%s", rendered)
	}

	renderedFile, fileForModel, skillForFile, err := service.RenderFileForModel(ctx, RenderSkillFileInput{
		ProjectID: "project-1",
		SkillID:   renderedSkill.ID,
		Path:      "templates/rewrite.md",
	})
	if err != nil {
		t.Fatalf("RenderFileForModel() error = %v", err)
	}
	if skillForFile.ID != renderedSkill.ID || fileForModel.RelativePath != "templates/rewrite.md" {
		t.Fatalf("RenderFileForModel() skill/file = %q/%q", skillForFile.ID, fileForModel.RelativePath)
	}
	if !strings.Contains(renderedFile, "Rewrite template") {
		t.Fatalf("RenderFileForModel() missing file body:\n%s", renderedFile)
	}
	if _, _, _, err := service.RenderFileForModel(ctx, RenderSkillFileInput{
		ProjectID: "project-1",
		SkillID:   renderedSkill.ID,
		Path:      "scripts/analyze.sh",
	}); err == nil {
		t.Fatal("RenderFileForModel(script) error = nil, want error")
	}
}

func TestServiceInstallRecordsSkillImportedActivity(t *testing.T) {
	db := openMigratedTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	installed, err := service.Install(ctx, InstallInput{
		ProjectID:   "project-1",
		SourceType:  SourceTypeUpload,
		SourceLabel: "style-pass.zip",
		Imported:    importedStylePass("Rewrite template"),
	})
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	events, err := store.New(db).ListActivityEvents(ctx, store.ListActivityEventsParams{ProjectID: "project-1", Limit: 10})
	if err != nil {
		t.Fatalf("ListActivityEvents() error = %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("activity event count = %d, want 1", len(events))
	}
	event := events[0]
	if event.EventType != "skill_imported" {
		t.Fatalf("event type = %q, want skill_imported", event.EventType)
	}
	if event.SessionID.Valid {
		t.Fatalf("event session ID = %#v, want empty", event.SessionID)
	}
	if event.Summary != "Imported skill style-pass" {
		t.Fatalf("event summary = %q", event.Summary)
	}
	var metadata map[string]any
	if err := json.Unmarshal([]byte(event.MetadataJson), &metadata); err != nil {
		t.Fatalf("unmarshal metadata: %v", err)
	}
	assertMetadataString(t, metadata, "skillId", installed.ID)
	assertMetadataString(t, metadata, "name", "style-pass")
	assertMetadataFloat(t, metadata, "scriptCount", 1)
	assertMetadataBool(t, metadata, "scriptsDisabled", true)
	assertMetadataString(t, metadata, "sourceType", "upload")
}

func TestServiceSelectsSessionSkillsAndRejectsCrossProjectSkill(t *testing.T) {
	db := openMigratedTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	projectSkill, err := service.Install(ctx, InstallInput{
		ProjectID:   "project-1",
		SourceType:  SourceTypeUpload,
		SourceLabel: "style-pass.zip",
		Imported:    importedStylePass("Rewrite template"),
	})
	if err != nil {
		t.Fatalf("Install(project-1) error = %v", err)
	}
	otherSkill, err := service.Install(ctx, InstallInput{
		ProjectID:   "project-2",
		SourceType:  SourceTypeUpload,
		SourceLabel: "other.zip",
		Imported: ImportedSkill{
			Name:                 "other-project",
			DisplayName:          "Other Project",
			InstructionsMarkdown: "Other project only.",
			MetadataJSON:         "{}",
		},
	})
	if err != nil {
		t.Fatalf("Install(project-2) error = %v", err)
	}

	selected, err := service.SelectSessionSkills(ctx, SelectSessionSkillsInput{
		ProjectID: "project-1",
		SessionID: "session-1",
		SkillIDs:  []string{projectSkill.ID},
	})
	if err != nil {
		t.Fatalf("SelectSessionSkills() error = %v", err)
	}
	if len(selected) != 1 || selected[0].ID != projectSkill.ID {
		t.Fatalf("SelectSessionSkills() = %#v, want project skill", selected)
	}

	sessionSkills, err := service.ListSessionSkills(ctx, "project-1", "session-1")
	if err != nil {
		t.Fatalf("ListSessionSkills() error = %v", err)
	}
	if len(sessionSkills) != 1 || sessionSkills[0].ID != projectSkill.ID {
		t.Fatalf("ListSessionSkills() = %#v, want project skill", sessionSkills)
	}

	_, err = service.SelectSessionSkills(ctx, SelectSessionSkillsInput{
		ProjectID: "project-1",
		SessionID: "session-1",
		SkillIDs:  []string{otherSkill.ID},
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("SelectSessionSkills(cross-project) error = %v, want ErrInvalidInput", err)
	}
}

func TestServiceSelectSessionSkillsRecordsSkillSelectedActivity(t *testing.T) {
	db := openMigratedTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	selectedSkill, err := service.Install(ctx, InstallInput{
		ProjectID:   "project-1",
		SourceType:  SourceTypeUpload,
		SourceLabel: "style-pass.zip",
		Imported:    importedStylePass("Rewrite template"),
	})
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	if _, err := service.SelectSessionSkills(ctx, SelectSessionSkillsInput{
		ProjectID: "project-1",
		SessionID: "session-1",
		SkillIDs:  []string{selectedSkill.ID},
	}); err != nil {
		t.Fatalf("SelectSessionSkills() error = %v", err)
	}

	events, err := store.New(db).ListActivityEvents(ctx, store.ListActivityEventsParams{ProjectID: "project-1", Limit: 10})
	if err != nil {
		t.Fatalf("ListActivityEvents() error = %v", err)
	}
	var selectedEvent store.ActivityEvent
	found := false
	for _, event := range events {
		if event.EventType == "skill_selected" {
			selectedEvent = event
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("events = %#v, want skill_selected", events)
	}
	if !selectedEvent.SessionID.Valid || selectedEvent.SessionID.String != "session-1" {
		t.Fatalf("event session ID = %#v, want session-1", selectedEvent.SessionID)
	}
	var metadata map[string]any
	if err := json.Unmarshal([]byte(selectedEvent.MetadataJson), &metadata); err != nil {
		t.Fatalf("unmarshal metadata: %v", err)
	}
	assertMetadataString(t, metadata, "skillId", selectedSkill.ID)
	assertMetadataString(t, metadata, "name", "style-pass")
	assertMetadataFloat(t, metadata, "scriptCount", 1)
	assertMetadataBool(t, metadata, "scriptsDisabled", true)
}

func TestServiceSelectSessionSkillsDeduplicatesSkillSelectedActivity(t *testing.T) {
	db := openMigratedTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	selectedSkill, err := service.Install(ctx, InstallInput{
		ProjectID:   "project-1",
		SourceType:  SourceTypeUpload,
		SourceLabel: "style-pass.zip",
		Imported:    importedStylePass("Rewrite template"),
	})
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	if _, err := service.SelectSessionSkills(ctx, SelectSessionSkillsInput{
		ProjectID: "project-1",
		SessionID: "session-1",
		SkillIDs:  []string{selectedSkill.ID, selectedSkill.ID},
	}); err != nil {
		t.Fatalf("SelectSessionSkills() error = %v", err)
	}

	events, err := store.New(db).ListActivityEvents(ctx, store.ListActivityEventsParams{ProjectID: "project-1", Limit: 10})
	if err != nil {
		t.Fatalf("ListActivityEvents() error = %v", err)
	}
	if got := countActivityEvents(events, "skill_selected"); got != 1 {
		t.Fatalf("skill_selected event count = %d, want 1 in %#v", got, events)
	}
}

func TestServiceSelectSessionSkillsClearsWithoutSkillSelectedActivity(t *testing.T) {
	db := openMigratedTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	selectedSkill, err := service.Install(ctx, InstallInput{
		ProjectID:   "project-1",
		SourceType:  SourceTypeUpload,
		SourceLabel: "style-pass.zip",
		Imported:    importedStylePass("Rewrite template"),
	})
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	if _, err := service.SelectSessionSkills(ctx, SelectSessionSkillsInput{
		ProjectID: "project-1",
		SessionID: "session-1",
		SkillIDs:  []string{selectedSkill.ID},
	}); err != nil {
		t.Fatalf("SelectSessionSkills(select) error = %v", err)
	}

	if _, err := service.SelectSessionSkills(ctx, SelectSessionSkillsInput{
		ProjectID: "project-1",
		SessionID: "session-1",
		SkillIDs:  []string{},
	}); err != nil {
		t.Fatalf("SelectSessionSkills(clear) error = %v", err)
	}

	sessionSkills, err := service.ListSessionSkills(ctx, "project-1", "session-1")
	if err != nil {
		t.Fatalf("ListSessionSkills() error = %v", err)
	}
	if len(sessionSkills) != 0 {
		t.Fatalf("session skills after clear = %#v, want none", sessionSkills)
	}
	events, err := store.New(db).ListActivityEvents(ctx, store.ListActivityEventsParams{ProjectID: "project-1", Limit: 10})
	if err != nil {
		t.Fatalf("ListActivityEvents() error = %v", err)
	}
	if got := countActivityEvents(events, "skill_selected"); got != 1 {
		t.Fatalf("skill_selected event count after clear = %d, want 1 in %#v", got, events)
	}
}

func importedStylePass(templateBody string) ImportedSkill {
	return ImportedSkill{
		Name:                 "style-pass",
		DisplayName:          "Style Pass",
		Description:          "Tightens style",
		InstructionsMarkdown: "Always tighten prose.",
		MetadataJSON:         "{}",
		SourceLabel:          "style-pass.zip",
		ScriptCount:          1,
		ScriptsDisabled:      true,
		Files: []ImportedSkillFile{
			{
				RelativePath: "templates/rewrite.md",
				Purpose:      FilePurposeTemplate,
				MediaType:    "text/markdown; charset=utf-8",
				BodyText:     templateBody,
				Bytes:        int64(len(templateBody)),
			},
			{
				RelativePath:   "scripts/analyze.sh",
				Purpose:        FilePurposeScript,
				MediaType:      "text/x-shellscript; charset=utf-8",
				BodyText:       "echo should-not-execute",
				Bytes:          int64(len("echo should-not-execute")),
				ScriptDisabled: true,
			},
		},
		RoutingHints: []RoutingHint{
			{
				ActionKind:  "rewrite",
				ContentKind: "chapter",
				Tag:         "style",
				Priority:    10,
			},
		},
	}
}

func countActivityEvents(events []store.ActivityEvent, eventType string) int {
	count := 0
	for _, event := range events {
		if event.EventType == eventType {
			count++
		}
	}
	return count
}

func assertMetadataString(t *testing.T, metadata map[string]any, key, want string) {
	t.Helper()
	if metadata[key] != want {
		t.Fatalf("metadata[%s] = %#v, want %q", key, metadata[key], want)
	}
}

func assertMetadataFloat(t *testing.T, metadata map[string]any, key string, want float64) {
	t.Helper()
	if metadata[key] != want {
		t.Fatalf("metadata[%s] = %#v, want %v", key, metadata[key], want)
	}
}

func assertMetadataBool(t *testing.T, metadata map[string]any, key string, want bool) {
	t.Helper()
	if metadata[key] != want {
		t.Fatalf("metadata[%s] = %#v, want %v", key, metadata[key], want)
	}
}

func openMigratedTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := store.Open(filepath.Join(t.TempDir(), "writer.db"))
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatalf("set goose dialect: %v", err)
	}
	if err := goose.Up(db, filepath.Join("..", "migrations")); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO authors (id, email, password_hash, created_at)
		VALUES ('author-1', 'author@example.com', 'hash', '2026-06-13T00:00:00Z');
		INSERT INTO story_projects (id, author_id, title, slug, language, created_at, updated_at)
		VALUES
			('project-1', 'author-1', 'Test', 'test', 'en', '2026-06-13T00:00:00Z', '2026-06-13T00:00:00Z'),
			('project-2', 'author-1', 'Other', 'other', 'en', '2026-06-13T00:00:00Z', '2026-06-13T00:00:00Z');
		INSERT INTO provider_configs (id, author_id, name, base_url, api_key_encrypted, created_at, updated_at)
		VALUES ('provider-1', 'author-1', 'Test Provider', 'http://example.test', 'encrypted', '2026-06-13T00:00:00Z', '2026-06-13T00:00:00Z');
		INSERT INTO model_variants (
			id, provider_config_id, name, model, temperature, max_output_tokens,
			context_window_tokens, input_price_per_million, output_price_per_million,
			cache_read_price_per_million, cache_write_price_per_million,
			request_token_field, reasoning_format, compatibility_json, created_at, updated_at
		) VALUES (
			'model-1', 'provider-1', 'Test Model', 'test-model', 0.7, 2048,
			8192, 0, 0, 0, 0, 'max_tokens', '', '{}', '2026-06-13T00:00:00Z', '2026-06-13T00:00:00Z'
		);
		INSERT INTO agent_sessions (id, project_id, title, action_kind, model_variant_id, apply_mode, created_at, updated_at)
		VALUES ('session-1', 'project-1', 'Test Session', 'rewrite', 'model-1', 'direct_apply', '2026-06-13T00:00:00Z', '2026-06-13T00:00:00Z');
	`)
	if err != nil {
		t.Fatalf("seed project: %v", err)
	}

	return db
}
