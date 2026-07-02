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
  reason, created_by, created_at, agent_session_id, action_kind, model_variant_id, skill_id
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: ListRevisions :many
SELECT revisions.*
FROM revisions
JOIN content_items ON content_items.id = revisions.content_item_id
WHERE revisions.content_item_id = sqlc.arg(content_item_id)
  AND content_items.project_id = sqlc.arg(project_id)
ORDER BY revision_number DESC;

-- name: GetRevisionByNumber :one
SELECT revisions.*
FROM revisions
JOIN content_items ON content_items.id = revisions.content_item_id
WHERE revisions.content_item_id = sqlc.arg(content_item_id)
  AND content_items.project_id = sqlc.arg(project_id)
  AND revisions.revision_number = sqlc.arg(revision_number);

-- name: CreateEntrySection :exec
INSERT INTO entry_sections (id, content_item_id, heading, body_markdown, sort_order)
VALUES (?, ?, ?, ?, ?);

-- name: ListEntrySections :many
SELECT entry_sections.*
FROM entry_sections
JOIN content_items ON content_items.id = entry_sections.content_item_id
WHERE entry_sections.content_item_id = sqlc.arg(content_item_id)
  AND content_items.project_id = sqlc.arg(project_id)
ORDER BY entry_sections.sort_order ASC;

-- name: UpdateEntrySectionBody :execrows
UPDATE entry_sections
SET body_markdown = sqlc.arg(body_markdown),
    current_revision = entry_sections.current_revision + 1
WHERE entry_sections.content_item_id = sqlc.arg(content_item_id)
  AND entry_sections.heading = sqlc.arg(heading)
  AND entry_sections.current_revision = sqlc.arg(expected_revision)
  AND EXISTS (
    SELECT 1
    FROM content_items
    WHERE content_items.id = entry_sections.content_item_id
      AND content_items.project_id = sqlc.arg(project_id)
  );

-- name: GetEntrySection :one
SELECT entry_sections.*
FROM entry_sections
JOIN content_items ON content_items.id = entry_sections.content_item_id
WHERE entry_sections.content_item_id = sqlc.arg(content_item_id)
  AND entry_sections.heading = sqlc.arg(heading)
  AND content_items.project_id = sqlc.arg(project_id);

-- name: CreateEntryRelation :exec
INSERT INTO entry_relations (
  id, project_id, source_item_id, target_item_id, target_title, relation_type, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: ListEntryRelations :many
SELECT * FROM entry_relations
WHERE project_id = ? AND source_item_id = ?
ORDER BY target_title ASC;

-- name: CreateAttachedNote :exec
INSERT INTO attached_notes (
  id, project_id, content_item_id, selection_start, selection_end,
  title, body_markdown, source, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: SearchContent :many
SELECT content_items.*
FROM content_search(CAST(sqlc.arg(query) AS TEXT))
JOIN content_items ON content_items.rowid = content_search.rowid
WHERE content_items.project_id = sqlc.arg(project_id)
ORDER BY rank
LIMIT sqlc.arg(limit);

-- name: SearchContentCandidates :many
SELECT content_items.*
FROM content_search(CAST(sqlc.arg(query) AS TEXT))
JOIN content_items ON content_items.rowid = content_search.rowid
WHERE content_items.project_id = sqlc.arg(project_id)
ORDER BY rank
LIMIT 1000;
