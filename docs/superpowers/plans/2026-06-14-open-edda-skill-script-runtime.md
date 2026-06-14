# Open Edda Skill Script Runtime Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build Milestone 3.6: an admin-approved skill script runtime that runs audited helper scripts against database-backed Open Edda project inputs and returns reviewable reports, proposals, generated data, or drafts without directly mutating story content.

**Architecture:** Extend Skill Core with script audit, approval, and run records, then add a small runtime package that launches approved scripts through a strict JSON stdin/stdout envelope in an empty temporary working directory. Agent Core exposes a bounded `skill_script` tool only for enabled scripts, and the UI exposes admin inspection/enablement plus run history; all outputs are artifacts for review, not direct writes to Story Text or Story Bible records.

**Tech Stack:** Go 1.26, chi, sqlc, goose, SQLite, existing Skill Core and Agent Core tool plumbing, React, TypeScript, Bun, Vite. Runtime v1 uses local executable commands only when explicitly allowed by admin policy; no network access is granted by Open Edda.

---

## Scope

This plan implements Milestone 3.6 only:

- Store script audit records for imported script files.
- Let an admin enable or disable individual script files per project.
- Record runtime policy: runtime command, timeout, network flag, filesystem mode, input contract, output contract, and audit notes.
- Add a safe script execution envelope:
  - Open Edda builds JSON input from explicit project selections and skill assets.
  - The script receives only JSON on stdin and a temp working directory.
  - The script returns JSON on stdout.
  - Open Edda stores stdout/stderr/status/duration as a run record and tool artifact.
  - Outputs are reviewable reports, proposals, generated data, or draft text.
- Add model-facing `skill_script` tool for selected sessions.
- Add HTTP/admin UI surfaces for audits, approvals, and run history.
- Add compatibility/degradation behavior when a script is missing runtime support, disabled, times out, returns invalid JSON, or requests unsupported mutation.
- Update docs and roadmap tracking.

This plan does not implement:

- Arbitrary remote marketplace script execution.
- Live filesystem project access.
- Silent mutation of Story Text, Story Bible Entries, Entry Sections, Project Notes, Attached Notes, or project structure.
- Network sandbox enforcement beyond policy rejection. Scripts that require network remain disabled in v1.
- Long-running watchers such as `story-zoom/watcher.ts`.
- Full reverse-outliner orchestration as a product feature. This plan creates the runtime scaffolding and a single report/proposal path that later work can build on.

## Product Rules

- `$` remains the skill mention prefix. Script execution is not invoked by `$` alone; the model or UI must explicitly call an enabled script.
- `/` remains command syntax and is not reused for skills.
- `@` remains entity mention syntax and is the preferred way for authors to select script inputs.
- Admin enablement is per script file, not per whole skill.
- Imported script files are never automatically executable after import.
- Runtime-backed scripts must degrade clearly: disabled means the agent explains that the script is unavailable and continues with normal Edda-native guidance.
- Subsystems introduced by this work may use Scandinavian names only where the boundary helps. The runtime package remains `skill/runtime` because it is a technical implementation detail; a later user-facing orchestration surface may receive a Scandinavian name.

## File Structure

Create:

- `migrations/00004_skill_script_runtime.sql` - script audit, approval, and run tables.
- `queries/skill_script_runtime.sql` - sqlc queries for audits, approvals, and run history.
- `skill/runtime/types.go` - JSON envelope, policy, input/output, and run result types.
- `skill/runtime/runner.go` - approved command runner with timeout, temp directory, stdin/stdout limits, and output validation.
- `skill/runtime/runner_test.go` - pure runtime tests using local shell helper commands.
- `skill/script_runtime.go` - Skill Service methods for audit creation, policy updates, run validation, and run persistence.
- `skill/script_runtime_test.go` - service-level tests against SQLite.
- `frontend/src/scriptRuntimeTypes.ts` - TypeScript types for audits, approvals, and run history.
- `frontend/src/scriptRuntimeApi.ts` - API calls for script runtime admin and history.

Modify:

- `skill/types.go` - add script audit, policy, and run DTOs.
- `skill/service.go` - add audit creation during install and expose runtime dependencies.
- `skill/http.go` - add script audit, approval, and run history endpoints.
- `agent/types.go` - add `skill_script` tool result DTOs if needed.
- `agent/service.go` - extend `SkillProvider` with runtime methods.
- `agent/tools.go` - add `skill_script` tool definition and execution branch.
- `agent/tools_test.go` - add enabled/disabled script tool tests.
- `agent/prompt.go` - disclose enabled script helpers in skill context.
- `agent/prompt_test.go` - verify script status appears in prompt context.
- `app/app.go` - no route shape change beyond `skill.RegisterRoutes`, but ensure runtime service dependencies are wired.
- `main.go` - construct runtime runner and pass it into Skill Service.
- `store/db_test.go` - migration checks for runtime tables.
- Generated by sqlc: `store/skill_script_runtime.sql.go`, `store/models.go`, `store/querier.go`.
- `frontend/src/App.tsx` - add compact admin controls in Skill detail.
- `frontend/src/styles.css` - add script runtime controls and run history styles.
- `docs/roadmap.md` - link this plan and keep status as Planned until implementation finishes.
- `docs/skills/script-audit.md` - add a 3.6 implementation note once the runtime exists.
- `docs/adr/0005-skill-scripts-require-admin-approval.md` - update from "reserved for later" to the 3.6 permission model after implementation.

## Runtime Contract

The runner sends this JSON to script stdin:

```json
{
  "runtimeVersion": "skill-script-runtime/v1",
  "project": {
    "id": "project_123",
    "title": "Elysium",
    "language": "en"
  },
  "skill": {
    "id": "skill_123",
    "name": "character-names"
  },
  "script": {
    "fileId": "skillfile_123",
    "relativePath": "scripts/character-name.ts"
  },
  "inputs": {
    "contentIds": ["content_1"],
    "entrySections": [{"contentId": "content_2", "heading": "Names"}],
    "assets": [{"path": "data/cultures/anglo-given.json", "bodyText": "[\"Ada\"]"}],
    "arguments": {"mode": "suggest_names"}
  }
}
```

The script must return this JSON on stdout:

```json
{
  "kind": "report",
  "title": "Name collision report",
  "markdown": "No collisions found.",
  "proposals": [],
  "generatedData": {},
  "metadata": {"confidence": "medium"}
}
```

Allowed output kinds:

- `report` - visible analysis or diagnostic Markdown.
- `proposal` - reviewable changes, not applied.
- `draft` - generated prose or notes for copy/apply review.
- `generated_data` - structured JSON such as name candidates or scoring rows.

Rejected output:

- Requests to write files.
- Requests to directly modify content.
- Non-JSON stdout.
- Output exceeding configured byte limits.
- Any output kind outside the allowed list.

