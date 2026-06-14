# Writer Skill Library Rewrite Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task. Do not start implementation until the skill library decisions in this plan are accepted.

**Goal:** Build Milestone 3.5: curate and rewrite Writer's Built-In Skill Library from the copied fiction skill collection plus the important project skills, so first-run Writer has useful daily writing skills, optional specialized skills, clear `$` mention names, script-status disclosure, and no unrelated terminal-agent or GM workflow clutter.

**Product Principle:** The `$` picker should feel like a writing assistant's skill shelf, not a dump of every copied prompt. Daily fiction-writing help is enabled by default. Specialized but useful fiction skills are installed as Optional Skills. Skills outside Writer's focus, or skills whose core value depends on script execution, are Archived Skills or deferred until the Skill Script Runtime.

**Inputs:**

- `docs/skills/important/worldbuilding-brainstorm`
- `docs/skills/important/writing-skills`
- `docs/skills/suggested/fiction`
- `docs/roadmap.md`
- `CONTEXT.md`
- `docs/adr/0005-skill-scripts-require-admin-approval.md`

---

## Scope

This plan covers:

- Curating every skill in `docs/skills/important` and `docs/skills/suggested/fiction`.
- Renaming skills when the source name is vague or misleading.
- Rewriting `SKILL.md` files into Writer-native instructions, routing metadata, and author-facing descriptions.
- Creating one new optional Writer-native skill: `$children-stories`.
- Creating a manifest that records each skill's source, final name, default/optional/archive status, script status, and rewrite notes.
- Auditing scripts and deferring script-dependent skills until Milestone 3.6.
- Preserving useful `data/`, `templates/`, and `references/` assets when they improve the skill without forcing script execution.
- Updating docs so admins and authors understand Built-In Skill Library behavior.

This plan does not cover:

- Implementing the Skill Script Runtime.
- Running bundled skill scripts inside Writer.
- Building a remote skill marketplace.
- Building genre packs beyond `$children-stories`.
- Building the deferred Story Consistency Dashboard from `story-zoom`.
- Adding RPG or GM workflows.

## Source Decisions

- Default Skills are enabled on first run because they support common daily fiction writing.
- Optional Skills are installed but disabled by default because they support specialized formats, genres, publishing-adjacent work, or advanced workflows.
- Archived Skills are reviewed and intentionally not shipped in the Built-In Skill Library.
- Script-bearing skills must disclose script status in the skill browser.
- A skill whose core value depends on running scripts is deferred until Milestone 3.6.
- A skill can ship in Milestone 3.5 with scripts present only when the script is an optional accelerator and the skill has a clear Writer-native path without script execution.
- Skill names should describe author intent in the `$` picker. Preserve source names only when they are already clear.

## Built-In Skill Library Decisions

### Default Daily Writing Skills

