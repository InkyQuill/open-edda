-- name: UpsertSkill :exec
INSERT INTO skills (
  id, project_id, name, display_name, description, instructions_markdown,
  source_type, source_label, script_count, scripts_disabled, metadata_json,
  installed_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(project_id, name) DO UPDATE SET
  display_name = excluded.display_name,
  description = excluded.description,
  instructions_markdown = excluded.instructions_markdown,
  source_type = excluded.source_type,
  source_label = excluded.source_label,
  script_count = excluded.script_count,
  scripts_disabled = excluded.scripts_disabled,
  metadata_json = excluded.metadata_json,
  updated_at = excluded.updated_at;

-- name: GetSkillByProjectName :one
SELECT * FROM skills
WHERE project_id = ? AND name = ?;

-- name: GetSkillByProjectID :one
SELECT * FROM skills
WHERE project_id = ? AND id = ?;

-- name: ListSkillsByProject :many
SELECT * FROM skills
WHERE project_id = ?
ORDER BY display_name ASC, name ASC;

-- name: CountSkillsByProject :one
SELECT COUNT(*) FROM skills
WHERE project_id = ?;

-- name: DeleteSkillFiles :exec
DELETE FROM skill_files
WHERE skill_id = ?;

-- name: CreateSkillFile :exec
INSERT INTO skill_files (
  id, skill_id, relative_path, purpose, media_type, body_text, bytes, script_disabled, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: ListSkillFiles :many
SELECT * FROM skill_files
WHERE skill_id = ?
ORDER BY relative_path ASC;

-- name: GetSkillFile :one
SELECT skill_files.*
FROM skill_files
JOIN skills ON skills.id = skill_files.skill_id
WHERE skills.project_id = sqlc.arg(project_id)
  AND skill_files.skill_id = sqlc.arg(skill_id)
  AND skill_files.relative_path = sqlc.arg(relative_path);

-- name: DeleteSkillRoutingHints :exec
DELETE FROM skill_routing_hints
WHERE skill_id = ?;

-- name: CreateSkillRoutingHint :exec
INSERT INTO skill_routing_hints (
  id, skill_id, action_kind, content_kind, tag, priority, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: ListSkillRoutingHints :many
SELECT * FROM skill_routing_hints
WHERE skill_id = ?
ORDER BY priority DESC, action_kind ASC, content_kind ASC, tag ASC;

-- name: ListRoutableSkills :many
SELECT DISTINCT skills.*
FROM skills
LEFT JOIN skill_routing_hints ON skill_routing_hints.skill_id = skills.id
WHERE skills.project_id = sqlc.arg(project_id)
  AND (
    skill_routing_hints.id IS NULL
    OR skill_routing_hints.action_kind = ''
    OR skill_routing_hints.action_kind = sqlc.arg(action_kind)
  )
  AND (
    skill_routing_hints.id IS NULL
    OR skill_routing_hints.content_kind = ''
    OR skill_routing_hints.content_kind = sqlc.arg(content_kind)
  )
ORDER BY skills.display_name ASC, skills.name ASC;

-- name: ReplaceSessionSkillsDelete :exec
DELETE FROM agent_session_skills
WHERE session_id = ?;

-- name: AddSessionSkill :execrows
INSERT INTO agent_session_skills (project_id, session_id, skill_id, selected_at)
SELECT sqlc.arg(project_id), agent_sessions.id, skills.id, sqlc.arg(selected_at)
FROM agent_sessions
JOIN skills ON skills.id = sqlc.arg(skill_id)
WHERE agent_sessions.id = sqlc.arg(session_id)
  AND agent_sessions.project_id = sqlc.arg(project_id)
  AND skills.project_id = sqlc.arg(project_id);

-- name: ListSessionSkills :many
SELECT skills.*
FROM agent_session_skills
JOIN skills ON skills.id = agent_session_skills.skill_id
JOIN agent_sessions ON agent_sessions.id = agent_session_skills.session_id
WHERE agent_session_skills.session_id = sqlc.arg(session_id)
  AND agent_sessions.project_id = sqlc.arg(project_id)
  AND skills.project_id = agent_sessions.project_id
ORDER BY skills.display_name ASC, skills.name ASC;

-- name: ListSessionSkillIDs :many
SELECT skills.id
FROM agent_session_skills
JOIN skills ON skills.id = agent_session_skills.skill_id
JOIN agent_sessions ON agent_sessions.id = agent_session_skills.session_id
WHERE agent_session_skills.session_id = sqlc.arg(session_id)
  AND agent_sessions.project_id = sqlc.arg(project_id)
  AND skills.project_id = agent_sessions.project_id
ORDER BY skills.display_name ASC, skills.name ASC;