## Task 1: Schema And Queries

**Files:**
- Create: `migrations/00004_skill_script_runtime.sql`
- Create: `queries/skill_script_runtime.sql`
- Modify: `store/db_test.go`
- Generated: `store/skill_script_runtime.sql.go`
- Generated: `store/models.go`
- Generated: `store/querier.go`

- [x] **Step 1: Write the failing migration test**

Add this test to `store/db_test.go`:

```go
func TestSkillScriptRuntimeTablesExist(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	tables := []string{
		"skill_script_audits",
		"skill_script_approvals",
		"skill_script_runs",
	}
	for _, table := range tables {
		var name string
		err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?`, table).Scan(&name)
		if err != nil {
			t.Fatalf("table %s missing: %v", table, err)
		}
	}
	if !tableHasColumn(t, db, "skill_script_audits", "destructive_operations") {
		t.Fatal("skill_script_audits.destructive_operations missing")
	}
	if !tableHasColumn(t, db, "skill_script_approvals", "runtime_command") {
		t.Fatal("skill_script_approvals.runtime_command missing")
	}
	if !tableHasColumn(t, db, "skill_script_runs", "output_kind") {
		t.Fatal("skill_script_runs.output_kind missing")
	}
}
```

- [x] **Step 2: Run the failing migration test**

Run:

```bash
go test -tags sqlite_fts5 ./store -run TestSkillScriptRuntimeTablesExist -count=1
```

Expected: FAIL because `skill_script_audits` does not exist.

- [x] **Step 3: Add the runtime migration**

Create `migrations/00004_skill_script_runtime.sql`:

```sql
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
```

- [x] **Step 4: Add sqlc queries**

Create `queries/skill_script_runtime.sql`:

```sql
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
```

- [x] **Step 5: Generate store code**

Run:

```bash
sqlc generate
```

Expected: generated files include `store/skill_script_runtime.sql.go`, and existing generated files compile.

- [x] **Step 6: Run migration tests**

Run:

```bash
go test -tags sqlite_fts5 ./store -run TestSkillScriptRuntimeTablesExist -count=1
```

Expected: PASS.

- [x] **Step 7: Commit**

```bash
git add migrations/00004_skill_script_runtime.sql queries/skill_script_runtime.sql store store/db_test.go
git commit -m "feat: add skill script runtime schema"
```

## Task 2: Runtime Types And Runner

**Files:**
- Create: `skill/runtime/types.go`
- Create: `skill/runtime/runner.go`
- Create: `skill/runtime/runner_test.go`

- [x] **Step 1: Write failing runner tests**

Create `skill/runtime/runner_test.go`:

```go
package runtime

