package project

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Version returns one immutable manifest; an empty versionID pins the current
// head in the first SELECT. Entries are then read by that pinned ID.
func (s *VersionStore) Version(ctx context.Context, authorID, projectID, versionID string) (ProjectVersion, error) {
	var v ProjectVersion
	err := s.db.QueryRowContext(ctx, `SELECT v.id,v.project_id,COALESCE(v.parent_id,''),v.operation_id,v.message,v.created_at
 FROM project_versions v JOIN story_projects p ON p.id=v.project_id
 WHERE p.author_id=? AND v.project_id=? AND v.id=COALESCE(NULLIF(?,''),(SELECT version_id FROM project_version_heads WHERE project_id=?))`, authorID, projectID, versionID, projectID).Scan(&v.ID, &v.ProjectID, &v.ParentID, &v.OperationID, &v.Message, &v.CreatedAt)
	if err != nil {
		return v, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id,path,kind,COALESCE(sha256,''),bytes FROM project_version_entries WHERE project_id=? AND version_id=? ORDER BY path`, projectID, v.ID)
	if err != nil {
		return v, err
	}
	defer rows.Close()
	v.Entries = []TreeEntry{}
	for rows.Next() {
		var entry TreeEntry
		if err := rows.Scan(&entry.ID, &entry.Path, &entry.Kind, &entry.SHA256, &entry.Bytes); err != nil {
			return ProjectVersion{}, err
		}
		v.Entries = append(v.Entries, entry)
	}
	return v, rows.Err()
}

// Publish atomically records the complete tree, advances its head, and records
// the idempotency receipt in project_versions. All bytes must already have been
// uploaded for this project. Object verification runs outside the write lock.
func (s *VersionStore) Publish(ctx context.Context, input PublishVersionInput) (ProjectVersion, error) {
	return s.publish(ctx, input, "")
}

func (s *VersionStore) publish(ctx context.Context, input PublishVersionInput, restoredFrom string) (ProjectVersion, error) {
	if err := s.authorize(ctx, input.AuthorID, input.ProjectID); err != nil {
		return ProjectVersion{}, err
	}
	entries, fingerprint, err := canonicalTree(input, s.limits, restoredFrom)
	if err != nil {
		return ProjectVersion{}, err
	}
	// A lost-response retry remains valid even if the head has moved since it.
	if v, found, err := s.receipt(ctx, input, fingerprint); found || err != nil {
		return v, err
	}
	verified := map[string]bool{}
	for _, entry := range entries {
		if entry.Kind != "file" {
			continue
		}
		var size int64
		if err := s.db.QueryRowContext(ctx, "SELECT bytes FROM project_objects WHERE project_id=? AND sha256=?", input.ProjectID, entry.SHA256).Scan(&size); err != nil {
			return ProjectVersion{}, err
		}
		if size != entry.Bytes {
			return ProjectVersion{}, ErrObjectIntegrity
		}
		if verified[entry.SHA256] {
			continue
		}
		file, err := s.openVerifiedObject(ctx, entry.SHA256, entry.Bytes)
		if err != nil {
			return ProjectVersion{}, err
		}
		if err := file.Close(); err != nil {
			return ProjectVersion{}, err
		}
		verified[entry.SHA256] = true
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ProjectVersion{}, err
	}
	defer tx.Rollback()
	// Acquire the SQLite write lock before reading the head. A deferred read then
	// upgrade can fail with SQLITE_BUSY even with a busy timeout under contention.
	result, err := tx.ExecContext(ctx, `INSERT INTO project_version_heads(project_id)
 SELECT id FROM story_projects WHERE id=? AND author_id=?
 ON CONFLICT(project_id) DO UPDATE SET project_id=excluded.project_id`, input.ProjectID, input.AuthorID)
	if err != nil {
		return ProjectVersion{}, err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return ProjectVersion{}, err
	}
	if count != 1 {
		return ProjectVersion{}, sql.ErrNoRows
	}
	var synchronous int
	if err := tx.QueryRowContext(ctx, "PRAGMA synchronous").Scan(&synchronous); err != nil {
		return ProjectVersion{}, err
	}
	if synchronous != 2 && synchronous != 3 {
		return ProjectVersion{}, errors.New("project publication requires SQLite synchronous FULL or EXTRA")
	}
	var existingID, existingHash string
	err = tx.QueryRowContext(ctx, "SELECT id,request_sha256 FROM project_versions WHERE project_id=? AND operation_id=?", input.ProjectID, input.OperationID).Scan(&existingID, &existingHash)
	if err == nil {
		if existingHash != fingerprint {
			return ProjectVersion{}, ErrOperationConflict
		}
		if err := tx.Rollback(); err != nil {
			return ProjectVersion{}, err
		}
		return s.Version(ctx, input.AuthorID, input.ProjectID, existingID)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return ProjectVersion{}, err
	}
	var head string
	if err := tx.QueryRowContext(ctx, "SELECT COALESCE(version_id,'') FROM project_version_heads WHERE project_id=?", input.ProjectID).Scan(&head); err != nil {
		return ProjectVersion{}, err
	}
	if head != input.ExpectedVersion {
		return ProjectVersion{}, ErrVersionConflict
	}
	id, err := versionID()
	if err != nil {
		return ProjectVersion{}, err
	}
	v := ProjectVersion{ID: id, ProjectID: input.ProjectID, ParentID: head, OperationID: input.OperationID, Message: input.Message, CreatedAt: time.Now().UTC().Format(time.RFC3339Nano), Entries: entries}
	_, err = tx.ExecContext(ctx, `INSERT INTO project_versions(project_id,id,parent_id,operation_id,request_sha256,message,created_at) VALUES (?,?,NULLIF(?,''),?,?,?,?)`, v.ProjectID, v.ID, v.ParentID, v.OperationID, fingerprint, v.Message, v.CreatedAt)
	if err != nil {
		return ProjectVersion{}, err
	}
	for _, entry := range entries {
		if _, err := tx.ExecContext(ctx, "INSERT INTO project_tree_ids(project_id,id,kind) VALUES (?,?,?) ON CONFLICT(project_id,id) DO NOTHING", input.ProjectID, entry.ID, entry.Kind); err != nil {
			return ProjectVersion{}, err
		}
		var kind string
		if err := tx.QueryRowContext(ctx, "SELECT kind FROM project_tree_ids WHERE project_id=? AND id=?", input.ProjectID, entry.ID).Scan(&kind); err != nil {
			return ProjectVersion{}, err
		}
		if kind != entry.Kind {
			return ProjectVersion{}, fmt.Errorf("%w: entry identity changed kind", ErrInvalidTree)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO project_version_entries(project_id,version_id,id,path,kind,sha256,bytes) VALUES (?,?,?,?,?,NULLIF(?,''),?)`, input.ProjectID, id, entry.ID, entry.Path, entry.Kind, entry.SHA256, entry.Bytes); err != nil {
			return ProjectVersion{}, err
		}
	}
	if _, err := tx.ExecContext(ctx, "UPDATE project_version_heads SET version_id=? WHERE project_id=?", id, input.ProjectID); err != nil {
		return ProjectVersion{}, err
	}
	if err := tx.Commit(); err != nil {
		return ProjectVersion{}, err
	}
	return v, nil
}

