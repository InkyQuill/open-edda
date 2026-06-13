# Writer Vision Design

## Purpose

Writer is a private, self-hosted AI Writing Studio for hobby novelists. Its core promise is that an author can work with an AI agent over the whole Story Project, not only the current text selection, while keeping ownership, privacy, Markdown portability, and revision safety. It also allows to not be tied to any human language, as models became good enough to produce good results in common ones. No more English-only like in Sudowrite/NovelAI.

This matters because the target workflow already proved itself in local terminal-agent editing: an AI assistant can navigate story text, story bible material, and project-specific skills, then perform large coherent revisions that current hosted writing products do not handle well. Writer preserves that breakthrough and makes it available through a browser-friendly writing environment.

## Scope

The first version targets a Single-Author Instance with support for multiple Story Projects. It is responsive from day one: desktop and tablet layouts prioritize full chapter editing, while phone-sized layouts support reading, chat, small edits, and triggering agent actions.

In scope for the first vertical slice:

- Create and manage Story Projects.
- Import and export Markdown using the Elysium Layout.
- Edit Markdown-based Chapters through Galley Editor.
- Manage Story Bible Entries with types, sections, and optional typed relations.
- Manage layered Writing Briefs.
- Configure OpenAI-compatible providers and model variants.
- Run project-aware chat and the first three agent actions: Continuation, Rewrite, and Read and Check.
- Preserve meaningful changes through Per-Item Revisions, diffs, Activity Trails, and Prompt Records.
- Install/import Skills as instructions, routing metadata, templates, and data files.

Out of scope for the first version:

- Multi-author collaboration.
- Public registration, invitations, roles, or organizations.
- Native Android or desktop apps.
- Publishing workflows.
- Marketplace/community features.
- Live local-folder sync.
- Arbitrary skill script execution.
- Vector embeddings as a required retrieval dependency.

## Product Model

### Story Projects

A Story Project is the complete body of material for one fiction work or series. It includes Chapters, Story Bible Entries, Writing Briefs, Project Notes, Attached Notes, Agent Sessions, Skills, revisions, and settings.

The Project Dashboard is the entry point for choosing, creating, importing, exporting, and configuring Story Projects. The Writing Workspace is the project-level surface where the author edits chapters, manages reference material, reviews notes/reports, works with skills, and runs agent sessions.

### Chapters

Chapters are the primary Story Text unit in v1. A Chapter is a continuous Markdown document with cursor and selection-aware agent actions. Scenes can be represented through Markdown headings or markers, but v1 does not model scenes as separate database entities.

### Story Bible Entries

Story Bible Entries are typed, Markdown-based reference entries. Each entry has a title, type, metadata, optional tags, optional status, optional typed Entry Relations, and a rich Markdown body.

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

Writer is one self-hosted deployable. The Go backend serves the API and the built React frontend. During development, the frontend and backend can run separately with a dev proxy.

The React frontend is required because Galley Editor is the intended Markdown-native editing foundation. Galley Editor is used for Chapters and other Markdown-based content.

The backend baseline is:

- Go for the service layer.
- `chi` for HTTP routing.
- `sqlc` for database access.
- `goose` for migrations.
- SQLite for the first database.

SQLite fits the single-author self-hosted workload if the implementation uses WAL mode, short transactions, and avoids long-running provider or agent work inside database transactions. Postgres remains a later option if collaboration or heavier concurrency demands it.

The backend owns:

- Auth and session handling.
- Persistence and migrations.
- Markdown import/export.
- Provider configuration and model variants.
- Prompt assembly.
- Agent tool definitions.
- Structured Writes.
- Revisions and diffs.
- Activity Trails.
- Prompt Records.

## Agent Workflow

Writer has one agent system with two entry styles:

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

### Continuation

Continuation generates new prose at the cursor or chapter end. The author can choose a word or sentence target and optional guidance. Without guidance, the model infers a plausible next direction from the chapter, Writing Brief, and project context.

Continuation may stream into a transient Generation Buffer in the UI, but it does not stream token-by-token into the saved Chapter. Once complete, the generated span commits as one insert or append operation.

### Rewrite

Rewrite transforms selected text using the author's guidance. It receives the selection, surrounding context, Writing Brief, and relevant project context. It produces a complete candidate before replacing text.

