# Open Edda Roadmap

This roadmap tracks product milestones separately from implementation plans. Detailed task plans live in `docs/superpowers/plans/`.

## Milestone Status

| Milestone | Name | Status | Tracking Plan |
| --- | --- | --- | --- |
| 1 | Project Core | Implemented | `docs/superpowers/plans/2026-06-13-writer-project-core.md` |
| 2 | Agent Core | Implemented | `docs/superpowers/plans/2026-06-13-writer-agent-core.md` |
| 3 | Skill Core | Implemented | `docs/superpowers/plans/2026-06-13-writer-skill-core.md` |
| 3.5 | Elysium Skill Library Rewrite | Implemented | `docs/superpowers/plans/2026-06-14-writer-skill-library-rewrite.md` |
| 3.6 | Skill Script Runtime | Implemented | `docs/superpowers/plans/2026-06-14-open-edda-skill-script-runtime.md` |
| 4 | Daily Writing Polish | In progress | See phase tracker below |
| 5 | File-First Projects And Checkpoints | In Progress | `docs/superpowers/specs/2026-07-01-file-first-checkpoints-design.md` |
| Later | Collaboration | Deferred | Needs specs after single-author file-first workflow is stable |

## Milestone 1: Project Core

Auth, Story Projects, Chapters, Story Bible Entries, Entry Sections, Entry Relations, Writing Briefs, Project Notes, Attached Notes, Per-Item Revisions, diffs, and Elysium Layout import/export. This milestone was implemented under the earlier database-backed content model.

Acceptance target:

- A self-hosted author can create/import/export a Story Project.
- Markdown content is currently database-backed, revisioned, and exportable through the Elysium layout.
- Chapters, Story Bible Entries, Writing Briefs, Project Notes, and Attached Notes exist as first-class content.

## Milestone 2: Agent Core

OpenAI-compatible provider configuration, model variants, prompt assembly, project map, structured retrieval/read tools, Continuation, Rewrite, Read and Check, Direct Apply, preview mode, Structured Writes, conflict handling, Activity Trails, and Prompt Records.

Acceptance target:

- An author can configure a model variant and run project-aware chat plus Continuation, Rewrite, and Read and Check.
- The agent can inspect project context through structured tools instead of receiving the whole project in prompt.
- Writes are revision-safe, conflicts are detected, and Prompt Records/Activity Trails explain what happened.

## Milestone 3: Skill Core

Skill import/install, skill browsing, routing/selecting skills for Agent Sessions, exposing skill instructions/assets to the agent, and clear handling for script-disabled skills.

Acceptance target:

- An author can import project skills from zip archives or server-local skill folders.
- The UI can browse installed skills, files, routing hints, and script-disabled status.
- Agent Sessions and quick actions can select skills.
- The model sees available/selected skill guidance and can load full skill content through a bounded `skill` tool.
- Skill scripts are imported with clear disabled status until the Skill Script Runtime milestone defines auditing, permissions, and service-prepared execution scaffolding.

## Milestone 3.5: Elysium Skill Library Rewrite

After Skill Core exists and before Daily Writing Polish, rewrite the current Elysium project skills from:

`/home/inky/Документы/elysium/.agents/skills`

The rewrite should make these skills first-class Open Edda skills instead of terminal-agent skills. The work is not just import cleanup: each skill needs an importance decision, routing metadata, updated tool names, and structure aligned with Open Edda's exposed agent tools.

### Rewrite Goals

- Convert each `SKILL.md` to Open Edda-compatible frontmatter: `name`, `description`, `route.actionKinds`, `route.contentKinds`, `route.tags`, and `route.priority`.
- Replace terminal/file-operation instructions with Open Edda tools:
  - `project_map`
  - `search_content`
  - `read_content`
  - `read_chapter`
  - `read_story_bible_entry`
  - `read_entry_section`
  - `list_revisions`
  - `append_to_chapter`
  - `insert_into_chapter`
  - `replace_selection`
  - `update_story_bible_entry`
  - `update_entry_section`
  - `skill`
