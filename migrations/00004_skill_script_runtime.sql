-- +goose Up
PRAGMA foreign_keys = ON;

CREATE TABLE skill_script_audits (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL REFERENCES story_projects(id) ON DELETE CASCADE,
  skill_id TEXT NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
  skill_file_id TEXT NOT NULL REFERENCES skill_files(id) ON DELETE CASCADE,
  relative_path TEXT NOT NULL,
  runtime TEXT NOT NULL DEFAULT '',
  destructive_operations INTEGER NOT NULL CHECK(destructive_operations IN (0, 1)) DEFAULT 0,
  filesystem_access TEXT NOT NULL DEFAULT 'none' CHECK(filesystem_access IN ('none', 'read_assets', 'temp_workspace', 'project_files')),
  network_access INTEGER NOT NULL CHECK(network_access IN (0, 1)) DEFAULT 0,
  external_dependencies TEXT NOT NULL DEFAULT '',
  expected_inputs_json TEXT NOT NULL DEFAULT '{}',
  expected_outputs_json TEXT NOT NULL DEFAULT '{}',
  risk_notes TEXT NOT NULL DEFAULT '',
  recommendation TEXT NOT NULL DEFAULT 'disabled' CHECK(recommendation IN ('disabled', 'approve_with_limits', 'defer')),
  audited_at TEXT NOT NULL,
  UNIQUE(project_id, skill_file_id)
);

CREATE TABLE skill_script_approvals (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL REFERENCES story_projects(id) ON DELETE CASCADE,
  skill_id TEXT NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
  skill_file_id TEXT NOT NULL REFERENCES skill_files(id) ON DELETE CASCADE,
  audit_id TEXT NOT NULL REFERENCES skill_script_audits(id) ON DELETE CASCADE,
  enabled INTEGER NOT NULL CHECK(enabled IN (0, 1)) DEFAULT 0,
  runtime_command TEXT NOT NULL DEFAULT '',
  timeout_ms INTEGER NOT NULL DEFAULT 5000,
  max_stdout_bytes INTEGER NOT NULL DEFAULT 65536,
  max_stderr_bytes INTEGER NOT NULL DEFAULT 16384,
  allow_network INTEGER NOT NULL CHECK(allow_network IN (0, 1)) DEFAULT 0,
  allow_project_files INTEGER NOT NULL CHECK(allow_project_files IN (0, 1)) DEFAULT 0,
  approved_by TEXT NOT NULL DEFAULT 'local-admin',
  approved_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE(project_id, skill_file_id)
);

CREATE TABLE skill_script_runs (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL REFERENCES story_projects(id) ON DELETE CASCADE,
  session_id TEXT REFERENCES agent_sessions(id) ON DELETE SET NULL,
  skill_id TEXT NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
  skill_file_id TEXT NOT NULL REFERENCES skill_files(id) ON DELETE CASCADE,
  approval_id TEXT NOT NULL REFERENCES skill_script_approvals(id) ON DELETE RESTRICT,
  tool_call_id TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL CHECK(status IN ('succeeded', 'failed', 'timed_out', 'rejected')),
  output_kind TEXT NOT NULL DEFAULT '',
  input_json TEXT NOT NULL DEFAULT '{}',
  output_json TEXT NOT NULL DEFAULT '{}',
  stdout_text TEXT NOT NULL DEFAULT '',
  stderr_text TEXT NOT NULL DEFAULT '',
  exit_code INTEGER NOT NULL DEFAULT 0,
  duration_ms INTEGER NOT NULL DEFAULT 0,
  error_message TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL
);

CREATE INDEX idx_skill_script_audits_project ON skill_script_audits(project_id, skill_id);
CREATE INDEX idx_skill_script_approvals_project ON skill_script_approvals(project_id, skill_id, enabled);
CREATE INDEX idx_skill_script_runs_project ON skill_script_runs(project_id, created_at DESC);
CREATE INDEX idx_skill_script_runs_session ON skill_script_runs(project_id, session_id, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_skill_script_runs_session;
DROP INDEX IF EXISTS idx_skill_script_runs_project;
DROP INDEX IF EXISTS idx_skill_script_approvals_project;
DROP INDEX IF EXISTS idx_skill_script_audits_project;
DROP TABLE IF EXISTS skill_script_runs;
DROP TABLE IF EXISTS skill_script_approvals;
DROP TABLE IF EXISTS skill_script_audits;