In preview mode, the author sees a diff and accepts or rejects the candidate. In Direct Apply mode, the replacement commits after completion and creates a revision.

### Read and Check

Read and Check evaluates a selection or Chapter and returns a report with observations and suggestions. It does not directly change prose. Reports are stored as Agent Session output and can become Attached Notes linked to the checked Chapter or selection.

### Structured Writes

Document-changing agent tools use Structured Writes rather than arbitrary raw patches. V1 write operations include:

- Insert at position.
- Append to Chapter.
- Replace selected range.
- Update Story Bible Entry body.
- Update Entry Section body.

Each Structured Write targets a known content revision. If the content changed after the agent read it, the backend returns a conflict error. The agent can re-read current content and retry, or ask the author when the merge is ambiguous.

## Revisions And Diffs

Revisions are per item, not project-wide. Chapters, Story Bible Entries, Writing Briefs, and Project Notes can have Per-Item Revisions.

Revision creation is grouped by meaningful operations:

- Direct Apply.
- Accepted preview.
- Import.
- Restore.
- Explicit save revision.
- Manual editing sessions that cross a chosen time or content threshold.

Autosave does not create noisy revision entries every few seconds.

Diffs compare item revisions and support restore. Agent-created revisions are labeled with action, model variant, skill if any, and the linked Agent Session.

## Import And Export

V1 uses the Elysium Layout as the first Markdown interoperability convention:

- `story/` maps to Chapters.
- `characters/` maps to character Story Bible Entries.
- `worldbuilding/` maps to worldbuilding Story Bible Entries.
- `braindump/` maps to Project Notes.
- Top-level guidance files such as `genre.md`, `synopsis.md`, and `conventions.md` map into Writing Briefs or Project Notes depending on role.

Frontmatter preserves metadata such as status, type, roles, pronouns, tags, and relations. Markdown headings become Entry Sections where applicable.

Import is conservative and creates a new Story Project from a clean Elysium Layout folder. Export is repeatable and loss-aware. V1 does not attempt ongoing merge/sync with a local folder.

A later Local Sync Tool will detect local Markdown changes and replay them into the service database with merge handling.

## Privacy And Trust

Writer is private by default. Story Project content, revisions, Prompt Records, Activity Trails, provider keys, exports, skills, and settings are private to the authenticated Author.

V1 uses normal password auth with one admin author by default. There is no public registration, invitation flow, role system, or collaboration model in v1.

Provider Disclosure is visible but not noisy. Settings show the configured provider and model variants, and agent actions indicate which provider/model is being used. The app does not interrupt every generation with a warning unless the provider configuration changes or the author is setting it up.

Activity Trails are compact by default, such as an `actions: 32` pill, and expand to show readable details: which project items were read, which action ran, which skill was used, which model variant was used, and which content item changed.

Prompt Records are separate advanced/debug records of raw assembled model inputs and outputs. They are subject to retention controls and must never include provider secrets.

## Roadmap

### Milestone 1: Project Core

Auth, Story Projects, Chapters, Story Bible Entries, Entry Sections, Entry Relations, Writing Briefs, Project Notes, Attached Notes, Per-Item Revisions, diffs, and Elysium Layout import/export.

### Milestone 2: Agent Core

OpenAI-compatible provider configuration, model variants, prompt assembly, project map, structured retrieval/read tools, Continuation, Rewrite, Read and Check, Direct Apply, preview mode, Structured Writes, conflict handling, Activity Trails, and Prompt Records.

### Milestone 3: Skill Core

Skill import/install, skill browsing, routing/selecting skills for Agent Sessions, exposing skill instructions/assets to the agent, and clear handling for script-disabled skills.

### Milestone 4: Daily Writing Polish

Editor ergonomics, mobile-friendly layouts, side-panel Attached Notes, better diff/restore UI, export polish, provider disclosure polish, and model-switching UX.

### Later: Local Sync And Collaboration

Build a Local Sync Tool that detects local Markdown changes and replays them into the service database with merge handling. Add multi-author collaboration only after the single-author workflow is stable.

## Open Decisions Deferred To Implementation Planning

- Exact React build tooling and component library.
- Exact SQLite schema and migration layout.
- Exact full-text search implementation.
- Exact auth/session library.
- Exact provider streaming protocol abstraction.
- Exact Galley Editor integration API.
