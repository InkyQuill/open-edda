# Agent Tools

This document describes the current Open Edda agent tool mechanism. It is written as both human-readable implementation documentation and agent-readable operating guidance for future work.

## Current State

Open Edda has real OpenAI-compatible function tools for chat sessions and quick-action generation. The full chat tool catalog is declared in `agent.ContextToolDefinitions` and passed to the model from `RunChatTurn`.

Quick actions use a read/context subset declared by `agent.QuickActionToolDefinitions`. Continuation, Rewrite, and Read and Check can call tools such as `project_map`, `search_content`, `read_chapter`, `read_story_bible_entry`, `read_entry_section`, `list_revisions`, `skill`, and `skill_script` before producing the final action output. Direct write tools are intentionally excluded from quick-action generation so preview/direct-apply semantics remain controlled by the action pipeline.

There is no tool named `read_file`, `read_skill`, or `read_project_file` yet. Skill loading is exposed as the `skill` tool. Project content reads are currently exposed through content-oriented tools such as `read_content`, `read_chapter`, `read_story_bible_entry`, and `read_entry_section`. Under the file-first architecture, those content tools should resolve through the project folder index and saved file hashes rather than treating database rows as canonical prose.

## Tool Catalog

The tool definitions live in `agent/tools.go`:

| Tool | Purpose | Notes |
| --- | --- | --- |
| `project_map` | Read a concise map of project content. | Uses `project.Service.ProjectMap`. |
| `search_content` | Search project content by text, kind, metadata, and tags. | Supports `chapter`, `story_bible_entry`, `writing_brief`, and `project_note`; limit is 1-100. |
| `read_content` | Read one content item by `contentId`. | Allows any content kind. |
| `read_chapter` | Read one chapter by `contentId`. | Rejects non-chapter content. |
| `read_story_bible_entry` | Read one Story Bible entry by `contentId`. | Rejects non-entry content. |
| `read_entry_section` | Read one Story Bible entry section by `contentId` and `heading`. | Rejects non-entry content and missing sections. |
| `list_revisions` | List revisions for one content item. | Read-only. |
| `skill` | Load one installed Edda skill by `skillId`. | Returns instructions and a manifest of supporting files; use `read_skill_file` for individual reference/template/data files. |
| `read_skill_file` | Read one non-script supporting file from an installed Edda skill by `skillId` and `path`. | Loads files on demand; rejects scripts and disabled script files. |
| `skill_script` | Run one enabled helper script from a selected skill. | Requires session selection, admin approval, and a safe runtime envelope; scripts cannot mutate project content directly. |
| `append_to_chapter` | Append generated Markdown to a chapter. | Write tool; requires session ID and `expectedRevision`. |
| `insert_into_chapter` | Insert generated Markdown into a chapter at a byte position. | Write tool; requires session ID and `expectedRevision`. |
| `replace_selection` | Replace a selected chapter byte range. | Write tool; requires session ID and `expectedRevision`. |
| `update_story_bible_entry` | Replace a Story Bible entry body. | Write tool; requires session ID and `expectedRevision`. |
| `update_entry_section` | Update a Story Bible entry section body. | Write tool; requires session ID and `expectedRevision`. |

All tools are declared as OpenAI-style `function` tools with JSON schemas. Argument decoding and validation happen server-side in `Service.ExecuteTool`.

## Invocation Flow

Chat turns use this flow:

1. `RunChatTurn` appends the user message to `agent_messages`.
2. It builds prompt context from the prompt profile, transcript, available skills, and selected skills.
3. It sends `ContextToolDefinitions()` in `CompletionRequest.Tools`.
4. If the model returns `tool_calls`, Open Edda executes each call through `Service.ExecuteTool`.
5. Each tool result is appended to the provider message list as a `tool` role message with the original `tool_call_id`.
6. The loop repeats until the model returns a final assistant message or exceeds `maxToolRounds`.

