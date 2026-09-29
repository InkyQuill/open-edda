package store

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/pressly/goose/v3"
)

func TestOpenEnablesForeignKeys(t *testing.T) {
	t.Parallel()

	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	var enabled int
	if err := db.QueryRow("PRAGMA foreign_keys").Scan(&enabled); err != nil {
		t.Fatalf("query foreign_keys: %v", err)
	}
	if enabled != 1 {
		t.Fatalf("foreign_keys = %d, want 1", enabled)
	}
}

func TestOpenEnforcesForeignKeysAcrossConnections(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, err := Open(filepath.Join(t.TempDir(), "writer.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(4)

	if _, err := db.ExecContext(ctx, `
		CREATE TABLE parents (
			id TEXT PRIMARY KEY
		);
		CREATE TABLE children (
			id TEXT PRIMARY KEY,
			parent_id TEXT NOT NULL REFERENCES parents(id)
		);
	`); err != nil {
		t.Fatalf("create foreign key tables: %v", err)
	}

	conns := make([]*sql.Conn, 0, 4)
	for i := 0; i < 4; i++ {
		conn, err := db.Conn(ctx)
		if err != nil {
			t.Fatalf("open connection %d: %v", i, err)
		}
		conns = append(conns, conn)
	}
	defer func() {
		for i, conn := range conns {
			if err := conn.Close(); err != nil {
				t.Errorf("close connection %d: %v", i, err)
			}
		}
	}()

	for i, conn := range conns {
		_, execErr := conn.ExecContext(ctx,
			"INSERT INTO children (id, parent_id) VALUES (?, ?)",
			"child-"+strconv.Itoa(i),
			"missing-parent",
		)

		if execErr == nil {
			t.Fatalf("connection %d allowed child row with missing parent", i)
		}
	}
}

func TestProjectCoreMigrationCreatesStoryProjects(t *testing.T) {
	db := openMigratedProjectCoreDB(t)
	defer db.Close()

	var tableName string
	if err := db.QueryRow(
		"SELECT name FROM sqlite_master WHERE type = 'table' AND name = 'story_projects'",
	).Scan(&tableName); err != nil {
		t.Fatalf("query story_projects table: %v", err)
	}
	if tableName != "story_projects" {
		t.Fatalf("table name = %q, want story_projects", tableName)
	}
}

func TestAgentCoreTablesExist(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	tables := []string{
		"provider_configs",
		"model_variants",
		"prompt_profiles",
		"agent_sessions",
		"agent_messages",
		"activity_events",
		"prompt_records",
		"generation_candidates",
		"prompt_context_snapshots",
		"tool_result_artifacts",
	}
	for _, table := range tables {
		var name string
		err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?`, table).Scan(&name)
		if err != nil {
			t.Fatalf("table %s missing: %v", table, err)
		}
	}
}

func TestSkillCoreTablesExist(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	tables := []string{
		"skills",
		"skill_files",
		"skill_routing_hints",
		"agent_session_skills",
	}
	for _, table := range tables {
		var name string
		err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?`, table).Scan(&name)
		if err != nil {
			t.Fatalf("table %s missing: %v", table, err)
		}
	}
	if !tableHasColumn(t, db, "generation_candidates", "skill_id") {
		t.Fatal("generation_candidates.skill_id missing")
	}
}

func TestSkillScriptRuntimeTablesExist(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	tables := []string{
		"skill_script_audits",
		"skill_script_approvals",
		"skill_script_runs",
	}
	for _, table := range tables {
		var name string
		err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?`, table).Scan(&name)
		if err != nil {
			t.Fatalf("table %s missing: %v", table, err)
		}
	}
	if !tableHasColumn(t, db, "skill_script_audits", "destructive_operations") {
		t.Fatal("skill_script_audits.destructive_operations missing")
	}
	if !tableHasColumn(t, db, "skill_script_approvals", "runtime_command") {
		t.Fatal("skill_script_approvals.runtime_command missing")
	}
	if !tableHasColumn(t, db, "skill_script_runs", "output_kind") {
		t.Fatal("skill_script_runs.output_kind missing")
	}
}

func TestSkillScriptRuntimeProjectForeignKeys(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	seedSkillScriptRuntimeProjectRows(t, db)

	expectForeignKeyFailure(t, db, "cross-project audit", `
		INSERT INTO skill_script_audits (
			id, project_id, skill_id, skill_file_id, relative_path, audited_at
		) VALUES (
			'audit-cross', 'project-2', 'skill-1', 'file-1', 'scripts/one.ts', '2026-06-13T00:00:00Z'
		);
	`)
	expectForeignKeyFailure(t, db, "cross-project approval audit", `
		INSERT INTO skill_script_approvals (
			id, project_id, skill_id, skill_file_id, audit_id, approved_at, updated_at
		) VALUES (
			'approval-cross-audit', 'project-2', 'skill-2', 'file-2', 'audit-1',
			'2026-06-13T00:00:00Z', '2026-06-13T00:00:00Z'
		);
	`)
	expectForeignKeyFailure(t, db, "cross-project run session", `
		INSERT INTO skill_script_runs (
			id, project_id, session_id, skill_id, skill_file_id, approval_id, status, created_at
		) VALUES (
			'run-cross-session', 'project-1', 'session-2', 'skill-1', 'file-1', 'approval-1',
			'succeeded', '2026-06-13T00:00:00Z'
		);
	`)
	expectForeignKeyFailure(t, db, "cross-project run approval", `
		INSERT INTO skill_script_runs (
			id, project_id, session_id, skill_id, skill_file_id, approval_id, status, created_at
		) VALUES (
			'run-cross-approval', 'project-2', 'session-2', 'skill-2', 'file-2', 'approval-1',
			'succeeded', '2026-06-13T00:00:00Z'
		);
	`)
}

func TestSkillScriptRuntimeRunSurvivesSessionDelete(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	seedSkillScriptRuntimeProjectRows(t, db)

	if _, err := db.Exec(`
		INSERT INTO skill_script_runs (
			id, project_id, session_id, skill_id, skill_file_id, approval_id, status, created_at
		) VALUES (
			'run-1', 'project-1', 'session-1', 'skill-1', 'file-1', 'approval-1',
			'succeeded', '2026-06-13T00:00:00Z'
		);
	`); err != nil {
		t.Fatalf("insert skill script run: %v", err)
	}

	if _, err := db.Exec(`DELETE FROM agent_sessions WHERE id = 'session-1'`); err != nil {
		t.Fatalf("delete agent session: %v", err)
	}

	var sessionID sql.NullString
	if err := db.QueryRow(`SELECT session_id FROM skill_script_runs WHERE id = 'run-1'`).Scan(&sessionID); err != nil {
		t.Fatalf("query skill script run after session delete: %v", err)
	}
	if sessionID.Valid {
		t.Fatalf("session_id = %q, want NULL after session delete", sessionID.String)
	}
}

func TestAddSessionSkillRejectsCrossProjectSkill(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	queries := New(db)
	ctx := context.Background()

	seedProjectScopedContent(t, db)
	now := "2026-06-13T00:00:00Z"
	if err := queries.CreateAgentSession(ctx, CreateAgentSessionParams{
		ID:             "session-1",
		ProjectID:      "project-1",
		Title:          "Rewrite",
		ActionKind:     "rewrite",
		ModelVariantID: sql.NullString{},
		ApplyMode:      "preview",
		CreatedAt:      now,
		UpdatedAt:      now,
	}); err != nil {
		t.Fatalf("create agent session: %v", err)
	}
	if err := queries.UpsertSkill(ctx, UpsertSkillParams{
		ID:                   "skill-2",
		ProjectID:            "project-2",
		Name:                 "style",
		DisplayName:          "Style",
		Description:          "",
		InstructionsMarkdown: "Write plainly.",
		SourceType:           "upload",
		SourceLabel:          "style.md",
		ScriptCount:          0,
		ScriptsDisabled:      1,
		MetadataJson:         "{}",
		InstalledAt:          now,
		UpdatedAt:            now,
	}); err != nil {
		t.Fatalf("create skill: %v", err)
	}
	rowsAffected, err := queries.AddSessionSkill(ctx, AddSessionSkillParams{
		ProjectID:  "project-1",
		SessionID:  "session-1",
		SkillID:    "skill-2",
		SelectedAt: now,
	})
	if err != nil {
		t.Fatalf("add cross-project session skill: %v", err)
	}
	if rowsAffected != 0 {
		t.Fatalf("AddSessionSkill() rows affected = %d, want 0", rowsAffected)
	}

	skills, err := queries.ListSessionSkills(ctx, ListSessionSkillsParams{
		SessionID: "session-1",
		ProjectID: "project-1",
	})
	if err != nil {
		t.Fatalf("list session skills: %v", err)
	}
	if len(skills) != 0 {
		t.Fatalf("ListSessionSkills() returned %d skills, want 0", len(skills))
	}
}

func TestAgentCoreDownMigrationRemovesRevisionColumns(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	columns := []string{"agent_session_id", "action_kind", "model_variant_id", "skill_id"}
	for _, column := range columns {
		if !tableHasColumn(t, db, "revisions", column) {
			t.Fatalf("revisions.%s missing before down migration", column)
		}
	}

	if err := goose.DownTo(db, filepath.Join("..", "migrations"), 1); err != nil {
		t.Fatalf("roll back agent core migration: %v", err)
	}

	for _, column := range columns {
		if tableHasColumn(t, db, "revisions", column) {
			t.Fatalf("revisions.%s still exists after down migration", column)
		}
	}
}

func TestAgentCoreQueriesScopeProviderAndModelByAuthor(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	queries := New(db)
	ctx := context.Background()

	seedAgentCoreAuthors(t, db)
	now := "2026-06-13T00:00:00Z"
	if err := queries.CreateProviderConfig(ctx, CreateProviderConfigParams{
		ID:              "provider-1",
		AuthorID:        "author-1",
		Name:            "OpenAI",
		BaseUrl:         "https://api.openai.example/v1",
		ApiKeyEncrypted: "secret-1",
		CreatedAt:       now,
		UpdatedAt:       now,
	}); err != nil {
		t.Fatalf("create provider config: %v", err)
	}
	if err := queries.CreateModelVariant(ctx, CreateModelVariantParams{
		ID:                        "model-1",
		ProviderConfigID:          "provider-1",
		Name:                      "Default",
		Model:                     "gpt-test",
		Temperature:               0.7,
		MaxOutputTokens:           2048,
		ContextWindowTokens:       128000,
		InputPricePerMillion:      1,
		OutputPricePerMillion:     2,
		CacheReadPricePerMillion:  0.1,
		CacheWritePricePerMillion: 0.2,
		RequestTokenField:         "max_tokens",
		ReasoningFormat:           "",
		CompatibilityJson:         "{}",
		CreatedAt:                 now,
		UpdatedAt:                 now,
	}); err != nil {
		t.Fatalf("create model variant: %v", err)
	}

	if _, err := queries.GetProviderConfig(ctx, GetProviderConfigParams{ID: "provider-1", AuthorID: "author-2"}); err != sql.ErrNoRows {
		t.Fatalf("GetProviderConfig() with wrong author error = %v, want sql.ErrNoRows", err)
	}
	if err := queries.UpdateProviderConfig(ctx, UpdateProviderConfigParams{
		BaseUrl:         "https://wrong-author.example/v1",
		ApiKeyEncrypted: "wrong-secret",
		UpdatedAt:       now,
		ID:              "provider-1",
		AuthorID:        "author-2",
	}); err != nil {
		t.Fatalf("update provider config with wrong author: %v", err)
	}
	provider, err := queries.GetProviderConfig(ctx, GetProviderConfigParams{ID: "provider-1", AuthorID: "author-1"})
	if err != nil {
		t.Fatalf("get provider config with owner: %v", err)
	}
	if provider.BaseUrl == "https://wrong-author.example/v1" {
		t.Fatal("wrong author updated provider config")
	}
	if err := queries.DeleteProviderConfig(ctx, DeleteProviderConfigParams{ID: "provider-1", AuthorID: "author-2"}); err != nil {
		t.Fatalf("delete provider config with wrong author: %v", err)
	}
	if _, err := queries.GetProviderConfig(ctx, GetProviderConfigParams{ID: "provider-1", AuthorID: "author-1"}); err != nil {
		t.Fatalf("wrong author deleted provider config: %v", err)
	}

	if _, err := queries.GetModelVariant(ctx, GetModelVariantParams{ID: "model-1", AuthorID: "author-2"}); err != sql.ErrNoRows {
		t.Fatalf("GetModelVariant() with wrong author error = %v, want sql.ErrNoRows", err)
	}
	models, err := queries.ListModelVariantsByProvider(ctx, ListModelVariantsByProviderParams{
		ProviderConfigID: "provider-1",
		AuthorID:         "author-2",
	})
	if err != nil {
		t.Fatalf("list model variants by provider with wrong author: %v", err)
	}
	if len(models) != 0 {
		t.Fatalf("ListModelVariantsByProvider() with wrong author returned %d models, want 0", len(models))
	}
	models, err = queries.ListModelVariantsByProvider(ctx, ListModelVariantsByProviderParams{
		ProviderConfigID: "provider-1",
		AuthorID:         "author-1",
	})
	if err != nil {
		t.Fatalf("list model variants by provider with owner: %v", err)
	}
	if len(models) != 1 || models[0].ID != "model-1" {
		t.Fatalf("ListModelVariantsByProvider() with owner returned %#v, want model-1", models)
	}
	if err := queries.UpdateModelVariant(ctx, UpdateModelVariantParams{
		Name:                      "Wrong author",
		Model:                     "wrong-model",
		Temperature:               1,
		MaxOutputTokens:           1,
		ContextWindowTokens:       1,
		InputPricePerMillion:      1,
		OutputPricePerMillion:     1,
		CacheReadPricePerMillion:  1,
		CacheWritePricePerMillion: 1,
		RequestTokenField:         "max_tokens",
		ReasoningFormat:           "",
		CompatibilityJson:         "{}",
		UpdatedAt:                 now,
		ID:                        "model-1",
		AuthorID:                  "author-2",
	}); err != nil {
		t.Fatalf("update model variant with wrong author: %v", err)
	}
	model, err := queries.GetModelVariant(ctx, GetModelVariantParams{ID: "model-1", AuthorID: "author-1"})
	if err != nil {
		t.Fatalf("get model variant with owner: %v", err)
	}
	if model.Model == "wrong-model" {
		t.Fatal("wrong author updated model variant")
	}
	if err := queries.DeleteModelVariant(ctx, DeleteModelVariantParams{ID: "model-1", AuthorID: "author-2"}); err != nil {
		t.Fatalf("delete model variant with wrong author: %v", err)
	}
	if _, err := queries.GetModelVariant(ctx, GetModelVariantParams{ID: "model-1", AuthorID: "author-1"}); err != nil {
		t.Fatalf("wrong author deleted model variant: %v", err)
	}
}

func TestCreateRevisionStoresAgentAttribution(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	queries := New(db)
	ctx := context.Background()

	seedProjectScopedContent(t, db)
	now := "2026-06-13T00:00:00Z"
	if err := queries.CreateProviderConfig(ctx, CreateProviderConfigParams{
		ID:              "provider-1",
		AuthorID:        "author-1",
		Name:            "OpenAI",
		BaseUrl:         "https://api.openai.example/v1",
		ApiKeyEncrypted: "secret-1",
		CreatedAt:       now,
		UpdatedAt:       now,
	}); err != nil {
		t.Fatalf("create provider config: %v", err)
	}
	if err := queries.CreateModelVariant(ctx, CreateModelVariantParams{
		ID:                        "model-1",
		ProviderConfigID:          "provider-1",
		Name:                      "Default",
		Model:                     "gpt-test",
		Temperature:               0.7,
		MaxOutputTokens:           2048,
		ContextWindowTokens:       128000,
		InputPricePerMillion:      1,
		OutputPricePerMillion:     2,
		CacheReadPricePerMillion:  0.1,
		CacheWritePricePerMillion: 0.2,
		RequestTokenField:         "max_tokens",
		ReasoningFormat:           "",
		CompatibilityJson:         "{}",
		CreatedAt:                 now,
		UpdatedAt:                 now,
	}); err != nil {
		t.Fatalf("create model variant: %v", err)
	}
	if err := queries.CreateAgentSession(ctx, CreateAgentSessionParams{
		ID:             "session-1",
		ProjectID:      "project-1",
		Title:          "Rewrite",
		ActionKind:     "rewrite",
		ModelVariantID: sql.NullString{String: "model-1", Valid: true},
		ApplyMode:      "preview",
		CreatedAt:      now,
		UpdatedAt:      now,
	}); err != nil {
		t.Fatalf("create agent session: %v", err)
	}
	if err := queries.CreateRevision(ctx, CreateRevisionParams{
		ID:             "revision-1",
		ContentItemID:  "item-1",
		RevisionNumber: 1,
		BodyMarkdown:   "Body",
		MetadataJson:   "{}",
		Reason:         "Agent rewrite",
		CreatedBy:      "agent",
		CreatedAt:      now,
		AgentSessionID: sql.NullString{String: "session-1", Valid: true},
		ActionKind:     "rewrite",
		ModelVariantID: sql.NullString{String: "model-1", Valid: true},
		SkillID:        "skill-1",
	}); err != nil {
		t.Fatalf("create revision: %v", err)
	}

	revisions, err := queries.ListRevisions(ctx, ListRevisionsParams{
		ContentItemID: "item-1",
		ProjectID:     "project-1",
	})
	if err != nil {
		t.Fatalf("list revisions: %v", err)
	}
	if len(revisions) != 1 {
		t.Fatalf("ListRevisions() returned %d revisions, want 1", len(revisions))
	}
	revision := revisions[0]
	if !revision.AgentSessionID.Valid || revision.AgentSessionID.String != "session-1" {
		t.Fatalf("AgentSessionID = %#v, want session-1", revision.AgentSessionID)
	}
	if revision.ActionKind != "rewrite" {
		t.Fatalf("ActionKind = %q, want rewrite", revision.ActionKind)
	}
	if !revision.ModelVariantID.Valid || revision.ModelVariantID.String != "model-1" {
		t.Fatalf("ModelVariantID = %#v, want model-1", revision.ModelVariantID)
	}
	if revision.SkillID != "skill-1" {
		t.Fatalf("SkillID = %q, want skill-1", revision.SkillID)
	}
}

func TestListToolResultArtifactsByProjectIncludesProjectLevelArtifacts(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	queries := New(db)
	ctx := context.Background()

	seedProjectScopedContent(t, db)
	now := "2026-06-13T00:00:00Z"
	if err := queries.CreateAgentSession(ctx, CreateAgentSessionParams{
		ID:             "session-1",
		ProjectID:      "project-1",
		Title:          "Chat",
		ActionKind:     "chat",
		ModelVariantID: sql.NullString{},
		ApplyMode:      "preview",
		CreatedAt:      now,
		UpdatedAt:      now,
	}); err != nil {
		t.Fatalf("create agent session: %v", err)
	}
	for _, artifact := range []CreateToolResultArtifactParams{
		{
			ID:                   "artifact-session",
			ProjectID:            "project-1",
			SessionID:            sql.NullString{String: "session-1", Valid: true},
			ToolCallID:           "tool-call-1",
			ToolName:             "read_content",
			FullResultJson:       "{}",
			ModelVisibleMarkdown: "visible",
			Truncated:            0,
			FullResultBytes:      2,
			CreatedAt:            now,
		},
		{
			ID:                   "artifact-project",
			ProjectID:            "project-1",
			SessionID:            sql.NullString{},
			ToolCallID:           "tool-call-2",
			ToolName:             "project_map",
			FullResultJson:       "{}",
			ModelVisibleMarkdown: "visible",
			Truncated:            0,
			FullResultBytes:      2,
			CreatedAt:            "2026-06-13T00:00:01Z",
		},
	} {
		if err := queries.CreateToolResultArtifact(ctx, artifact); err != nil {
			t.Fatalf("create tool result artifact %s: %v", artifact.ID, err)
		}
	}

	artifacts, err := queries.ListToolResultArtifactsByProject(ctx, "project-1")
	if err != nil {
		t.Fatalf("list tool result artifacts by project: %v", err)
	}
	if len(artifacts) != 2 {
		t.Fatalf("ListToolResultArtifactsByProject() returned %d artifacts, want 2", len(artifacts))
	}
	if artifacts[0].ID != "artifact-project" || artifacts[1].ID != "artifact-session" {
		t.Fatalf("artifact order = [%s, %s], want [artifact-project, artifact-session]", artifacts[0].ID, artifacts[1].ID)
	}
}

func TestSearchContentUsesFTSIndex(t *testing.T) {
	db := openMigratedProjectCoreDB(t)
	defer db.Close()

	if _, err := db.Exec(`
		INSERT INTO authors (id, email, password_hash, created_at)
		VALUES ('author-1', 'author@example.com', 'hash', '2026-06-13T00:00:00Z');
		INSERT INTO story_projects (id, author_id, title, slug, language, created_at, updated_at)
		VALUES
			('project-1', 'author-1', 'Project One', 'project-one', 'en', '2026-06-13T00:00:00Z', '2026-06-13T00:00:00Z'),
			('project-2', 'author-1', 'Project Two', 'project-two', 'en', '2026-06-13T00:00:00Z', '2026-06-13T00:00:00Z');
		INSERT INTO content_items (
			id, project_id, kind, title, slug, body_markdown, metadata_json,
			sort_order, current_revision, created_at, updated_at
		) VALUES
			(
				'item-1', 'project-1', 'chapter', 'Opening', 'opening', 'A precise lantern scene.', '{}',
				1, 1, '2026-06-13T00:00:00Z', '2026-06-13T00:00:00Z'
			),
			(
				'item-2', 'project-2', 'chapter', 'Opening', 'opening', 'Another lantern scene.', '{}',
				1, 1, '2026-06-13T00:00:00Z', '2026-06-13T00:00:00Z'
			);
	`); err != nil {
		t.Fatalf("seed content: %v", err)
	}

	items, err := New(db).SearchContent(context.Background(), SearchContentParams{
		Query:     "lantern",
		ProjectID: "project-1",
		Limit:     10,
	})
	if err != nil {
		t.Fatalf("SearchContent() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("SearchContent() returned %d items, want 1", len(items))
	}
	if items[0].ID != "item-1" {
		t.Fatalf("SearchContent() item ID = %q, want item-1", items[0].ID)
	}
}

func TestSearchContentCandidatesHasSafetyLimit(t *testing.T) {
	db := openMigratedProjectCoreDB(t)
	defer db.Close()

	if _, err := db.Exec(`
		INSERT INTO authors (id, email, password_hash, created_at)
		VALUES ('author-limit', 'limit@example.com', 'hash', '2026-06-13T00:00:00Z');
		INSERT INTO story_projects (id, author_id, title, slug, language, created_at, updated_at)
		VALUES ('project-limit', 'author-limit', 'Project Limit', 'project-limit', 'en', '2026-06-13T00:00:00Z', '2026-06-13T00:00:00Z');
	`); err != nil {
		t.Fatalf("seed project: %v", err)
	}

	stmt, err := db.Prepare(`
		INSERT INTO content_items (
			id, project_id, kind, title, slug, body_markdown, metadata_json,
			sort_order, current_revision, created_at, updated_at
		) VALUES (?, 'project-limit', 'chapter', ?, ?, ?, '{}', ?, 1, '2026-06-13T00:00:00Z', '2026-06-13T00:00:00Z');
	`)
	if err != nil {
		t.Fatalf("prepare content insert: %v", err)
	}
	defer stmt.Close()

	for i := 0; i < 1005; i++ {
		if _, err := stmt.Exec(
			fmt.Sprintf("limit-item-%04d", i),
			fmt.Sprintf("Needle Item %04d", i),
			fmt.Sprintf("needle-item-%04d", i),
			fmt.Sprintf("needle body %04d", i),
			i,
		); err != nil {
			t.Fatalf("insert content item %d: %v", i, err)
		}
	}

	items, err := New(db).SearchContentCandidates(context.Background(), SearchContentCandidatesParams{
		Query:     "needle",
		ProjectID: "project-limit",
	})
	if err != nil {
		t.Fatalf("SearchContentCandidates() error = %v", err)
	}
	if len(items) != 1000 {
		t.Fatalf("SearchContentCandidates() returned %d items, want 1000", len(items))
	}
}

func TestProjectCoreMigrationRejectsCrossProjectEntryRelations(t *testing.T) {
	db := openMigratedProjectCoreDB(t)
	defer db.Close()
	seedProjectScopedContent(t, db)

	_, err := db.Exec(`
		INSERT INTO entry_relations (
			id, project_id, source_item_id, target_item_id, target_title, relation_type, created_at
		) VALUES (
			'relation-cross-source', 'project-1', 'item-2', NULL, 'Imported target', 'mentions', '2026-06-13T00:00:00Z'
		);
	`)
	if err == nil {
		t.Fatal("cross-project source_item_id insert succeeded, want foreign key failure")
	}

	_, err = db.Exec(`
		INSERT INTO entry_relations (
			id, project_id, source_item_id, target_item_id, target_title, relation_type, created_at
		) VALUES (
			'relation-cross-target', 'project-1', 'item-1', 'item-2', 'Other project target', 'mentions', '2026-06-13T00:00:00Z'
		);
	`)
	if err == nil {
		t.Fatal("cross-project target_item_id insert succeeded, want foreign key failure")
	}

	if _, err := db.Exec(`
		INSERT INTO entry_relations (
			id, project_id, source_item_id, target_item_id, target_title, relation_type, created_at
		) VALUES (
			'relation-unresolved-target', 'project-1', 'item-1', NULL, 'Imported target', 'mentions', '2026-06-13T00:00:00Z'
		);
	`); err != nil {
		t.Fatalf("insert unresolved target relation: %v", err)
	}
}

func TestProjectCoreMigrationRejectsCrossProjectAttachedNotes(t *testing.T) {
	db := openMigratedProjectCoreDB(t)
	defer db.Close()
	seedProjectScopedContent(t, db)

	_, err := db.Exec(`
		INSERT INTO attached_notes (
			id, project_id, content_item_id, title, body_markdown, source, created_at, updated_at
		) VALUES (
			'note-cross-content', 'project-1', 'item-2', 'Cross project note', 'Body', 'author',
			'2026-06-13T00:00:00Z', '2026-06-13T00:00:00Z'
		);
	`)
	if err == nil {
		t.Fatal("cross-project attached note insert succeeded, want foreign key failure")
	}

	if _, err := db.Exec(`
		INSERT INTO attached_notes (
			id, project_id, content_item_id, title, body_markdown, source, created_at, updated_at
		) VALUES (
			'note-project-level', 'project-1', NULL, 'Project note', 'Body', 'author',
			'2026-06-13T00:00:00Z', '2026-06-13T00:00:00Z'
		);
	`); err != nil {
		t.Fatalf("insert project-level attached note: %v", err)
	}
}

func openMigratedProjectCoreDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := Open(filepath.Join(t.TempDir(), "writer.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	if err := goose.SetDialect("sqlite3"); err != nil {
		_ = db.Close()
		t.Fatalf("set goose dialect: %v", err)
	}
	if err := goose.Up(db, filepath.Join("..", "migrations")); err != nil {
		_ = db.Close()
		requireFTS5(t, err)
		t.Fatalf("apply migrations: %v", err)
	}

	return db
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()

	return openMigratedProjectCoreDB(t)
}

func tableHasColumn(t *testing.T, db *sql.DB, tableName string, columnName string) bool {
	t.Helper()

	rows, err := db.Query("PRAGMA table_info(" + tableName + ")")
	if err != nil {
		t.Fatalf("query %s columns: %v", tableName, err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			cid        int
			name       string
			columnType string
			notNull    int
			defaultVal sql.NullString
			pk         int
		)
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultVal, &pk); err != nil {
			t.Fatalf("scan %s column: %v", tableName, err)
		}
		if name == columnName {
			return true
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate %s columns: %v", tableName, err)
	}
	return false
}

func expectForeignKeyFailure(t *testing.T, db *sql.DB, label string, statement string) {
	t.Helper()

	_, err := db.Exec(statement)
	if err == nil {
		t.Fatalf("%s insert succeeded, want foreign key failure", label)
	}
	if !strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
		t.Fatalf("%s error = %v, want foreign key failure", label, err)
	}
}

func seedProjectScopedContent(t *testing.T, db *sql.DB) {
	t.Helper()

	if _, err := db.Exec(`
		INSERT INTO authors (id, email, password_hash, created_at)
		VALUES ('author-1', 'author@example.com', 'hash', '2026-06-13T00:00:00Z');
		INSERT INTO story_projects (id, author_id, title, slug, language, created_at, updated_at)
		VALUES
			('project-1', 'author-1', 'Project One', 'project-one', 'en', '2026-06-13T00:00:00Z', '2026-06-13T00:00:00Z'),
			('project-2', 'author-1', 'Project Two', 'project-two', 'en', '2026-06-13T00:00:00Z', '2026-06-13T00:00:00Z');
		INSERT INTO content_items (
			id, project_id, kind, title, slug, body_markdown, metadata_json,
			sort_order, current_revision, created_at, updated_at
		) VALUES
			('item-1', 'project-1', 'story_bible_entry', 'Item One', 'item-one', 'Body one', '{}', 1, 1, '2026-06-13T00:00:00Z', '2026-06-13T00:00:00Z'),
			('item-2', 'project-2', 'story_bible_entry', 'Item Two', 'item-two', 'Body two', '{}', 1, 1, '2026-06-13T00:00:00Z', '2026-06-13T00:00:00Z');
	`); err != nil {
		t.Fatalf("seed project-scoped content: %v", err)
	}
}

func seedSkillScriptRuntimeProjectRows(t *testing.T, db *sql.DB) {
	t.Helper()

	if _, err := db.Exec(`
		INSERT INTO authors (id, email, password_hash, created_at)
		VALUES ('author-1', 'author@example.com', 'hash', '2026-06-13T00:00:00Z');
		INSERT INTO story_projects (id, author_id, title, slug, language, created_at, updated_at)
		VALUES
			('project-1', 'author-1', 'Project One', 'project-one', 'en', '2026-06-13T00:00:00Z', '2026-06-13T00:00:00Z'),
			('project-2', 'author-1', 'Project Two', 'project-two', 'en', '2026-06-13T00:00:00Z', '2026-06-13T00:00:00Z');
		INSERT INTO skills (
			id, project_id, name, display_name, source_type, installed_at, updated_at
		) VALUES
			('skill-1', 'project-1', 'one', 'One', 'upload', '2026-06-13T00:00:00Z', '2026-06-13T00:00:00Z'),
			('skill-2', 'project-2', 'two', 'Two', 'upload', '2026-06-13T00:00:00Z', '2026-06-13T00:00:00Z');
		INSERT INTO skill_files (
			id, skill_id, relative_path, purpose, body_text, created_at
		) VALUES
			('file-1', 'skill-1', 'scripts/one.ts', 'script', 'one', '2026-06-13T00:00:00Z'),
			('file-2', 'skill-2', 'scripts/two.ts', 'script', 'two', '2026-06-13T00:00:00Z');
		INSERT INTO agent_sessions (
			id, project_id, title, action_kind, apply_mode, created_at, updated_at
		) VALUES
			('session-1', 'project-1', 'One', 'chat', 'preview', '2026-06-13T00:00:00Z', '2026-06-13T00:00:00Z'),
			('session-2', 'project-2', 'Two', 'chat', 'preview', '2026-06-13T00:00:00Z', '2026-06-13T00:00:00Z');
		INSERT INTO skill_script_audits (
			id, project_id, skill_id, skill_file_id, relative_path, audited_at
		) VALUES
			('audit-1', 'project-1', 'skill-1', 'file-1', 'scripts/one.ts', '2026-06-13T00:00:00Z'),
			('audit-2', 'project-2', 'skill-2', 'file-2', 'scripts/two.ts', '2026-06-13T00:00:00Z');
		INSERT INTO skill_script_approvals (
			id, project_id, skill_id, skill_file_id, audit_id, approved_at, updated_at
		) VALUES
			('approval-1', 'project-1', 'skill-1', 'file-1', 'audit-1', '2026-06-13T00:00:00Z', '2026-06-13T00:00:00Z');
	`); err != nil {
		t.Fatalf("seed skill script runtime project rows: %v", err)
	}
}

func seedAgentCoreAuthors(t *testing.T, db *sql.DB) {
	t.Helper()

	if _, err := db.Exec(`
		INSERT INTO authors (id, email, password_hash, created_at)
		VALUES
			('author-1', 'author-1@example.com', 'hash', '2026-06-13T00:00:00Z'),
			('author-2', 'author-2@example.com', 'hash', '2026-06-13T00:00:00Z');
	`); err != nil {
		t.Fatalf("seed agent core authors: %v", err)
	}
}

func requireFTS5(t *testing.T, err error) {
	t.Helper()

	if strings.Contains(err.Error(), "no such module: fts5") {
		t.Fatalf("sqlite FTS5 support is required; run tests with: go test -tags sqlite_fts5 ./...: %v", err)
	}
}
