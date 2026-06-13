# Writer Agent Core Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build Milestone 2: OpenAI-compatible model configuration, agent sessions, prompt assembly, project context tools, Continuation, Rewrite, Read and Check, structured writes, conflict handling, activity trails, and prompt records.

**Architecture:** The Go backend owns provider calls, prompt assembly, tool execution, structured writes, activity events, and prompt records. The React frontend adds a small Agent panel and settings surface on top of the existing Project Dashboard and Writing Workspace shell. Model calls are non-streaming in this milestone; generated text is committed only after a complete provider response and a revision check.

**Tech Stack:** Go 1.26, chi, sqlc, goose, SQLite FTS5, OpenAI-compatible `/v1/chat/completions`, React, TypeScript, Bun, Vite.

---

## Scope

This plan implements Agent Core only. It does not implement Skill Core, script execution, Galley Editor integration, embeddings, collaboration, or local sync.

Agent Core must support:

- OpenAI-compatible provider configuration and model variants.
- Chat sessions per Story Project.
- Prompt assembly from project settings, writing briefs, selected chapter/selection, and tool-accessible project context.
- Structured context tools: project map, content search, read content, read story bible section, read revisions metadata.
- Structured write tools: append to chapter, insert into chapter, replace selected range, update Story Bible Entry body, update Entry Section body.
- Continuation, Rewrite, and Read and Check quick actions.
- Explicit per-action apply mode: `preview` or `direct_apply`.
- Preview candidates that can be accepted later with the original expected revision.
- Revision conflict errors when content changed since the agent read it.
- Agent-created revision labels linking action kind, model variant, and Agent Session.
- Activity Trails as compact events.
- Prompt Records for raw assembled request/response debugging without provider secrets.
- Prompt Record retention controls.

## File Map

Create:

- `migrations/00002_agent_core.sql` - provider, model variant, session, message, activity, prompt record, prompt profile, candidate tables, and agent revision label columns.
- `queries/agent_core.sql` - sqlc queries for Agent Core tables.
- `agent/types.go` - domain types for provider config, model variants, sessions, messages, actions, context bundles, structured writes, activity events, and prompt records.
- `agent/service.go` - orchestration service for sessions, chat turns, quick actions, prompt assembly, tool execution, and structured writes.
- `agent/provider.go` - OpenAI-compatible provider client interface and HTTP implementation.
- `agent/prompt.go` - prompt profile and action prompt assembly.
- `agent/tools.go` - tool definitions and execution over `project.Service`.
- `agent/http.go` - HTTP routes for provider settings, sessions, messages, quick actions, activity events, and prompt records.
- `agent/service_test.go` - service tests using a fake provider.
- `agent/http_test.go` - HTTP tests using a fake provider.
- `frontend/src/agentTypes.ts` - frontend Agent Core types.
- `frontend/src/agentApi.ts` - Agent Core API functions.

Modify:

- `migrations/00001_project_core.sql` - no direct edits unless a missing constraint is discovered; prefer migration 00002.
- `queries/project_core.sql` - add only missing project/content lookup queries needed by Agent Core.
- `project/service.go` - expose narrow read/write helpers for project maps, metadata-filtered search, revisions, structured writes, and attached-note creation.
- `project/types.go` - add structured-write input/result types if they belong with content revisions.
- `project/http.go` - no Agent routes; keep Agent routes in `agent/http.go`.
- `app/app.go` - add `AgentService` dependency and register `/api` Agent routes.
- `app/app_test.go` - add route-level tests for Agent Core registration.
- `main.go` - wire Agent Core service and provider HTTP client.
- `frontend/src/App.tsx` - add Agent side panel and quick action controls.
- `frontend/src/api.ts` - leave project API functions here; keep Agent API in `agentApi.ts`.
- `frontend/src/types.ts` - leave project content types here; keep Agent types in `agentTypes.ts`.
- `frontend/src/styles.css` - add focused Agent panel styles.

## Design Decisions