| Writer Skill | Source | Decision |
| --- | --- | --- |
| `$story-coach` | `story-coach` | Default. Coaching mode that asks questions and does not write prose for the author. |
| `$story-collaborator` | `story-collaborator` | Default. Active co-writing mode that can generate prose, alternatives, continuations, and examples. |
| `$story-sense` | `story-sense` | Default. Broad "what kind of story problem is this?" diagnostic router. |
| `$story-analysis` | `story-analysis` | Default. Structured report for completed chapters, short stories, or selected story text. |
| `$worldbuilding-brainstorm` | `important/worldbuilding-brainstorm` | Default, non-negotiable. Guided lore brainstorming with canon-safe confirmation before recording decisions. |
| `$worldbuilding-check` | `worldbuilding` | Default, renamed. Diagnostic review for thin, inconsistent, or consequence-light settings. |
| `$dialogue-check` | `dialogue` | Default, renamed. Checks flat dialogue, same-voice characters, exposition dumps, and missing subtext. |
| `$character-arc` | `character-arc` | Default. Designs and diagnoses character transformation arcs. |
| `$scene-pacing` | `scene-sequencing` | Default, renamed. Diagnoses scene function, momentum, pacing, and scene/sequel rhythm. |
| `$prose-polish` | `prose-style` | Default, renamed. Sentence-level polish after structure is solid. |
| `$drafting` | `drafting` | Default. Helps break blocks and move first drafts forward. |
| `$revision` | `revision` | Default. Guides ordinary edit passes without endless tinkering. |
| `$outline-coach` | `outline-coach` | Default. Guided outlining without generating outline content for the author. |
| `$outline-collaborator` | `outline-collaborator` | Default. Active outline partner that proposes beats, structures, and alternatives. |
| `$avoid-cliches` | `cliche-transcendence` + `statistical-distance` | Default, merged. Makes generic elements fresher while preserving their story function. |
| `$genre-check` | `genre-conventions` | Default, renamed. Checks whether genre promise, tone, and story ingredients align. |
| `$ending-check` | `endings` | Default, renamed. Diagnoses weak, rushed, arbitrary, predictable, or unearned endings. |
| `$skill-writer` | `important/writing-skills` | Default installed and available to everyone, routed as an authoring tool rather than ordinary prose-editing help. |

### Optional Specialized Skills

| Writer Skill | Source | Decision |
| --- | --- | --- |
| `$revision-planner` | `novel-revision` | Optional, renamed. Major novel-scale revision planning and cascade control. |
| `$emotional-beats` | `key-moments` | Optional, renamed. Builds story structure around essential emotional moments. |
| `$character-names` | `character-naming` | Optional, renamed. Uses data references and manual/agent guidance for culturally and narratively coherent names. |
| `$conlang` | `conlang` + `language-evolution` | Optional. Keeps language construction and language-history guidance together. Script features deferred. |
| `$systemic-worldbuilding` | `systemic-worldbuilding` | Optional. Traces consequences from speculative changes. |
| `$belief-systems` | `belief-systems` | Optional. Designs religions, philosophies, rituals, and belief institutions. |
| `$economic-systems` | `economic-systems` | Optional. Designs scarcity, trade, currencies, markets, and material pressures. |
| `$governance-systems` | `governance-systems` | Optional. Designs states, laws, legitimacy, power transfer, and institutions. |
| `$settlement-design` | `settlement-design` | Optional. Designs towns, cities, stations, districts, and layered places. |
| `$worldbuilding-fragments` | `oblique-worldbuilding` | Optional, renamed. Creates indirect worldbuilding through documents, quotes, inscriptions, and limited perspectives. |
| `$cultural-depth` | `memetic-depth` | Optional, renamed. Builds implied cultural history and texture. |
| `$multi-pov` | `perspectival-constellation` | Optional, renamed. Structures multi-POV stories around meaningful intersections. |
| `$ordinary-pivot` | `positional-revelation` | Optional, renamed. Builds stories where ordinary people become crucial through structural position. |
| `$identity-denial` | `identity-denial` | Optional. Supports self-deception and transformation arcs. |
| `$moral-parallax` | `moral-parallax` | Optional. Supports systemic-exploitation stories and collapsed moral distance. |
| `$underdog-team` | `underdog-unit` | Optional, renamed. Builds misfit teams, rejected departments, and last-chance units facing impossible work. |
| `$flash-fiction` | `flash-fiction` | Optional. Supports compressed stories and microfiction. |
| `$interactive-fiction` | `interactive-fiction` | Optional. Supports branching narrative and agency problems. |
| `$sleep-story` | `sleep-story` | Optional. Supports calm bedtime or meditation-style stories. |
| `$paradox-fables` | `paradox-fables` | Optional. Supports fables that preserve paradox instead of reducing to simple morals. |
| `$societal-evolution` | `multi-order-evolution` | Optional, renamed. Supports multi-generational civilization changes. |
| `$children-stories` | New Writer-native skill | Optional. Supports age-appropriate children's fiction, read-aloud rhythm, gentle tension, and age bands. |
| `$story-dna` | `dna-extraction` | Optional, renamed. Extracts functional story mechanics from existing works for study. |
| `$adapt-story` | `adaptation-synthesis` + `media-adaptation` | Optional, renamed and merged. Adapts functional story DNA into new contexts. |
| `$book-marketing` | `book-marketing` | Optional. Publishing-adjacent blurbs, taglines, query copy, and descriptions. |
| `$sensitivity-check` | `sensitivity-check` | Optional. Flags representation and sensitive-subject risks. |