import (
	"context"
	"encoding/json"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestRunnerExecutesJSONEnvelope(t *testing.T) {
	runner := NewRunner()
	command := "cat"
	if runtime.GOOS == "windows" {
		t.Skip("cat-based runner test is Unix-only")
	}

	result, err := runner.Run(context.Background(), RunRequest{
		Command:        command,
		Timeout:        time.Second,
		MaxStdoutBytes: 4096,
		MaxStderrBytes: 1024,
		Input: Envelope{
			RuntimeVersion: RuntimeVersion,
			Inputs: EnvelopeInputs{
				Arguments: map[string]any{"mode": "echo"},
			},
		},
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Status != StatusSucceeded {
		t.Fatalf("status = %s, want %s; stderr=%s", result.Status, StatusSucceeded, result.StderrText)
	}
	if !json.Valid([]byte(result.StdoutText)) {
		t.Fatalf("stdout is not JSON: %q", result.StdoutText)
	}
}

func TestRunnerRejectsInvalidOutputKind(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("sh-based runner test is Unix-only")
	}
	runner := NewRunner()
	result, err := runner.Run(context.Background(), RunRequest{
		Command:        "sh -c 'printf %s \"{\\\"kind\\\":\\\"mutation\\\",\\\"title\\\":\\\"x\\\",\\\"markdown\\\":\\\"x\\\"}\"'",
		Timeout:        time.Second,
		MaxStdoutBytes: 4096,
		MaxStderrBytes: 1024,
		Input:          Envelope{RuntimeVersion: RuntimeVersion},
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Status != StatusRejected {
		t.Fatalf("status = %s, want %s", result.Status, StatusRejected)
	}
	if !strings.Contains(result.ErrorMessage, "unsupported output kind") {
		t.Fatalf("error = %q, want unsupported output kind", result.ErrorMessage)
	}
}

func TestRunnerTimesOut(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("sh-based runner test is Unix-only")
	}
	runner := NewRunner()
	result, err := runner.Run(context.Background(), RunRequest{
		Command:        "sh -c 'sleep 2'",
		Timeout:        20 * time.Millisecond,
		MaxStdoutBytes: 4096,
		MaxStderrBytes: 1024,
		Input:          Envelope{RuntimeVersion: RuntimeVersion},
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Status != StatusTimedOut {
		t.Fatalf("status = %s, want %s", result.Status, StatusTimedOut)
	}
}
```

- [x] **Step 2: Run the failing runner tests**

Run:

```bash
go test ./skill/runtime -count=1
```

Expected: FAIL because package `skill/runtime` does not exist.

- [x] **Step 3: Add runtime types**

Create `skill/runtime/types.go`:

```go
package runtime

import "time"

const RuntimeVersion = "skill-script-runtime/v1"

type Status string

const (
	StatusSucceeded Status = "succeeded"
	StatusFailed    Status = "failed"
	StatusTimedOut  Status = "timed_out"
	StatusRejected  Status = "rejected"
)

type OutputKind string

const (
	OutputKindReport        OutputKind = "report"
	OutputKindProposal      OutputKind = "proposal"
	OutputKindDraft         OutputKind = "draft"
	OutputKindGeneratedData OutputKind = "generated_data"
)

type Envelope struct {
	RuntimeVersion string         `json:"runtimeVersion"`
	Project        EnvelopeRef    `json:"project"`
	Skill          EnvelopeRef    `json:"skill"`
	Script         ScriptRef      `json:"script"`
	Inputs         EnvelopeInputs `json:"inputs"`
}

type EnvelopeRef struct {
	ID       string `json:"id"`
	Name     string `json:"name,omitempty"`
	Title    string `json:"title,omitempty"`
	Language string `json:"language,omitempty"`
}

type ScriptRef struct {
	FileID       string `json:"fileId"`
	RelativePath string `json:"relativePath"`
}

type EnvelopeInputs struct {
	ContentIDs    []string          `json:"contentIds,omitempty"`
	EntrySections []EntrySectionRef `json:"entrySections,omitempty"`
	Assets        []AssetInput      `json:"assets,omitempty"`
	Arguments     map[string]any    `json:"arguments,omitempty"`
}

type EntrySectionRef struct {
	ContentID string `json:"contentId"`
	Heading   string `json:"heading"`
}

type AssetInput struct {
	Path     string `json:"path"`
	BodyText string `json:"bodyText"`
}

type ScriptOutput struct {
	Kind          OutputKind     `json:"kind"`
	Title         string         `json:"title"`
	Markdown      string         `json:"markdown"`
	Proposals     []Proposal     `json:"proposals,omitempty"`
	GeneratedData map[string]any `json:"generatedData,omitempty"`
	Metadata      map[string]any `json:"metadata,omitempty"`
}

type Proposal struct {
	TargetType string `json:"targetType"`
	TargetID   string `json:"targetId"`
	Title      string `json:"title"`
	Markdown   string `json:"markdown"`
}

type RunRequest struct {
	Command        string
	Timeout        time.Duration
	MaxStdoutBytes int64
	MaxStderrBytes int64
	Input          Envelope
}

type RunResult struct {
	Status       Status
	Output       ScriptOutput
	StdoutText   string
	StderrText   string
	ExitCode     int
	Duration     time.Duration
	ErrorMessage string
}
```

- [x] **Step 4: Add runner implementation**

Create `skill/runtime/runner.go`:

```go
package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

var ErrInvalidRequest = errors.New("invalid script runtime request")

type Runner struct{}

func NewRunner() *Runner {
	return &Runner{}
}

func (r *Runner) Run(ctx context.Context, request RunRequest) (RunResult, error) {
	if strings.TrimSpace(request.Command) == "" {
		return RunResult{}, ErrInvalidRequest
	}
	if request.Timeout <= 0 {
		request.Timeout = 5 * time.Second
	}
	if request.MaxStdoutBytes <= 0 {
		request.MaxStdoutBytes = 64 * 1024
	}
	if request.MaxStderrBytes <= 0 {
		request.MaxStderrBytes = 16 * 1024
	}
	if request.Input.RuntimeVersion == "" {
		request.Input.RuntimeVersion = RuntimeVersion
	}

	inputBytes, err := json.Marshal(request.Input)
	if err != nil {
		return RunResult{}, fmt.Errorf("marshal script input: %w", err)
	}
	workdir, err := os.MkdirTemp("", "open-edda-skill-script-*")
	if err != nil {
		return RunResult{}, fmt.Errorf("create runtime temp dir: %w", err)
	}
	defer os.RemoveAll(workdir)

	ctx, cancel := context.WithTimeout(ctx, request.Timeout)
	defer cancel()

	commandName, commandArgs := shellCommand(request.Command)
	cmd := exec.CommandContext(ctx, commandName, commandArgs...)
	cmd.Dir = workdir
	cmd.Env = []string{
		"OPEN_EDDA_SKILL_RUNTIME=1",
		"NO_COLOR=1",
	}
	cmd.Stdin = bytes.NewReader(inputBytes)

	var stdout, stderr limitedBuffer
	stdout.limit = request.MaxStdoutBytes
	stderr.limit = request.MaxStderrBytes
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err = cmd.Run()
	duration := time.Since(start)

	result := RunResult{
		Status:     StatusSucceeded,
		StdoutText: stdout.String(),
		StderrText: stderr.String(),
		Duration:   duration,
	}
	if cmd.ProcessState != nil {
		result.ExitCode = cmd.ProcessState.ExitCode()
	}
	if ctx.Err() == context.DeadlineExceeded {
		result.Status = StatusTimedOut
		result.ErrorMessage = "script timed out"
		return result, nil
	}
	if err != nil {
		result.Status = StatusFailed
		result.ErrorMessage = err.Error()
		return result, nil
	}

	var output ScriptOutput
	if err := json.Unmarshal([]byte(result.StdoutText), &output); err != nil {
		result.Status = StatusRejected
		result.ErrorMessage = "script stdout must be valid JSON"
		return result, nil
	}
	if err := validateOutput(output); err != nil {
		result.Status = StatusRejected
		result.ErrorMessage = err.Error()
		return result, nil
	}
	result.Output = output
	return result, nil
}

func validateOutput(output ScriptOutput) error {
	switch output.Kind {
	case OutputKindReport, OutputKindProposal, OutputKindDraft, OutputKindGeneratedData:
	default:
		return fmt.Errorf("unsupported output kind %q", output.Kind)
	}
	if strings.TrimSpace(output.Title) == "" {
		return errors.New("script output title is required")
	}
	if strings.TrimSpace(output.Markdown) == "" && len(output.GeneratedData) == 0 && len(output.Proposals) == 0 {
		return errors.New("script output must include markdown, proposals, or generatedData")
	}
	for _, proposal := range output.Proposals {
		if strings.TrimSpace(proposal.TargetType) == "" || strings.TrimSpace(proposal.Title) == "" {
			return errors.New("proposal targetType and title are required")
		}
	}
	return nil
}

func shellCommand(command string) (string, []string) {
	if runtime.GOOS == "windows" {
		return "cmd", []string{"/C", command}
	}
	return "sh", []string{"-c", command}
}

type limitedBuffer struct {
	buf   bytes.Buffer
	limit int64
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	if b.limit <= 0 {
		return len(p), nil
	}
	remaining := b.limit - int64(b.buf.Len())
	if remaining <= 0 {
		return len(p), nil
	}
	if int64(len(p)) > remaining {
		p = p[:remaining]
	}
	_, _ = b.buf.Write(p)
	return len(p), nil
}

func (b *limitedBuffer) String() string {
	return b.buf.String()
}
```

- [x] **Step 5: Run runner tests**

Run:

```bash
go test ./skill/runtime -count=1
```

Expected: PASS.

- [x] **Step 6: Commit**

```bash
git add skill/runtime
git commit -m "feat: add skill script runner"
```

## Task 3: Skill Service Runtime API

**Files:**
- Modify: `skill/types.go`
- Modify: `skill/service.go`
- Create: `skill/script_runtime.go`
- Create: `skill/script_runtime_test.go`

- [x] **Step 1: Write failing service tests**

Create `skill/script_runtime_test.go` with:

```go
package skill

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	scriptruntime "git.inkyquill.net/inky/writer/skill/runtime"
	"git.inkyquill.net/inky/writer/project"
	"git.inkyquill.net/inky/writer/store"
)

func TestInstallCreatesDisabledScriptAudit(t *testing.T) {
	ctx := context.Background()
	db := openSkillTestDB(t)
	defer db.Close()
	projectService := project.NewService(db)
	storyProject := createSkillTestProject(t, ctx, projectService)

	service := NewService(db)
	installed, err := service.Install(ctx, InstallInput{
		ProjectID:  storyProject.ID,
		SourceType: SourceTypeUpload,
		Imported: ImportedSkill{
			Name:                 "name-helper",
			DisplayName:          "Name Helper",
			Description:          "Suggests names",
			InstructionsMarkdown: "Use the script only when enabled.",
			ScriptCount:          1,
			ScriptsDisabled:      true,
			Files: []ImportedSkillFile{{
				RelativePath:   "scripts/name.ts",
				Purpose:        FilePurposeScript,
				MediaType:      "text/plain; charset=utf-8",
				BodyText:       "console.log('{}')",
				Bytes:          int64(len("console.log('{}')")),
				ScriptDisabled: true,
			}},
		},
	})
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	audits, err := service.ListScriptAudits(ctx, storyProject.ID, installed.ID)
	if err != nil {
		t.Fatalf("ListScriptAudits() error = %v", err)
	}
	if len(audits) != 1 {
		t.Fatalf("audit count = %d, want 1", len(audits))
	}
	if audits[0].Recommendation != ScriptRecommendationDisabled {
		t.Fatalf("recommendation = %s, want disabled", audits[0].Recommendation)
	}
}

func TestRunScriptRequiresApproval(t *testing.T) {
	ctx := context.Background()
	service, storyProject, installed := setupScriptRuntimeSkill(t, ctx)

	_, err := service.RunScript(ctx, RunScriptInput{
		ProjectID: storyProject.ID,
		SkillID:   installed.ID,
		ScriptPath: "scripts/name.ts",
	})
	if err == nil {
		t.Fatal("RunScript() error = nil, want disabled error")
	}
}

func TestRunScriptPersistsReportRun(t *testing.T) {
	ctx := context.Background()
	service, storyProject, installed := setupScriptRuntimeSkill(t, ctx)
	audits, err := service.ListScriptAudits(ctx, storyProject.ID, installed.ID)
	if err != nil {
		t.Fatalf("ListScriptAudits() error = %v", err)
	}
	_, err = service.UpdateScriptApproval(ctx, UpdateScriptApprovalInput{
		ProjectID:       storyProject.ID,
		SkillID:         installed.ID,
		SkillFileID:     audits[0].SkillFileID,
		Enabled:         true,
		RuntimeCommand:  `printf '{"kind":"report","title":"Names","markdown":"Ada works."}'`,
		TimeoutMS:       1000,
		MaxStdoutBytes:  4096,
		MaxStderrBytes:  1024,
		ApprovedBy:      "test-admin",
	})
	if err != nil {
		t.Fatalf("UpdateScriptApproval() error = %v", err)
	}

	run, err := service.RunScript(ctx, RunScriptInput{
		ProjectID:  storyProject.ID,
		SkillID:    installed.ID,
		ScriptPath: "scripts/name.ts",
		Arguments: map[string]any{"mode": "suggest"},
	})
	if err != nil {
		t.Fatalf("RunScript() error = %v", err)
	}
	if run.Status != ScriptRunStatusSucceeded {
		t.Fatalf("status = %s, want succeeded", run.Status)
	}
	if run.OutputKind != "report" {
		t.Fatalf("output kind = %s, want report", run.OutputKind)
	}
	if run.DurationMS <= 0 && run.CreatedAt == "" {
		t.Fatalf("run did not persist timing/created metadata: %#v", run)
	}
}
```

- [x] **Step 2: Run failing service tests**

Run:

```bash
go test -tags sqlite_fts5 ./skill -run 'TestInstallCreatesDisabledScriptAudit|TestRunScriptRequiresApproval|TestRunScriptPersistsReportRun' -count=1
```

Expected: FAIL because runtime service types and methods do not exist.

- [x] **Step 3: Add service DTOs**

Append to `skill/types.go`:

```go
type ScriptRecommendation string

const (
	ScriptRecommendationDisabled          ScriptRecommendation = "disabled"
	ScriptRecommendationApproveWithLimits ScriptRecommendation = "approve_with_limits"
	ScriptRecommendationDefer             ScriptRecommendation = "defer"
)

type ScriptRunStatus string

const (
	ScriptRunStatusSucceeded ScriptRunStatus = "succeeded"
	ScriptRunStatusFailed    ScriptRunStatus = "failed"
	ScriptRunStatusTimedOut  ScriptRunStatus = "timed_out"
	ScriptRunStatusRejected  ScriptRunStatus = "rejected"
)

type ScriptAudit struct {
	ID                    string               `json:"id"`
	ProjectID             string               `json:"projectId"`
	SkillID               string               `json:"skillId"`
	SkillFileID           string               `json:"skillFileId"`
	RelativePath          string               `json:"relativePath"`
	Runtime               string               `json:"runtime"`
	DestructiveOperations bool                 `json:"destructiveOperations"`
	FilesystemAccess      string               `json:"filesystemAccess"`
	NetworkAccess         bool                 `json:"networkAccess"`
	ExternalDependencies  string               `json:"externalDependencies"`
	ExpectedInputsJSON    string               `json:"expectedInputsJson"`
	ExpectedOutputsJSON   string               `json:"expectedOutputsJson"`
	RiskNotes             string               `json:"riskNotes"`
	Recommendation        ScriptRecommendation `json:"recommendation"`
	AuditedAt             string               `json:"auditedAt"`
	Approval              *ScriptApproval      `json:"approval,omitempty"`
}

type ScriptApproval struct {
	ID                 string `json:"id"`
	ProjectID          string `json:"projectId"`
	SkillID            string `json:"skillId"`
	SkillFileID        string `json:"skillFileId"`
	AuditID            string `json:"auditId"`
	Enabled            bool   `json:"enabled"`
	RuntimeCommand     string `json:"runtimeCommand"`
	TimeoutMS          int64  `json:"timeoutMs"`
	MaxStdoutBytes     int64  `json:"maxStdoutBytes"`
	MaxStderrBytes     int64  `json:"maxStderrBytes"`
	AllowNetwork       bool   `json:"allowNetwork"`
	AllowProjectFiles  bool   `json:"allowProjectFiles"`
	ApprovedBy         string `json:"approvedBy"`
	ApprovedAt         string `json:"approvedAt"`
	UpdatedAt          string `json:"updatedAt"`
}

type UpdateScriptApprovalInput struct {
	ProjectID         string
	SkillID           string
	SkillFileID       string
	Enabled           bool
	RuntimeCommand    string
	TimeoutMS         int64
	MaxStdoutBytes    int64
	MaxStderrBytes    int64
	AllowNetwork      bool
	AllowProjectFiles bool
	ApprovedBy        string
}

type RunScriptInput struct {
	ProjectID       string
	SessionID       string
	ToolCallID      string
	SkillID         string
	SkillFileID     string
	ScriptPath      string
	ContentIDs      []string
	EntrySections   []ScriptEntrySectionInput
	AssetPaths      []string
	Arguments       map[string]any
}

type ScriptEntrySectionInput struct {
	ContentID string `json:"contentId"`
	Heading   string `json:"heading"`
}

type ScriptRun struct {
	ID           string          `json:"id"`
	ProjectID    string          `json:"projectId"`
	SessionID    string          `json:"sessionId,omitempty"`
	SkillID      string          `json:"skillId"`
	SkillFileID  string          `json:"skillFileId"`
	ApprovalID   string          `json:"approvalId"`
	ToolCallID   string          `json:"toolCallId"`
	Status       ScriptRunStatus `json:"status"`
	OutputKind   string          `json:"outputKind"`
	InputJSON    string          `json:"inputJson"`
	OutputJSON   string          `json:"outputJson"`
	StdoutText   string          `json:"stdoutText"`
	StderrText   string          `json:"stderrText"`
	ExitCode     int64           `json:"exitCode"`
	DurationMS   int64           `json:"durationMs"`
	ErrorMessage string          `json:"errorMessage"`
	CreatedAt    string          `json:"createdAt"`
}
```

- [x] **Step 4: Add runtime dependency to Skill Service**

Modify `skill/service.go`:

```go
type ScriptRunner interface {
	Run(ctx context.Context, request scriptruntime.RunRequest) (scriptruntime.RunResult, error)
}

type Service struct {
	db           *sql.DB
	queries      *store.Queries
	scriptRunner ScriptRunner
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db, queries: store.New(db)}
}

func (s *Service) SetScriptRunner(runner ScriptRunner) {
	s.scriptRunner = runner
}
```

Add import:

```go
scriptruntime "git.inkyquill.net/inky/writer/skill/runtime"
```

- [x] **Step 5: Create audits during install**

In `skill/service.go`, after creating each `skill_files` row, add:

```go
if file.Purpose == FilePurposeScript {
	if err := q.UpsertSkillScriptAudit(ctx, store.UpsertSkillScriptAuditParams{
		ID:                    newID("scriptaudit"),
		ProjectID:             input.ProjectID,
		SkillID:               skillID,
		SkillFileID:           fileID,
		RelativePath:          file.RelativePath,
		Runtime:               runtimeFromScriptPath(file.RelativePath),
		DestructiveOperations: 0,
		FilesystemAccess:      "temp_workspace",
		NetworkAccess:         0,
		ExternalDependencies:  "",
		ExpectedInputsJson:    "{}",
		ExpectedOutputsJson:   `{"allowedKinds":["report","proposal","draft","generated_data"]}`,
		RiskNotes:             "Imported script is disabled until an admin reviews and enables it.",
		Recommendation:        string(ScriptRecommendationDisabled),
		AuditedAt:             now,
	}); err != nil {
		return fmt.Errorf("create script audit %s: %w", file.RelativePath, err)
	}
}
```

Store `fileID := newID("skillfile")` before `CreateSkillFile` so the audit references the exact file row.

- [x] **Step 6: Add `skill/script_runtime.go`**

Create methods:

```go
func (s *Service) ListScriptAudits(ctx context.Context, projectID, skillID string) ([]ScriptAudit, error)
func (s *Service) UpdateScriptApproval(ctx context.Context, input UpdateScriptApprovalInput) (ScriptApproval, error)
func (s *Service) RunScript(ctx context.Context, input RunScriptInput) (ScriptRun, error)
func (s *Service) ListScriptRunsByProject(ctx context.Context, projectID string, limit int64) ([]ScriptRun, error)
func (s *Service) ListScriptRunsBySession(ctx context.Context, projectID, sessionID string, limit int64) ([]ScriptRun, error)
```

Implementation requirements:

- `ListScriptAudits` loads audits and best-effort matching approvals.
- `UpdateScriptApproval` rejects:
  - empty `ProjectID`, `SkillID`, or `SkillFileID`,
  - enabled approval with empty `RuntimeCommand`,
  - enabled approval where audit has `destructive_operations = 1`,
  - enabled approval where `AllowNetwork = true`,
  - enabled approval where `AllowProjectFiles = true`,
  - timeout outside `100..30000` ms.
- `RunScript` resolves script by `SkillFileID` or `ScriptPath`, requires enabled approval, builds a `scriptruntime.Envelope`, calls the runner, and persists `skill_script_runs`.
- `RunScript` never writes Story Text or Story Bible content.
- `RunScript` stores invalid/failed/timed-out results as run records instead of dropping them.

- [x] **Step 7: Run service tests**

Run:

```bash
go test -tags sqlite_fts5 ./skill -run 'TestInstallCreatesDisabledScriptAudit|TestRunScriptRequiresApproval|TestRunScriptPersistsReportRun' -count=1
```

Expected: PASS.

- [x] **Step 8: Commit**

```bash
git add skill/types.go skill/service.go skill/script_runtime.go skill/script_runtime_test.go
git commit -m "feat: add skill script runtime service"
```

## Task 4: HTTP Admin API

**Files:**
- Modify: `skill/http.go`
- Modify: `skill/http_test.go`

- [x] **Step 1: Write failing HTTP tests**

Add tests to `skill/http_test.go`:

```go
func TestScriptAuditRoutesListDisabledScripts(t *testing.T) {
	harness := newSkillHTTPHarness(t)
	skill := harness.importSkillArchive(t, "scripted.zip", map[string]string{
		"SKILL.md":        "---\nname: scripted\ndescription: scripted\n---\nUse cautiously.",
		"scripts/run.ts":  "console.log('{}')",
	})

	req := httptest.NewRequest(http.MethodGet, "/api/projects/"+harness.project.ID+"/skills/"+skill.ID+"/scripts", nil)
	rec := httptest.NewRecorder()
	harness.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	assertBodyContains(t, rec.Body.String(), `"relativePath":"scripts/run.ts"`)
	assertBodyContains(t, rec.Body.String(), `"enabled":false`)
}

func TestScriptApprovalRejectsNetwork(t *testing.T) {
	harness := newSkillHTTPHarness(t)
	skill := harness.importSkillArchive(t, "scripted.zip", map[string]string{
		"SKILL.md":       "---\nname: scripted\ndescription: scripted\n---\nUse cautiously.",
		"scripts/run.ts": "console.log('{}')",
	})
	audit := harness.firstScriptAudit(t, skill.ID)

	body := strings.NewReader(`{
		"enabled": true,
		"runtimeCommand": "deno run scripts/run.ts",
		"timeoutMs": 1000,
		"maxStdoutBytes": 4096,
		"maxStderrBytes": 1024,
		"allowNetwork": true,
		"allowProjectFiles": false
	}`)
	req := httptest.NewRequest(http.MethodPut, "/api/projects/"+harness.project.ID+"/skills/"+skill.ID+"/scripts/"+audit.SkillFileID+"/approval", body)
	rec := httptest.NewRecorder()
	harness.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body = %s", rec.Code, rec.Body.String())
	}
}
```

- [x] **Step 2: Run failing HTTP tests**

Run:

```bash
go test -tags sqlite_fts5 ./skill -run 'TestScriptAuditRoutesListDisabledScripts|TestScriptApprovalRejectsNetwork' -count=1
```

Expected: FAIL because routes do not exist.

- [x] **Step 3: Add HTTP routes**

In `skill.RegisterRoutes`, add:

```go
r.Get("/projects/{projectID}/skills/{skillID}/scripts", h.listScriptAudits)
r.Put("/projects/{projectID}/skills/{skillID}/scripts/{skillFileID}/approval", h.updateScriptApproval)
r.Get("/projects/{projectID}/skill-script-runs", h.listProjectScriptRuns)
r.Get("/projects/{projectID}/agent/sessions/{sessionID}/skill-script-runs", h.listSessionScriptRuns)
```

Add request type:

```go
type updateScriptApprovalRequest struct {
	Enabled           bool   `json:"enabled"`
	RuntimeCommand    string `json:"runtimeCommand"`
	TimeoutMS         int64  `json:"timeoutMs"`
	MaxStdoutBytes    int64  `json:"maxStdoutBytes"`
	MaxStderrBytes    int64  `json:"maxStderrBytes"`
	AllowNetwork      bool   `json:"allowNetwork"`
	AllowProjectFiles bool   `json:"allowProjectFiles"`
}
```

Add handlers that call the service methods and return JSON. Use existing `writeError` mapping so invalid approvals return HTTP 400.

- [x] **Step 4: Run HTTP tests**

Run:

```bash
go test -tags sqlite_fts5 ./skill -run 'TestScriptAuditRoutesListDisabledScripts|TestScriptApprovalRejectsNetwork' -count=1
```

Expected: PASS.

- [x] **Step 5: Commit**

```bash
git add skill/http.go skill/http_test.go
git commit -m "feat: expose skill script admin api"
```

## Task 5: Agent `skill_script` Tool

**Files:**
- Modify: `agent/service.go`
- Modify: `agent/tools.go`
- Modify: `agent/tools_test.go`
- Modify: `agent/prompt.go`
- Modify: `agent/prompt_test.go`

- [x] **Step 1: Write failing tool tests**

Add to `agent/tools_test.go`:

```go
func TestSkillScriptToolRequiresEnabledScript(t *testing.T) {
	ctx := context.Background()
	harness := newAgentToolHarness(t)
	session := harness.createSession(t, ctx)

	_, err := harness.service.ExecuteTool(ctx, ToolCallInput{
		ProjectID:     harness.project.ID,
		SessionID:     session.ID,
		ToolCallID:   "tool-script-1",
		ToolName:     "skill_script",
		ArgumentsJSON: `{"skillId":"missing","scriptPath":"scripts/x.ts"}`,
	})
	if err == nil {
		t.Fatal("ExecuteTool() error = nil, want invalid script")
	}
}

func TestSkillScriptToolStoresReportArtifact(t *testing.T) {
	ctx := context.Background()
	harness := newAgentToolHarness(t)
	skill := harness.installEnabledScriptSkill(t, ctx, `printf '{"kind":"report","title":"Script report","markdown":"Looks good."}'`)
	session := harness.createSessionWithSkills(t, ctx, []string{skill.ID})

	result, err := harness.service.ExecuteTool(ctx, ToolCallInput{
		ProjectID:   harness.project.ID,
		SessionID:   session.ID,
		ToolCallID: "tool-script-1",
		ToolName:   "skill_script",
		ArgumentsJSON: `{
			"skillId": "` + skill.ID + `",
			"scriptPath": "scripts/report.ts",
			"arguments": {"mode": "check"}
		}`,
	})
	if err != nil {
		t.Fatalf("ExecuteTool() error = %v", err)
	}
	assertContains(t, result.ModelVisibleMarkdown, "Script report")
	assertContains(t, result.ModelVisibleMarkdown, "Looks good.")
}
```

- [x] **Step 2: Run failing tool tests**

Run:

```bash
go test -tags sqlite_fts5 ./agent -run 'TestSkillScriptToolRequiresEnabledScript|TestSkillScriptToolStoresReportArtifact' -count=1
```

Expected: FAIL because `skill_script` is not registered.

- [x] **Step 3: Extend `SkillProvider`**

In `agent/service.go`, update the interface:

```go
type SkillProvider interface {
	List(ctx context.Context, projectID string) ([]skill.Skill, error)
	ListSessionSkills(ctx context.Context, projectID, sessionID string) ([]skill.Skill, error)
	SelectSessionSkills(ctx context.Context, input skill.SelectSessionSkillsInput) ([]skill.Skill, error)
	RenderForModel(ctx context.Context, input skill.RenderSkillInput) (string, skill.Skill, error)
	ListScriptAudits(ctx context.Context, projectID, skillID string) ([]skill.ScriptAudit, error)
	RunScript(ctx context.Context, input skill.RunScriptInput) (skill.ScriptRun, error)
}
```

- [x] **Step 4: Add tool definition**

In `agent.ContextToolDefinitions`, after `skill`, add:

```go
contextTool("skill_script", "Run one admin-enabled Edda skill helper script with explicit JSON inputs. Returns a reviewable report, proposal, draft, or generated data. It cannot directly mutate project content.", objectSchema(map[string]any{
	"skillId":    map[string]any{"type": "string"},
	"scriptPath": map[string]any{"type": "string"},
	"contentIds": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
	"entrySections": map[string]any{
		"type": "array",
		"items": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"contentId": map[string]any{"type": "string"},
				"heading":   map[string]any{"type": "string"},
			},
			"required": []string{"contentId", "heading"},
		},
	},
	"assetPaths": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
	"arguments":  map[string]any{"type": "object", "additionalProperties": true},
}, "skillId", "scriptPath"))
```

- [x] **Step 5: Execute the tool**

In `executeContextTool`, add:

```go
case "skill_script":
	if s.skillService == nil {
		return nil, false, nil, fmt.Errorf("skill service is not configured")
	}
	var args struct {
		SkillID       string                          `json:"skillId"`
		ScriptPath    string                          `json:"scriptPath"`
		ContentIDs     []string                        `json:"contentIds"`
		EntrySections  []skill.ScriptEntrySectionInput `json:"entrySections"`
		AssetPaths     []string                        `json:"assetPaths"`
		Arguments      map[string]any                  `json:"arguments"`
	}
	if err := decodeToolArgs(input.ArgumentsJSON, &args); err != nil {
		return nil, false, nil, err
	}
	run, err := s.skillService.RunScript(ctx, skill.RunScriptInput{
		ProjectID:     input.ProjectID,
		SessionID:     input.SessionID,
		ToolCallID:    input.ToolCallID,
		SkillID:       args.SkillID,
		ScriptPath:    args.ScriptPath,
		ContentIDs:     args.ContentIDs,
		EntrySections:  args.EntrySections,
		AssetPaths:     args.AssetPaths,
		Arguments:      args.Arguments,
	})
	if err != nil {
		return nil, false, nil, err
	}
	return map[string]any{
		"status": run.Status,
		"outputKind": run.OutputKind,
		"outputJson": json.RawMessage(run.OutputJSON),
		"errorMessage": run.ErrorMessage,
	}, false, map[string]any{
		"skillId": args.SkillID,
		"scriptPath": args.ScriptPath,
		"scriptRunId": run.ID,
		"status": string(run.Status),
	}, nil
