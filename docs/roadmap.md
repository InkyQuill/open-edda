# Writer Roadmap

This roadmap tracks product milestones separately from implementation plans. Detailed task plans live in `docs/superpowers/plans/`.

## Milestone Status

| Milestone | Name | Status | Tracking Plan |
| --- | --- | --- | --- |
| 1 | Project Core | Implemented | `docs/superpowers/plans/2026-06-13-writer-project-core.md` |
| 2 | Agent Core | Implemented | `docs/superpowers/plans/2026-06-13-writer-agent-core.md` |
| 3 | Skill Core | Implemented | `docs/superpowers/plans/2026-06-13-writer-skill-core.md` |
| 3.5 | Elysium Skill Library Rewrite | Implemented | `docs/superpowers/plans/2026-06-14-writer-skill-library-rewrite.md` |
| 3.6 | Skill Script Runtime | Planned | `docs/superpowers/plans/2026-06-14-open-edda-skill-script-runtime.md` |
| 4 | Daily Writing Polish | Planned | Needs dedicated plan |
| Later | Local Sync And Collaboration | Deferred | Needs specs after v1 is stable |

## Milestone 1: Project Core

Auth, Story Projects, Chapters, Story Bible Entries, Entry Sections, Entry Relations, Writing Briefs, Project Notes, Attached Notes, Per-Item Revisions, diffs, and Elysium Layout import/export.

Acceptance target:

- A self-hosted author can create/import/export a Story Project.
- Markdown content is database-backed, revisioned, and exportable through the Elysium layout.
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
- Skill scripts are imported with clear disabled status until the Skill Script Runtime milestone defines auditing, permissions, and database-backed execution scaffolding.

## Milestone 3.5: Elysium Skill Library Rewrite

After Skill Core exists and before Daily Writing Polish, rewrite the current Elysium project skills from:

`/home/inky/Документы/elysium/.agents/skills`

The rewrite should make these skills first-class Writer skills instead of terminal-agent skills. The work is not just import cleanup: each skill needs an importance decision, routing metadata, updated tool names, and structure aligned with Writer's exposed agent tools.

### Rewrite Goals

- Convert each `SKILL.md` to Writer-compatible frontmatter: `name`, `description`, `route.actionKinds`, `route.contentKinds`, `route.tags`, and `route.priority`.
- Replace terminal/file-operation instructions with Writer tools:
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
- Audit every bundled script for destructive behavior, filesystem assumptions, runtime requirements, and whether it remains useful in a database-backed Writer project.
- Defer any skill whose core value depends on running scripts until the Skill Script Runtime exists.
- Keep a script-bearing skill in Milestone 3.5 only when the script is an optional accelerator and the skill can provide clear value through Writer-native agent instructions, data, templates, or references without executing the script.
- Keep scripts only when they are safe, useful, and can be made runnable through Writer's future script runtime.
- Decide which skills should ship with Writer as default authoring aids, which should be optional, and which should be archived.

### Skill Library Decisions

The accepted Default Skill, Optional Skill, Archived Skill, rename, merge, and script-dependence decisions live in `docs/superpowers/plans/2026-06-14-writer-skill-library-rewrite.md`.

The high-level policy is:

- Daily fiction-writing skills are enabled by default.
- Specialized but useful writing skills are installed as optional skills and disabled by default.
- Skills outside Writer's fiction-writing focus, or skills whose core value depends on scripts, are archived or deferred.
- `$` is the skill mention prefix, `/` remains a command prefix, and `@` is reserved for entity mentions.

### Required Output

- A rewritten skill library under a Writer-compatible source folder.
- A manifest documenting default, optional, and archived skills.
- Tests or fixture imports proving every rewritten skill imports through Milestone 3 Skill Core.
- A short compatibility note for each script-heavy skill explaining whether the script was removed, retained for the future script runtime, converted to data/template/reference material, or deferred until database-backed script adapters exist.
- Skill browser disclosure for script-bearing skills, including whether script support is deferred and whether the skill still works through Writer-native agent guidance.

## Milestone 3.6: Skill Script Runtime

Add a safe runtime for audited skill helper scripts after the built-in skill library has been reviewed. The runtime should let useful scripts run against Writer project data without assuming a local Markdown folder is the project source of truth.

Acceptance target:

- Each runnable skill script has an audit record covering destructive operations, filesystem access, network access, runtime dependencies, expected inputs, and expected outputs.
- Scripts run through Writer-provided scaffolding that can fetch chapters, Story Bible Entries, Entry Sections, Project Notes, Attached Notes, and skill assets from the database.
- Scripts cannot directly mutate Story Text or Story Bible content. They return structured proposals, reports, generated data, or draft outputs that the author can review before applying.
- Admin controls can enable, disable, and inspect runnable scripts per built-in or imported skill.
- Skills with missing or disabled script support degrade clearly in the agent session instead of asking the author to run terminal commands.

## Milestone 4: Daily Writing Polish

Editor ergonomics, mobile-friendly layouts, side-panel Attached Notes, better diff/restore UI, export polish, provider disclosure polish, and model-switching UX.

Acceptance target:

- The Writing Workspace feels usable for daily chapter work, not just API validation.
- Selection/cursor workflows are comfortable on desktop and tablet, with phone support for reading, chat, small edits, and triggering actions.
- Revision review, attached notes, model switching, and provider disclosure are visible without dominating the writing surface.

## Later: Local Sync And Collaboration

Build a Local Sync Tool that detects local Markdown changes and replays them into the service database with merge handling. Add multi-author collaboration only after the single-author workflow is stable.

Acceptance target:

- Local Markdown workflows can coexist with the database source of truth through explicit sync/replay.
- Collaboration is considered only after single-author privacy, revisions, agent activity, and export flows are stable.

## Later: Story Consistency Dashboard

Adapt the useful parts of the deferred `story-zoom` skill into a Writer-native consistency dashboard after the main writing workflow is stable. The dashboard should help authors see drift between the Writing Brief, Story Text, Story Bible Entries, Entry Sections, Project Notes, and recent revisions without relying on file watchers or script execution.

Acceptance target:

- Writer can surface likely inconsistencies across story levels using database-backed project state and agent-readable summaries.
- Authors can review suggested consistency fixes before any Story Text or Story Bible changes are applied.
- The workflow remains optional and does not block ordinary drafting or revision.
