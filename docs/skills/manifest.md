# Edda Built-In Skill Library Manifest

This manifest records the Milestone 3.5 disposition for every source `SKILL.md` under `docs/skills/important` and `docs/skills/suggested/fiction`. It is the planning surface for the Edda-native rewrites; it is not the final installable library yet.

Invocation syntax reminder for future workers: `$` is a skill mention, `/` is a command, and `@` is an entity mention.

## Default Skills

| Edda skill/source skill | Source path(s) | Status | Category | Script status | Asset/source dependency notes | Rewrite notes |
| --- | --- | --- | --- | --- | --- | --- |
| `$story-coach` | `docs/skills/suggested/fiction/core/story-coach` | Default built-in | Core coaching | No scripts | Skill text only | Keep it in coaching mode only; no prose generation default. |
| `$story-collaborator` | `docs/skills/suggested/fiction/core/story-collaborator` | Default built-in | Core co-writing | No scripts | Skill text only | Rewrite as active co-writing partner with Edda guardrails. |
| `$story-sense` | `docs/skills/suggested/fiction/core/story-sense` | Default built-in | Core diagnosis/router | Optional helper scripts | `data/` genre and function references; optional `scripts/` helpers | Preserve broad diagnostic routing; scripts stay non-required in 3.5. |
| `$story-analysis` | `docs/skills/suggested/fiction/core/story-analysis` | Default built-in | Core analysis | No scripts | Skill text only | Keep structured story and chapter analysis workflow. |
| `$worldbuilding-brainstorm` | `docs/skills/important/worldbuilding-brainstorm` | Default built-in | Worldbuilding | No scripts | `WORLD-FILE-FORMAT.md`, agent config | Non-negotiable default. Rewrite as canon-safe lore brainstorming inside Edda. |
| `$worldbuilding-check` | `docs/skills/suggested/fiction/worldbuilding/worldbuilding` | Default built-in | Worldbuilding diagnosis | No scripts | Skill text only | Rename from source skill to a diagnostic worldbuilding check. |
| `$dialogue-check` | `docs/skills/suggested/fiction/character/dialogue` | Default built-in | Character/dialogue | Optional helper scripts | Optional `scripts/` audits | Rewrite around dialogue diagnosis and improvement, not autonomous rewrite. |
| `$character-arc` | `docs/skills/suggested/fiction/character/character-arc` | Default built-in | Character | No scripts | Skill text only | Keep arc tracking and lie/want/need analysis. |
| `$scene-pacing` | `docs/skills/suggested/fiction/structure/scene-sequencing` | Default built-in | Structure | Optional helper scripts | Optional `scripts/` analyzer | Rename around scene pacing and sequencing. |
| `$prose-polish` | `docs/skills/suggested/fiction/craft/prose-style` | Default built-in | Craft/prose | Optional helper scripts | Optional `scripts/` rhythm and prose checks | Rewrite as Edda polish/editing aid, not a detached prose-style prompt. |
| `$drafting` | `docs/skills/suggested/fiction/craft/drafting` | Default built-in | Craft/drafting | No scripts | Skill text only | Keep practical drafting support for getting words on the page. |
| `$revision` | `docs/skills/suggested/fiction/craft/revision` | Default built-in | Craft/revision | Optional helper scripts | Optional `scripts/` revision audit | Preserve revision workflow; script remains helper-only until 3.6. |
| `$outline-coach` | `docs/skills/suggested/fiction/structure/outline-coach` | Default built-in | Structure/outline | No scripts | Skill text only | Keep advisory outlining mode. |
| `$outline-collaborator` | `docs/skills/suggested/fiction/structure/outline-collaborator` | Default built-in | Structure/outline | No scripts | Skill text only | Keep active outlining collaboration mode. |
| `$avoid-cliches` | `docs/skills/suggested/fiction/craft/cliche-transcendence`<br>`docs/skills/suggested/fiction/character/statistical-distance` | Default built-in | Craft/originality | Optional helper scripts | Optional `scripts/` orthogonality check from `cliche-transcendence` | Merge both sources into one practical originality skill focused on avoiding defaults and flattening. |
| `$genre-check` | `docs/skills/suggested/fiction/craft/genre-conventions` | Default built-in | Craft/genre | Optional helper scripts | `data/genre-elements.json`; optional `scripts/` | Rename around genre fit and expectation checks. |
| `$ending-check` | `docs/skills/suggested/fiction/structure/endings` | Default built-in | Structure/endings | Optional helper scripts | Optional `scripts/` payoff checks | Rewrite as ending/payoff diagnosis rather than a standalone endings prompt. |
| `$skill-writer` | `docs/skills/important/writing-skills` | Default built-in | Authoring tool | Source scripts/assets are not Milestone 3.5 runtime helpers; executable support is deferred and not required to ship | Reference docs and local helper script (`render-graphs.js`) | Planned Default Skill authoring tool available to everyone after rewrite/import integration. Script guidance must stay conservative, and no executable helper should be required for the 3.5 rewrite to ship. |

## Optional Skills

