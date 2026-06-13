package store

import (
	"context"
	"path/filepath"
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

func TestProjectCoreMigrationCreatesStoryProjects(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "writer.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatalf("set goose dialect: %v", err)
	}
	if err := goose.Up(db, filepath.Join("..", "migrations")); err != nil {
		if strings.Contains(err.Error(), "no such module: fts5") {
			t.Skip("sqlite3 driver was built without FTS5 support")
		}
		t.Fatalf("apply migrations: %v", err)
	}

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

func TestSearchContentUsesFTSIndex(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "writer.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatalf("set goose dialect: %v", err)
	}
	if err := goose.Up(db, filepath.Join("..", "migrations")); err != nil {
		if strings.Contains(err.Error(), "no such module: fts5") {
			t.Skip("sqlite3 driver was built without FTS5 support")
		}
		t.Fatalf("apply migrations: %v", err)
	}

	if _, err := db.Exec(`
		INSERT INTO authors (id, email, password_hash, created_at)
		VALUES ('author-1', 'author@example.com', 'hash', '2026-06-13T00:00:00Z');
		INSERT INTO story_projects (id, author_id, title, slug, language, created_at, updated_at)
		VALUES ('project-1', 'author-1', 'Project', 'project', 'en', '2026-06-13T00:00:00Z', '2026-06-13T00:00:00Z');
		INSERT INTO content_items (
			id, project_id, kind, title, slug, body_markdown, metadata_json,
			sort_order, current_revision, created_at, updated_at
		) VALUES (
			'item-1', 'project-1', 'chapter', 'Opening', 'opening', 'A precise lantern scene.', '{}',
			1, 1, '2026-06-13T00:00:00Z', '2026-06-13T00:00:00Z'
		);
	`); err != nil {
		t.Fatalf("seed content: %v", err)
	}

	items, err := New(db).SearchContent(context.Background(), SearchContentParams{
		Query: "lantern",
		Limit: 10,
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