- Provider calls are synchronous HTTP requests in this milestone. Streaming can wrap the same provider interface later.
- Provider secrets are stored in SQLite for the first self-hosted version. They must never be returned by API responses or written into Prompt Records.
- A model variant is the author-facing unit used by actions. It references a provider config and contains model name plus generation defaults.
- Prompt Records store JSON request and response bodies for debugging. They include provider name, model name, action kind, and session ID, but not API keys.
- Activity Trails are normalized, compact events. The UI can collapse them into an `actions: N` pill later; this milestone exposes the data.
- Direct Apply uses existing revision creation. Preview mode stores a `generation_candidates` row and does not change content until the author accepts it.
- Preview accept is a backend action, not just a UI action. It reuses the candidate's stored operation and original `expectedRevision`, then either commits one structured write or returns a conflict.
- Read and Check creates an Agent Session assistant message and an Attached Note linked to the checked chapter/selection.
- Tool calls are executed server-side. The frontend never receives provider credentials.
- Skill retrieval is intentionally limited to the future Skill Core tables. Agent Core reserves `skill_id` fields in revision/activity metadata but does not load skill instructions or assets yet.

---

## Task 1: Agent Core Schema And Queries

**Files:**
- Create: `migrations/00002_agent_core.sql`
- Create: `queries/agent_core.sql`
- Generated: `store/agent_core.sql.go`
- Generated: `store/models.go`
- Generated: `store/querier.go`
- Test: `store/db_test.go`

- [ ] **Step 1: Add migration test expectations**

Extend `store/db_test.go` with a test that opens a migrated in-memory database and verifies the new tables exist:

```go
func TestAgentCoreTablesExist(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	tables := []string{
		"provider_configs",
		"model_variants",
		"prompt_profiles",
		"agent_sessions",
		"agent_messages",
		"activity_events",
		"prompt_records",
		"generation_candidates",
	}
	for _, table := range tables {
		var name string
		err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?`, table).Scan(&name)
		if err != nil {
			t.Fatalf("table %s missing: %v", table, err)
		}
	}
}
```

- [ ] **Step 2: Run the test and verify it fails**

Run:

```bash
go test -tags sqlite_fts5 ./store -run TestAgentCoreTablesExist -count=1
```

Expected: FAIL because the new tables do not exist.

- [ ] **Step 3: Add migration**

Create `migrations/00002_agent_core.sql`:

```sql
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
  status TEXT NOT NULL CHECK(status IN ('pending', 'accepted', 'rejected', 'conflict')),
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

