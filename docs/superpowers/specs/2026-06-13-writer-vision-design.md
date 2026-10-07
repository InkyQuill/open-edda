# Open Edda Vision Design

## Purpose

Open Edda is a private, self-hosted AI writing studio for hobby novelists. Its core promise is that an author can work with an AI agent over the whole Story Project, not only the current text selection, while keeping ownership, privacy, Markdown portability, local file ownership, and version safety. It also allows the author to write in any human language the configured model handles well. No more English-only like in Sudowrite/NovelAI.

This matters because the target workflow already proved itself in local terminal-agent editing: an AI assistant can navigate story text, story bible material, and project-specific skills, then perform large coherent revisions that current hosted writing products do not handle well. Open Edda preserves that breakthrough and makes it available through a browser-friendly writing environment.

## Scope

The first version targets a Single-Author Instance with support for multiple Story Projects. It is responsive from day one: desktop and tablet layouts prioritize full chapter editing, while phone-sized layouts support reading, chat, small edits, and triggering agent actions.

In scope for the first vertical slice and file-first migration:

- Create and manage Story Projects.
- Open or create Markdown project folders using the defined Edda layout.
- Import and export Markdown using the Edda layout, with conversion helpers for Elysium and other existing projects.
- Edit Markdown-based Chapters through Galley Editor.
- Manage Story Bible Entries with types, sections, and optional typed relations.
- Manage layered Writing Briefs.
- Configure OpenAI-compatible providers and model variants.
- Run project-aware chat and the first three agent actions: Continuation, Rewrite, and Read and Check.
- Preserve meaningful changes through explicit saves, project-wide checkpoints, diffs, Activity Trails, and Prompt Records.
- Support a small local CLI for `edda get`, `edda save`, status, `edda send`, `edda take`, diff, history, and restore.
- Install/import Skills as instructions, routing metadata, templates, and data files.

Out of scope for the first version:

- Multi-author collaboration.
- Public registration, invitations, roles, or organizations.
- Native Android or desktop apps.
- Publishing workflows.
- Marketplace/community features.
- Keystroke-level local-folder sync.
- Git-compatible branching, staging, rebasing, remote management, or merge commits.
- Arbitrary skill script execution.
- Vector embeddings as a required retrieval dependency.

## Product Model

### Story Projects

A Story Project is the complete body of material for one fiction work or series. In the target model it is an Edda project folder plus `.edda/` metadata. It includes story prose files, story bible files, writing briefs, project notes, attached-note records, Agent Sessions, Skills, checkpoints, and settings.

The Project Dashboard is the entry point for choosing, creating, fetching, importing, exporting, and configuring Story Projects. The Writing Workspace is the project-level surface where the author edits chapters, manages reference material, reviews notes/reports, works with skills, and runs agent sessions.

### Chapters

Chapters are the primary Story Text unit in v1. A Chapter is a continuous Markdown file in the Edda layout with cursor and selection-aware agent actions. Scenes can be represented through Markdown headings or markers, but v1 does not model scenes as separate entities.

### Story Bible Entries

Story Bible Entries are typed, Markdown-based reference files in the Edda layout. Each entry has a title, type, metadata, optional tags, optional status, optional typed Entry Relations, and a rich Markdown body.

Entry Sections are freeform named sections inside an entry. They are addressable independently so the agent can retrieve focused context such as `Dwarves > Player Perspective` without reading the full entry.

Entry Relations connect entries to each other. The relation label is optional, allowing both simple imported `related` lists and richer relationships such as `mentor of`, `located in`, `member of`, `contradicts`, or `source for`.

### Writing Briefs

Writing Briefs are durable instructions describing what the agent is writing and how it should write it: genre, tense, point of view, style, conventions, and prose constraints.

Briefs are layered. Each Story Project has project-wide defaults, and individual Chapters may add focused overrides such as POV, tense, or document form.

### Notes

Project Notes are exploratory or temporary material that can inform the author and agent but is not canon by default.

Attached Notes are linked to a Chapter or text selection. Read and Check reports become Agent Session output and may be represented as Attached Notes. They do not become Project Notes or Story Bible Entries unless the author promotes them.

### Skills

Skills are project-installable agent procedures. A Skill can include instructions, routing metadata, templates, and data files. V1 imports and exposes these materials to the agent, but does not execute arbitrary bundled scripts. Later script execution requires admin approval.

## Architecture

Open Edda is one self-hosted deployable plus a small local CLI. The Go backend serves the API and the built React frontend. During development, the frontend and backend can run separately with a dev proxy. The CLI works with Edda project folders and talks to the service for connected projects.

The React frontend is required because Galley Editor is the intended Markdown-native editing foundation. Galley Editor is used for Chapters and other Markdown-based content.

The backend baseline is:

- Go for the service layer.
- `chi` for HTTP routing.
- `sqlc` for database access.
- `goose` for migrations.
- SQLite for the first database.

