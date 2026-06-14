# Edda Milestone 3.5 Script And Asset Audit

## 1. Summary and policy

Milestone 3.5 does not execute repository scripts. Script-bearing source skills can still ship in 3.5 when their value survives as Edda-native guidance, built-in data, note templates, or human/model-readable references. Skills whose core workflow depends on filesystem mutation, long-running watchers, multi-step orchestration, or direct CLI pipelines defer to Milestone 3.6.

Milestone 3.6 implements the runtime scaffolding for `retained-for-runtime` and future approved helpers. It does not automatically revive deferred skills. `reverse-outliner`, `story-zoom`, and `world-fates` still need product-specific adapters before they should become installed skills.

Milestone 3.5 script policy used in this audit:

- `removed`: not useful for Edda, unsafe, outside product scope, or replaced cleanly by instructions.
- `retained-for-runtime`: safe and useful, but should wait for the 3.6 script runtime.
- `converted-to-reference`: keep the logic as prose guidance, checklists, scoring rules, or response formats.
- `converted-to-data-template`: keep the useful content as built-in data, schemas, or note templates rather than code.
- `deferred-skill`: the surrounding skill is materially script-dependent and should not ship in 3.5.

Hard product constraints carried into every decision:

- No script should directly mutate Story Text or Story Bible content in the future runtime.
- Future runtime adapters should prefer proposals, reports, generated metadata, or drafts for review.
- Database-backed Edda inputs are the preferred future source for structured records instead of arbitrary local file reads.

Audit scope note: the required `find ... -name scripts` inventory catches every conventional `scripts/` directory under the source roots, but `docs/skills/important/writing-skills/render-graphs.js` is also a script-bearing source and is included here explicitly so source coverage is complete.

## 2. Script inventory table

