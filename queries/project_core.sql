-- name: CreateStoryProject :exec
INSERT INTO story_projects (id, author_id, title, slug, language, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: ListStoryProjects :many
SELECT * FROM story_projects
WHERE author_id = ?
ORDER BY updated_at DESC, title ASC;

-- name: GetStoryProject :one
SELECT * FROM story_projects
WHERE id = ? AND author_id = ?;

-- name: GetStoryProjectByID :one
SELECT * FROM story_projects
WHERE id = ?;

-- name: CreateContentItem :exec
INSERT INTO content_items (
  id, project_id, kind, title, slug, body_markdown, metadata_json,
  sort_order, current_revision, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: ListContentItems :many
SELECT * FROM content_items
WHERE project_id = ? AND kind = ?
ORDER BY sort_order ASC, title ASC;

-- name: ListProjectContentItems :many
SELECT * FROM content_items
WHERE project_id = ?
ORDER BY kind ASC, sort_order ASC, title ASC;

-- name: GetContentItem :one
SELECT * FROM content_items
WHERE id = ? AND project_id = ?;

-- name: UpdateContentItemBody :execrows
UPDATE content_items
SET body_markdown = sqlc.arg(body_markdown),
    metadata_json = sqlc.arg(metadata_json),
    current_revision = sqlc.arg(next_revision),
    updated_at = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id)
  AND project_id = sqlc.arg(project_id)
  AND current_revision = sqlc.arg(expected_revision);

-- name: CreateRevision :exec
INSERT INTO revisions (
  id, content_item_id, revision_number, body_markdown, metadata_json,
  reason, created_by, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: ListRevisions :many
SELECT revisions.*
FROM revisions
JOIN content_items ON content_items.id = revisions.content_item_id
WHERE revisions.content_item_id = sqlc.arg(content_item_id)
  AND content_items.project_id = sqlc.arg(project_id)
ORDER BY revision_number DESC;

-- name: CreateEntrySection :exec
INSERT INTO entry_sections (id, content_item_id, heading, body_markdown, sort_order)
VALUES (?, ?, ?, ?, ?);

-- name: ListEntrySections :many
SELECT * FROM entry_sections
WHERE content_item_id = ?
ORDER BY sort_order ASC;

-- name: CreateEntryRelation :exec
INSERT INTO entry_relations (
  id, project_id, source_item_id, target_item_id, target_title, relation_type, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: ListEntryRelations :many
SELECT * FROM entry_relations
WHERE project_id = ? AND source_item_id = ?
ORDER BY target_title ASC;

-- name: SearchContent :many
SELECT content_items.*
FROM content_search(CAST(sqlc.arg(query) AS TEXT))
JOIN content_items ON content_items.rowid = content_search.rowid
WHERE content_items.project_id = sqlc.arg(project_id)
ORDER BY rank
LIMIT sqlc.arg(limit);
