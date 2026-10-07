package project

import (
	"context"
	"database/sql"
	"errors"
	"testing"
)

func TestDeleteProjectConfirmationOwnershipAndCascade(t *testing.T) {
	versions, db, _ := newTestVersions(t)
	ctx := context.Background()
	service := NewService(db)
	entry := uploadTestEntry(t, versions, "draft", "draft.md", []byte("text"))
	first := publishTestVersion(t, versions, "", "first", []TreeEntry{entry})
	publishTestVersion(t, versions, first.ID, "second", []TreeEntry{})
	createStructuredWriteContent(t, ctx, service, ContentKind("chapter"), "Chapter", "text", "{}")
	if err := service.DeleteProject(ctx, "another-author", "project-1", "Test"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("ownership: %v", err)
	}
	if err := service.DeleteProject(ctx, "author-1", "project-1", "wrong"); !errors.Is(err, ErrConflict) {
		t.Fatalf("confirmation: %v", err)
	}
	if _, err := versions.Version(ctx, "author-1", "project-1", ""); err != nil {
		t.Fatal("failed confirmation deleted project", err)
	}
	if err := service.DeleteProject(ctx, "author-1", "project-1", "Test"); err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{"story_projects", "project_versions", "project_version_entries", "project_objects", "project_tree_ids", "project_version_heads", "content_items", "agent_sessions"} {
		var count int
		if err := db.QueryRow("SELECT count(*) FROM " + table).Scan(&count); err != nil || count != 0 {
			t.Fatalf("%s: %d %v", table, count, err)
		}
	}
	if err := service.DeleteProject(ctx, "author-1", "project-1", "Test"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("repeated: %v", err)
	}
}
