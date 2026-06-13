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
	service := NewService(store.New(db))

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