### Archived Or Deferred Skills

| Source Skill | Decision |
| --- | --- |
| `story-zoom` | Defer as future Story Consistency Dashboard, not a Milestone 3.5 skill. |
| `shared-world` | Archive for now. Its useful parts overlap Story Bible and future collaboration/shared-universe work. |
| `list-builder` | Archive for now. Revisit only if Writer gets random-table or entropy tools. |
| `game-facilitator` | Archive. GM session-running is outside Writer's core. |
| `table-tone` | Archive. Useful for GMs, not core fiction-writing workflow. |
| `world-fates` | Archive/defer. Campaign mechanics and fate rolls depend on runtime/data workflows outside v1. |
| `reverse-outliner` | Defer after script audit. Current value depends on a file/script analysis pipeline that needs a Writer-native runtime and structured intermediate storage. |
| Any script-dependent skill | Defer until Skill Script Runtime if it cannot provide clear value without running scripts. |

## Naming And Description Rules

Every shipped skill must have:

- A `$` mention name that describes author intent.
- A short human description for the skill browser and `$` picker.
- A longer description that explains when to use it and when not to use it.
- Clear distinction between coaching and collaboration modes.
- Clear distinction between brainstorming/building and checking/diagnosis.
- Route metadata aligned to Writer actions and content kinds.

Example:

```yaml
name: dialogue-check
description: Check and improve conversations for distinct voices, subtext, exposition pressure, and dialogue that does more than one job.
route:
  actionKinds: [read_check, rewrite]
  contentKinds: [chapter, selection]
  tags: [dialogue, voice, subtext]
  priority: 80
```

## Writer-Native Rewrite Rules

Rewrite source skills away from terminal-agent assumptions:

- Replace files/folders with Writer concepts:
  - `story/`, `manuscript/`, scene files -> Story Text and Chapters.
  - `characters/`, `worldbuilding/`, wiki files -> Story Bible Entries and Entry Sections.
  - scratch folders and exploration files -> Project Notes.
  - reports tied to selected text -> Attached Notes.
- Replace direct edits with Structured Writes:
  - `append_to_chapter`
  - `insert_into_chapter`
  - `replace_selection`
  - `update_story_bible_entry`
  - `update_entry_section`
- Replace "save output to file" with one of:
  - return in Agent Session;
  - create an Attached Note;
  - create/update a Project Note;
  - propose a Story Bible Entry or Entry Section update after author confirmation.
- Remove instructions that ask the author or agent to run terminal commands.
- Preserve the author as creative authority, especially for canon changes.
- Do not let a skill write or revise Story Text unless the action is explicitly a writing/rewrite/continuation workflow.

## Script Policy For Milestone 3.5

Every script-bearing skill needs a compatibility note.

Classify each script as:

- `removed`: not useful, unsafe, outside product scope, or replaced by instructions.
- `retained-for-runtime`: safe and useful, but not runnable until Milestone 3.6.
- `converted-to-reference`: algorithm is useful but better as human/model-readable instructions.
- `converted-to-data-template`: useful content should be represented as data or templates instead of executable code.
- `deferred-skill`: the skill's core value depends on scripts and should not ship yet.

Audit each retained script for:

- destructive file operations;
- arbitrary filesystem reads/writes;
- network access;
- runtime dependency such as Deno, Node, Graphviz, or external CLIs;
- expected inputs;
- expected outputs;
- whether it can use database-backed Writer inputs later;
- whether output can be a proposal/report instead of a direct mutation.