SQLite fits the single-author self-hosted workload if the implementation uses WAL mode, short transactions, and avoids long-running provider or agent work inside database transactions. In the file-first model, SQLite stores indexes, operational state, prompt/activity records, and cache data; story files and `.edda/` metadata remain rebuildable source material. Postgres remains a later option if collaboration or heavier concurrency demands it.

The backend owns:

- Auth and session handling.
- Persistence and migrations.
- Edda layout validation, file indexing, and Markdown import/export.
- Draft autosave, explicit Save, and checkpoint orchestration.
- Local/server sync for saved files and checkpoint metadata.
- Provider configuration and model variants.
- Prompt assembly.
- Agent tool definitions.
- Structured Writes.
- Version checks, checkpoint history, and diffs.
- Activity Trails.
- Prompt Records.

## Agent Workflow

Open Edda has one agent system with two entry styles:

- Conversational chat.
- Editor quick actions.

Quick actions are structured shortcuts that start or continue an Agent Session with a preset intent. Chat can call the same tools when the conversation naturally turns into action.

### Whole-Project Context

Whole-Project Context means tool-accessible context, not placing the entire project into every prompt. Each action starts with a task-focused default context bundle, and the agent can expand from there through tools.

V1 retrieval uses:

- A structured project map.
- Metadata filters.
- Full-text search over Markdown.
- Chapter reads.
- Story Bible Entry reads.
- Entry Section reads.
- Skill instructions/assets.

Embeddings can be added later if structured and lexical retrieval prove insufficient.

Prompt context should be assembled from named Context Sources rather than one ad hoc string. Each source has a stable key, deterministic rendering, and a stored snapshot for the Agent Session. Stable sources include the prompt profile, layered Writing Briefs, project map, selected content window, available tool names, provider/model disclosure, and later Skill guidance. When a source changes during a session, the backend admits the new rendered value only at the next provider-turn boundary and records that admission as session context history. This preserves debuggability while avoiding stale invisible prompt changes.

Tool results returned to the model are bounded. The backend stores the complete structured result for audit/debug use, but the model-visible result is a size-limited Markdown/JSON projection that identifies when it was truncated. This prevents one broad search or large Story Bible read from flooding the next provider turn.

### Continuation

Continuation generates new prose at the cursor or chapter end. The author can choose a word or sentence target and optional guidance. Without guidance, the model infers a plausible next direction from the chapter, Writing Brief, and project context.

Continuation may stream into a transient Generation Buffer in the UI, but it does not stream token-by-token into the saved Chapter file. Once complete, the generated span commits as one insert or append operation into a draft or saved file according to the active action.

### Rewrite

Rewrite transforms selected text using the author's guidance. It receives the selection, surrounding context, Writing Brief, and relevant project context. It produces a complete candidate before replacing text.

In preview mode, the author sees a diff and accepts or rejects the candidate. In Direct Apply mode, the replacement commits after completion into the current draft or saved file. A checkpoint is created only when the author explicitly saves one.

### Read and Check

Read and Check evaluates a selection or Chapter and returns a report with observations and suggestions. It does not directly change prose. Reports are stored as Agent Session output and can become Attached Notes linked to the checked Chapter or selection.

### Structured Writes

Document-changing agent tools use Structured Writes rather than arbitrary raw patches. In the file-first model, each write targets a known saved file hash, current implementation revision token, byte range, or structured location. V1 write operations include:

- Insert at position.
- Append to Chapter.
- Replace selected range.
- Update Story Bible Entry body.
- Update Entry Section body.

Each Structured Write targets a known file-backed version. If the content changed after the agent read it, the backend returns a conflict error. The agent can re-read current content and retry, or ask the author when the merge is ambiguous.

## Saves, Checkpoints, And Diffs

Open Edda separates draft autosave, Save, and Checkpoint.

- Draft autosave protects web-editor work in progress without updating canonical Markdown files.
- Save writes the current draft into the working project state.
- Checkpoint records a named project-wide snapshot of saved files for history, comparison, restore, recovery, and sync.

Checkpoints are linear and writer-facing. They are not git branches or commits in the product vocabulary. The CLI should support commands such as `edda save "Note"`, `edda status`, `edda send`, `edda take`, `edda history`, `edda diff`, and `edda restore`.

Current per-item database revisions can remain as migration scaffolding, audit detail, or implementation tokens, but the product-facing history model should move to explicit project-wide checkpoints.

Diffs compare current saved files with checkpoints or compare two checkpoints. Agent-created saved changes should be labeled with action, model variant, skill if any, and the linked Agent Session in activity/checkpoint metadata.

## Import And Export

Open Edda uses one defined layout to classify Markdown files. The target is the `alchemist`-style Edda layout:

- `AGENTS.md` and `BOOTSTRAP.md` hold project guidance and bootstrap instructions.
- `.agents/skills/` holds project-local skill source material.
- `story/` maps to Chapters and includes `_index.md` plus chapter files.
- `storyline/` maps to plans, chapter plans, sequence maps, continuity matrices, draft status, style guides, and other planning material.
- `characters/` maps to character Story Bible Entries and includes `_index.md`.
- `worldbuilding/` maps to lore, with subdirectories such as `culture/`, `magic/`, `monsters/`, and `places/`.
- `drafts/` maps to draft material and alternate chapter versions.

