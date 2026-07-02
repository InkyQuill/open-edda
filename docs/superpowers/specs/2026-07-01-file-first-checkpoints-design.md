# File-First Projects And Linear Checkpoints Design

## Purpose

Open Edda should support a writer who mostly works in ordinary Markdown files across several computers, while still offering the web editor, AI assistants, search, history, and recovery. The goal is not to turn Open Edda into a git client. The goal is a simple writing workflow: open or fetch a project, edit files, save current work, create named checkpoints, sync with the server, compare changes, and recover earlier versions when needed.

This design supersedes the earlier database-source-of-truth model for story content. The Edda project folder becomes canonical. SQLite remains important, but as a rebuildable index, cache, and operational store.

## Goals

- Treat Edda project folders as first-class Open Edda projects.
- Keep story prose, characters, worldbuilding, planning, drafts, and other author material as normal files.
- Allow projects to start either in the web app/server or as a local folder.
- Preserve a comfortable web editing workflow with draft autosave and explicit Save.
- Provide linear checkpoints for diff, history, restore, and recovery without branches, staging, remotes, rebases, or git conflict terminology.
- Provide a writer-facing CLI with simple commands such as `edda get URL` and `edda save "Note"`.
- Keep AI assistants and project-aware tools fast through SQLite indexes and derived context.
- Let the database be rebuilt from the folder plus `.edda/` metadata.

## Non-Goals

- Git compatibility as a product requirement.
- Branching, rebasing, staging, remote management, or merge commits.
- Multi-author collaborative editing.
- Keystroke-level sync between devices.
- Making every Open Edda feature work without any database at runtime.
- Supporting arbitrary user folder structures without conversion into the Edda layout.

## Source Of Truth

The Edda project folder is canonical. Edda opens a folder such as `/writing/alchemist`, reads Markdown files in place, and writes normal Markdown files when the user saves in the web UI or local CLI workflow.

SQLite stores derived and operational data:

- current file index,
- search rows,
- project maps,
- assistant context caches,
- editor/session state,
- checkpoint metadata mirrors,
- sync state,
- prompt/activity records,
- script runtime records.

If the database is deleted, Edda can rescan the folder and recover the project content and checkpoint history from files and `.edda/`. Data that is not represented in the folder may be lost unless a later design explicitly persists it into `.edda/`.

## Folder Layout

Human-authored content remains ordinary files in one defined layout. The target structure follows the stronger `alchemist` project shape:

```text
AGENTS.md
BOOTSTRAP.md
.agents/
  skills/
story/
  _index.md
  chapter-01.md
characters/
  _index.md
  Protagonist.md
worldbuilding/
  _index.md
  culture/
  magic/
  monsters/
  places/
storyline/
  _index.md
  chapter-01-plan.md
  vol-01-plan.md
drafts/
  chapter-01-draft.md
```

This is the structure Edda is comfortable with. Existing projects can be imported, moved, renamed, or manually refined into this layout before Edda treats them as fully managed projects. The core mapping is:

- `story/` to prose,
- `characters/` to character notes,
- `worldbuilding/**` to lore,
- `storyline/` to planning,
- `drafts/` to draft material,
- top-level guidance files to project instructions and writing guidance,
- `.agents/skills/` to project-local skill source material.

Edda-owned metadata lives in `.edda/`:

```text
.edda/
  project.json
  ids.json
  state.local.json
  drafts/
  checkpoints/
  conflicts/
```

- `.edda/project.json` stores the project title, server link, schema version, and layout version.
- `.edda/ids.json` maps stable Edda item IDs to file paths so renames can be tracked without treating every rename as delete plus add.
- `.edda/state.local.json` stores last-seen hashes, device identity, sync cursors, and pending upload state. It is device-local and should be ignored by cloud-drive sync and checkpoint capture.
- `.edda/drafts/` stores web-editor autosaves that have not been saved into canonical files.
- `.edda/checkpoints/` stores checkpoint manifests and snapshots or deltas.
- `.edda/conflicts/` stores base/local/server versions and resolution records for conflicted files.

