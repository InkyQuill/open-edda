# Writer Roadmap

This roadmap tracks product milestones separately from implementation plans. Detailed task plans live in `docs/superpowers/plans/`.

## Milestone Status

| Milestone | Name | Status | Tracking Plan |
| --- | --- | --- | --- |
| 1 | Project Core | Implemented | `docs/superpowers/plans/2026-06-13-writer-project-core.md` |
| 2 | Agent Core | Implemented | `docs/superpowers/plans/2026-06-13-writer-agent-core.md` |
| 3 | Skill Core | Planned | `docs/superpowers/plans/2026-06-13-writer-skill-core.md` |
| 3.5 | Elysium Skill Library Rewrite | Planned | Needs dedicated plan after Milestone 3 lands |
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
- Skill scripts are imported only as inert reference files. Writer does not execute bundled skill scripts in v1.

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
- Decide which skills should ship with Writer as default authoring aids, which should be optional, and which should be archived.

### Current Skill Triage

| Skill | Initial Importance | Rewrite Notes |
| --- | --- | --- |
| `story-collaborator` | Critical | Core fit for Writer chat and generative ideation. Rewrite around Writer quick actions and selected project context. |
| `story-coach` | Critical | Useful as a non-generative coaching mode. Make routing explicit so it does not activate for Continuation/Rewrite unless selected. |
| `story-analysis` | Critical | Strong fit for Read and Check. Convert checklists into report templates and link outputs to Attached Notes. |
| `dialogue` | Critical | High-value Rewrite/Read and Check support. Has scripts; keep them disabled or convert their logic into model-facing checklists. |
| `character-arc` | Critical | Core revision/planning skill. Route to Read and Check, Rewrite, and Story Bible work. |
| `story-sense` | Important | Broad diagnostic framework with data/scripts. Keep data, disable scripts, and split large logic into focused references if needed. |
| `worldbuilding` | Important | Broad worldbuilding diagnostic. Keep as an umbrella skill or routing entry point to narrower worldbuilding skills. |
| `worldbuilding-brainstorm` | Important | Strong fit for project-aware brainstorming. Replace file-update assumptions with `update_story_bible_entry` and `update_entry_section`. |
| `oblique-worldbuilding` | Important | Likely useful for making lore feel indirect and lived-in. Route to Rewrite and Story Bible work. |
| `systemic-worldbuilding` | Important | Useful as architecture for setting logic. Consider merging with `worldbuilding` if overlap is high. |
| `belief-systems` | Optional Important | Good specialized worldbuilding skill. Keep if project needs religion/spiritual institutions. |
| `economic-systems` | Optional Important | Good specialized worldbuilding skill. Keep if economy drives plot or setting plausibility. |
| `governance-systems` | Optional Important | Good specialized worldbuilding skill. Keep for factions, states, law, institutions. |
| `settlement-design` | Optional Important | Useful for place design. Route to Story Bible entries and Read and Check. |
| `memetic-depth` | Optional Important | Useful texture skill; may overlap with `oblique-worldbuilding`. Decide whether to merge. |
| `statistical-distance` | Optional Important | Useful anti-cliche method. Could become a general Rewrite/brainstorming support skill. |
| `character-naming` | Optional | Has data/templates/scripts. Keep only if disabled-script workflow remains useful; otherwise rewrite as a lightweight naming checklist plus data references. |
| `conlang` | Optional | Has data/scripts. Valuable for language-heavy projects, but not core daily writing. Needs disabled-script redesign. |
| `language-evolution` | Optional | Keep as a worldbuilding specialty; may pair with `conlang`. |
| `world-fates` | Optional | Script/data/template heavy and campaign-like. Keep only if the target story workflow needs probabilistic world-state change tracking. |
| `story-idea-generator` | Optional | Useful for new projects, less important for the current editor loop. Route mostly to chat, not document writes. |
| `metabolic-cultures` | Niche | Strong but setting-specific. Keep as optional package, not default. |
| `underdog-unit` | Needs Review | Determine whether it is project-specific, reusable, or archive-only. |

### Skill Decisions To Make

- Default bundle: likely `story-collaborator`, `story-coach`, `story-analysis`, `dialogue`, `character-arc`, `story-sense`, `worldbuilding`, and `worldbuilding-brainstorm`.
- Optional worldbuilding pack: `belief-systems`, `economic-systems`, `governance-systems`, `settlement-design`, `systemic-worldbuilding`, `oblique-worldbuilding`, `memetic-depth`, `statistical-distance`, `language-evolution`.
- Optional generator pack: `character-naming`, `conlang`, `story-idea-generator`, `world-fates`.
- Niche/archive candidates: `metabolic-cultures`, `underdog-unit` unless current projects need them.

### Required Output

- A rewritten skill library under a Writer-compatible source folder.
- A manifest documenting default, optional, and archived skills.
- Tests or fixture imports proving every rewritten skill imports through Milestone 3 Skill Core.
- A short compatibility note for each script-heavy skill explaining whether the script was removed, retained as inert reference, or replaced by a Writer-native workflow.

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