- Remove or rewrite references to shell commands, local script execution, direct filesystem edits, and terminal-agent-only assumptions.
- Move reusable examples/checklists into `templates/`, `references/`, or `data/` where useful.
- Preserve script files only when they are useful as reference algorithms; mark them as disabled and rewrite instructions so the model does not ask to execute them.
- Audit every bundled script for destructive behavior, filesystem assumptions, runtime requirements, and whether it remains useful through Open Edda's service-prepared script inputs.
- Defer any skill whose core value depends on running scripts until the Skill Script Runtime exists.
- Keep a script-bearing skill in Milestone 3.5 only when the script is an optional accelerator and the skill can provide clear value through Edda-native agent instructions, data, templates, or references without executing the script.
- Keep scripts only when they are safe, useful, and can be made runnable through Open Edda's future script runtime.
- Decide which skills should ship with Open Edda as default authoring aids, which should be optional, and which should be archived.

### Skill Library Decisions

The accepted Default Skill, Optional Skill, Archived Skill, rename, merge, and script-dependence decisions live in `docs/superpowers/plans/2026-06-14-writer-skill-library-rewrite.md`.

The high-level policy is:

- Daily fiction-writing skills are enabled by default.
- Specialized but useful writing skills are installed as optional skills and disabled by default.
- Skills outside Open Edda's fiction-writing focus, or skills whose core value depends on scripts, are archived or deferred.
- `$` is the skill mention prefix, `/` remains a command prefix, and `@` is reserved for entity mentions.

### Required Output

- A rewritten skill library under a Open Edda-compatible source folder.
- A manifest documenting default, optional, and archived skills.
- Tests or fixture imports proving every rewritten skill imports through Milestone 3 Skill Core.
- A short compatibility note for each script-heavy skill explaining whether the script was removed, retained for the future script runtime, converted to data/template/reference material, or deferred until service-backed script adapters exist.
- Skill browser disclosure for script-bearing skills, including whether script support is deferred and whether the skill still works through Edda-native agent guidance.

## Milestone 3.6: Skill Script Runtime

Add a safe runtime for audited skill helper scripts after the built-in skill library has been reviewed. The runtime should let useful scripts run against service-prepared Open Edda project data without giving scripts direct write access to project files or `.edda/` metadata.

Acceptance target:

- Each runnable skill script has an audit record covering destructive operations, filesystem access, network access, runtime dependencies, expected inputs, and expected outputs.
- Scripts run through Open Edda-provided scaffolding that can fetch chapters, Story Bible Entries, Entry Sections, Project Notes, Attached Notes, and skill assets from indexed project data.
- Scripts cannot directly mutate Story Text, Story Bible content, project files, or `.edda/` metadata. They return structured proposals, reports, generated data, or draft outputs that the author can review before applying.
- Admin controls can enable, disable, and inspect runnable scripts per built-in or imported skill.
- Skills with missing or disabled script support degrade clearly in the agent session instead of asking the author to run terminal commands.

## Milestone 4: Daily Writing Polish

Editor ergonomics, mobile-friendly layouts, system settings, project/content creation flows, side-panel Attached Notes, better diff/restore UI, export polish, provider disclosure polish, and assistant chat UX.

Acceptance target:

- The Writing Workspace feels usable for daily chapter work, not just API validation.
- Selection/cursor workflows are comfortable on desktop and tablet, with phone support for reading, chat, small edits, and triggering actions.
- Revision review, attached notes, model availability, and provider disclosure are visible without dominating the writing surface.
- Provider configuration, model catalog selection, and skill administration live in system/project settings, not in the writing workspace's assistant drawer.
- Assistant mode keeps the right panel focused on chat only.

### Milestone 4 Phase Tracker

Detailed design context lives in `docs/superpowers/specs/2026-06-14-milestone-4-daily-writing-polish-design.md`.