The first checkpoint implementation should favor simple full snapshots or content-addressed blobs over clever delta storage. Reliability and easy recovery matter more than storage optimization.

The syncable `.edda/` files are `project.json`, `ids.json`, `checkpoints/`, and resolved conflict records. Drafts, unresolved conflicts, and `state.local.json` are local operational data until explicitly saved or resolved.

## Save Model

Edda separates three concepts:

- **Draft autosave:** while editing in the web UI, changes autosave into an Edda draft store. This protects work in progress but does not update the canonical Markdown file and does not create a checkpoint.
- **Save:** when the user hits the normal Save button, the current draft is written into the project as the current working version. For a local connected project, this updates the Markdown file. For a server-started project, this updates the server's current file state.
- **Checkpoint:** when the user chooses `edda save "Note"` or Save checkpoint in the UI, Edda records a named project-wide snapshot of the current saved files.

Autosave prevents data loss. Save makes intentional file changes. Checkpoint is the larger "remember this state" action used for history, comparison, recovery, and sync.

## CLI Workflow

The CLI should use writing-oriented language:

- `edda get URL` downloads or connects a server project into a local Edda folder.
- `edda status` shows unsaved drafts, changed files, pending uploads, and conflicts in plain language.
- `edda save "Note"` creates a named checkpoint from the current saved folder state and records it in the local pending upload queue.
- `edda send` retries or completes queued pending uploads when a server URL exists.
- `edda take` downloads saved server changes and checkpoint metadata.
- `edda history` lists checkpoints.
- `edda diff` shows changes since the last checkpoint or between two checkpoints.
- `edda restore` restores a file, folder, or project state from a checkpoint.

`edda save` is the main habit: name the moment, preserve it, and share it. `edda send` is for repair and catch-up.

## Server And Local Mobility

A project can begin in the web app or as a local folder.

For a server-started project, `edda get URL` creates a local folder containing current Markdown files and `.edda/` metadata. The local folder then behaves like any other Edda folder.

For a local-started project, Edda initializes `.edda/`, indexes files, and can later connect the folder to a server project. Once connected, checkpoints and saved file state can move between local and server through `edda save`, `edda send`, and `edda take`.

Sync operates on saved file state and checkpoints, not editor drafts or every keystroke.

Server-started projects should generate local `.edda/` metadata during `edda get`. The server owns project identity and checkpoint records; the CLI creates device-local state when it materializes the folder.

## Conflict Handling

Conflict handling is linear and writer-facing.

- If only one side changed a file since the last shared state, Edda applies it automatically.
- If both local and server changed the same file, Edda preserves both versions and marks the file conflicted.
- If one side renames a file while the other edits it, Edda preserves the edited content and records the rename as part of the conflict so the writer chooses the final path.
- If one side deletes a file while the other edits it, Edda keeps the edited version as a conflict candidate instead of silently deleting it.
- If both sides delete the same file, Edda accepts the deletion without creating a conflict.
- Edda stores or exposes base, local, and server versions for comparison.
- The UI and CLI should say what happened in plain language, such as "This file changed both here and on the server."
- Resolution produces a normal saved file.
- After resolution, `edda save "Resolved ..."` records and pushes the resolved state.

There are no branches and no merge vocabulary in the main workflow.

## Backend Components

- **Layout service:** reads `.edda/project.json`, validates the Edda project layout, discovers Markdown files, and classifies them by path.
- **File indexer:** hashes saved files, updates SQLite rows, and detects added, changed, renamed, and deleted files.
- **Draft service:** stores autosaved web edits separately from canonical files until Save.
- **Checkpoint service:** creates named project-wide snapshots, records manifests, supports diff/history/restore, and tracks upload state.
- **Sync service:** exchanges saved files, checkpoints, and metadata with the server using a linear shared-state cursor.
- **Conflict service:** detects divergent local/server edits, preserves base/local/server versions, and exposes resolution operations.
- **Project API:** serves current indexed project content to the web UI and agent tools while respecting file-first ownership.

