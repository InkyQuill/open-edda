-- +goose Up
PRAGMA foreign_keys = ON;

CREATE TABLE provider_configs (
  id TEXT PRIMARY KEY,
  author_id TEXT NOT NULL REFERENCES authors(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  base_url TEXT NOT NULL,
  api_key_encrypted TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE(author_id, name)
);

CREATE TABLE model_variants (
  id TEXT PRIMARY KEY,
  provider_config_id TEXT NOT NULL REFERENCES provider_configs(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  model TEXT NOT NULL,
  temperature REAL NOT NULL DEFAULT 0.7,
  max_output_tokens INTEGER NOT NULL DEFAULT 2048,
  context_window_tokens INTEGER NOT NULL DEFAULT 0,
  input_price_per_million REAL NOT NULL DEFAULT 0,
  output_price_per_million REAL NOT NULL DEFAULT 0,
  cache_read_price_per_million REAL NOT NULL DEFAULT 0,
  cache_write_price_per_million REAL NOT NULL DEFAULT 0,
  request_token_field TEXT NOT NULL DEFAULT 'max_tokens',
  reasoning_format TEXT NOT NULL DEFAULT '',
  compatibility_json TEXT NOT NULL DEFAULT '{}',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE(provider_config_id, name)
);

CREATE TABLE prompt_profiles (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL REFERENCES story_projects(id) ON DELETE CASCADE,
  genre TEXT NOT NULL DEFAULT '',
  tense TEXT NOT NULL DEFAULT '',
  pov TEXT NOT NULL DEFAULT '',
  voice TEXT NOT NULL DEFAULT '',
  instructions_markdown TEXT NOT NULL DEFAULT '',
  prompt_record_retention_days INTEGER NOT NULL DEFAULT 30,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE(project_id)
);

CREATE TABLE agent_sessions (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL REFERENCES story_projects(id) ON DELETE CASCADE,
  title TEXT NOT NULL,
  action_kind TEXT NOT NULL CHECK(action_kind IN ('chat', 'continuation', 'rewrite', 'read_check')),
  model_variant_id TEXT REFERENCES model_variants(id) ON DELETE SET NULL,
  apply_mode TEXT NOT NULL CHECK(apply_mode IN ('preview', 'direct_apply')),
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE agent_messages (
  id TEXT PRIMARY KEY,
  session_id TEXT NOT NULL REFERENCES agent_sessions(id) ON DELETE CASCADE,
  role TEXT NOT NULL CHECK(role IN ('user', 'assistant', 'tool', 'system')),
  body_markdown TEXT NOT NULL,
  metadata_json TEXT NOT NULL DEFAULT '{}',
  created_at TEXT NOT NULL
);

CREATE TABLE activity_events (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL REFERENCES story_projects(id) ON DELETE CASCADE,
  session_id TEXT REFERENCES agent_sessions(id) ON DELETE SET NULL,
  event_type TEXT NOT NULL,
  summary TEXT NOT NULL,
  metadata_json TEXT NOT NULL DEFAULT '{}',
  created_at TEXT NOT NULL
);

CREATE TABLE prompt_records (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL REFERENCES story_projects(id) ON DELETE CASCADE,
  session_id TEXT REFERENCES agent_sessions(id) ON DELETE SET NULL,
  provider_name TEXT NOT NULL,
  model_name TEXT NOT NULL,
  action_kind TEXT NOT NULL,
  request_json TEXT NOT NULL,
  response_json TEXT NOT NULL,
  input_tokens INTEGER NOT NULL DEFAULT 0,
  output_tokens INTEGER NOT NULL DEFAULT 0,
  cache_read_tokens INTEGER NOT NULL DEFAULT 0,
  cache_write_tokens INTEGER NOT NULL DEFAULT 0,
  total_tokens INTEGER NOT NULL DEFAULT 0,
  input_cost REAL NOT NULL DEFAULT 0,
  output_cost REAL NOT NULL DEFAULT 0,
  cache_read_cost REAL NOT NULL DEFAULT 0,
  cache_write_cost REAL NOT NULL DEFAULT 0,
  total_cost REAL NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL
);

CREATE TABLE prompt_context_snapshots (
  id TEXT PRIMARY KEY,
  prompt_record_id TEXT NOT NULL REFERENCES prompt_records(id) ON DELETE CASCADE,
  source_key TEXT NOT NULL,
  source_version TEXT NOT NULL DEFAULT '',
  rendered_markdown TEXT NOT NULL,
  value_json TEXT NOT NULL DEFAULT '{}',
  created_at TEXT NOT NULL,
  UNIQUE(prompt_record_id, source_key)
);

CREATE TABLE tool_result_artifacts (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL REFERENCES story_projects(id) ON DELETE CASCADE,
  session_id TEXT REFERENCES agent_sessions(id) ON DELETE SET NULL,
  tool_call_id TEXT NOT NULL,
  tool_name TEXT NOT NULL,
  full_result_json TEXT NOT NULL,
  model_visible_markdown TEXT NOT NULL,
  truncated INTEGER NOT NULL CHECK(truncated IN (0, 1)),
  full_result_bytes INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL
);

CREATE TABLE generation_candidates (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL REFERENCES story_projects(id) ON DELETE CASCADE,
  session_id TEXT NOT NULL REFERENCES agent_sessions(id) ON DELETE CASCADE,
  content_item_id TEXT NOT NULL,
  action_kind TEXT NOT NULL CHECK(action_kind IN ('continuation', 'rewrite')),
  operation_kind TEXT NOT NULL CHECK(operation_kind IN ('append', 'insert', 'replace')),
  expected_revision INTEGER NOT NULL,
  selection_start INTEGER,
  selection_end INTEGER,
  insert_position INTEGER,
  original_markdown TEXT NOT NULL,
  generated_markdown TEXT NOT NULL,
  reason TEXT NOT NULL,
  model_variant_id TEXT REFERENCES model_variants(id) ON DELETE SET NULL,
  status TEXT NOT NULL CHECK(status IN ('pending', 'applying', 'accepted', 'rejected', 'conflict')),
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (content_item_id, project_id) REFERENCES content_items(id, project_id) ON DELETE CASCADE
);

ALTER TABLE revisions ADD COLUMN agent_session_id TEXT REFERENCES agent_sessions(id) ON DELETE SET NULL;
ALTER TABLE revisions ADD COLUMN action_kind TEXT NOT NULL DEFAULT '';
ALTER TABLE revisions ADD COLUMN model_variant_id TEXT REFERENCES model_variants(id) ON DELETE SET NULL;
ALTER TABLE revisions ADD COLUMN skill_id TEXT NOT NULL DEFAULT '';

CREATE INDEX idx_model_variants_provider_name ON model_variants(provider_config_id, name);
CREATE INDEX idx_agent_sessions_project_updated ON agent_sessions(project_id, updated_at DESC);
CREATE INDEX idx_agent_messages_session_created ON agent_messages(session_id, created_at ASC);
CREATE INDEX idx_activity_events_project_created ON activity_events(project_id, created_at DESC);
CREATE INDEX idx_prompt_records_project_created ON prompt_records(project_id, created_at DESC);
CREATE INDEX idx_generation_candidates_project_session ON generation_candidates(project_id, session_id, created_at DESC);
CREATE INDEX idx_tool_result_artifacts_project_session ON tool_result_artifacts(project_id, session_id, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_prompt_records_project_created;
DROP INDEX IF EXISTS idx_generation_candidates_project_session;
DROP INDEX IF EXISTS idx_tool_result_artifacts_project_session;
DROP INDEX IF EXISTS idx_activity_events_project_created;
DROP INDEX IF EXISTS idx_agent_messages_session_created;
DROP INDEX IF EXISTS idx_agent_sessions_project_updated;
DROP INDEX IF EXISTS idx_model_variants_provider_name;
ALTER TABLE revisions DROP COLUMN skill_id;
ALTER TABLE revisions DROP COLUMN model_variant_id;
ALTER TABLE revisions DROP COLUMN action_kind;
ALTER TABLE revisions DROP COLUMN agent_session_id;
DROP TABLE IF EXISTS prompt_context_snapshots;
DROP TABLE IF EXISTS tool_result_artifacts;
DROP TABLE IF EXISTS prompt_records;
DROP TABLE IF EXISTS generation_candidates;
DROP TABLE IF EXISTS activity_events;
DROP TABLE IF EXISTS agent_messages;
DROP TABLE IF EXISTS agent_sessions;
DROP TABLE IF EXISTS prompt_profiles;
DROP TABLE IF EXISTS model_variants;
DROP TABLE IF EXISTS provider_configs;
