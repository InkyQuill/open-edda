package project

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
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

func TestStructuredAppendToContentCreatesAgentRevision(t *testing.T) {
	db := openMigratedTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	chapter := createStructuredWriteContent(t, ctx, service, KindChapter, "Opening", "The lantern burned blue.", `{"status":"draft"}`)

	updated, err := service.AppendToContent(ctx, StructuredWriteInput{
		ProjectID:         "project-1",
		ContentID:         chapter.ID,
		ExpectedRevision:  1,
		GeneratedMarkdown: "\nMira listened.",
		Reason:            "continue opening",
		AgentSessionID:    "session-1",
		ActionKind:        "continuation",
		ModelVariantID:    "model-1",
	})
	if err != nil {
		t.Fatalf("AppendToContent() error = %v", err)
	}
	if updated.BodyMarkdown != "The lantern burned blue.\nMira listened." {
		t.Fatalf("body = %q", updated.BodyMarkdown)
	}
	if updated.CurrentRevision != 2 {
		t.Fatalf("revision = %d, want 2", updated.CurrentRevision)
	}
	assertLatestRevision(t, ctx, db, chapter.ID, revisionAssertion{
		RevisionNumber:  2,
		BodyMarkdown:    "The lantern burned blue.\nMira listened.",
		Reason:          "continue opening",
		CreatedBy:       "agent",
		AgentSessionID:  "session-1",
		ActionKind:      "continuation",
		ModelVariantID:  "model-1",
		SkillID:         "",
		WantAgentFields: true,
	})
}

func TestStructuredInsertIntoContentCreatesRevision(t *testing.T) {
	db := openMigratedTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	chapter := createStructuredWriteContent(t, ctx, service, KindChapter, "Opening", "The lantern burned blue.", `{}`)

	updated, err := service.InsertIntoContent(ctx, StructuredWriteInput{
		ProjectID:         "project-1",
		ContentID:         chapter.ID,
		ExpectedRevision:  1,
		GeneratedMarkdown: " quietly",
		InsertPosition:    int64(len("The lantern")),
		Reason:            "insert tone",
		AgentSessionID:    "session-1",
		ActionKind:        "rewrite",
		ModelVariantID:    "model-1",
	})
	if err != nil {
		t.Fatalf("InsertIntoContent() error = %v", err)
	}
	if updated.BodyMarkdown != "The lantern quietly burned blue." {
		t.Fatalf("body = %q", updated.BodyMarkdown)
	}
	if updated.CurrentRevision != 2 {
		t.Fatalf("revision = %d, want 2", updated.CurrentRevision)
	}
	assertRevisionCount(t, ctx, db, chapter.ID, 2)
}

func TestStructuredReplaceContentRangeCreatesRevision(t *testing.T) {
	db := openMigratedTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	chapter := createStructuredWriteContent(t, ctx, service, KindChapter, "Opening", "The lantern burned blue.", `{}`)

	updated, err := service.ReplaceContentRange(ctx, StructuredWriteInput{
		ProjectID:         "project-1",
		ContentID:         chapter.ID,
		ExpectedRevision:  1,
		GeneratedMarkdown: "glowed green",
		SelectionStart:    int64(len("The lantern ")),
		SelectionEnd:      int64(len("The lantern burned blue")),
		Reason:            "replace image",
		AgentSessionID:    "session-1",
		ActionKind:        "rewrite",
		ModelVariantID:    "model-1",
	})
	if err != nil {
		t.Fatalf("ReplaceContentRange() error = %v", err)
	}
	if updated.BodyMarkdown != "The lantern glowed green." {
		t.Fatalf("body = %q", updated.BodyMarkdown)
	}
	if updated.CurrentRevision != 2 {
		t.Fatalf("revision = %d, want 2", updated.CurrentRevision)
	}
	assertRevisionCount(t, ctx, db, chapter.ID, 2)
}

func TestStructuredInsertRejectsNonUTF8BoundaryByteOffsetWithoutPersisting(t *testing.T) {
	db := openMigratedTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	chapter := createStructuredWriteContent(t, ctx, service, KindChapter, "Opening", "éclair", `{}`)

	_, err := service.InsertIntoContent(ctx, StructuredWriteInput{
		ProjectID:         "project-1",
		ContentID:         chapter.ID,
		ExpectedRevision:  1,
		GeneratedMarkdown: " bright",
		InsertPosition:    1,
		Reason:            "insert at invalid byte offset",
		AgentSessionID:    "session-1",
		ActionKind:        "rewrite",
		ModelVariantID:    "model-1",
	})
	if err == nil {
		t.Fatal("InsertIntoContent() error = nil, want UTF-8 boundary error")
	}
	if !strings.Contains(err.Error(), "UTF-8 rune boundary") {
		t.Fatalf("InsertIntoContent() error = %v, want UTF-8 boundary error", err)
	}

	assertContentUnchanged(t, ctx, service, chapter.ID, "éclair", 1)
	assertRevisionCount(t, ctx, db, chapter.ID, 1)
}

