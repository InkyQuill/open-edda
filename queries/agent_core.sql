-- name: CreateProviderConfig :exec
INSERT INTO provider_configs (
  id, author_id, name, base_url, api_key_encrypted, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: ListProviderConfigs :many
SELECT * FROM provider_configs
WHERE author_id = ?
ORDER BY name ASC;

-- name: GetProviderConfig :one
SELECT * FROM provider_configs
WHERE id = sqlc.arg(id)
  AND author_id = sqlc.arg(author_id);

-- name: UpdateProviderConfig :exec
UPDATE provider_configs
SET base_url = sqlc.arg(base_url),
    api_key_encrypted = sqlc.arg(api_key_encrypted),
    updated_at = sqlc.arg(updated_at)
WHERE id = sqlc.arg(id)
  AND author_id = sqlc.arg(author_id);

-- name: DeleteProviderConfig :exec
DELETE FROM provider_configs
WHERE id = sqlc.arg(id)
  AND author_id = sqlc.arg(author_id);

-- name: CreateModelVariant :exec
INSERT INTO model_variants (
  id, provider_config_id, name, model, temperature, max_output_tokens,
  context_window_tokens, input_price_per_million, output_price_per_million,
  cache_read_price_per_million, cache_write_price_per_million,
  request_token_field, reasoning_format, compatibility_json, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: ListModelVariantsByProvider :many
SELECT model_variants.*
FROM model_variants
JOIN provider_configs ON provider_configs.id = model_variants.provider_config_id
WHERE model_variants.provider_config_id = sqlc.arg(provider_config_id)
  AND provider_configs.author_id = sqlc.arg(author_id)
ORDER BY model_variants.name ASC;

-- name: ListModelVariantsByAuthor :many
SELECT model_variants.*
FROM model_variants
JOIN provider_configs ON provider_configs.id = model_variants.provider_config_id
WHERE provider_configs.author_id = ?
ORDER BY provider_configs.name ASC, model_variants.name ASC;

-- name: GetModelVariant :one
SELECT model_variants.*
FROM model_variants
JOIN provider_configs ON provider_configs.id = model_variants.provider_config_id
WHERE model_variants.id = sqlc.arg(id)
  AND provider_configs.author_id = sqlc.arg(author_id);

-- name: GetModelVariantForProject :one
SELECT model_variants.*
FROM model_variants
JOIN provider_configs ON provider_configs.id = model_variants.provider_config_id
JOIN story_projects ON story_projects.author_id = provider_configs.author_id
WHERE story_projects.id = sqlc.arg(project_id)
  AND model_variants.id = sqlc.arg(model_variant_id);

-- name: UpdateModelVariant :exec
UPDATE model_variants
SET name = sqlc.arg(name),
    model = sqlc.arg(model),
    temperature = sqlc.arg(temperature),
    max_output_tokens = sqlc.arg(max_output_tokens),
    context_window_tokens = sqlc.arg(context_window_tokens),
    input_price_per_million = sqlc.arg(input_price_per_million),
    output_price_per_million = sqlc.arg(output_price_per_million),
    cache_read_price_per_million = sqlc.arg(cache_read_price_per_million),
    cache_write_price_per_million = sqlc.arg(cache_write_price_per_million),
    request_token_field = sqlc.arg(request_token_field),
    reasoning_format = sqlc.arg(reasoning_format),
    compatibility_json = sqlc.arg(compatibility_json),
    updated_at = sqlc.arg(updated_at)
WHERE model_variants.id = sqlc.arg(id)
  AND provider_config_id IN (
    SELECT provider_configs.id FROM provider_configs
    WHERE provider_configs.author_id = sqlc.arg(author_id)
  );

-- name: DeleteModelVariant :exec
DELETE FROM model_variants
WHERE model_variants.id = sqlc.arg(id)
  AND provider_config_id IN (
    SELECT provider_configs.id FROM provider_configs
    WHERE provider_configs.author_id = sqlc.arg(author_id)
  );

-- name: UpsertPromptProfile :exec
INSERT INTO prompt_profiles (
  id, project_id, genre, tense, pov, voice, instructions_markdown,
  prompt_record_retention_days, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(project_id) DO UPDATE SET
  genre = excluded.genre,
  tense = excluded.tense,
  pov = excluded.pov,
  voice = excluded.voice,
  instructions_markdown = excluded.instructions_markdown,
  prompt_record_retention_days = excluded.prompt_record_retention_days,
  updated_at = excluded.updated_at;

-- name: GetPromptProfile :one
SELECT * FROM prompt_profiles
WHERE project_id = ?;

-- name: CreateAgentSession :exec
INSERT INTO agent_sessions (
  id, project_id, title, action_kind, model_variant_id, apply_mode, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: ListAgentSessions :many
SELECT * FROM agent_sessions
WHERE project_id = ?
ORDER BY updated_at DESC
LIMIT sqlc.arg(limit);

-- name: GetAgentSession :one
SELECT * FROM agent_sessions
WHERE project_id = ? AND id = ?;

-- name: TouchAgentSession :exec
UPDATE agent_sessions
SET updated_at = sqlc.arg(updated_at)
WHERE project_id = sqlc.arg(project_id)
  AND id = sqlc.arg(id);

-- name: CreateAgentMessage :exec
INSERT INTO agent_messages (
  id, session_id, role, body_markdown, metadata_json, created_at
) VALUES (?, ?, ?, ?, ?, ?);

-- name: ListAgentMessages :many
SELECT * FROM agent_messages
WHERE session_id = ?
ORDER BY created_at ASC;

-- name: CreateAgentMessageForProject :execrows
INSERT INTO agent_messages (
  id, session_id, role, body_markdown, metadata_json, created_at
)
SELECT
  sqlc.arg(id),
  agent_sessions.id,
  sqlc.arg(role),
  sqlc.arg(body_markdown),
  sqlc.arg(metadata_json),
  sqlc.arg(created_at)
FROM agent_sessions
WHERE agent_sessions.id = sqlc.arg(session_id)
  AND agent_sessions.project_id = sqlc.arg(project_id);

-- name: ListAgentMessagesForProject :many
SELECT agent_messages.*
FROM agent_messages
JOIN agent_sessions ON agent_sessions.id = agent_messages.session_id
WHERE agent_messages.session_id = sqlc.arg(session_id)
  AND agent_sessions.project_id = sqlc.arg(project_id)
ORDER BY agent_messages.created_at ASC;

-- name: CreateActivityEvent :exec
INSERT INTO activity_events (
  id, project_id, session_id, event_type, summary, metadata_json, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: ListActivityEvents :many
SELECT * FROM activity_events
WHERE project_id = ?
ORDER BY created_at DESC
LIMIT sqlc.arg(limit);

-- name: CreatePromptRecord :exec
INSERT INTO prompt_records (
  id, project_id, session_id, provider_name, model_name, action_kind,
  request_json, response_json, input_tokens, output_tokens, cache_read_tokens,
  cache_write_tokens, total_tokens, input_cost, output_cost, cache_read_cost,
  cache_write_cost, total_cost, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: ListPromptRecords :many
SELECT * FROM prompt_records
WHERE project_id = ?
ORDER BY created_at DESC
LIMIT sqlc.arg(limit);

-- name: DeleteExpiredPromptRecords :execrows
DELETE FROM prompt_records
WHERE project_id = sqlc.arg(project_id)
  AND created_at < sqlc.arg(cutoff);

-- name: CreatePromptContextSnapshot :exec
INSERT INTO prompt_context_snapshots (
  id, prompt_record_id, source_key, source_version, rendered_markdown, value_json, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: ListPromptContextSnapshots :many
SELECT * FROM prompt_context_snapshots
WHERE prompt_record_id = ?
ORDER BY source_key ASC;

-- name: CreateToolResultArtifact :exec
INSERT INTO tool_result_artifacts (
  id, project_id, session_id, tool_call_id, tool_name, full_result_json,
  model_visible_markdown, truncated, full_result_bytes, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: ListToolResultArtifacts :many
SELECT * FROM tool_result_artifacts
WHERE project_id = sqlc.arg(project_id)
  AND session_id = sqlc.arg(session_id)
ORDER BY created_at DESC;

-- name: ListToolResultArtifactsByProject :many
SELECT * FROM tool_result_artifacts
WHERE project_id = ?
ORDER BY created_at DESC;

-- name: CreateGenerationCandidate :exec
INSERT INTO generation_candidates (
  id, project_id, session_id, content_item_id, action_kind, operation_kind,
  expected_revision, selection_start, selection_end, insert_position,
  original_markdown, generated_markdown, reason, model_variant_id, status,
  created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetGenerationCandidate :one
SELECT * FROM generation_candidates
WHERE project_id = ? AND id = ?;

-- name: ListGenerationCandidates :many
SELECT * FROM generation_candidates
WHERE project_id = sqlc.arg(project_id)
  AND session_id = sqlc.arg(session_id)
ORDER BY created_at DESC;

-- name: UpdateGenerationCandidateStatus :exec
UPDATE generation_candidates
SET status = sqlc.arg(status),
    updated_at = sqlc.arg(updated_at)
WHERE project_id = sqlc.arg(project_id)
  AND id = sqlc.arg(id);