-- +goose Down
DROP INDEX IF EXISTS idx_prompt_records_project_created;
DROP INDEX IF EXISTS idx_generation_candidates_project_session;
DROP INDEX IF EXISTS idx_activity_events_project_created;
DROP INDEX IF EXISTS idx_agent_messages_session_created;
DROP INDEX IF EXISTS idx_agent_sessions_project_updated;
DROP INDEX IF EXISTS idx_model_variants_provider_name;
DROP TABLE IF EXISTS prompt_records;
DROP TABLE IF EXISTS generation_candidates;
DROP TABLE IF EXISTS activity_events;
DROP TABLE IF EXISTS agent_messages;
DROP TABLE IF EXISTS agent_sessions;
DROP TABLE IF EXISTS prompt_profiles;
DROP TABLE IF EXISTS model_variants;
DROP TABLE IF EXISTS provider_configs;
```

- [ ] **Step 4: Add sqlc queries**

Create `queries/agent_core.sql` with CRUD queries for:

- Provider configs: create, list by author, get by ID, update secret/base URL, delete.
- Model variants: create, list by provider, list by author, get by ID, update defaults, delete.
- Prompt profile: upsert and get by project.
- Agent sessions: create, list by project, get by project/session, touch updated time.
- Agent messages: create and list by session.
- Activity events: create and list by project with limit.
- Prompt records: create and list by project with limit.
- Generation candidates: create, get by project/candidate, mark accepted/rejected/conflict, list by session.

Use explicit sqlc names:

```sql
-- name: CreateAgentSession :exec
INSERT INTO agent_sessions (
  id, project_id, title, action_kind, model_variant_id, apply_mode, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: ListAgentSessions :many
SELECT * FROM agent_sessions
WHERE project_id = ?
ORDER BY updated_at DESC
LIMIT ?;

-- name: CreateAgentMessage :exec
INSERT INTO agent_messages (id, session_id, role, body_markdown, metadata_json, created_at)
VALUES (?, ?, ?, ?, ?, ?);

-- name: ListAgentMessages :many
SELECT * FROM agent_messages
WHERE session_id = ?
ORDER BY created_at ASC;
```

Add these explicit query names in the same style: `CreateProviderConfig`, `ListProviderConfigs`, `GetProviderConfig`, `UpdateProviderConfig`, `DeleteProviderConfig`, `CreateModelVariant`, `ListModelVariantsByProvider`, `ListModelVariantsByAuthor`, `GetModelVariant`, `UpdateModelVariant`, `DeleteModelVariant`, `UpsertPromptProfile`, `GetPromptProfile`, `CreateAgentSession`, `ListAgentSessions`, `GetAgentSession`, `TouchAgentSession`, `CreateAgentMessage`, `ListAgentMessages`, `CreateActivityEvent`, `ListActivityEvents`, `CreatePromptRecord`, `ListPromptRecords`, `DeleteExpiredPromptRecords`, `CreateGenerationCandidate`, `GetGenerationCandidate`, `ListGenerationCandidates`, and `UpdateGenerationCandidateStatus`.

- [ ] **Step 5: Generate sqlc code**

Run:

```bash
go run github.com/sqlc-dev/sqlc/cmd/sqlc generate
```

Expected: generated files update without errors.

- [ ] **Step 6: Verify**

Run:

```bash
go test -tags sqlite_fts5 ./store
go test -tags sqlite_fts5 ./...
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add migrations/00002_agent_core.sql queries/agent_core.sql store/db_test.go
git add store/agent_core.sql.go store/models.go store/querier.go
git commit -m "feat: add agent core schema"
```

---

## Task 2: Agent Domain Service Skeleton

**Files:**
- Create: `agent/types.go`
- Create: `agent/service.go`
- Test: `agent/service_test.go`

- [ ] **Step 1: Add service tests for sessions and messages**

Create tests that:

- Create a session for a project.
- Append a user message.
- Append an assistant message.
- List the session transcript in creation order.

Use a helper that creates an in-memory migrated DB, a `project.Service`, one Story Project, and an `agent.Service`.

- [ ] **Step 2: Run tests and verify they fail**

Run:

```bash
go test -tags sqlite_fts5 ./agent -run TestCreateSessionAndMessages -count=1
```

Expected: FAIL because package `agent` does not exist.

- [ ] **Step 3: Define domain types**

Create `agent/types.go` with:

```go
package agent

type ActionKind string

const (
	ActionKindChat         ActionKind = "chat"
	ActionKindContinuation ActionKind = "continuation"
	ActionKindRewrite      ActionKind = "rewrite"
	ActionKindReadCheck    ActionKind = "read_check"
)

type ApplyMode string

const (
	ApplyModePreview     ApplyMode = "preview"
	ApplyModeDirectApply ApplyMode = "direct_apply"
)

type MessageRole string

const (
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
	MessageRoleTool      MessageRole = "tool"
	MessageRoleSystem    MessageRole = "system"
)

type Session struct {
	ID             string     `json:"id"`
	ProjectID      string     `json:"projectId"`
	Title          string     `json:"title"`
	ActionKind     ActionKind `json:"actionKind"`
	ModelVariantID string     `json:"modelVariantId"`
	ApplyMode      ApplyMode  `json:"applyMode"`
	CreatedAt      string     `json:"createdAt"`
	UpdatedAt      string     `json:"updatedAt"`
}

type Message struct {
	ID           string      `json:"id"`
	SessionID    string      `json:"sessionId"`
	Role         MessageRole `json:"role"`
	BodyMarkdown string      `json:"bodyMarkdown"`
	MetadataJSON string      `json:"metadataJson"`
	CreatedAt    string      `json:"createdAt"`
}
```

Add input/result types for session creation, message appends, provider config, model variant, prompt profile, quick actions, structured writes, activity events, and prompt records.

- [ ] **Step 4: Implement service skeleton**

Create `agent/service.go` with `NewService(db *sql.DB, projectService *project.Service, provider Provider) *Service`, session CRUD methods, message append/list methods, and conversion helpers from sqlc models.

- [ ] **Step 5: Verify**

Run:

```bash
go test -tags sqlite_fts5 ./agent
go test -tags sqlite_fts5 ./...
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add agent/types.go agent/service.go agent/service_test.go
git commit -m "feat: add agent session service"
```

---

## Task 3: Provider Configuration And OpenAI-Compatible Client

**Files:**
- Create: `agent/provider.go`
- Modify: `agent/service.go`
- Test: `agent/provider_test.go`
- Test: `agent/service_test.go`

- [ ] **Step 1: Add provider config service tests**

Test that the service can:

- Save a provider config with `name`, `baseURL`, and API key.
- Return provider configs without the API key.
- Create two model variants for one provider.
- Select a model variant by ID for an action.

- [ ] **Step 2: Add provider client tests**

Use `httptest.Server` to assert the OpenAI-compatible client sends:

- `POST /v1/chat/completions`
- `Authorization: Bearer <api-key>`
- JSON body containing `model`, `messages`, `tools`, `tool_choice`, `temperature`, and `max_tokens`.

Return a minimal assistant response and assert it is parsed.

- [ ] **Step 3: Implement Provider interface**

Create:

```go
type Provider interface {
	Complete(ctx context.Context, request CompletionRequest) (CompletionResponse, error)
}
```

Define `CompletionRequest`, `CompletionMessage`, `CompletionTool`, `CompletionToolCall`, and `CompletionResponse` using OpenAI-compatible names internally but project-owned types externally.

- [ ] **Step 4: Implement HTTP client**

Implement `OpenAICompatibleClient` with base URL normalization, context-aware HTTP requests, bearer auth, non-2xx error handling, and JSON decoding.

- [ ] **Step 5: Implement config service methods**

Provider config API responses must never include `api_key_encrypted`. For this milestone, store the key as plaintext in `api_key_encrypted` and add a short code comment: `Encryption is handled in the later auth/security milestone.`

- [ ] **Step 6: Verify**

Run:

```bash
go test -tags sqlite_fts5 ./agent
go test -tags sqlite_fts5 ./...
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add agent/provider.go agent/provider_test.go agent/service.go agent/service_test.go agent/types.go
git commit -m "feat: add OpenAI-compatible provider config"
```

---

## Task 4: Prompt Profiles And Prompt Assembly

**Files:**
- Create: `agent/prompt.go`
- Modify: `agent/service.go`
- Test: `agent/prompt_test.go`

- [ ] **Step 1: Add prompt profile tests**

Test that a project prompt profile stores and retrieves:

- genre
- tense
- point of view
- voice
- freeform writing instructions
- prompt record retention days

- [ ] **Step 2: Add prompt assembly tests**

Create a project with:

- one project-wide Writing Brief containing synopsis,
- one chapter-specific Writing Brief whose metadata JSON contains `{"contentItemId":"<chapter id>","scope":"chapter"}`,
- one Chapter,
- one Story Bible Entry.

Assert assembled Continuation prompt includes:

- task instruction,
- genre/tense/POV/voice profile,
- project-wide Writing Brief text,
- chapter-specific Writing Brief text after the project-wide brief,
- target chapter title,
- cursor location summary,
- instruction to use tools for additional context rather than assuming the whole project is in prompt.

- [ ] **Step 3: Implement prompt profile service methods**

Add `GetPromptProfile`, `UpsertPromptProfile`, and conversion helpers.

- [ ] **Step 4: Implement prompt assembly**

`BuildActionPrompt(ctx, input BuildPromptInput) (PromptBundle, error)` returns:

- system message,
- developer/context message,
- user action message,
- tool definitions,
- prompt metadata JSON.

Prompt assembly must load layered Writing Briefs in deterministic order:

1. Prompt profile fields.
2. Project-wide `writing_brief` content items whose metadata JSON has no `contentItemId`.
3. Target-content `writing_brief` items whose metadata JSON has `contentItemId` equal to the target Chapter ID.
4. User guidance for the current action.

The system prompt must include:

```text
You are a fiction writing assistant working inside Writer. Preserve the author's intent, respect established project facts, and use available tools to inspect project context before making claims. Do not invent durable worldbuilding facts unless the author asks you to brainstorm.
```

- [ ] **Step 5: Verify**

Run:

```bash
go test -tags sqlite_fts5 ./agent -run 'TestPrompt'
go test -tags sqlite_fts5 ./...
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add agent/prompt.go agent/prompt_test.go agent/service.go agent/types.go
git commit -m "feat: add agent prompt assembly"
```

---

## Task 5: Context Tools And Activity Events

**Files:**
- Create: `agent/tools.go`
- Modify: `agent/service.go`
- Test: `agent/tools_test.go`

- [ ] **Step 1: Add tool execution tests**

Test these tool names:

- `project_map`
- `search_content`
- `read_content`
- `read_chapter`
- `read_story_bible_entry`
- `read_entry_section`
- `list_revisions`

Each tool must create an activity event containing event type, summary, and metadata.

`search_content` must support metadata filters in addition to FTS query text. The test must create content with metadata such as `{"type":"character","status":"draft","tags":["alchemy"]}` and assert a search with filters returns matching items only.

- [ ] **Step 2: Add required project service helpers**

Add these narrow helpers to `project.Service`:

- `ProjectMap(ctx, projectID string) (ProjectMap, error)`
- `SearchContent(ctx, input SearchContentInput) ([]ContentItem, error)`
- `ListRevisions(ctx, projectID, contentID string) ([]Revision, error)`
- `CreateAttachedNote(ctx, input CreateAttachedNoteInput) (AttachedNote, error)`

`SearchContentInput` must include `ProjectID`, `Query`, `Kind`, `MetadataFilters map[string]string`, `Tags []string`, and `Limit`.

- [ ] **Step 3: Implement tool definitions**

Each tool definition must include a JSON schema matching the arguments accepted by execution. Keep schemas small and explicit.

The `search_content` schema must expose `query`, optional `kind`, optional `metadataFilters`, optional `tags`, and `limit`. `read_chapter` and `read_story_bible_entry` can delegate to the same underlying content read helper, but they must exist as separate tool names so prompts can choose them explicitly.

- [ ] **Step 4: Implement tool execution**

`ExecuteTool(ctx, input ToolCallInput) (ToolResult, error)` validates project/session scope, runs the matching project service method, records an activity event, and returns JSON Markdown-friendly output.

- [ ] **Step 5: Verify**

Run:

```bash
go test -tags sqlite_fts5 ./agent -run 'Test.*Tool'
go test -tags sqlite_fts5 ./project
go test -tags sqlite_fts5 ./...
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add agent/tools.go agent/tools_test.go agent/service.go agent/types.go project/service.go project/types.go project/service_test.go queries/project_core.sql store
git commit -m "feat: add agent context tools"
```

---

## Task 6: Structured Writes And Conflict Handling

**Files:**
- Modify: `agent/tools.go`
- Modify: `project/service.go`
- Modify: `project/types.go`
- Test: `agent/tools_test.go`
- Test: `project/service_test.go`

- [ ] **Step 1: Add structured write tests**

Test:

- Append to Chapter creates a new content revision.
- Insert into Chapter creates a new content revision.
- Replace selected range creates a new content revision.
- Update Story Bible Entry body creates a new content revision.
- Update Entry Section body updates the section and records an activity event.
- A stale expected revision returns `project.ErrConflict`.
- Agent-created content revisions persist `agent_session_id`, `action_kind`, `model_variant_id`, and empty `skill_id`.

- [ ] **Step 2: Implement project structured writes**

Add project service methods:

- `AppendToContent(ctx, input StructuredWriteInput) (ContentItem, error)`
- `InsertIntoContent(ctx, input StructuredWriteInput) (ContentItem, error)`
- `ReplaceContentRange(ctx, input StructuredWriteInput) (ContentItem, error)`
- `UpdateEntrySectionBody(ctx, input UpdateEntrySectionInput) (EntrySection, error)`

All content-item writes must reuse the existing revision-checked update path and set `CreatedBy: "agent"`.

The structured write input must carry `AgentSessionID`, `ActionKind`, `ModelVariantID`, and `SkillID`. `SkillID` is empty in Agent Core and reserved for Milestone 3.

- [ ] **Step 3: Implement write tools**

Add tool names:

- `append_to_chapter`
- `insert_into_chapter`
- `replace_selection`
- `update_story_bible_entry`
- `update_entry_section`

Every write tool requires `projectId`, target ID, `expectedRevision` where applicable, and a human-readable `reason`.

Every write tool records an activity event whose metadata includes target content item ID, operation kind, previous revision, next revision, action kind, model variant ID, and agent session ID.

- [ ] **Step 4: Verify**

Run:

```bash
go test -tags sqlite_fts5 ./project -run 'Test.*Structured|Test.*Conflict'
go test -tags sqlite_fts5 ./agent -run 'Test.*Write'
go test -tags sqlite_fts5 ./...
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add agent/tools.go agent/tools_test.go project/service.go project/types.go project/service_test.go queries/project_core.sql store
git commit -m "feat: add revision-safe agent writes"
```

---

## Task 7: Chat Turns, Tool Loop, And Prompt Records

**Files:**
- Modify: `agent/service.go`
- Modify: `agent/provider.go`
- Test: `agent/service_test.go`

- [ ] **Step 1: Add fake-provider chat tests**

Use a fake provider that first returns a tool call and then returns a final assistant message. Assert:

- user message is stored,
- tool message is stored,
- assistant message is stored,
- activity event is stored,
- prompt record is stored,
- request JSON does not contain provider secret.
- prompt records older than the profile retention window are deleted when `PrunePromptRecords` runs.

- [ ] **Step 2: Implement chat turn orchestration**

`RunChatTurn(ctx, input ChatTurnInput) (ChatTurnResult, error)` must:

1. Validate session and model variant.
2. Store the user message.
3. Build prompt messages from prompt profile and transcript.
4. Call provider with tool definitions.
5. Execute tool calls if present.
6. Call provider again with tool results, up to `maxToolRounds = 4`.
7. Store final assistant message.
8. Store prompt record and activity events.

- [ ] **Step 3: Implement prompt record redaction**

Prompt records must include raw provider request/response JSON, but never API keys or auth headers. Store only provider display name and model name.

- [ ] **Step 4: Implement prompt record retention**

Add `PrunePromptRecords(ctx, projectID string) (int64, error)`. It loads the project's `prompt_record_retention_days` from `prompt_profiles`, deletes records where `created_at` is older than that many days, and returns the deleted row count. A retention value of `0` disables prompt record storage for new turns and deletes existing records for the project when pruning runs.

- [ ] **Step 5: Verify**

Run:

```bash
go test -tags sqlite_fts5 ./agent -run 'TestRunChatTurn|TestPromptRecord|TestPrunePromptRecords'
go test -tags sqlite_fts5 ./...
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add agent/service.go agent/provider.go agent/service_test.go agent/types.go
git commit -m "feat: run agent chat turns"
```

---

## Task 8: Quick Actions

**Files:**
- Modify: `agent/service.go`
- Modify: `agent/prompt.go`
- Test: `agent/service_test.go`

- [ ] **Step 1: Add quick action tests**

Test:

- Continuation direct apply appends generated prose to a Chapter and creates a revision.
- Continuation preview creates a pending `generation_candidates` row and returns it without changing the Chapter.
- Accepting a Continuation preview commits exactly one structured write and marks the candidate accepted.
- Rewrite direct apply replaces selected text and creates a revision.
- Rewrite preview creates a pending `generation_candidates` row with original/generated text and a diff-friendly payload without changing the Chapter.
- Rejecting a Rewrite preview marks the candidate rejected without changing the Chapter.
- Accepting a Rewrite preview with a stale expected revision marks the candidate conflict and returns `project.ErrConflict`.
- Read and Check stores assistant report and creates an Attached Note with source `read_and_check`.

- [ ] **Step 2: Implement quick action input types**

Inputs must include:

- `projectId`
- `contentId`
- `modelVariantId`
- `applyMode`
- optional `guidance`
- `expectedRevision`
- selection range for rewrite/read-check
- cursor/insert position for continuation insert mode
- word or sentence target for continuation

- [ ] **Step 3: Implement Continuation**

Continuation uses prompt assembly, calls provider once or through tools, and either:

- returns `GeneratedCandidate` for preview, or
- commits one append/insert structured write for direct apply.

- [ ] **Step 4: Implement Rewrite**

Rewrite requires selected text and surrounding context. Preview returns original/generated text plus a diff-friendly payload. Direct Apply commits one replace structured write.

- [ ] **Step 5: Implement preview accept/reject**

Add:

- `AcceptCandidate(ctx, input AcceptCandidateInput) (AcceptCandidateResult, error)`
- `RejectCandidate(ctx, input RejectCandidateInput) (GenerationCandidate, error)`

Accept must re-run the stored structured write using the candidate's original `expectedRevision`. Reject must only update candidate status. Both operations must create activity events.

- [ ] **Step 6: Implement Read and Check**

Read and Check never changes prose. It stores the report as an assistant message and creates an attached note linked to the target chapter/selection.

Use `project.Service.CreateAttachedNote` with `Source: "read_and_check"`, `ContentItemID`, `SelectionStart`, `SelectionEnd`, and a title derived from the checked content title.

- [ ] **Step 7: Verify**

Run:

```bash
go test -tags sqlite_fts5 ./agent -run 'Test.*Continuation|Test.*Rewrite|Test.*ReadAndCheck'
go test -tags sqlite_fts5 ./...
```

Expected: PASS.

- [ ] **Step 8: Commit**

```bash
git add agent/service.go agent/prompt.go agent/service_test.go agent/types.go
git commit -m "feat: add agent quick actions"
```

---

## Task 9: Agent HTTP API

**Files:**
- Create: `agent/http.go`
- Create: `agent/http_test.go`
- Modify: `app/app.go`
- Modify: `app/app_test.go`
- Modify: `main.go`

- [ ] **Step 1: Add HTTP tests**

Test endpoints:

- `GET /api/provider-configs`
- `POST /api/provider-configs`
- `PUT /api/provider-configs/{providerID}`
- `GET /api/provider-configs/{providerID}/model-variants`
- `POST /api/provider-configs/{providerID}/model-variants`
- `GET /api/projects/{projectID}/agent/prompt-profile`
- `PUT /api/projects/{projectID}/agent/prompt-profile`
- `GET /api/projects/{projectID}/agent/sessions`
- `POST /api/projects/{projectID}/agent/sessions`
- `POST /api/projects/{projectID}/agent/sessions/{sessionID}/messages`
- `POST /api/projects/{projectID}/agent/actions/continuation`
- `POST /api/projects/{projectID}/agent/actions/rewrite`
- `POST /api/projects/{projectID}/agent/actions/read-check`
- `POST /api/projects/{projectID}/agent/candidates/{candidateID}/accept`
- `POST /api/projects/{projectID}/agent/candidates/{candidateID}/reject`
- `GET /api/projects/{projectID}/agent/activity`
- `GET /api/projects/{projectID}/agent/prompt-records`
- `POST /api/projects/{projectID}/agent/prompt-records/prune`

- [ ] **Step 2: Run tests and verify they fail**

Run:

```bash
go test -tags sqlite_fts5 ./agent ./app -run 'Test.*Agent|Test.*Provider'
```

Expected: FAIL because routes do not exist.

- [ ] **Step 3: Implement routes**

Register Agent routes under `/api`. Keep project routes in `project.RegisterRoutes` and Agent routes in `agent.RegisterRoutes`.

- [ ] **Step 4: Wire app and main**

Add `AgentService` to `app.Dependencies`. In `main.go`, create:

```go
projectService := project.NewService(db)
agentService := agent.NewService(db, projectService, agent.NewOpenAICompatibleClient(http.DefaultClient))
handler := app.New(&app.Dependencies{
	ProjectService: projectService,
	AgentService:   agentService,
	StaticFS:       staticFS,
})
```

- [ ] **Step 5: Verify**

Run:

```bash
go test -tags sqlite_fts5 ./agent ./app
go test -tags sqlite_fts5 ./...
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add agent/http.go agent/http_test.go app/app.go app/app_test.go main.go
git commit -m "feat: expose agent core API"
```

---

## Task 10: React Agent Settings And Panel

**Files:**
- Create: `frontend/src/agentTypes.ts`
- Create: `frontend/src/agentApi.ts`
- Modify: `frontend/src/App.tsx`
- Modify: `frontend/src/styles.css`

- [ ] **Step 1: Add frontend Agent types**

Create types for provider config summaries, model variants, sessions, messages, quick action requests/results, activity events, and prompt records.

- [ ] **Step 2: Add Agent API functions**

Implement functions for the endpoints from Task 9. Use `encodeURIComponent` for path segments and `URLSearchParams` for query strings.

- [ ] **Step 3: Add provider/model settings surface**

Add a compact settings area that lets the author:

- view provider configs,
- create/update one provider config,
- create/list model variants,
- select active model variant.
- edit prompt profile fields: genre, tense, point of view, voice, writing instructions, and Prompt Record retention days.

Do not show provider API keys after save.

Provider disclosure must be visible but not noisy: the active provider/model variant is shown in the Agent panel header and near quick action controls, but no warning banner appears unless there is no configured model variant.

- [ ] **Step 4: Add Agent side panel**

For the selected project/content item, add:

- chat transcript,
- message input,
- apply mode checkbox/select remembered in React state,
- quick action buttons for Continuation, Rewrite, Read and Check,
- current provider/model label,
- activity count pill that expands to event rows.

- [ ] **Step 5: Add quick action controls**

Continuation controls:

- target type: words or sentences,
- target count,
- optional guidance.

Rewrite and Read and Check controls can use the whole previewed content until Galley Editor selection integration exists.

Preview results must show generated text and Accept/Reject buttons. Accept calls the candidate accept endpoint; Reject calls the candidate reject endpoint. Direct Apply results must refresh the selected content item after success.

- [ ] **Step 6: Verify frontend**

Run:

```bash
cd frontend
bun run build
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add frontend/src/agentTypes.ts frontend/src/agentApi.ts frontend/src/App.tsx frontend/src/styles.css
git commit -m "feat: add agent workspace panel"
```

---

## Task 11: End-To-End Smoke Test

**Files:**
- Modify docs only if setup commands changed.

- [ ] **Step 1: Run backend verification**

```bash
gofmt -w .
go test -tags sqlite_fts5 ./...
go vet -tags sqlite_fts5 ./...
```

Expected: PASS.

- [ ] **Step 2: Run frontend verification**

```bash
cd frontend
bun run build
```

Expected: PASS.

- [ ] **Step 3: Run API smoke test with fake provider**

If the app supports a built-in fake provider only in tests, run the HTTP tests from Task 9. If a development fake provider mode is added, start the app with that mode and run:

```bash
go run -tags sqlite_fts5 .
curl http://localhost:8080/api/health
```

Expected:

```json
{"status":"ok"}
```

- [ ] **Step 4: Commit docs adjustments**

If commands changed:

```bash
git add docs
git commit -m "docs: update agent core plan notes"
```

Skip this commit if no docs changed.

## Self-Review

Spec coverage:

- OpenAI-compatible provider configuration: Tasks 1, 3, 9, 10.
- Model variants and quick switching foundation: Tasks 1, 3, 9, 10.
- Prompt assembly with genre/tense/POV/instructions and layered Writing Briefs: Task 4.
- Project map and retrieval/read tools: Task 5.
- Metadata-filtered retrieval: Task 5.
- Chapter reads and Story Bible Entry reads: Task 5.
- Continuation, Rewrite, Read and Check: Task 8.
- Direct Apply, preview mode, and preview accept/reject: Tasks 6, 8, 9, 10.
- Structured Writes and conflict handling: Task 6.
- Agent-created revision labels for action/model/session/skill: Tasks 1 and 6.
- Activity Trails: Tasks 1, 5, 7, 9, 10.
- Prompt Records and retention controls: Tasks 1, 7, 9, 10.
- Provider Disclosure: Task 10.

Deferred by design:

- Skill import/install and skill routing are Milestone 3.
- Galley Editor selection wiring is Daily Writing Polish; this milestone can use whole-content actions from the preview pane.
- Streaming generation is deferred. Provider calls are non-streaming and commit complete candidates.
- Auth remains placeholder from Project Core unless explicitly pulled forward into a separate auth/security plan.
