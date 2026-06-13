package store

import (
	"context"
	"database/sql"
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

func requireFTS5(t *testing.T, err error) {
	t.Helper()

	if strings.Contains(err.Error(), "no such module: fts5") {
		t.Fatalf("sqlite FTS5 support is required; run tests with: go test -tags sqlite_fts5 ./...: %v", err)
	}
}
