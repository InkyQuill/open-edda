-- name: UpsertSkillScriptAudit :exec
INSERT INTO skill_script_audits (
  id, project_id, skill_id, skill_file_id, relative_path, runtime,
  destructive_operations, filesystem_access, network_access, external_dependencies,
  expected_inputs_json, expected_outputs_json, risk_notes, recommendation, audited_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(project_id, skill_file_id) DO UPDATE SET
  runtime = excluded.runtime,
  destructive_operations = excluded.destructive_operations,
  filesystem_access = excluded.filesystem_access,
  network_access = excluded.network_access,
  external_dependencies = excluded.external_dependencies,
  expected_inputs_json = excluded.expected_inputs_json,
  expected_outputs_json = excluded.expected_outputs_json,
  risk_notes = excluded.risk_notes,
  recommendation = excluded.recommendation,
  audited_at = excluded.audited_at;

-- name: ListSkillScriptAudits :many
SELECT * FROM skill_script_audits
WHERE project_id = ? AND skill_id = ?
ORDER BY relative_path ASC;

-- name: GetSkillScriptAuditByFile :one
SELECT * FROM skill_script_audits
WHERE project_id = ? AND skill_file_id = ?;

-- name: UpsertSkillScriptApproval :exec
INSERT INTO skill_script_approvals (
  id, project_id, skill_id, skill_file_id, audit_id, enabled, runtime_command,
  timeout_ms, max_stdout_bytes, max_stderr_bytes, allow_network,
  allow_project_files, approved_by, approved_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(project_id, skill_file_id) DO UPDATE SET
  audit_id = excluded.audit_id,
  enabled = excluded.enabled,
  runtime_command = excluded.runtime_command,
  timeout_ms = excluded.timeout_ms,
  max_stdout_bytes = excluded.max_stdout_bytes,
  max_stderr_bytes = excluded.max_stderr_bytes,
  allow_network = excluded.allow_network,
  allow_project_files = excluded.allow_project_files,
  approved_by = excluded.approved_by,
  updated_at = excluded.updated_at;

-- name: GetSkillScriptApprovalByFile :one
SELECT * FROM skill_script_approvals
WHERE project_id = ? AND skill_file_id = ?;

-- name: ListEnabledSkillScriptApprovals :many
SELECT skill_script_approvals.*
FROM skill_script_approvals
JOIN skills ON skills.id = skill_script_approvals.skill_id
WHERE skill_script_approvals.project_id = ?
  AND skill_script_approvals.enabled = 1
ORDER BY skills.name ASC, skill_script_approvals.skill_file_id ASC;

-- name: CreateSkillScriptRun :exec
INSERT INTO skill_script_runs (
  id, project_id, session_id, skill_id, skill_file_id, approval_id, tool_call_id,
  status, output_kind, input_json, output_json, stdout_text, stderr_text,
  exit_code, duration_ms, error_message, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: ListSkillScriptRunsByProject :many
SELECT * FROM skill_script_runs
WHERE project_id = ?
ORDER BY created_at DESC
LIMIT ?;

-- name: ListSkillScriptRunsBySession :many
SELECT * FROM skill_script_runs
WHERE project_id = ? AND session_id = ?
ORDER BY created_at DESC
LIMIT ?;