```

- [x] **Step 6: Add special visible Markdown for script output**

Update `modelVisibleMarkdownPayload` or add a new branch so `skill_script` output renders:

```markdown
# Skill Script: Script report

Looks good.

_Output kind: report. This result is reviewable and did not modify project content._
```

The implementation should parse `outputJson` into `runtime.ScriptOutput`; if parsing fails, show the status and error message.

- [x] **Step 7: Update prompt script disclosure**

In `agent/prompt.go`, when selected skills are rendered, include enabled script summaries:

```markdown
Enabled runtime helpers are available only through the `skill_script` tool. They return reports, proposals, generated data, or drafts and cannot directly apply project changes.
```

Disabled scripts should still be described as unavailable.

- [x] **Step 8: Run agent tests**

Run:

```bash
go test -tags sqlite_fts5 ./agent -run 'TestSkillScriptToolRequiresEnabledScript|TestSkillScriptToolStoresReportArtifact|TestBuildActionPrompt' -count=1
```

Expected: PASS.

- [x] **Step 9: Commit**

```bash
git add agent/service.go agent/tools.go agent/tools_test.go agent/prompt.go agent/prompt_test.go
git commit -m "feat: expose enabled skill scripts to agent tools"
```

## Task 6: Frontend Admin Controls

**Files:**
- Create: `frontend/src/scriptRuntimeTypes.ts`
- Create: `frontend/src/scriptRuntimeApi.ts`
- Modify: `frontend/src/App.tsx`
- Modify: `frontend/src/styles.css`

- [x] **Step 1: Add TypeScript types**

Create `frontend/src/scriptRuntimeTypes.ts`:

```ts
export type ScriptRecommendation = "disabled" | "approve_with_limits" | "defer";
export type ScriptRunStatus = "succeeded" | "failed" | "timed_out" | "rejected";

