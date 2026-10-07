package project

import (
	"context"
	"database/sql"
)

// DeleteProject removes the project and dependent database records atomically.
// Immutable object bytes remain for a future garbage collector; no filesystem
// removal may race an in-flight reader or remove another project's objects.
func (s *Service) DeleteProject(ctx context.Context, authorID, projectID, confirmationTitle string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	// Version heads, parents and entries reference one another while cascading.
	if _, err = tx.ExecContext(ctx, "PRAGMA defer_foreign_keys = ON"); err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, "DELETE FROM story_projects WHERE id=? AND author_id=? AND title=?", projectID, authorID, confirmationTitle)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		var id string
		if err := tx.QueryRowContext(ctx, "SELECT id FROM story_projects WHERE id=? AND author_id=?", projectID, authorID).Scan(&id); err != nil {
			return err
		}
		return ErrConflict
	}
	if count != 1 {
		return sql.ErrNoRows
	}
	return tx.Commit()
}