`maxToolRounds` is currently 4. If the model still asks for tools after the final allowed round, the chat turn fails.

Quick actions use this flow:

1. `runQuickActionCompletion` creates a session for `continuation`, `rewrite`, or `read_check`.
2. `BuildActionPrompt` creates system, developer, and user messages.
3. The provider receives `QuickActionToolDefinitions()` in `CompletionRequest.Tools`.
4. If the model returns `tool_calls`, Open Edda rejects direct write tools and executes allowed context tools through `Service.ExecuteTool`.
5. Each tool result is appended to the provider message list as a `tool` role message with the original `tool_call_id`.
6. The loop repeats until the model returns final action output or exceeds `maxToolRounds`.
7. The returned text becomes a generation candidate, a direct write, or a read/check report depending on the action.

## Prompt Guidance

The system prompt is intentionally short:

```txt
You are a fiction writing assistant working inside Edda. Preserve the author's intent, respect established project facts, and use available tools to inspect project context before making claims. Do not invent durable worldbuilding facts unless the author asks you to brainstorm.
```

Additional tool guidance is in system/developer context sources rather than a long system prompt.

For chat, `chatContextSources` records a tool catalog source with the instruction that project context tools are available for chat turns. The actual callable catalog is still the `tools` array in the completion request.

For action prompts, `BuildActionPrompt` renders the provider disclosure:

```txt
Use tools for additional project context instead of assuming the whole project is present in this prompt.
```

The action prompt's tool catalog also tells the model that project context tools are available for the action and that direct write tools are not available. `runQuickActionCompletion` passes `bundle.Tools` to the provider and runs the same bounded tool-round pattern as chat.

Skill guidance is rendered separately:

- Available skills are listed by ID, name, and description.
- The prompt tells the model to use the `skill` tool when a task matches a skill description.
- Selected skills are listed with enabled runtime helpers, if any.
- Runtime helpers are described as available only through `skill_script`.

This guidance is callable in chat and quick actions. Chat receives the full tool catalog, including write tools. Quick actions receive only the read/context subset plus selected safe skill-script helpers.

## Skill Tool Semantics

`skill` is the current equivalent of loading the main skill instructions.

Inputs:

```json
{ "skillId": "skill-..." }
```

Behavior:

- Loads a skill from the database through `skill.Service.RenderForModel`.
- Returns the main instructions and a manifest of supporting files.
- Lists templates, references, data files, and other non-script supporting files as readable paths without loading their bodies.
- Omits script file bodies and marks scripts as not readable through skill-file loading.
- Bounds model-visible output to avoid over-large tool responses.
- Records an activity event with event type `skill_loaded`.

Important boundary: loading a skill does not execute bundled scripts.

## Skill File Semantics

`read_skill_file` loads one supporting file listed by `skill`.

Inputs:

```json
{ "skillId": "skill-...", "path": "references/checklist.md" }
```

Behavior:

- Loads exactly one stored skill file by relative path.
- Returns the file body as model-visible Markdown.
- Allows non-script files such as `templates/`, `references/`, `data/`, and other inert text files.
- Rejects scripts and disabled script files; executable helpers belong to `skill_script`.

Use this tool only after the skill manifest shows that the file exists and is readable.

## Skill Script Semantics

`skill_script` is the only model-callable script execution path.

Inputs include:

```json
{
  "skillId": "skill-...",
  "scriptPath": "scripts/example.ts",
  "contentIds": ["content-..."],
  "entrySections": [{ "contentId": "content-...", "heading": "History" }],
  "assetPaths": ["data/example.json"],
  "arguments": {}
}
```

Execution requirements:

- The call must belong to a session.
- The skill must be selected for that session.
- The script must be a stored skill file with purpose `script`.
- A matching `skill_script_approval` must exist and be enabled.
- The approval must have a runtime command.
- Destructive operations, network access, and project-file access are rejected in the current runtime path.