| Edda skill/source skill | Source path(s) | Status | Category | Script status | Asset/source dependency notes | Rewrite notes |
| --- | --- | --- | --- | --- | --- | --- |
| `$revision-planner` | `docs/skills/suggested/fiction/structure/novel-revision` | Optional built-in | Structure/revision | No scripts | Skill text only | Keep as a revision-planning workflow, disabled by default. |
| `$emotional-beats` | `docs/skills/suggested/fiction/structure/key-moments` | Optional built-in | Structure/emotion | No scripts | Skill text only | Rename around emotional beat shaping. |
| `$character-names` | `docs/skills/suggested/fiction/character/character-naming` | Optional built-in | Character/naming | Deferred helper scripts | Large `data/` sets, `templates/`, framework guide, optional `scripts/` | Rename for practical naming help; keep datasets valuable, defer runtime helpers to 3.6. |
| `$conlang` | `docs/skills/suggested/fiction/worldbuilding/conlang`<br>`docs/skills/suggested/fiction/worldbuilding/language-evolution` | Optional built-in | Worldbuilding/language | Deferred helper scripts | `conlang/data/`; optional `conlang/scripts/` | Merge language invention plus evolution guidance. Scripts remain deferred helpers. |
| `$systemic-worldbuilding` | `docs/skills/suggested/fiction/worldbuilding/systemic-worldbuilding`<br>`docs/skills/suggested/fiction/worldbuilding/metabolic-cultures` | Optional built-in | Worldbuilding/systems | No required scripts | Text-only sources | Use `metabolic-cultures` as a specialized systems/world-society source within the broader rewrite. |
| `$belief-systems` | `docs/skills/suggested/fiction/worldbuilding/belief-systems` | Optional built-in | Worldbuilding/culture | No scripts | Skill text only | Keep as optional deep worldbuilding support. |
| `$economic-systems` | `docs/skills/suggested/fiction/worldbuilding/economic-systems` | Optional built-in | Worldbuilding/culture | No scripts | Skill text only | Keep as optional economy/world logic support. |
| `$governance-systems` | `docs/skills/suggested/fiction/worldbuilding/governance-systems` | Optional built-in | Worldbuilding/culture | No scripts | Skill text only | Keep as optional governance/power-structure support. |
| `$settlement-design` | `docs/skills/suggested/fiction/worldbuilding/settlement-design` | Optional built-in | Worldbuilding/place | No scripts | Skill text only | Keep as optional place-design workflow. |
| `$worldbuilding-fragments` | `docs/skills/suggested/fiction/worldbuilding/oblique-worldbuilding` | Optional built-in | Worldbuilding | No scripts | Skill text only | Rename around fragmentary, indirect worldbuilding. |
| `$cultural-depth` | `docs/skills/suggested/fiction/character/memetic-depth` | Optional built-in | Character/culture | No scripts | Skill text only | Reframe as culture and inherited-pattern depth for characters/societies. |
| `$multi-pov` | `docs/skills/suggested/fiction/structure/perspectival-constellation` | Optional built-in | Structure/POV | No scripts | Skill text only | Rename for clear Edda-facing POV guidance. |
| `$ordinary-pivot` | `docs/skills/suggested/fiction/structure/positional-revelation` | Optional built-in | Structure | No scripts | Skill text only | Keep as a specialized reveal/turn skill. |
| `$identity-denial` | `docs/skills/suggested/fiction/structure/identity-denial` | Optional built-in | Structure/character | No scripts | Skill text only | Keep specialized identity-conflict framing. |
| `$moral-parallax` | `docs/skills/suggested/fiction/structure/moral-parallax` | Optional built-in | Structure/theme | No scripts | Skill text only | Keep as optional moral-perspective shaping skill. |
| `$underdog-team` | `docs/skills/suggested/fiction/character/underdog-unit` | Optional built-in | Character/team dynamics | No scripts | Skill text only | Rename for clearer Edda-facing team-story use. |
| `$flash-fiction` | `docs/skills/suggested/fiction/application/flash-fiction` | Optional built-in | Application/form | No scripts | Skill text only | Keep as a form-specific optional skill. |
| `$interactive-fiction` | `docs/skills/suggested/fiction/application/interactive-fiction` | Optional built-in | Application/form | No scripts | Skill text only | Keep as optional branching-fiction support. |
| `$sleep-story` | `docs/skills/suggested/fiction/application/sleep-story` | Optional built-in | Application/form | No scripts | Skill text only | Keep as optional soothing-form support. |
| `$paradox-fables` | `docs/skills/suggested/fiction/application/paradox-fables` | Optional built-in | Application/form | No scripts | Skill text only | Keep as optional specialized form skill. |
| `$societal-evolution` | `docs/skills/suggested/fiction/application/multi-order-evolution` | Optional built-in | Application/worldbuilding | No scripts | Skill text only | Rename for clearer long-horizon civilization design support. |
| `$children-stories` | No source folder yet | Optional built-in (Edda-native) | Application/form | No scripts yet | New Edda-native skill; source TBD | Reserve slot now so later workers can author the rewrite in place. |
| `$story-dna` | `docs/skills/suggested/fiction/application/dna-extraction` | Optional built-in | Application/analysis | Deferred helper scripts | `data/` extraction templates; optional `scripts/` | Rename around extracting reusable story DNA. Keep scripts optional/deferred. |
| `$adapt-story` | `docs/skills/suggested/fiction/application/adaptation-synthesis`<br>`docs/skills/suggested/fiction/application/media-adaptation` | Optional built-in | Application/adaptation | Deferred helper scripts | `adaptation-synthesis/data/`; optional `scripts/` | Merge adaptation planning sources. Preserve useful framework, defer script runtime dependencies. |
| `$book-marketing` | `docs/skills/suggested/fiction/application/book-marketing` | Optional built-in | Publishing-adjacent | No scripts | `templates/` for blurbs/query/taglines | Keep optional and clearly publishing-adjacent. |
| `$sensitivity-check` | `docs/skills/suggested/fiction/application/sensitivity-check` | Optional built-in | Application/review | Deferred helper scripts | Optional `scripts/` audits | Keep as optional review aid; scripts remain deferred helpers until 3.6. |