func TestStructuredReplaceRejectsNonUTF8BoundaryByteOffsetWithoutPersisting(t *testing.T) {
	db := openMigratedTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	chapter := createStructuredWriteContent(t, ctx, service, KindChapter, "Opening", "café noir", `{}`)

	_, err := service.ReplaceContentRange(ctx, StructuredWriteInput{
		ProjectID:         "project-1",
		ContentID:         chapter.ID,
		ExpectedRevision:  1,
		GeneratedMarkdown: "e",
		SelectionStart:    3,
		SelectionEnd:      4,
		Reason:            "replace at invalid byte offset",
		AgentSessionID:    "session-1",
		ActionKind:        "rewrite",
		ModelVariantID:    "model-1",
	})
	if err == nil {
		t.Fatal("ReplaceContentRange() error = nil, want UTF-8 boundary error")
	}
	if !strings.Contains(err.Error(), "UTF-8 rune boundary") {
		t.Fatalf("ReplaceContentRange() error = %v, want UTF-8 boundary error", err)
	}

	assertContentUnchanged(t, ctx, service, chapter.ID, "café noir", 1)
	assertRevisionCount(t, ctx, db, chapter.ID, 1)
}

func TestStructuredUpdateStoryBibleEntryBodyCreatesRevision(t *testing.T) {
	db := openMigratedTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	entry := createStructuredWriteContent(t, ctx, service, KindStoryBibleEntry, "Mira", "An alchemist.", `{"type":"character"}`)

	updated, err := service.ReplaceContentRange(ctx, StructuredWriteInput{
		ProjectID:         "project-1",
		ContentID:         entry.ID,
		ExpectedRevision:  1,
		GeneratedMarkdown: "An alchemist with a precise memory.",
		SelectionStart:    0,
		SelectionEnd:      int64(len("An alchemist.")),
		Reason:            "update bible entry",
		AgentSessionID:    "session-1",
		ActionKind:        "rewrite",
		ModelVariantID:    "model-1",
	})
	if err != nil {
		t.Fatalf("ReplaceContentRange(entry) error = %v", err)
	}
	if updated.BodyMarkdown != "An alchemist with a precise memory." {
		t.Fatalf("body = %q", updated.BodyMarkdown)
	}
	if updated.CurrentRevision != 2 {
		t.Fatalf("revision = %d, want 2", updated.CurrentRevision)
	}
	assertRevisionCount(t, ctx, db, entry.ID, 2)
}

func TestStructuredWriteStaleExpectedRevisionReturnsConflict(t *testing.T) {
	db := openMigratedTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	chapter := createStructuredWriteContent(t, ctx, service, KindChapter, "Opening", "The lantern burned blue.", `{}`)
	if _, err := service.AppendToContent(ctx, StructuredWriteInput{
		ProjectID:         "project-1",
		ContentID:         chapter.ID,
		ExpectedRevision:  1,
		GeneratedMarkdown: "\nMira listened.",
		Reason:            "first append",
		AgentSessionID:    "session-1",
		ActionKind:        "continuation",
		ModelVariantID:    "model-1",
	}); err != nil {
		t.Fatalf("AppendToContent(first) error = %v", err)
	}

	_, err := service.AppendToContent(ctx, StructuredWriteInput{
		ProjectID:         "project-1",
		ContentID:         chapter.ID,
		ExpectedRevision:  1,
		GeneratedMarkdown: "\nStale text.",
		Reason:            "stale append",
		AgentSessionID:    "session-1",
		ActionKind:        "continuation",
		ModelVariantID:    "model-1",
	})
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("AppendToContent(stale) error = %v, want ErrConflict", err)
	}
}

