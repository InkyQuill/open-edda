package fileproject

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type IndexResult struct {
	Layout ProjectLayout
	Files  []StableFile
	Counts map[LayoutKind]int
}

func RebuildIndex(ctx context.Context, db *sql.DB, projectID string, root string) (IndexResult, error) {
	if projectID == "" {
		return IndexResult{}, fmt.Errorf("project id is required")
	}

	layout, err := Scan(root)
	if err != nil {
		return IndexResult{}, err
	}
	idMap, files, err := AssignStableIDs(root, layout)
	if err != nil {
		return IndexResult{}, err
	}
	if err := WriteIDMap(root, idMap); err != nil {
		return IndexResult{}, err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return IndexResult{}, fmt.Errorf("begin file index transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	if _, err := tx.ExecContext(ctx, `DELETE FROM project_files WHERE project_id = ?`, projectID); err != nil {
		return IndexResult{}, fmt.Errorf("clear project file index: %w", err)
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	for _, file := range files {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO project_files (
				id, project_id, relative_path, kind, title, sha256, bytes, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`, file.ID, projectID, file.Path, string(file.Kind), file.Title, file.SHA256, file.Size, now); err != nil {
			return IndexResult{}, fmt.Errorf("insert project file index row %s: %w", file.Path, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return IndexResult{}, fmt.Errorf("commit file index transaction: %w", err)
	}
	committed = true

	return IndexResult{
		Layout: layout,
		Files:  files,
		Counts: CountByKind(layout.Files),
	}, nil
}
