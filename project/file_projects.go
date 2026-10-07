package project

import (
	"context"
	"strings"

	"github.com/InkyQuill/open-edda/store"
	"github.com/mattn/go-sqlite3"
)

// Creating a file project commits its empty tree together with the project row.
// Prototype content projects use a separate service; no legacy migration is required.
func (s *Service) createFileProject(ctx context.Context, input CreateProjectInput) (StoryProject, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" || len(title) > 512 || len(input.Language) > 64 {
		return StoryProject{}, ErrInvalidTree
	}
	p := StoryProject{ID: newID("project"), Title: title, Slug: slugify(title), Language: input.Language, StorageMode: "files"}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return StoryProject{}, err
	}
	defer tx.Rollback()
	now := nowString()
	err = store.New(tx).CreateStoryProject(ctx, store.CreateStoryProjectParams{ID: p.ID, AuthorID: input.AuthorID, Title: p.Title, Slug: p.Slug, Language: p.Language, CreatedAt: now, UpdatedAt: now})
	if isSQLiteConstraint(err, sqlite3.ErrConstraintUnique) {
		return StoryProject{}, ErrConflict
	}
	if err != nil {
		return StoryProject{}, err
	}
	if _, err = tx.ExecContext(ctx, "UPDATE story_projects SET storage_mode='files' WHERE id=?", p.ID); err != nil {
		return StoryProject{}, err
	}
	version := newID("version")
	_, fingerprint, err := canonicalTree(PublishVersionInput{OperationID: "create", Entries: []TreeEntry{}}, VersionLimits{}, "")
	if err != nil {
		return StoryProject{}, err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO project_versions(project_id,id,parent_id,operation_id,request_sha256,message,created_at) VALUES(?,?,NULL,'create',?,'',?)`, p.ID, version, fingerprint, now); err != nil {
		return StoryProject{}, err
	}
	if _, err = tx.ExecContext(ctx, "INSERT INTO project_version_heads(project_id,version_id) VALUES(?,?)", p.ID, version); err != nil {
		return StoryProject{}, err
	}
	if err = tx.Commit(); err != nil {
		return StoryProject{}, err
	}
	return p, nil
}
