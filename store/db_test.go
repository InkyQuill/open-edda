package store

import (
	"context"
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

	for i := 0; i < 4; i++ {
		conn, err := db.Conn(ctx)
		if err != nil {
			t.Fatalf("open connection %d: %v", i, err)
		}

		_, execErr := conn.ExecContext(ctx,
			"INSERT INTO children (id, parent_id) VALUES (?, ?)",
			"child-"+strconv.Itoa(i),
			"missing-parent",
		)
		closeErr := conn.Close()

		if execErr == nil {
			t.Fatalf("connection %d allowed child row with missing parent", i)
		}
		if closeErr != nil {
			t.Fatalf("close connection %d: %v", i, closeErr)
		}
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
		requireFTS5(t, err)
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
		requireFTS5(t, err)
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

func requireFTS5(t *testing.T, err error) {
	t.Helper()

	if strings.Contains(err.Error(), "no such module: fts5") {
		t.Fatalf("sqlite FTS5 support is required; run tests with: go test -tags sqlite_fts5 ./...: %v", err)
	}
}
