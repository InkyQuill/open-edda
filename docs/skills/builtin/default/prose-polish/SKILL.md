---
name: prose-polish
description: Sentence-level fiction polish for stable passages with weak rhythm, vague diction, filtering, abstraction, redundancy, weak verbs, or voice drift.
route:
  actionKinds:
    - chat
    - read_check
    - rewrite
  contentKinds:
    - chapter
    - story_text
    - attached_note
    - project_note
  tags:
    - fiction
    - prose
    - line-edit
    - voice
  priority: 70
metadata:
  useCases:
    - Structure, scene order, and character intent are already stable.
    - The author wants line-level diagnosis of rhythm, specificity, clarity, filtering, abstraction, weak verbs, redundancy, or POV voice.
    - A passage feels flat, overwritten, monotonous, generic, indirect, distant, or inconsistent in voice.
    - The author wants a focused polish pass or a scoped rewrite that preserves the existing author voice.
  doNotUse:
    - The story still has major structural, pacing, or arc problems.
    - The author wants broad story diagnosis instead of line work.
    - The author wants lore building or continuity planning.
    - The author wants dialogue craft, scene sequencing, or developmental revision as the primary task.
  status: default
  source:
    - docs > skills > suggested > fiction > craft > prose-style > SKILL.md
  scriptStatus: source-helpers-converted-to-guidance
---

# $prose-polish

Line-level prose diagnosis and scoped polishing for structurally stable fiction while preserving the author's existing voice.

## Edda Workflow

1. Confirm that the requested passage is ready for sentence-level work. If unresolved structure, scene purpose, pacing, or character-arc problems would make line edits premature, say that briefly and route the author toward developmental revision.
2. Read the target selection or chapter. Also read nearby chapter context, the Writing Brief, and any relevant voice or POV notes when available through Edda project tools.
3. Identify the dominant line-level state before prescribing fixes: flat prose, unclear writing, overwrought prose, monotonous rhythm, default passive construction, weak verbs, filtering distance, abstraction, redundancy, inconsistent POV voice, or an image system that distracts from the scene.
4. Diagnose at multiple levels:
   - Sentence rhythm: sentence length spread, repeated openings, paragraph length variation, punch position, and whether short/long sentences are being used intentionally.
   - Specificity: generic nouns, vague modifiers, imprecise verbs, thesaurus abuse, elegant variation that hurts clarity, and concrete detail that could replace vague abstraction.
   - Filtering and POV distance: saw, heard, felt, noticed, realized, thought, knew, watched, wondered, decided, and other constructions that report perception instead of rendering experience when close POV is intended.
   - Abstraction and clarity: compressed logic, unclear pronoun antecedents, missing context, vague emotional labels, and conceptual language that avoids the observable moment.
   - Image systems: metaphors, similes, sensory details, repeated images, and motif language; check whether they clarify character and scene or create mixed, ornamental, or off-register effects.
   - POV voice: diction level, syntax, sentence music, distance from the reader, narrator attitude, and intrusive author phrasing that does not belong to the viewpoint.
   - Redundancy: repeated information, doubled modifiers, restated emotions, explanatory tags after clear action, and paragraphs that say the same thing twice.
   - Weak verbs and default constructions: be, have, do, get, make, seem, appear, become, passive forms, and adverb-plus-weak-verb combinations that could become precise action.
5. Distinguish line-level diagnosis from rewriting. First explain the issue with brief evidence from the passage. Rewrite only when the author explicitly asks for applied edits or when the action kind is a rewrite.
6. When rewriting, preserve author voice over generic polish. Keep the passage's POV, register, cadence, image logic, character vocabulary, and intended level of richness. Do not flatten a lyrical passage merely because it is rich; do not decorate a spare passage merely to make it sound more literary.
7. Make every recommendation conditional on purpose. Passive voice, abstraction, repetition, filtering, adverbs, and long sentences are problems only when they are accidental or weaken the intended effect.
8. Keep edits tightly scoped to the selected passage. Do not change plot facts, character decisions, timeline, lore, names, setting rules, or durable canon while polishing prose.

## Criteria For Diagnosis

- Rhythm is stronger when sentence and paragraph lengths vary for emphasis, sentence openings do not drone, related words stay together, and emphatic words land at the end when possible.
- Specificity is stronger when concrete nouns and exact verbs carry the meaning instead of vague abstractions, stacked adjectives, generic emotional labels, or approximate adverbs.
- Filtering is worth cutting when it distances the reader from direct experience in close POV; it may stay when the act of perception, uncertainty, or interpretation matters.
- Abstraction is acceptable when it names an idea the scene needs; it weakens prose when it replaces action, sensation, choice, or visible consequence.
- Image systems should belong to the character, world, genre, and moment. Flag mixed metaphors, ornamental comparisons, repeated images that no longer develop, and images that contradict the passage's emotional logic.
- POV voice should remain consistent in diction level, rhythm, attitude, and distance. Flag shifts that feel accidental rather than character- or scene-driven.
- Redundancy includes repeated beats, explanatory restatement, doubled modifiers, duplicate sentence functions, and repeated information that does not create rhythm or emphasis.
- Weak verbs and passive constructions should be treated as diagnostic signals, not automatic errors. Replace them when a more specific verb, named agent, or clearer subject would increase force or clarity.

## Edda Output Handling

- Return a short diagnosis in chat by default: dominant issue, evidence, why it matters, and a few targeted fixes.
- For line-level review, group findings by criterion and include compact before/after examples only for the lines needed to demonstrate the fix.
- For an applied rewrite, return the revised passage plus a brief change note covering rhythm, specificity, filtering, abstraction, image system, POV voice, redundancy, and weak-verb decisions that materially changed the text.
- Create an Attached Note when the polish report belongs to one chapter, scene, or selected passage.
- Create or update a Project Note only when the author asks for reusable voice guardrails, a chapter-wide polish checklist, or a cross-chapter prose pattern report.
- Use Structured Writes only when the author explicitly asks to apply a replacement to the selected passage.
- If line editing reveals a canon contradiction, report it separately as a question or Story Bible proposal. Do not silently change canon in prose-polish output.

## Script Compatibility

The source `prose-check.ts` and `rhythm.ts` helpers are converted to guidance and deferred policy for Edda. Do not ask the runtime agent to read or run source scripts.

Use the script methodology manually in diagnosis:

- From `rhythm.ts`: consider sentence length distribution, paragraph length variation, repeated sentence openings, missing short punch sentences, missing longer flow sentences, and monotony caused by low variation.
- From `prose-check.ts`: consider passive voice concentration, weak verb frequency, adverb density, filter words, and adjective stacking.

These are heuristics, not rules. If an approved `skill_script` helper is later enabled for this skill, use it only as a non-mutating diagnostic aid and still make final craft judgments from the passage, author intent, and POV voice.
