package project

import "context"

type VersionPage struct {
	Versions []ProjectVersion `json:"versions"`
	Next     string           `json:"next,omitempty"`
}

// History follows immutable parent links, so new publications cannot reorder
// pages. Summaries omit entry arrays; each version can be fetched independently.
func (s *VersionStore) History(ctx context.Context, authorID, projectID, cursor string) (VersionPage, error) {
	page := VersionPage{Versions: []ProjectVersion{}}
	for i := 0; i < 30; i++ {
		var v ProjectVersion
		err := s.db.QueryRowContext(ctx, `SELECT v.id,v.project_id,COALESCE(v.parent_id,''),v.operation_id,v.message,v.created_at
 FROM project_versions v JOIN story_projects p ON p.id=v.project_id
 WHERE p.author_id=? AND v.project_id=? AND v.id=COALESCE(NULLIF(?,''),(SELECT version_id FROM project_version_heads WHERE project_id=?))`, authorID, projectID, cursor, projectID).Scan(&v.ID, &v.ProjectID, &v.ParentID, &v.OperationID, &v.Message, &v.CreatedAt)
		if err != nil {
			return page, err
		}
		page.Versions = append(page.Versions, v)
		cursor = v.ParentID
		if cursor == "" {
			return page, nil
		}
	}
	page.Next = cursor
	return page, nil
}