export type ScriptApproval = {
  id: string;
  projectId: string;
  skillId: string;
  skillFileId: string;
  auditId: string;
  enabled: boolean;
  runtimeCommand: string;
  timeoutMs: number;
  maxStdoutBytes: number;
  maxStderrBytes: number;
  allowNetwork: boolean;
  allowProjectFiles: boolean;
  approvedBy: string;
  approvedAt: string;
  updatedAt: string;
};

export type ScriptAudit = {
  id: string;
  projectId: string;
  skillId: string;
  skillFileId: string;
  relativePath: string;
  runtime: string;
  destructiveOperations: boolean;
  filesystemAccess: string;
  networkAccess: boolean;
  externalDependencies: string;
  expectedInputsJson: string;
  expectedOutputsJson: string;
  riskNotes: string;
  recommendation: ScriptRecommendation;
  auditedAt: string;
  approval?: ScriptApproval;
};

export type ScriptRun = {
  id: string;
  projectId: string;
  sessionId?: string;
  skillId: string;
  skillFileId: string;
  approvalId: string;
  toolCallId: string;
  status: ScriptRunStatus;
  outputKind: string;
  inputJson: string;
  outputJson: string;
  stderrText: string;
  errorMessage: string;
  durationMs: number;
  createdAt: string;
};
```

- [x] **Step 2: Add API functions**

Create `frontend/src/scriptRuntimeApi.ts`:

```ts
import { requestJSON } from "./api";
import type { ScriptApproval, ScriptAudit, ScriptRun } from "./scriptRuntimeTypes";

