-- +goose Up
PRAGMA foreign_keys = ON;

CREATE TABLE skills (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL REFERENCES story_projects(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  display_name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  instructions_markdown TEXT NOT NULL DEFAULT '',
  source_type TEXT NOT NULL CHECK(source_type IN ('upload', 'local_directory')),
  source_label TEXT NOT NULL DEFAULT '',
  script_count INTEGER NOT NULL DEFAULT 0,
  scripts_disabled INTEGER NOT NULL CHECK(scripts_disabled IN (0, 1)) DEFAULT 1,
  metadata_json TEXT NOT NULL DEFAULT '{}',
  installed_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE(project_id, name)
);

CREATE TABLE skill_files (
  id TEXT PRIMARY KEY,
  skill_id TEXT NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
  relative_path TEXT NOT NULL,
  purpose TEXT NOT NULL CHECK(purpose IN ('instruction', 'template', 'reference', 'data', 'script', 'other')),
  media_type TEXT NOT NULL DEFAULT 'text/plain; charset=utf-8',
  body_text TEXT NOT NULL,
  bytes INTEGER NOT NULL DEFAULT 0,
  script_disabled INTEGER NOT NULL CHECK(script_disabled IN (0, 1)) DEFAULT 0,
  created_at TEXT NOT NULL,
  UNIQUE(skill_id, relative_path)
);

CREATE TABLE skill_routing_hints (
  id TEXT PRIMARY KEY,
  skill_id TEXT NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
  action_kind TEXT NOT NULL DEFAULT '',
  content_kind TEXT NOT NULL DEFAULT '',
  tag TEXT NOT NULL DEFAULT '',
  priority INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL
);

CREATE TABLE agent_session_skills (
  session_id TEXT NOT NULL REFERENCES agent_sessions(id) ON DELETE CASCADE,
  skill_id TEXT NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
  selected_at TEXT NOT NULL,
  PRIMARY KEY (session_id, skill_id)
);

ALTER TABLE generation_candidates ADD COLUMN skill_id TEXT NOT NULL DEFAULT '';

CREATE INDEX idx_skills_project_name ON skills(project_id, name);
CREATE INDEX idx_skill_files_skill_path ON skill_files(skill_id, relative_path);
CREATE INDEX idx_skill_routing_hints_lookup ON skill_routing_hints(action_kind, content_kind, tag, priority DESC);
CREATE INDEX idx_agent_session_skills_skill ON agent_session_skills(skill_id, session_id);

-- +goose Down
DROP INDEX IF EXISTS idx_agent_session_skills_skill;
DROP INDEX IF EXISTS idx_skill_routing_hints_lookup;
DROP INDEX IF EXISTS idx_skill_files_skill_path;
DROP INDEX IF EXISTS idx_skills_project_name;
ALTER TABLE generation_candidates DROP COLUMN skill_id;
DROP TABLE IF EXISTS agent_session_skills;
DROP TABLE IF EXISTS skill_routing_hints;
DROP TABLE IF EXISTS skill_files;
DROP TABLE IF EXISTS skills;