| Skill/source path | Scripts | Runtime | Risk notes | Classification | Runtime adapter need |
| --- | --- | --- | --- | --- | --- |
| `docs/skills/important/writing-skills` | `render-graphs.js` -> Node + Graphviz `dot`; reads `SKILL.md`, shells out with `execSync`, writes SVG and `.dot` files | Node, Graphviz CLI | Child process execution; arbitrary local reads within chosen skill dir; filesystem writes; authoring-only helper outside Edda end-user scope | `removed` | None in 3.5. If ever revived, it belongs in an internal authoring toolchain, not built-in skill runtime. |
| `docs/skills/suggested/fiction/application/adaptation-synthesis` | `form-options.ts` -> Deno, read-only generator from embedded/data-backed form suggestions; `synthesize.ts` -> Deno interactive synthesis CLI over DNA JSON inputs; `validate-synthesis.ts` -> Deno validator for synthesis docs | Deno | Read-only today; expects external JSON files/CLI args; no direct mutation; future fit is structured Edda inputs plus report output | `form-options.ts`: `converted-to-data-template`.<br>`synthesize.ts`: `converted-to-reference`.<br>`validate-synthesis.ts`: `converted-to-reference`. | No 3.6 adapter required for 3.5 shipping. If runtime exists later, expose proposal/report generation against database-backed DNA records. |
| `docs/skills/suggested/fiction/application/dna-extraction` | `emotional-beat-map.ts` -> Deno beat-map/template generator; `extract-functions.ts` -> Deno structured extractor/classifier; `structural-stylistic.ts` -> Deno structure/style analyzer | Deno | Read-only analyzers over manuscript text or structured extraction files; no destructive ops; CLI-oriented inputs | `emotional-beat-map.ts`: `converted-to-data-template`.<br>`extract-functions.ts`: `converted-to-reference`.<br>`structural-stylistic.ts`: `converted-to-reference`. | No 3.6 adapter required for base skill. Later adapter could emit draft DNA notes and reviewable extraction records from selected chapters/documents. |
| `docs/skills/suggested/fiction/application/game-facilitator` | `complication-generator.ts` -> Deno prompt/entropy generator; `npc-generator.ts` -> Deno NPC generator; `session-notes.ts` -> Deno note-template generator with optional file write | Deno | `session-notes.ts` writes markdown; manifest archives the parent skill as GM/runtime-play content outside Edda 3.5 core scope; useful note structure can survive without shipping the skill | `complication-generator.ts`: `removed`.<br>`npc-generator.ts`: `removed`.<br>`session-notes.ts`: `converted-to-data-template`. | No adapter for 3.5. Revisit only if Edda later grows TTRPG support. |
| `docs/skills/suggested/fiction/application/list-builder` | `validate-list.ts` -> Deno quality report for JSON entropy lists | Deno | Read-only validator, but the manifest archives the parent skill as utility/reference-building outside the 3.5 built-in library; keep only the criteria as reference guidance | `removed` | No 3.6 adapter planned for the built-in library unless the archived skill is later revived. |
| `docs/skills/suggested/fiction/application/sensitivity-check` | `representation-map.ts` -> Deno structured identity/agency map from character JSON; `sensitivity-audit.ts` -> Deno manuscript audit | Deno | Read-only audit/reporting; sensitive domain means future output should remain advisory, never auto-editing text; expects JSON/files today | `representation-map.ts`: `converted-to-data-template`.<br>`sensitivity-audit.ts`: `converted-to-reference`. | Optional 3.6 adapter could read Edda character records and draft a review memo instead of touching manuscript text. |
| `docs/skills/suggested/fiction/application/shared-world` | `add-entry.ts` -> Deno writes entry/changelog/index files; `build-index.ts` -> Deno rewrites index from folder walk; `check-conflicts.ts` -> Deno read-only world-bible conflict scan; `init-world.ts` -> Deno scaffolds full world-bible directory tree | Deno plus std fs/walk helpers | Direct filesystem mutation, arbitrary tree reads, project scaffolding, and index/changelog rewrites; manifest archives the parent skill for 3.5 because it overlaps future Story Bible/collaboration product surfaces | `add-entry.ts`: `removed`.<br>`build-index.ts`: `removed`.<br>`check-conflicts.ts`: `removed`.<br>`init-world.ts`: `removed`. | No 3.5 adapter. Any revival belongs to a future Story Bible surface with structured entities, conflict reports, and review-gated proposals. |
| `docs/skills/suggested/fiction/character/character-naming` | `cast-tracker.ts` -> Deno cast JSON init/add/check/distribution with writes; `character-name.ts` -> Deno generator over local culture/pool/preset data | Deno | `cast-tracker.ts` writes project-local JSON; `character-name.ts` performs broad local reads across bundled datasets and optional cast file; both are safe conceptually but should use Edda records later | `cast-tracker.ts`: `converted-to-data-template`.<br>`character-name.ts`: `converted-to-data-template`. | No required adapter for 3.5 shipping. A later adapter could query Edda character roster and generate candidate names/collision reports from built-in datasets. |
| `docs/skills/suggested/fiction/character/dialogue` | `dialogue-audit.ts` -> Deno function/anti-pattern audit; `voice-check.ts` -> Deno voice differentiation audit | Deno | Read-only analysis of scene/manuscript text; no mutation; output is already a report | `dialogue-audit.ts`: `converted-to-reference`.<br>`voice-check.ts`: `converted-to-reference`. | No 3.6 adapter required. Optional later adapter could produce structured review cards from selected scenes. |
| `docs/skills/suggested/fiction/core/story-sense` | `entropy.ts` -> Deno entropy/randomness diagnostics over text/files; `functions.ts` -> Deno story-function lookup/classifier using bundled data | Deno | Read-only analyzers; no destructive ops; most durable value is in rules plus data tables rather than CLI | `entropy.ts`: `converted-to-reference`.<br>`functions.ts`: `converted-to-data-template`. | No required adapter for 3.5. Later adapter could compute diagnostics from selected Edda documents and emit reports only. |
| `docs/skills/suggested/fiction/craft/cliche-transcendence` | `orthogonality-check.ts` -> Deno interactive questionnaire/scoring helper | Deno | Interactive stdin/stdout only; no mutation; logic is better expressed as guided questioning inside the skill itself | `converted-to-reference` | No 3.6 adapter needed. |
| `docs/skills/suggested/fiction/craft/genre-conventions` | `genre-blend.ts` -> Deno combination guidance; `genre-check.ts` -> Deno genre-fit analyzer; `genre-elements.ts` -> Deno random element picker from bundled data | Deno | Read-only analyzers/generators; no destructive ops; `genre-elements.ts` depends mainly on bundled dataset | `genre-blend.ts`: `converted-to-reference`.<br>`genre-check.ts`: `converted-to-reference`.<br>`genre-elements.ts`: `converted-to-data-template`. | No required adapter for 3.5. Later adapter could surface check reports and randomized prompts from built-in genre tables. |
| `docs/skills/suggested/fiction/craft/prose-style` | `prose-check.ts` -> Deno prose audit; `rhythm.ts` -> Deno sentence rhythm/variance analyzer | Deno | Read-only text analysis; no mutation; algorithmic heuristics can become guidance/report formats | `prose-check.ts`: `converted-to-reference`.<br>`rhythm.ts`: `converted-to-reference`. | No 3.6 adapter required. |
| `docs/skills/suggested/fiction/craft/revision` | `revision-audit.ts` -> Deno revision-pass audit/report | Deno | Read-only analysis; no mutation; maps cleanly to Edda review workflow | `converted-to-reference` | No 3.6 adapter required. |
| `docs/skills/suggested/fiction/orchestrators/chapter-drafter` | `score-scene.ts` -> Deno scene scoring helper for multi-pass drafting workflow | Deno | Read-only, but the manifest archives the parent orchestrator skill as outside the 3.5 built-in library shape; the score only matters inside that archived workflow | `removed` | No 3.5 adapter. Any future use belongs to an orchestrator product surface with explicit review checkpoints. |
| `docs/skills/suggested/fiction/structure/endings` | `ending-check.ts` -> Deno ending evaluation; `setup-payoff.ts` -> Deno setup/payoff tracker and report | Deno | Read-only manuscript analysis; no destructive ops; output naturally becomes a checklist/report | `ending-check.ts`: `converted-to-reference`.<br>`setup-payoff.ts`: `converted-to-reference`. | No required adapter for 3.5. |
| `docs/skills/suggested/fiction/structure/reverse-outliner` | `analyze-scene-batch.ts` -> Deno batch scene analysis with optional JSON write; `detect-genre.ts` -> Deno genre detection with optional JSON write; `generate-outline.ts` -> Deno outline synthesis with optional markdown write; `reverse-outline.ts` -> Deno pipeline orchestrator that shells out to other Deno scripts and imports remote std modules; `segment-book.ts` -> Deno segmentation with optional JSON write; `track-characters.ts` -> Deno character tracking with optional JSON write | Deno; `reverse-outline.ts` also invokes `deno` subprocess and remote std imports | Multi-step pipeline, filesystem reads/writes, child process orchestration, and derived intermediate files are core to the current skill design; not suitable for 3.5 | `analyze-scene-batch.ts`: `deferred-skill`.<br>`detect-genre.ts`: `deferred-skill`.<br>`generate-outline.ts`: `deferred-skill`.<br>`reverse-outline.ts`: `deferred-skill`.<br>`segment-book.ts`: `deferred-skill`.<br>`track-characters.ts`: `deferred-skill`. | Needs a 3.6 analysis pipeline adapter that reads manuscript selections from Edda, stores intermediates in structured records, and emits outline/report artifacts for review instead of ad hoc files. |
| `docs/skills/suggested/fiction/structure/scene-sequencing` | `analyze-scene.ts` -> Deno single-scene analyzer from file, text, or stdin | Deno | Read-only analysis; no mutation; easy to restate as guided scene review rubric | `converted-to-reference` | No required adapter for 3.5. |
| `docs/skills/suggested/fiction/structure/story-zoom` | `init.ts` -> Deno scaffolds dashboard/log files; `status.ts` -> Deno reads change logs and state files; `watcher.ts` -> Deno long-running filesystem watcher that appends JSONL log entries | Deno with `watchFs` | Long-running watcher, scaffolding writes, arbitrary project-path reads, and runtime/dashboard dependence make the skill non-viable for 3.5 | `init.ts`: `deferred-skill`.<br>`status.ts`: `deferred-skill`.<br>`watcher.ts`: `deferred-skill`. | Needs a 3.6 event/log adapter tied to Edda documents and Story Bible entities rather than raw directory watchers. |
| `docs/skills/suggested/fiction/worldbuilding/conlang` | `phonology.ts` -> Deno phonology generator from bundled frequency/template data; `words.ts` -> Deno word generator from phonology JSON and bundled syllable data | Deno plus remote std path imports | Read-only generators today; no mutation; main durable value is bundled phonology data plus the generation procedure | `phonology.ts`: `converted-to-data-template`.<br>`words.ts`: `converted-to-data-template`. | No required adapter for 3.5 shipping. A later adapter could generate candidate lexicons from Edda world settings and reviewable phonology parameters. |
| `docs/skills/suggested/fiction/worldbuilding/world-fates` | `exposure-log.ts` -> Deno exposure report; `fate-choice.ts` -> Deno choice generation from bundled data; `fate-pressure.ts` -> Deno pressure scoring; `fate-roll.ts` -> Deno fate outcome roll; `propose-shift.ts` -> Deno proposal writer with optional markdown output | Deno plus remote std path imports | Mostly safe report/proposal logic, but the skill is a runtime-like fate system with stateful mechanics and future world-bible integration. It is already marked deferred at the skill level. | `exposure-log.ts`: `deferred-skill`.<br>`fate-choice.ts`: `deferred-skill`.<br>`fate-pressure.ts`: `deferred-skill`.<br>`fate-roll.ts`: `deferred-skill`.<br>`propose-shift.ts`: `deferred-skill`. | Needs a 3.6 structured simulation/proposal adapter backed by Edda world entities and explicit approval flow. |