## Archived / Deferred

| Edda skill/source skill | Source path(s) | Status | Category | Script status | Asset/source dependency notes | Rewrite notes |
| --- | --- | --- | --- | --- | --- | --- |
| `None - deferred` | `docs/skills/suggested/fiction/structure/story-zoom` | Archived/deferred | Structure/dashboard | Deferred; runtime dependent | `templates/` plus `scripts/` watcher/status tools | No final 3.5 built-in `$` name. Hold for a future Story Consistency Dashboard instead of a 3.5 built-in skill. |
| `None - deferred` | `docs/skills/suggested/fiction/structure/reverse-outliner` | Archived/deferred | Structure/analysis | Deferred; script pipeline dependent | `data/` plus multiple `scripts/` | No final 3.5 built-in `$` name. Defer until a Edda-native analysis pipeline can replace the current file/script workflow. |
| `None - archived` | `docs/skills/suggested/fiction/application/shared-world` | Archived/deferred | Collaboration/world bible | Deferred; script-heavy | `templates/` and multi-script workflow | No final 3.5 built-in `$` name. Archive for now; overlaps future Story Bible/collaboration surfaces. |
| `None - archived` | `docs/skills/suggested/fiction/application/list-builder` | Archived/deferred | Utility/reference building | Deferred; script-assisted | Validation script plus reference docs | No final 3.5 built-in `$` name. Archive for now; revisit with entropy/random-table support later. |
| `None - archived` | `docs/skills/suggested/fiction/application/game-facilitator` | Archived/deferred | GM/runtime play | Deferred; script-heavy | Session/NPC/complication scripts | No final 3.5 built-in `$` name. Out of Edda core scope; archive. |
| `None - archived` | `docs/skills/suggested/fiction/application/table-tone` | Archived/deferred | GM/runtime play | No scripts, but out of scope | Skill text only | No final 3.5 built-in `$` name. Archive as GM-oriented rather than fiction-writing core. |
| `None - deferred` | `docs/skills/suggested/fiction/worldbuilding/world-fates` | Archived/deferred | World simulation | Deferred; script/runtime dependent | `data/`, `templates/`, and multiple `scripts/` | No final 3.5 built-in `$` name. Defer until Skill Script Runtime exists. |
| `None - archived` | `docs/skills/suggested/fiction/core/story-idea-generator` | Archived/deferred | Core ideation | No scripts | Skill text only | No final 3.5 built-in `$` name. Reviewed intentionally but not shipping in 3.5; overlaps default coaching/collaboration/drafting set and needs Edda-native reframing. |
| `None - archived` | `docs/skills/suggested/fiction/orchestrators/chapter-drafter` | Archived/deferred | Orchestrator/autodrafting | Deferred; script and orchestration dependent | `templates/` and scoring script | No final 3.5 built-in `$` name. Archive for now; autonomous multi-pass drafting is outside the 3.5 built-in library shape. |

## Source Coverage Notes

- Verified source inventory: `56` source `SKILL.md` files under `docs/skills/important` and `docs/skills/suggested/fiction`.
- Every source `SKILL.md` is represented here as one of:
  - a planned Default Skill rewrite,
  - a planned Optional Skill rewrite,
  - a merged source feeding a rewritten Edda skill,
  - or an archived/deferred source reviewed intentionally.
- Merge coverage recorded explicitly:
  - `$avoid-cliches` covers `cliche-transcendence` and `statistical-distance`.
  - `$conlang` covers `conlang` and `language-evolution`.
  - `$systemic-worldbuilding` absorbs `metabolic-cultures` as a specialized source.
  - `$adapt-story` covers `adaptation-synthesis` and `media-adaptation`.
- Deferred coverage recorded explicitly:
  - `reverse-outliner` is deferred after script audit because the current value depends on a file/script analysis pipeline.
- Edda-native placeholder:
  - `$children-stories` has no source folder yet and is included so the built-in library shape matches the accepted optional set without implying source coverage from existing folders.
- Milestone 3.5 rule:
  - Script-dependent helpers are documentation inputs only in this milestone unless the manifest row explicitly says they are optional helpers rather than runtime requirements.