function projectPath(projectId: string, suffix: string): string {
  return `/api/projects/${encodeURIComponent(projectId)}/${suffix}`;
}

export function listScriptAudits(projectId: string, skillId: string, signal?: AbortSignal): Promise<ScriptAudit[]> {
  return requestJSON<ScriptAudit[]>(projectPath(projectId, `skills/${encodeURIComponent(skillId)}/scripts`), { signal });
}

export function updateScriptApproval(
  projectId: string,
  skillId: string,
  skillFileId: string,
  body: {
    enabled: boolean;
    runtimeCommand: string;
    timeoutMs: number;
    maxStdoutBytes: number;
    maxStderrBytes: number;
    allowNetwork: boolean;
    allowProjectFiles: boolean;
  },
): Promise<ScriptApproval> {
  return requestJSON<ScriptApproval>(
    projectPath(projectId, `skills/${encodeURIComponent(skillId)}/scripts/${encodeURIComponent(skillFileId)}/approval`),
    {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    },
  );
}

export function listProjectScriptRuns(projectId: string, signal?: AbortSignal): Promise<ScriptRun[]> {
  return requestJSON<ScriptRun[]>(projectPath(projectId, "skill-script-runs"), { signal });
}
```

- [x] **Step 3: Add Skill detail runtime panel**

In `frontend/src/App.tsx`, update `SkillDetail` so when `skill.scriptCount > 0` it loads `listScriptAudits(projectId, skill.id)` and renders:

```tsx
<section className="script-runtime-panel" aria-label="Skill scripts">
  <h3>Runtime helpers</h3>
  <p>Scripts are disabled until an admin enables them. Enabled helpers return reviewable reports, proposals, data, or drafts.</p>
  {audits.map((audit) => (
    <article className="script-runtime-row" key={audit.skillFileId}>
      <strong>{audit.relativePath}</strong>
      <span>{audit.recommendation}</span>
      <span>{audit.approval?.enabled ? "Enabled" : "Disabled"}</span>
      <label>
        Runtime command
        <input value={draftCommands[audit.skillFileId] ?? audit.approval?.runtimeCommand ?? ""} onChange={...} />
      </label>
      <button type="button" onClick={() => saveApproval(audit, true)}>Enable</button>
      <button type="button" onClick={() => saveApproval(audit, false)}>Disable</button>
    </article>
  ))}
