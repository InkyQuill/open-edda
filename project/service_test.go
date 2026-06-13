package project

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"git.inkyquill.net/inky/writer/store"
	"github.com/pressly/goose/v3"
)

func TestUpdateContentRejectsStaleRevision(t *testing.T) {
	db := openMigratedTestDB(t)
	service := NewService(db)

	created, err := service.CreateContent(context.Background(), CreateContentInput{
		ProjectID:    "project-1",
		Kind:         KindChapter,
		Title:        "Opening",
		BodyMarkdown: "first draft",
		CreatedBy:    "author",
	})
	if err != nil {
		t.Fatalf("CreateContent() error = %v", err)
	}

	_, err = service.UpdateContent(context.Background(), UpdateContentInput{
		ProjectID:        "project-1",
		ContentID:        created.ID,
		ExpectedRevision: 1,
		BodyMarkdown:     "second draft",
		CreatedBy:        "author",
	})
	if err != nil {
		t.Fatalf("UpdateContent() error = %v", err)
	}

	_, err = service.UpdateContent(context.Background(), UpdateContentInput{
		ProjectID:        "project-1",
		ContentID:        created.ID,
		ExpectedRevision: 1,
		BodyMarkdown:     "stale draft",
		CreatedBy:        "author",
	})
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("UpdateContent() error = %v, want ErrConflict", err)
	}
}

func TestCreateContentRollsBackWhenRevisionInsertFails(t *testing.T) {
	db := openMigratedTestDB(t)
	service := NewService(db)

	_, err := db.Exec(`
		CREATE TRIGGER block_revision_insert
		BEFORE INSERT ON revisions
		BEGIN
			SELECT RAISE(ABORT, 'revision insert blocked');
		END;
	`)
	if err != nil {
		t.Fatalf("create revision-blocking trigger: %v", err)
	}

	_, err = service.CreateContent(context.Background(), CreateContentInput{
		ProjectID:    "project-1",
		Kind:         KindChapter,
		Title:        "Rollback Check",
		BodyMarkdown: "draft that must roll back",
		CreatedBy:    "author",
	})
	if err == nil {
		t.Fatal("CreateContent() error = nil, want revision insert failure")
	}

	var count int
	err = db.QueryRow(`
		SELECT COUNT(*)
		FROM content_items
		WHERE project_id = 'project-1' AND title = 'Rollback Check'
	`).Scan(&count)
	if err != nil {
		t.Fatalf("count content items: %v", err)
	}
	if count != 0 {
		t.Fatalf("content rows after failed CreateContent() = %d, want 0", count)
	}
}

func TestUpdateContentRollsBackWhenRevisionInsertFails(t *testing.T) {
	db := openMigratedTestDB(t)
	service := NewService(db)

	created, err := service.CreateContent(context.Background(), CreateContentInput{
		ProjectID:    "project-1",
		Kind:         KindChapter,
		Title:        "Opening",
		BodyMarkdown: "first draft",
		CreatedBy:    "author",
	})
	if err != nil {
		t.Fatalf("CreateContent() error = %v", err)
	}

	_, err = db.Exec(`
		CREATE TRIGGER block_revision_insert
		BEFORE INSERT ON revisions
		BEGIN
			SELECT RAISE(ABORT, 'revision insert blocked');
		END;
	`)
	if err != nil {
		t.Fatalf("create revision-blocking trigger: %v", err)
	}

	_, err = service.UpdateContent(context.Background(), UpdateContentInput{
		ProjectID:        "project-1",
		ContentID:        created.ID,
		ExpectedRevision: 1,
		BodyMarkdown:     "second draft that must roll back",
		CreatedBy:        "author",
	})
	if err == nil {
		t.Fatal("UpdateContent() error = nil, want revision insert failure")
	}

	var body string
	var revision int64
	err = db.QueryRow(`
		SELECT body_markdown, current_revision
		FROM content_items
		WHERE id = ? AND project_id = ?
	`, created.ID, "project-1").Scan(&body, &revision)
	if err != nil {
		t.Fatalf("load content item: %v", err)
	}
	if body != "first draft" {
		t.Fatalf("body after failed UpdateContent() = %q, want first draft", body)
	}
	if revision != 1 {
		t.Fatalf("revision after failed UpdateContent() = %d, want 1", revision)
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
		VALUES ('project-1', 'author-1', 'Test', 'test', 'en', '2026-06-13T00:00:00Z', '2026-06-13T00:00:00Z');
	`)
	if err != nil {
		t.Fatalf("seed project: %v", err)
	}

	return db
}