## Web UI

The web UI should keep authoring comfortable:

- Draft autosave is quiet and frequent.
- Draft autosaves are local to the browser/session or server editor session until Save. They do not sync to other devices as drafts in the first version.
- Save is explicit and writes the current draft into the working project state.
- Save checkpoint is available but not required before ordinary editing.
- History, Compare, Restore, and Conflicts are project tools, not constant obstacles in the editor.
- Assistant actions should work against saved content or an explicit draft context, depending on the action. Any document-changing action must resolve through the same draft/save/checkpoint model.

## AI And Structured Writes

Structured Writes remain valuable, but their target changes from database-owned content revisions to file-backed item versions.

For a document-changing assistant action:

1. The agent reads a known file version or saved content hash.
2. The proposed write targets that version/hash plus a byte range or structured location.
3. If the saved file changed before apply, Edda reports a conflict and asks for reread, preview, or manual resolution.
4. Accepted or direct-applied changes update the draft or saved file according to the active UI action.
5. Checkpoint history records the result only when the user creates a checkpoint.

This keeps version safety without making every text edit a database revision.

## Documentation Migration

The existing docs should be updated so the project has one coherent model:

- Replace `docs/adr/0001-database-source-of-truth.md` with a file-first source-of-truth ADR.
- Add a new ADR for Edda linear checkpoints instead of git.
- Update `docs/superpowers/specs/2026-06-13-writer-vision-design.md` to describe the Edda layout, `.edda/`, drafts, saves, checkpoints, and CLI/server mobility.
- Update `docs/superpowers/specs/2026-06-14-milestone-4-daily-writing-polish-design.md` so editor conflict and revision language matches the draft/save/checkpoint model.
- Update `README.md` so it no longer says story text is durably kept in SQLite.
- Reframe `docs/adr/0008-elysium-layout-for-markdown-interoperability.md` around the `alchemist`-style Edda project layout. Elysium is a conversion source, not the target.
- Review ADRs 0002, 0003, 0007, and 0011 for wording that assumes database-owned revisions. Their core decisions still stand.
- Treat completed implementation plans as historical records. Do not rewrite them as if they are still future plans; create a migration implementation plan that supersedes their database-first assumptions.

## Testing Strategy

Tests should prioritize data safety:

- Import or refine source folders into the Edda layout, then scan that layout reliably.
- Round-trip Markdown files without content drift.
- Rebuild SQLite project indexes from folder state.
- Preserve stable IDs across renames.
- Keep draft autosaves separate from saved files.
- Create checkpoints and restore individual files, folders, and whole project state.
- Show diffs between current files and checkpoints.
- Retry failed checkpoint uploads through `edda send`.
- Pull server changes through `edda take`.
- Detect divergent local/server edits and preserve base/local/server versions.
- Resolve conflicts into normal saved files.
- Verify assistant writes refuse to apply against stale file hashes.

## Initial Policy Decisions

- Syncable `.edda/` data: `project.json`, `ids.json`, `checkpoints/`, and resolved conflict records.
- Device-local `.edda/` data: `state.local.json`, draft autosaves, and unresolved conflict working files.
- Server-started projects generate local `.edda/` metadata during `edda get`; the server does not need to expose a literal `.edda/` archive.
- Browser draft autosaves remain local to the editor session until Save.
- Project-local `.agents/skills` are project files and should be included in checkpoints by default. A later skill dependency model can refine this if checkpoint size or duplication becomes a problem.
- Full snapshots are acceptable for the first implementation. Add content-addressed blobs when real project checkpoint histories show storage or transfer pain.