</section>
```

Implementation details:

- Default timeout: `5000`.
- Default max stdout: `65536`.
- Default max stderr: `16384`.
- Always send `allowNetwork: false`.
- Always send `allowProjectFiles: false`.
- Disable the Enable button when the command field is empty.
- Show API errors in the existing skill error area.

- [x] **Step 4: Add compact styles**

In `frontend/src/styles.css`, add:

```css
.script-runtime-panel {
  display: grid;
  gap: 0.75rem;
  border-top: 1px solid var(--border-muted);
  padding-top: 1rem;
}

.script-runtime-row {
  display: grid;
  gap: 0.5rem;
  border: 1px solid var(--border-muted);
  border-radius: 0.5rem;
  padding: 0.75rem;
}

.script-runtime-row label {
  display: grid;
  gap: 0.25rem;
  font-size: 0.85rem;
}

.script-runtime-row input {
  min-width: 0;
  width: 100%;
}
```

Use existing CSS variables if these exact names exist; otherwise use the nearest existing border/text variables in `styles.css`.

- [x] **Step 5: Run frontend checks**

Run:

```bash
bun run typecheck
bun run build
```

Expected: both PASS.

- [x] **Step 6: Commit**

```bash
git add frontend/src/App.tsx frontend/src/styles.css frontend/src/scriptRuntimeTypes.ts frontend/src/scriptRuntimeApi.ts
git commit -m "feat: add skill script admin controls"
```

## Task 7: Runtime Documentation And Roadmap

**Files:**
- Modify: `docs/roadmap.md`
- Modify: `docs/skills/script-audit.md`
- Modify: `docs/adr/0005-skill-scripts-require-admin-approval.md`
- Modify: `docs/superpowers/plans/2026-06-14-open-edda-skill-script-runtime.md`

- [x] **Step 1: Update roadmap status text**

In `docs/roadmap.md`, update Milestone 3.6 tracking plan to:

```markdown
| 3.6 | Skill Script Runtime | Planned | `docs/superpowers/plans/2026-06-14-open-edda-skill-script-runtime.md` |
```

After implementation finishes, the finisher should change the status from `Planned` to `Implemented`.

- [x] **Step 2: Update ADR 0005**

Replace the ADR body with:

```markdown
# Skill Scripts Require Admin Approval

