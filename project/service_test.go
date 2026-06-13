package project

import (
	"context"
	"database/sql"
	"encoding/json"
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

func TestProjectMapReturnsContentSectionsAndRelations(t *testing.T) {
	db := openMigratedTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	entry, err := service.CreateContent(ctx, CreateContentInput{
		ProjectID:    "project-1",
		Kind:         KindStoryBibleEntry,
		Title:        "Mira",
		BodyMarkdown: "An alchemist with a precise memory.",
		CreatedBy:    "author",
	})
	if err != nil {
		t.Fatalf("CreateContent(entry) error = %v", err)
	}
	if err := service.CreateEntrySection(ctx, CreateEntrySectionInput{
		ProjectID:     "project-1",
		ContentItemID: entry.ID,
		Heading:       "Motivation",
		BodyMarkdown:  "Find the ash compass.",
		SortOrder:     1,
	}); err != nil {
		t.Fatalf("CreateEntrySection() error = %v", err)
	}
	if err := service.CreateEntryRelation(ctx, CreateEntryRelationInput{
		ProjectID:    "project-1",
		SourceItemID: entry.ID,
		TargetTitle:  "Ash Compass",
		RelationType: "seeks",
	}); err != nil {
		t.Fatalf("CreateEntryRelation() error = %v", err)
	}

	projectMap, err := service.ProjectMap(ctx, "project-1")
	if err != nil {
		t.Fatalf("ProjectMap() error = %v", err)
	}

	if projectMap.Project.ID != "project-1" {
		t.Fatalf("project ID = %q, want project-1", projectMap.Project.ID)
	}
	if len(projectMap.Content) != 1 {
		t.Fatalf("content count = %d, want 1", len(projectMap.Content))
	}
	if len(projectMap.Content[0].Sections) != 1 || projectMap.Content[0].Sections[0].Heading != "Motivation" {
		t.Fatalf("sections = %#v, want Motivation section", projectMap.Content[0].Sections)
	}
	if len(projectMap.Content[0].Relations) != 1 || projectMap.Content[0].Relations[0].TargetTitle != "Ash Compass" {
		t.Fatalf("relations = %#v, want Ash Compass relation", projectMap.Content[0].Relations)
	}
}

func TestSearchContentFiltersMetadataAndTags(t *testing.T) {
	db := openMigratedTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	items := []CreateContentInput{
		{
			ProjectID:    "project-1",
			Kind:         KindStoryBibleEntry,
			Title:        "Mira",
			BodyMarkdown: "Alchemy notes mention the lantern.",
			MetadataJSON: `{"type":"character","status":"draft","tags":["alchemy"]}`,
			CreatedBy:    "author",
		},
		{
			ProjectID:    "project-1",
			Kind:         KindStoryBibleEntry,
			Title:        "Ash Compass",
			BodyMarkdown: "Alchemy notes mention the lantern.",
			MetadataJSON: `{"type":"artifact","status":"draft","tags":["alchemy"]}`,
			CreatedBy:    "author",
		},
		{
			ProjectID:    "project-1",
			Kind:         KindStoryBibleEntry,
			Title:        "Sol",
			BodyMarkdown: "Alchemy notes mention the lantern.",
			MetadataJSON: `{"type":"character","status":"final","tags":["alchemy"]}`,
			CreatedBy:    "author",
		},
	}
	for _, input := range items {
		if _, err := service.CreateContent(ctx, input); err != nil {
			t.Fatalf("CreateContent(%q) error = %v", input.Title, err)
		}
	}

	matches, err := service.SearchContent(ctx, SearchContentInput{
		ProjectID:       "project-1",
		Query:           "alchemy",
		Kind:            KindStoryBibleEntry,
		MetadataFilters: map[string]string{"type": "character", "status": "draft"},
		Tags:            []string{"alchemy"},
		Limit:           10,
	})
	if err != nil {
		t.Fatalf("SearchContent() error = %v", err)
	}

	if len(matches) != 1 {
		t.Fatalf("SearchContent() returned %d matches, want 1: %#v", len(matches), matches)
	}
	if matches[0].Title != "Mira" {
		t.Fatalf("match title = %q, want Mira", matches[0].Title)
	}
}

func TestCreateAttachedNoteStoresScopedNote(t *testing.T) {
	db := openMigratedTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	chapter, err := service.CreateContent(ctx, CreateContentInput{
		ProjectID:    "project-1",
		Kind:         KindChapter,
		Title:        "Opening",
		BodyMarkdown: "The lantern burned blue.",
		CreatedBy:    "author",
	})
	if err != nil {
		t.Fatalf("CreateContent() error = %v", err)
	}

	note, err := service.CreateAttachedNote(ctx, CreateAttachedNoteInput{
		ProjectID:      "project-1",
		ContentItemID:  chapter.ID,
		SelectionStart: 4,
		SelectionEnd:   11,
		Title:          "Check image",
		BodyMarkdown:   "Clarify lantern color.",
		Source:         "read_and_check",
	})
	if err != nil {
		t.Fatalf("CreateAttachedNote() error = %v", err)
	}

	if note.ID == "" {
		t.Fatal("note ID is empty")
	}
	if note.ContentItemID != chapter.ID {
		t.Fatalf("note content ID = %q, want %q", note.ContentItemID, chapter.ID)
	}

	var stored string
	if err := db.QueryRow(`SELECT body_markdown FROM attached_notes WHERE id = ?`, note.ID).Scan(&stored); err != nil {
		t.Fatalf("load attached note: %v", err)
	}
	if stored != "Clarify lantern color." {
		t.Fatalf("stored note body = %q", stored)
	}
}

func TestSearchContentRejectsInvalidMetadataJSON(t *testing.T) {
	db := openMigratedTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	item, err := service.CreateContent(ctx, CreateContentInput{
		ProjectID:    "project-1",
		Kind:         KindProjectNote,
		Title:        "Loose Note",
		BodyMarkdown: "lantern",
		CreatedBy:    "author",
	})
	if err != nil {
		t.Fatalf("CreateContent() error = %v", err)
	}
	_, err = db.Exec(`UPDATE content_items SET metadata_json = ? WHERE id = ?`, `{bad json`, item.ID)
	if err != nil {
		t.Fatalf("corrupt metadata: %v", err)
	}

	_, err = service.SearchContent(ctx, SearchContentInput{
		ProjectID:       "project-1",
		Query:           "lantern",
		MetadataFilters: map[string]string{"type": "note"},
	})
	if err == nil {
		t.Fatal("SearchContent() error = nil, want invalid metadata error")
	}

	var syntaxErr *json.SyntaxError
	if !errors.As(err, &syntaxErr) {
		t.Fatalf("SearchContent() error = %T %v, want json.SyntaxError", err, err)
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
