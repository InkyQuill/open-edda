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

	unlock := lockProjectSave(root)
	defer unlock()
	layout, idMap, files, err := scanStableFiles(root)
	if err != nil {
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

	rows, err := tx.QueryContext(ctx, `SELECT id FROM project_files WHERE project_id = ?`, projectID)
	if err != nil {
		return IndexResult{}, fmt.Errorf("load existing file index rows: %w", err)
	}
	currentIDs := make(map[string]bool, len(files))
	for _, file := range files {
		currentIDs[file.ID] = true
	}
	var staleIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			_ = rows.Close()
			return IndexResult{}, fmt.Errorf("scan existing file index row: %w", err)
		}
		if !currentIDs[id] {
			staleIDs = append(staleIDs, id)
		}
	}
	if err := rows.Close(); err != nil {
		return IndexResult{}, fmt.Errorf("close existing file index rows: %w", err)
	}
	for _, id := range staleIDs {
		if _, err := tx.ExecContext(ctx, `DELETE FROM project_files WHERE project_id = ? AND id = ?`, projectID, id); err != nil {
			return IndexResult{}, fmt.Errorf("delete stale file index row %s: %w", id, err)
		}
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO project_files (
			id, project_id, relative_path, kind, title, sha256, bytes, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(project_id, id) DO UPDATE SET
			relative_path = excluded.relative_path,
			kind = excluded.kind,
			title = excluded.title,
			sha256 = excluded.sha256,
			bytes = excluded.bytes,
			updated_at = CASE
				WHEN project_files.sha256 != excluded.sha256 THEN excluded.updated_at
				ELSE project_files.updated_at
			END
	`)
	if err != nil {
		return IndexResult{}, fmt.Errorf("prepare project file index upsert: %w", err)
	}
	defer stmt.Close()
	for _, file := range files {
		if _, err := stmt.ExecContext(ctx, file.ID, projectID, file.Path, string(file.Kind), file.Title, file.SHA256, file.Size, now); err != nil {
			return IndexResult{}, fmt.Errorf("upsert project file index row %s: %w", file.Path, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return IndexResult{}, fmt.Errorf("commit file index transaction: %w", err)
	}
	committed = true
	if err := WriteIDMap(root, idMap); err != nil {
		return IndexResult{}, err
	}

	return IndexResult{
		Layout: layout,
		Files:  files,
		Counts: CountByKind(layout.Files),
	}, nil
}