### Cross-cutting observations

- No audited script performs destructive deletes, but several perform direct filesystem writes or project scaffolding. Those are the clearest deferral boundary.
- No audited fiction scripts make outbound network requests, but some Deno scripts import remote std modules and therefore still depend on runtime/network behavior to execute in practice.
- The future-safe pattern is report/proposal generation. The unsafe pattern is direct project mutation or orchestration around intermediate files.

## 3. Asset inventory table

| Skill/source path | Asset type | Likely disposition |
| --- | --- | --- |
| `docs/skills/suggested/fiction/application/adaptation-synthesis/data` | `data/form-suggestions.json` | Likely copied into built-in skill data. Use as built-in function-to-form option table for adaptation prompts and constrained suggestion generation. |
| `docs/skills/suggested/fiction/application/book-marketing/templates` | `templates/amazon.md`, `blurb.md`, `query.md`, `taglines.md` | Keep as Edda note/response templates for the optional publishing-adjacent skill. No script dependency. |
| `docs/skills/suggested/fiction/application/dna-extraction/data` | `data/extraction-templates.json`, `function-categories.json` | Likely copied into built-in skill data. These are structured extraction schemas and function taxonomies, not runtime-only assets. |
| `docs/skills/suggested/fiction/application/list-builder/references` | `references/dataset-quality-criteria.md` | Keep as reference only. Fold its criteria into skill guidance and review rubrics; no runtime adapter needed. |
| `docs/skills/suggested/fiction/application/shared-world/templates` | `templates/discovery.md`, `entry.md`, `index.md`, `style-guide.md` | Archived with the parent skill for 3.5. Preserve as future Story Bible/Edda note references; do not copy into 3.5 built-ins. |
| `docs/skills/suggested/fiction/character/character-naming/data` | culture lists, mixed pools, phoneme presets, `_meta.json` | Strong candidate for built-in data copy. This is the main value of the naming skill and can power Edda-native naming help later without exposing arbitrary file reads. |
| `docs/skills/suggested/fiction/character/character-naming/templates` | `templates/cast-tracker.json` | Convert to a Edda-managed structured note/schema for cast tracking rather than keeping it as a file the script mutates. |
| `docs/skills/suggested/fiction/core/story-sense/data` | `data/functions-forms.json`, `genre-elements.json` | Likely copied into built-in skill data. These are reusable function and genre reference tables. |
| `docs/skills/suggested/fiction/craft/genre-conventions/data` | `data/genre-elements.json` | Likely copied into built-in skill data, though it may be deduplicated with `story-sense/data/genre-elements.json` during later import work. |
| `docs/skills/suggested/fiction/orchestrators/chapter-drafter/templates` | `templates/pass-criteria.md`, `progress-tracker.md` | Archived reference only. Do not copy into 3.5 Edda response formats or note templates while `chapter-drafter` remains archived. |
| `docs/skills/suggested/fiction/structure/reverse-outliner/data` | `data/key-moments-by-genre.json`, `scene-markers.json` | Defer with `reverse-outliner`. Do not copy into 3.5 built-ins; preserve as source data for a later Edda-native analysis pipeline/runtime decision. |
| `docs/skills/suggested/fiction/structure/story-zoom/templates` | `templates/README.md`, `templates/story-state/state.md` | Useful future dashboard/state templates, but defer with the skill. Not part of 3.5 built-in shipping set. |
| `docs/skills/suggested/fiction/worldbuilding/conlang/data` | `data/phoneme-frequencies.json`, `syllable-templates.json` | Strong candidate for built-in data copy. These assets carry most of the reusable conlang value and can later feed Edda-native generators. |
| `docs/skills/suggested/fiction/worldbuilding/world-fates/data` | `data/fate-choices.json`, `fate-tracking.json`, `shift-types.json` | Preserve for future 3.6 runtime adapter, not 3.5 shipping. These are state/mechanics tables for a deferred skill. |
| `docs/skills/suggested/fiction/worldbuilding/world-fates/templates` | `templates/fate-tracking.md` | Keep as future Story Bible template tied to the deferred world-fates workflow. |