Scripts receive a JSON envelope on stdin. The envelope can include project and skill refs, selected content IDs, entry section refs, requested skill assets, and free-form arguments.

Scripts run in a temporary workspace with a minimal environment. They must write valid JSON to stdout using the `skill-script-output/v1` contract. Valid output kinds are `report`, `proposal`, `draft`, and `generated_data`.

Scripts cannot directly mutate Story Text or Story Bible content. Their output is returned to the model and stored as a run artifact. Applying content changes still requires normal write tools or user-approved candidate flows.

## Result Storage

Every successful `ExecuteTool` call stores:

- An `activity_events` row.
- A `tool_result_artifacts` row containing full JSON, model-visible Markdown, truncation status, byte size, tool name, and tool call ID.

For `skill_script`, the skill service also stores a `skill_script_runs` row with the input envelope, stdout, stderr, exit code, duration, output JSON, and error message.

Model-visible tool output is bounded:

- `maxToolVisibleLines`: 2000
- `maxToolVisibleBytes`: 50 KiB

The full JSON remains available in `tool_result_artifacts`.

## Write Tool Safety

Write tools require a session ID and a tool call ID. They validate target content kind and required arguments.

Structured write tools require:

- `contentId`
- `expectedRevision`
- `generatedMarkdown`
- `reason`

This keeps writes version-aware. If content changed since the model read it, project-service checks can reject the write rather than silently overwriting newer author work. During the file-first migration, `expectedRevision` should evolve toward an expected saved file hash or equivalent file-backed version token.

Write activity metadata records target content, operation kind, session, action kind, model variant, and revision movement when available.

Quick-action generation does not expose write tools. If a model attempts a direct write tool call anyway, the server rejects it with an error. Continuation and Rewrite still write through their existing preview/direct-apply flow after the final model response, and Read and Check stores a report/attached note after the final response.

## Agent Operating Guidance

When acting as an Open Edda chat or quick-action agent:

1. Use `project_map` or `search_content` before making claims about project-wide facts.
2. Use `read_chapter`, `read_story_bible_entry`, or `read_entry_section` before giving detailed feedback on specific content.
3. Use `skill` when an available skill description matches the task.
4. Use `read_skill_file` for listed reference, template, or data files only when the loaded skill says that extra file is relevant.
5. Use `skill_script` only for selected skills with enabled runtime helpers.
6. Treat script results as proposals, reports, drafts, or generated data. Do not assume they changed project content.
7. In chat, use write tools only when the user asked for an applied change and you have the current revision or saved file version token.
8. Include a clear `reason` for every write.

When running a quick action, inspect whatever project context is needed before final output. Do not try to directly apply edits through write tools; quick-action writes happen after final output through the action pipeline.

## Implementation Pointers

- Tool definitions and execution: `agent/tools.go`
- Chat tool loop: `agent/service.go`, `RunChatTurn`
- Quick-action completion path: `agent/service.go`, `runQuickActionCompletion` and `completeQuickActionWithTools`
- Prompt construction and skill prompt sources: `agent/prompt.go`
- OpenAI-compatible request serialization: `agent/provider.go`
- Skill loading: `skill/service.go`, `RenderForModel`
- Skill script runtime: `skill/script_runtime.go` and `skill/runtime/`
- Tool artifacts schema: `migrations/00002_agent_core.sql`
- Skill schema: `migrations/00003_skill_core.sql`
- Skill script runtime schema: `migrations/00005_skill_script_runtime.sql`

## Known Gaps

- There are no file-system tools for agents yet. Reads are content-oriented, but the product source of truth is now the project folder plus `.edda/` metadata; the content tools need a file-index-backed implementation.
- `skill` loads by ID only. The model sees skill IDs in the available/selected skill prompt sources; there is no model-facing `read_skill_by_name`.
- Tool errors currently fail the chat turn instead of returning a model-visible tool error that the model can recover from.