Milestone 3.5 does not implement script execution. Milestone 3.6 provides the runtime and database-backed script adapters.

## Task 1: Create Library Manifest

**Files:**

- Create: `docs/skills/manifest.md`
- Create or choose final built-in library root, for example `docs/skills/builtin/`

- [ ] Add a manifest row for every skill in `docs/skills/important` and `docs/skills/suggested/fiction`.
- [ ] Record source path, final `$` name, status, category, source dependencies, script status, and rewrite notes.
- [ ] Include Archived Skills so future workers know they were reviewed intentionally.
- [ ] Mark `reverse-outliner` as deferred because the script audit found its current value depends on a file/script analysis pipeline.

## Task 2: Establish Built-In Skill Folder Shape

**Files:**

- Create: `docs/skills/builtin/README.md`
- Create: `docs/skills/builtin/default/`
- Create: `docs/skills/builtin/optional/`
- Create: `docs/skills/builtin/archive-notes/`

- [ ] Document the difference between Default Skills, Optional Skills, and Archived Skills.
- [ ] Document that script-dependent skills are deferred until Milestone 3.6.
- [ ] Document that built-in skills can be disabled by admins or authors through Skill Core.
- [ ] Keep archived notes as documentation, not installed skills.

## Task 3: Rewrite Default Skills

**Files:**

- Create one folder per Default Skill under `docs/skills/builtin/default/`.

- [ ] Rewrite `$story-coach`.
- [ ] Rewrite `$story-collaborator`.
- [ ] Rewrite `$story-sense`.
- [ ] Rewrite `$story-analysis`.
- [ ] Rewrite `$worldbuilding-brainstorm`.
- [ ] Rewrite `$worldbuilding-check`.
- [ ] Rewrite `$dialogue-check`.
- [ ] Rewrite `$character-arc`.
- [ ] Rewrite `$scene-pacing`.
- [ ] Rewrite `$prose-polish`.
- [ ] Rewrite `$drafting`.
- [ ] Rewrite `$revision`.
- [ ] Rewrite `$outline-coach`.
- [ ] Rewrite `$outline-collaborator`.
- [ ] Merge and rewrite `$avoid-cliches`.
- [ ] Rewrite `$genre-check`.
- [ ] Rewrite `$ending-check`.
- [ ] Rewrite `$skill-writer`.

For each rewritten skill:

- [ ] Add Writer-compatible frontmatter.
- [ ] Add clear author-facing description.
- [ ] Add "Use when" and "Do not use when" sections.
- [ ] Remove unsupported terminal/file instructions.
- [ ] Add Writer-native output handling.
- [ ] Add routing metadata.
- [ ] Add script compatibility note if any source script existed.

## Task 4: Rewrite Optional Skills

**Files:**

- Create one folder per Optional Skill under `docs/skills/builtin/optional/`.

- [ ] Rewrite `$revision-planner`.
- [ ] Rewrite `$emotional-beats`.
- [ ] Rewrite `$character-names`.
- [ ] Rewrite `$conlang`, merging `language-evolution`.
- [ ] Rewrite `$systemic-worldbuilding`.
- [ ] Rewrite `$belief-systems`.
- [ ] Rewrite `$economic-systems`.
- [ ] Rewrite `$governance-systems`.
- [ ] Rewrite `$settlement-design`.
- [ ] Rewrite `$worldbuilding-fragments`.
- [ ] Rewrite `$cultural-depth`.
- [ ] Rewrite `$multi-pov`.
- [ ] Rewrite `$ordinary-pivot`.
- [ ] Rewrite `$identity-denial`.
- [ ] Rewrite `$moral-parallax`.
- [ ] Rewrite `$underdog-team`.
- [ ] Rewrite `$flash-fiction`.
- [ ] Rewrite `$interactive-fiction`.
- [ ] Rewrite `$sleep-story`.
- [ ] Rewrite `$paradox-fables`.
- [ ] Rewrite `$societal-evolution`.
- [ ] Create `$children-stories`.
- [ ] Rewrite `$story-dna`.
- [ ] Rewrite `$adapt-story`, merging `media-adaptation`.
- [ ] Rewrite `$book-marketing`.
- [ ] Rewrite `$sensitivity-check`.