Frontmatter preserves metadata such as status, type, roles, pronouns, tags, and relations. Markdown headings become Entry Sections where applicable.

Import is conservative and can create a new Story Project from a clean Edda layout folder or archive. Export is repeatable and loss-aware. Elysium is one older structure that can be converted into Edda's layout, but it is not the target shape. For projects that use a different personal structure, Edda should provide conversion helpers and clear validation errors rather than trying to support every personal organization style. The target model also supports opening a local Edda folder in place, creating `.edda/` metadata, indexing files, and connecting that folder to a server project.

Saved file state and checkpoints move between local folders and the server through the Edda CLI and service sync. Sync operates on saved files and checkpoint metadata, not editor drafts or every keystroke.

## Privacy And Trust

Open Edda is private by default. Story Project files, checkpoints, Prompt Records, Activity Trails, provider keys, exports, skills, and settings are private to the authenticated Author.

V1 uses normal password auth with one admin author by default. There is no public registration, invitation flow, role system, or collaboration model in v1.

Provider Disclosure is visible but not noisy. Settings show the configured provider and model variants, and agent actions indicate which provider/model is being used. The app does not interrupt every generation with a warning unless the provider configuration changes or the author is setting it up.

Activity Trails are compact by default, such as an `actions: 32` pill, and expand to show readable details: which project files or indexed items were read, which action ran, which skill was used, which model variant was used, and which content item changed.

Prompt Records are separate advanced/debug records of raw assembled model inputs and outputs. They are subject to retention controls and must never include provider secrets.

Provider usage accounting is part of Prompt Records and Activity Trails. Open Edda stores provider-neutral token buckets: uncached input, output, cache read, cache write, total tokens, and per-bucket estimated cost. This is required for OpenAI-compatible providers with prompt-cache reporting, including DeepSeek-style APIs where cache-hit tokens are billed differently from ordinary input tokens. Model variants therefore store per-million-token prices for input, output, cache read, and cache write, plus provider compatibility metadata such as the request token field and thinking/reasoning format.

## Research Notes

Two existing agent systems informed the Agent Core design:

- OpenCode separates durable session history from provider-turn System Context, using stable Context Sources and safe provider-turn boundaries. Open Edda adopts the useful part: named prompt context sources with stored snapshots and deterministic admission, but keeps the implementation smaller and project-specific.
- OpenCode also bounds model-visible tool output and stores full output separately. Open Edda applies the same principle to retrieval/read tools so a large Story Bible or search result cannot dominate the next prompt.
- Pi keeps provider usage in normalized buckets: input, output, cache read, cache write, total tokens, and cost for each bucket. Open Edda adopts this accounting shape because it maps cleanly to DeepSeek/OpenAI-compatible cache usage and makes spend visible without provider-specific UI.
- Pi’s agent loop treats model calls, tool execution, and follow-up turns as explicit events. Open Edda does not need Pi’s full streaming loop in Milestone 2, but Activity Trails should use similarly explicit events for provider turn started/finished, tool called, tool result bounded, structured write applied, candidate accepted/rejected, and prompt record stored.

## Roadmap

### Milestone 1: Project Core

Auth, Story Projects, Chapters, Story Bible Entries, Entry Sections, Entry Relations, Writing Briefs, Project Notes, Attached Notes, Per-Item Revisions, diffs, and Elysium Layout import/export. This milestone is implemented under the earlier database-backed model and is historical context for the file-first migration.

### Milestone 2: Agent Core

OpenAI-compatible provider configuration, model variants, prompt assembly, project map, structured retrieval/read tools, Continuation, Rewrite, Read and Check, Direct Apply, preview mode, Structured Writes, conflict handling, Activity Trails, and Prompt Records.

### Milestone 3: Skill Core

Skill import/install, skill browsing, routing/selecting skills for Agent Sessions, exposing skill instructions/assets to the agent, and clear handling for script-disabled skills.

### Milestone 4: Daily Writing Polish

Editor ergonomics, mobile-friendly layouts, side-panel Attached Notes, better diff/restore UI, export polish, provider disclosure polish, and model-switching UX.

### Milestone 5: File-First Projects And Checkpoints

The defined Edda project layout, `.edda/` metadata, file indexing, draft/save/checkpoint semantics, local CLI, linear sync, diff/history/restore, and migration from database-owned prose to file-backed project content.

### Later: Collaboration

Add multi-author collaboration only after the single-author file-first workflow is stable.

## Open Decisions Deferred To Implementation Planning

- Exact React build tooling and component library.
- Exact SQLite schema and migration layout.
- Exact full-text search implementation.
- Exact auth/session library.
- Exact provider streaming protocol abstraction.
- Exact Galley Editor integration API.