Open Edda imports skill instructions, routing metadata, templates, data files, and script files as usable agent context, but imported scripts are disabled by default. Script execution requires an explicit admin approval record for the individual script file.

Approved scripts run through the Skill Script Runtime. Open Edda supplies database-backed JSON inputs and an empty temporary working directory, then accepts only structured JSON outputs: reports, proposals, generated data, or drafts. Scripts must not directly mutate Story Text, Story Bible Entries, Entry Sections, Project Notes, Attached Notes, or project structure.

The first runtime rejects network-enabled and project-file-enabled approvals. If a script needs those capabilities, it remains disabled until a later runtime policy can provide a stronger sandbox and review model.
```

- [x] **Step 3: Add script audit implementation note**

At the top of `docs/skills/script-audit.md`, after the summary, add:

```markdown
Milestone 3.6 implements the runtime scaffolding for `retained-for-runtime` and future approved helpers. It does not automatically revive deferred skills. `reverse-outliner`, `story-zoom`, and `world-fates` still need product-specific adapters before they should become installed skills.
```

- [x] **Step 4: Run documentation scan**

Run:

```bash
rg -n "Skill Script Runtime|skill_script|admin approval|directly mutate" docs/roadmap.md docs/skills/script-audit.md docs/adr/0005-skill-scripts-require-admin-approval.md
```

Expected: output shows the runtime link, `skill_script` where relevant, and the no-direct-mutation policy.

- [x] **Step 5: Commit**

```bash
git add docs/roadmap.md docs/skills/script-audit.md docs/adr/0005-skill-scripts-require-admin-approval.md docs/superpowers/plans/2026-06-14-open-edda-skill-script-runtime.md
git commit -m "docs: document skill script runtime policy"
```

## Task 8: End-To-End Verification

**Files:**
- No new files.
- Verify all modified files from Tasks 1-7.

- [ ] **Step 1: Run backend tests**

Run:

```bash
go test -tags sqlite_fts5 ./...
```

Expected: PASS for all packages.

- [ ] **Step 2: Run frontend checks**

Run:

```bash
cd frontend
bun run typecheck
bun run build
```

Expected: both PASS.

- [ ] **Step 3: Run focused grep checks**

Run:

```bash
rg -n "directly mutate|allowProjectFiles.*true|allowNetwork.*true|skill_script|Runtime helpers" skill agent frontend/src docs
```

Expected:

- `directly mutate` appears only in policy/disclosure strings.
- `allowProjectFiles` and `allowNetwork` are present in DTOs/UI but service validation rejects `true`.
- `skill_script` appears in tool definition, execution, tests, and docs.
- `Runtime helpers` appears in the Skill detail UI.

- [ ] **Step 4: Manual browser smoke test**

Run:

```bash
cd frontend
bun run build
cd ..
OPEN_EDDA_DB_PATH=/tmp/open-edda-script-runtime-smoke.db OPEN_EDDA_STATIC_PATH=frontend/dist go run .
```

Expected:

- Server starts with `starting open edda`.
- Skill browser still loads.
- A script-bearing skill detail shows Runtime helpers.
- Enabling a script with an empty command is blocked.
- Enabling a script with `allowNetwork` or `allowProjectFiles` cannot be done through UI defaults.

- [ ] **Step 5: Commit final fixes if any**

If Step 1-4 required fixes:

```bash
git add .
git commit -m "fix: polish skill script runtime"
```

If no fixes were required, do not create an empty commit.

## Self-Review

Spec coverage:

- Audit records: Task 1 and Task 3.
- Admin enable/disable: Task 3, Task 4, Task 6.
- Database-backed input scaffolding: Task 2 and Task 3 runtime envelope.
- No direct mutation: Product Rules, Task 2 output validation, Task 3 service validation, Task 5 tool disclosure.
- Disabled/missing scripts degrade clearly: Task 3 service errors, Task 5 tool errors, Task 6 UI disclosure.
- Run history and artifacts: Task 1 run table, Task 3 run persistence, Task 5 tool artifact path, Task 6 run API.
- Docs/roadmap: Task 7.

Placeholder scan:

- Placeholder scan completed: no red-flag placeholder instructions remain.
- Each implementation task includes concrete file paths, test commands, expected results, and code shape.

Type consistency:

- Runtime package uses `RunRequest`, `RunResult`, `Envelope`, and `ScriptOutput`.
- Skill service uses `RunScriptInput`, `ScriptRun`, `ScriptAudit`, and `ScriptApproval`.
- Agent tool name is consistently `skill_script`.
- Frontend types match the Go JSON names.