### Assets most likely to be copied into built-in skills

- `application/adaptation-synthesis/data/form-suggestions.json`
- `application/dna-extraction/data/extraction-templates.json`
- `application/dna-extraction/data/function-categories.json`
- `character/character-naming/data/**`
- `core/story-sense/data/functions-forms.json`
- `core/story-sense/data/genre-elements.json`
- `craft/genre-conventions/data/genre-elements.json` after dedupe review
- `worldbuilding/conlang/data/phoneme-frequencies.json`
- `worldbuilding/conlang/data/syllable-templates.json`

### Templates that should become Edda response formats or note templates

- `application/book-marketing/templates/*.md`
- `character/character-naming/templates/cast-tracker.json` as a structured cast note

### Archived or deferred templates not copied into 3.5 built-ins

- `application/shared-world/templates/*.md` stay archived with the parent skill and remain future Story Bible references only.
- `orchestrators/chapter-drafter/templates/*.md` stay archived reference material and are not 3.5 Edda response or note templates.
- `structure/story-zoom/templates/story-state/state.md` stays deferred with the future dashboard surface.
- `worldbuilding/world-fates/templates/fate-tracking.md` stays deferred with the future world-fates workflow.

## 4. Script-bearing skill disclosure language

Shipped 3.5 skills that inherit script-bearing source material should disclose the following behavior in their rewritten docs or system prompts:

> This skill includes logic adapted from source helper scripts, but Edda Milestone 3.5 does not run those scripts. The skill works through Edda-native analysis, guidance, built-in data, and reviewable draft outputs only.

> Any structured output produced by this skill is a proposal, checklist, report, or draft for review. It does not directly edit story text, story bible entries, or project files.

> When this skill refers to source datasets, taxonomies, genre tables, phoneme inventories, or templates, those assets are treated as built-in references rather than executable tooling.

For future 3.6 runtime-backed skills, add this stronger guardrail:

> Runtime helpers may read selected Edda records and generate reports, drafts, or metadata proposals. They must not perform silent mutations of manuscripts, canon records, or project structure.

## 5. Archived and deferred outcomes

### Archived / outside Milestone 3.5 product scope

These sources are intentionally not part of the 3.5 built-in library shape, matching the accepted manifest’s archived outcomes:

- `docs/skills/suggested/fiction/application/shared-world`
- `docs/skills/suggested/fiction/application/list-builder`
- `docs/skills/suggested/fiction/application/game-facilitator`
- `docs/skills/suggested/fiction/application/table-tone`
- `docs/skills/suggested/fiction/orchestrators/chapter-drafter`
- `docs/skills/important/writing-skills/render-graphs.js` authoring helper

### Deferred pending Milestone 3.6 runtime or follow-up audit

These sources remain deferred rather than archived, matching the manifest’s distinctions:

- `docs/skills/suggested/fiction/structure/reverse-outliner` - deferred after script audit because the current value depends on a file/script analysis pipeline
- `docs/skills/suggested/fiction/structure/story-zoom` - deferred runtime-dependent dashboard workflow
- `docs/skills/suggested/fiction/worldbuilding/world-fates` - deferred until the script runtime and world-entity proposal flow exist

## 6. Follow-up requirements for Milestone 3.6

1. Define a script runtime contract that only accepts explicit Edda-selected inputs and only returns reports, metadata proposals, generated notes, or draft text.
2. Add a path-independent input adapter layer so scripts consume Edda document/entity records instead of arbitrary local file paths.
3. Add an output adapter layer that writes to review queues, draft notes, or structured records rather than directly mutating manuscripts or canon data.
4. Ban silent filesystem mutation in runtime-backed skills; require human approval for any persisted draft or note creation.
5. Replace raw project scaffolding flows (`shared-world`, `story-zoom`) with Edda-native Story Bible and dashboard surfaces before enabling any helper logic.
6. Replace multi-file orchestration pipelines (`reverse-outliner`) with internal structured jobs and temporary storage managed by Edda rather than ad hoc JSON/markdown intermediates.
7. Vendor or normalize any retained runtime dependencies. Deno remote std imports and external CLIs should be pinned and mediated by the product runtime, not invoked ad hoc from skill docs.
8. Decide which copied datasets become first-class built-in assets and where deduplication happens, especially for genre tables and naming datasets.
9. Add test fixtures for every retained future runtime helper: expected inputs, expected report output, approval boundaries, and failure modes.
10. Add manifest cross-links from each deferred skill row to this audit so 3.6 implementers can distinguish deferred runtime work from 3.5-safe built-in rewrites.
