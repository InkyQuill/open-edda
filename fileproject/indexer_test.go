package fileproject

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"git.inkyquill.net/inky/writer/store"
	"github.com/pressly/goose/v3"
)

func TestRebuildIndexMirrorsCurrentFiles(t *testing.T) {
	db := openFileIndexTestDB(t)
	seedFileIndexProject(t, db)
	root := t.TempDir()
	mustWrite(t, root, "story/chapter-01.md", "# Chapter 1\n")
	mustWrite(t, root, "characters/Protagonist.md", "# Protagonist\n")

	result, err := RebuildIndex(context.Background(), db, "project-1", root)
	if err != nil {
		t.Fatalf("RebuildIndex() error = %v", err)
	}
	if len(result.Files) != 2 {
		t.Fatalf("indexed files = %d, want 2", len(result.Files))
	}

	rows := loadIndexedFiles(t, db)
	story := rows["story/chapter-01.md"]
	if story.ID == "" {
		t.Fatalf("story row missing: %#v", rows)
	}
	if story.Kind != string(LayoutKindStory) || story.Title != "chapter 01" || story.SHA256 != result.Files[1].SHA256 || story.Bytes != result.Files[1].Size {
		t.Fatalf("story row = %#v, files = %#v", story, result.Files)
	}
	character := rows["characters/Protagonist.md"]
	if character.ID == "" {
		t.Fatalf("character row missing: %#v", rows)
	}
	if character.Kind != string(LayoutKindCharacter) || character.Title != "Protagonist" || character.SHA256 != result.Files[0].SHA256 || character.Bytes != result.Files[0].Size {
		t.Fatalf("character row = %#v, files = %#v", character, result.Files)
	}
}

func TestRebuildIndexRemovesDisappearedFiles(t *testing.T) {
	db := openFileIndexTestDB(t)
	seedFileIndexProject(t, db)
	root := t.TempDir()
	mustWrite(t, root, "story/chapter-01.md", "# Chapter 1\n")
	mustWrite(t, root, "story/chapter-02.md", "# Chapter 2\n")

	if _, err := RebuildIndex(context.Background(), db, "project-1", root); err != nil {
		t.Fatalf("first RebuildIndex() error = %v", err)
	}
	if err := os.Remove(filepath.Join(root, "story", "chapter-02.md")); err != nil {
		t.Fatalf("remove chapter 2: %v", err)
	}
	if _, err := RebuildIndex(context.Background(), db, "project-1", root); err != nil {
		t.Fatalf("second RebuildIndex() error = %v", err)
	}

	rows := loadIndexedFiles(t, db)
	if _, ok := rows["story/chapter-02.md"]; ok {
		t.Fatalf("deleted file remains indexed: %#v", rows)
	}
	if rows["story/chapter-01.md"].ID == "" {
		t.Fatalf("remaining file missing: %#v", rows)
	}
}

func TestRebuildIndexPreservesUpdatedAtForUnchangedFiles(t *testing.T) {
	db := openFileIndexTestDB(t)
	seedFileIndexProject(t, db)
	root := t.TempDir()
	mustWrite(t, root, "story/chapter-01.md", "# Chapter 1\n")

	if _, err := RebuildIndex(context.Background(), db, "project-1", root); err != nil {
		t.Fatalf("first RebuildIndex() error = %v", err)
	}
	rows := loadIndexedFiles(t, db)
	firstUpdatedAt := rows["story/chapter-01.md"].UpdatedAt
	time.Sleep(time.Nanosecond)
	if _, err := RebuildIndex(context.Background(), db, "project-1", root); err != nil {
		t.Fatalf("second RebuildIndex() error = %v", err)
	}
	rows = loadIndexedFiles(t, db)
	if rows["story/chapter-01.md"].UpdatedAt != firstUpdatedAt {
		t.Fatalf("updated_at changed for unchanged file: first=%q second=%q", firstUpdatedAt, rows["story/chapter-01.md"].UpdatedAt)
	}
}

func openFileIndexTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := store.Open(filepath.Join(t.TempDir(), "writer.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatalf("set goose dialect: %v", err)
	}
	if err := goose.Up(db, filepath.Join("..", "migrations")); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
	return db
}

func seedFileIndexProject(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec(`
		INSERT INTO authors (id, email, password_hash, created_at)
		VALUES ('author-1', 'author@example.com', 'hash', '2026-07-01T00:00:00Z');
		INSERT INTO story_projects (id, author_id, title, slug, language, created_at, updated_at)
		VALUES ('project-1', 'author-1', 'Alchemy Draft', 'alchemy-draft', 'en', '2026-07-01T00:00:00Z', '2026-07-01T00:00:00Z');
	`)
	if err != nil {
		t.Fatalf("seed project: %v", err)
	}
}

type indexedFileRow struct {
	ID        string
	Kind      string
	Title     string
	SHA256    string
	Bytes     int64
	UpdatedAt string
}

func loadIndexedFiles(t *testing.T, db *sql.DB) map[string]indexedFileRow {
	t.Helper()
	rows, err := db.Query(`SELECT relative_path, id, kind, title, sha256, bytes, updated_at FROM project_files WHERE project_id = 'project-1'`)
	if err != nil {
		t.Fatalf("query project_files: %v", err)
	}
	defer rows.Close()

	result := map[string]indexedFileRow{}
	for rows.Next() {
		var path string
		var row indexedFileRow
		if err := rows.Scan(&path, &row.ID, &row.Kind, &row.Title, &row.SHA256, &row.Bytes, &row.UpdatedAt); err != nil {
			t.Fatalf("scan project file: %v", err)
		}
		result[path] = row
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("project file rows: %v", err)
	}
	return result
}