func TestUpdateEntrySectionBodyUpdatesSectionAndRecordsActivity(t *testing.T) {
	db := openMigratedTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	entry := createStructuredWriteContent(t, ctx, service, KindStoryBibleEntry, "Mira", "An alchemist.", `{"type":"character"}`)
	if err := service.CreateEntrySection(ctx, CreateEntrySectionInput{
		ProjectID:     "project-1",
		ContentItemID: entry.ID,
		Heading:       "Motivation",
		BodyMarkdown:  "Find the ash compass.",
		SortOrder:     1,
	}); err != nil {
		t.Fatalf("CreateEntrySection() error = %v", err)
	}

	section, err := service.UpdateEntrySectionBody(ctx, UpdateEntrySectionInput{
		ProjectID:      "project-1",
		ContentID:      entry.ID,
		Heading:        "Motivation",
		BodyMarkdown:   "Protect the glass city.",
		AgentSessionID: "session-1",
		ActionKind:     "rewrite",
		ModelVariantID: "model-1",
	})
	if err != nil {
		t.Fatalf("UpdateEntrySectionBody() error = %v", err)
	}
	if section.BodyMarkdown != "Protect the glass city." {
		t.Fatalf("section body = %q", section.BodyMarkdown)
	}

	sections, err := service.ListEntrySections(ctx, "project-1", entry.ID)
	if err != nil {
		t.Fatalf("ListEntrySections() error = %v", err)
	}
	if len(sections) != 1 || sections[0].BodyMarkdown != "Protect the glass city." {
		t.Fatalf("sections = %#v", sections)
	}
	assertSectionActivity(t, ctx, db, entry.ID, "Motivation")
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

func TestSearchContentAppliesLimitAfterMetadataFilters(t *testing.T) {
	db := openMigratedTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	for _, input := range []CreateContentInput{
		{
			ProjectID:    "project-1",
			Kind:         KindStoryBibleEntry,
			Title:        "Ash Compass",
			BodyMarkdown: "Alchemy lantern.",
			MetadataJSON: `{"type":"artifact","status":"draft","tags":["alchemy"]}`,
			CreatedBy:    "author",
		},
		{
			ProjectID:    "project-1",
			Kind:         KindStoryBibleEntry,
			Title:        "Mira",
			BodyMarkdown: "Alchemy lantern.",
			MetadataJSON: `{"type":"character","status":"draft","tags":["alchemy"]}`,
			CreatedBy:    "author",
		},
	} {
		if _, err := service.CreateContent(ctx, input); err != nil {
			t.Fatalf("CreateContent(%q) error = %v", input.Title, err)
		}
	}

	matches, err := service.SearchContent(ctx, SearchContentInput{
		ProjectID:       "project-1",
		Query:           "alchemy",
		MetadataFilters: map[string]string{"type": "character"},
		Limit:           1,
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

func TestSearchContentValidatesKindAndClampsLimit(t *testing.T) {
	db := openMigratedTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	for i := range 105 {
		if _, err := service.CreateContent(ctx, CreateContentInput{
			ProjectID:    "project-1",
			Kind:         KindProjectNote,
			Title:        fmt.Sprintf("Note %03d", i),
			BodyMarkdown: "lantern",
			SortOrder:    int64(i),
			CreatedBy:    "author",
		}); err != nil {
			t.Fatalf("CreateContent(%d) error = %v", i, err)
		}
	}

	_, err := service.SearchContent(ctx, SearchContentInput{
		ProjectID: "project-1",
		Kind:      ContentKind("invalid"),
	})
	if err == nil {
		t.Fatal("SearchContent() error = nil, want invalid kind error")
	}

	matches, err := service.SearchContent(ctx, SearchContentInput{
		ProjectID: "project-1",
		Query:     "lantern",
		Limit:     250,
	})
	if err != nil {
		t.Fatalf("SearchContent() error = %v", err)
	}
	if len(matches) != 100 {
		t.Fatalf("SearchContent() returned %d matches, want capped 100", len(matches))
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

func createStructuredWriteContent(t *testing.T, ctx context.Context, service *Service, kind ContentKind, title, body, metadata string) ContentItem {
	t.Helper()

	item, err := service.CreateContent(ctx, CreateContentInput{
		ProjectID:    "project-1",
		Kind:         kind,
		Title:        title,
		BodyMarkdown: body,
		MetadataJSON: metadata,
		CreatedBy:    "author",
	})
	if err != nil {
		t.Fatalf("CreateContent(%q) error = %v", title, err)
	}
	return item
}

type revisionAssertion struct {
	RevisionNumber  int64
	BodyMarkdown    string
	Reason          string
	CreatedBy       string
	AgentSessionID  string
	ActionKind      string
	ModelVariantID  string
	SkillID         string
	WantAgentFields bool
}

func assertLatestRevision(t *testing.T, ctx context.Context, db *sql.DB, contentID string, want revisionAssertion) {
	t.Helper()

	var revisionNumber int64
	var bodyMarkdown, reason, createdBy, actionKind, skillID string
	var agentSessionID, modelVariantID sql.NullString
	err := db.QueryRowContext(ctx, `
		SELECT revision_number, body_markdown, reason, created_by, agent_session_id, action_kind, model_variant_id, skill_id
		FROM revisions
		WHERE content_item_id = ?
		ORDER BY revision_number DESC
		LIMIT 1
	`, contentID).Scan(&revisionNumber, &bodyMarkdown, &reason, &createdBy, &agentSessionID, &actionKind, &modelVariantID, &skillID)
	if err != nil {
		t.Fatalf("load latest revision: %v", err)
	}
	if revisionNumber != want.RevisionNumber {
		t.Fatalf("revision number = %d, want %d", revisionNumber, want.RevisionNumber)
	}
	if bodyMarkdown != want.BodyMarkdown {
		t.Fatalf("revision body = %q, want %q", bodyMarkdown, want.BodyMarkdown)
	}
	if reason != want.Reason {
		t.Fatalf("revision reason = %q, want %q", reason, want.Reason)
	}
	if createdBy != want.CreatedBy {
		t.Fatalf("revision created_by = %q, want %q", createdBy, want.CreatedBy)
	}
	if want.WantAgentFields {
		if !agentSessionID.Valid || agentSessionID.String != want.AgentSessionID {
			t.Fatalf("agent_session_id = %#v, want %q", agentSessionID, want.AgentSessionID)
		}
		if actionKind != want.ActionKind {
			t.Fatalf("action_kind = %q, want %q", actionKind, want.ActionKind)
		}
		if !modelVariantID.Valid || modelVariantID.String != want.ModelVariantID {
			t.Fatalf("model_variant_id = %#v, want %q", modelVariantID, want.ModelVariantID)
		}
		if skillID != want.SkillID {
			t.Fatalf("skill_id = %q, want %q", skillID, want.SkillID)
		}
	}
}

func assertRevisionCount(t *testing.T, ctx context.Context, db *sql.DB, contentID string, want int) {
	t.Helper()

	var count int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM revisions WHERE content_item_id = ?`, contentID).Scan(&count); err != nil {
		t.Fatalf("count revisions: %v", err)
	}
	if count != want {
		t.Fatalf("revision count = %d, want %d", count, want)
	}
}

func assertContentUnchanged(t *testing.T, ctx context.Context, service *Service, contentID, wantBody string, wantRevision int64) {
	t.Helper()

	item, err := service.GetContent(ctx, "project-1", contentID)
	if err != nil {
		t.Fatalf("GetContent() error = %v", err)
	}
	if item.BodyMarkdown != wantBody {
		t.Fatalf("body = %q, want %q", item.BodyMarkdown, wantBody)
	}
	if item.CurrentRevision != wantRevision {
		t.Fatalf("revision = %d, want %d", item.CurrentRevision, wantRevision)
	}
}

func assertSectionActivity(t *testing.T, ctx context.Context, db *sql.DB, contentID, heading string) {
	t.Helper()

	events, err := store.New(db).ListActivityEvents(ctx, store.ListActivityEventsParams{
		ProjectID: "project-1",
		Limit:     10,
	})
	if err != nil {
		t.Fatalf("ListActivityEvents() error = %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("activity event count = %d, want 1", len(events))
	}
	event := events[0]
	if event.EventType != "agent_entry_section_updated" {
		t.Fatalf("event type = %q, want agent_entry_section_updated", event.EventType)
	}
	if !event.SessionID.Valid || event.SessionID.String != "session-1" {
		t.Fatalf("event session = %#v, want session-1", event.SessionID)
	}
	var metadata map[string]any
	if err := json.Unmarshal([]byte(event.MetadataJson), &metadata); err != nil {
		t.Fatalf("unmarshal activity metadata: %v", err)
	}
	if metadata["targetContentId"] != contentID {
		t.Fatalf("targetContentId = %#v, want %q", metadata["targetContentId"], contentID)
	}
	if metadata["heading"] != heading {
		t.Fatalf("heading = %#v, want %q", metadata["heading"], heading)
	}
	if metadata["operationKind"] != "update_entry_section" {
		t.Fatalf("operationKind = %#v, want update_entry_section", metadata["operationKind"])
	}
	if metadata["actionKind"] != "rewrite" {
		t.Fatalf("actionKind = %#v, want rewrite", metadata["actionKind"])
	}
	if metadata["modelVariantId"] != "model-1" {
		t.Fatalf("modelVariantId = %#v, want model-1", metadata["modelVariantId"])
	}
	if metadata["agentSessionId"] != "session-1" {
		t.Fatalf("agentSessionId = %#v, want session-1", metadata["agentSessionId"])
	}
}