| Phase | Scope | Status | Plan |
| --- | --- | --- | --- |
| 1 | Routed workspace foundation: React Router, Redux, Tailwind v4, shadcn/ui, responsive shell, editor-local action shells | Implemented | `docs/superpowers/plans/2026-06-14-milestone-4-workspace-foundation.md` |
| 2 | Behavior parity and data slices: move old monolithic assistant/settings/skills/activity behavior into routed vertical slices | Implemented | `docs/superpowers/plans/2026-06-14-milestone-4-behavior-parity.md` |
| 3 | Editor adapter: replace read-only textarea assumptions with an editor boundary prepared for Galley integration and mutation-safe cursor/selection APIs | Implemented | `docs/superpowers/plans/2026-06-17-milestone-4-editor-adapter.md` |
| 3.5 | Information architecture correction: move provider/model/skill administration to settings, make assistant drawer chat-only, add project/content creation controls, and redesign the projects page | Implemented | `docs/superpowers/plans/2026-06-17-milestone-4-system-settings-and-ia.md` |
| 4 | Assistant actions: wire Generate, Rewrite, Check, preview, accept/reject, and version-safe conflict handling from the editor-local controls | Implemented | `docs/superpowers/plans/2026-06-17-milestone-4-assistant-actions.md` |
| 5 | Review surfaces: checkpoints/history, diff/restore, attached notes, activity, prompt records, and review-oriented drawer workflows | Implemented | `docs/superpowers/plans/2026-07-01-milestone-4-review-surfaces.md` |
| 6 | Mobile and browser smoke hardening: sheet behavior, persistence, responsive ergonomics, and Playwright/browser coverage | Implemented | `docs/superpowers/plans/2026-07-01-milestone-4-mobile-browser-smoke.md` |

## Milestone 5: File-First Projects And Checkpoints

Move Open Edda from database-owned prose toward the defined Edda project layout plus `.edda/` metadata. Add the local CLI and lightweight linear checkpoint model described in `docs/superpowers/specs/2026-07-01-file-first-checkpoints-design.md`.

Acceptance target:

- A writer can start from the web app or from a local folder that already follows, or is converted into, the `alchemist`-style Edda layout.
- Story prose, storyline/planning material, characters, worldbuilding, drafts, project guidance, and project-local skills are ordinary files in the defined layout.
- SQLite indexes and caches project data, but project content and checkpoint history can be rebuilt from the folder plus `.edda/`.
- `edda get`, `edda status`, `edda save`, `edda send`, `edda take`, `edda history`, `edda diff`, and `edda restore` cover the main local workflow.
- Checkpoints provide project-wide history, comparison, restore, recovery, and sync without exposing git branches, staging, rebases, or remote-management concepts.
- Conflicts preserve base/local/server versions and resolve back into normal saved files.

### Milestone 5 Phase Tracker

| Phase | Scope | Status | Plan |
| --- | --- | --- | --- |
| 1 | File-first layout foundation: scan the `alchemist`-style Edda folder structure, read/write `.edda/project.json`, and add `edda init/status` CLI skeleton | Implemented | `docs/superpowers/plans/2026-07-01-milestone-5-layout-foundation.md` |
| 2 | File index and stable IDs: rebuild SQLite index rows from files, hash saved content, and preserve IDs across renames via `.edda/ids.json` | Planned | Create after Phase 1 is implemented |
| 3 | Draft/save model: separate browser/server draft autosaves from canonical file writes and update web Save semantics | Planned | Create after Phase 2 is implemented |
| 4 | Linear checkpoints: create, list, diff, and restore project-wide snapshots using `.edda/checkpoints/` | Planned | Create after Phase 3 is implemented |
| 5 | CLI sync workflow: implement `edda get`, `send`, `take`, server connection metadata, pending upload state, and retry behavior | Planned | Create after Phase 4 is implemented |
| 6 | Conflict preservation and resolution: detect divergent saved file edits, preserve base/local/server versions, and resolve back to normal files | Planned | Create after Phase 5 is implemented |
| 7 | Agent/write migration: move structured writes and review surfaces from database revisions toward saved file hashes and checkpoints | Planned | Create after Phase 6 is implemented |

## Later: Collaboration

Add multi-author collaboration only after the single-author file-first workflow is stable.

Acceptance target:

- Collaboration is considered only after single-author privacy, checkpoints, agent activity, file mobility, and recovery flows are stable.

## Later: Story Consistency Dashboard

Adapt the useful parts of the deferred `story-zoom` skill into an Edda-native consistency dashboard after the main writing workflow is stable. The dashboard should help authors see drift between the Writing Brief, Story Text, Story Bible Entries, Entry Sections, Project Notes, and recent checkpoints without relying on raw file watchers or script execution.

Acceptance target:

- Open Edda can surface likely inconsistencies across story levels using indexed file-backed project state and agent-readable summaries.
- Authors can review suggested consistency fixes before any Story Text or Story Bible changes are applied.
- The workflow remains optional and does not block ordinary drafting or revision.