## Task 5: Archive Or Defer Out-Of-Scope Skills

**Files:**

- Create notes under `docs/skills/builtin/archive-notes/`.

- [ ] Archive `story-zoom` with pointer to the roadmap's Story Consistency Dashboard.
- [ ] Archive `shared-world` with pointer to future collaboration/shared-universe possibilities.
- [ ] Archive `list-builder` with pointer to future entropy/random-table support.
- [ ] Archive `game-facilitator`.
- [ ] Archive `table-tone`.
- [ ] Archive `world-fates`.
- [ ] Defer `reverse-outliner` with a pointer to a future Writer-native analysis pipeline.
- [ ] Archive `story-idea-generator` with a pointer to future genre/idea expansion.
- [ ] Archive `chapter-drafter` with a pointer to future orchestration support.
- [ ] Archive any skill that fails script-dependence review.

Each archive note should explain why the skill is not included and what would need to change for reconsideration.

## Task 6: Audit Assets And Scripts

**Files:**

- Create: `docs/skills/script-audit.md`

- [ ] Inventory every `scripts/` directory in the source collection.
- [ ] Inventory useful `data/`, `templates/`, and `references/` assets.
- [ ] Decide which data assets are copied into built-in skills.
- [ ] Decide which templates become skill response formats or Writer note templates.
- [ ] Classify every script using the Milestone 3.5 script policy.
- [ ] Add disclosure text to every shipped script-bearing skill.
- [ ] Mark script-dependent skills as deferred.

## Task 7: Verify Import Compatibility

**Files:**

- Create or update Skill Core fixtures/tests after Milestone 3 implementation exists.

- [ ] Add fixtures for representative Default Skills.
- [ ] Add fixtures for representative Optional Skills.
- [ ] Include at least one script-bearing skill with script status disclosure.
- [ ] Include one merged skill such as `$avoid-cliches`.
- [ ] Verify frontmatter parses.
- [ ] Verify route metadata imports.
- [ ] Verify skill files classify as instruction, data, template, reference, script, or other.
- [ ] Verify disabled script status appears in the skill browser API.

## Task 8: Update Product Documentation

**Files:**

- Modify: `docs/roadmap.md`
- Modify: `CONTEXT.md`
- Create or update: `docs/skills/README.md`

- [ ] Link Milestone 3.5 to this plan from the roadmap.
- [ ] Document the default/optional/archive policy.
- [ ] Document the `$` skill mention naming policy.
- [ ] Document that slash commands are commands, `@` is entity mention, and `$` is skill mention.
- [ ] Document that genre-specific skills are a future expansion track, seeded by `$children-stories`.
- [ ] Document that script runtime work belongs to Milestone 3.6.

## Acceptance Criteria

- Every source skill under `docs/skills/important` and `docs/skills/suggested/fiction` is accounted for in the manifest.
- Every shipped skill has a clear Writer-facing `$` name and human description.
- Default Skills are useful for daily fiction writing and do not depend on scripts.
- Optional Skills are useful but specialized, and disabled by default.
- Archived Skills are documented with reasons.
- Script-bearing skills disclose script status.
- Script-dependent skills are deferred until Milestone 3.6.
- `$worldbuilding-brainstorm` ships as a Default Skill.
- `$skill-writer` ships for everyone as a skill-authoring tool with strong script caution.
- `$children-stories` exists as a new Optional Skill.
- Skill fixtures import cleanly through Skill Core once Milestone 3 is available.