func (s *VersionStore) receipt(ctx context.Context, input PublishVersionInput, fingerprint string) (ProjectVersion, bool, error) {
	var id, hash string
	err := s.db.QueryRowContext(ctx, "SELECT id,request_sha256 FROM project_versions WHERE project_id=? AND operation_id=?", input.ProjectID, input.OperationID).Scan(&id, &hash)
	if errors.Is(err, sql.ErrNoRows) {
		return ProjectVersion{}, false, nil
	}
	if err != nil {
		return ProjectVersion{}, false, err
	}
	if hash != fingerprint {
		return ProjectVersion{}, true, ErrOperationConflict
	}
	v, err := s.Version(ctx, input.AuthorID, input.ProjectID, id)
	return v, true, err
}

// Restore publishes the retained target tree as a new version against the
// caller's current base. It never moves head backwards or deletes history.
func (s *VersionStore) Restore(ctx context.Context, input PublishVersionInput, targetVersion string) (ProjectVersion, error) {
	if targetVersion == "" {
		return ProjectVersion{}, fmt.Errorf("%w: restore target required", ErrInvalidTree)
	}
	target, err := s.Version(ctx, input.AuthorID, input.ProjectID, targetVersion)
	if err != nil {
		return ProjectVersion{}, err
	}
	input.Entries = target.Entries
	return s.publish(ctx, input, targetVersion)
}
