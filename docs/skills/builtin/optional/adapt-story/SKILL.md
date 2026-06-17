---
name: adapt-story
description: Function-first story adaptation for translating source DNA into a new form, setting, or medium without surface reskinning.
route:
  actionKinds:
    - chat
    - read_check
    - continuation
    - rewrite
  contentKinds:
    - chapter
    - story_text
    - attached_note
    - project_note
    - story_bible_entry
    - entry_section
  tags:
    - fiction
    - adaptation
    - synthesis
    - optional
  priority: 48
metadata:
  useCases:
    - The author wants to adapt a story, media influence, trope cluster, or prior draft into a new setting, form, genre, or medium.
    - A proposed adaptation sounds like "source but in new setting" and needs function-first transformation instead of cosmetic reskinning.
    - Multiple sources, inspirations, or DNA extractions need a primary-source hierarchy and conflict resolution.
    - The author needs validation that adapted forms preserve emotional experience, power dynamics, conflict architecture, and context logic.
  doNotUse:
    - The author only wants to summarize, review, or identify influences in a source work without planning an adapted work.
    - No source material, extracted story DNA, or source analysis is available and the next needed step is DNA extraction.
    - The author wants a beat-for-beat copy, parody, homage checklist, or cosmetic name-and-setting swap.
    - The task is only line editing or polishing already-adapted prose.
  status: optional
  source:
    - docs > skills > suggested > fiction > application > adaptation-synthesis > SKILL.md
    - docs > skills > suggested > fiction > application > media-adaptation > SKILL.md
  scriptStatus: source-helpers-deferred-data-retained
---

# $adapt-story

Function-first adaptation support for moving source DNA into a new setting, medium, genre, or formal container while preserving what makes the original work.

## Edda Workflow

1. Establish adaptation state before generating material:
   - `SYN0: No DNA Documents` means the author wants an adaptation but has not identified what functions the source serves. Do not synthesize yet; ask for source analysis or route toward DNA extraction.
   - `SYN1: DNA Ready` means story functions, emotional experience, tone, power dynamics, relationship patterns, or structural requirements are available. Begin context mapping.
   - `SYN2: Surface Translation Mode` means the plan maps forms 1:1, such as king to CEO or castle to spaceship. Stop and convert every copied form into a function question.
   - `SYN3: Context Mismatch` means the function is known but the target world does not naturally support the chosen form. Rebuild the form from target-context pressures.
   - `SYN4: Function Gap` means a new form exists but fails to create the same structural, character, or emotional outcome. Add, replace, or split forms until coverage is real.
   - `SYN5: Genre Drift` means the adaptation has changed the audience promise by accident. Recheck emotional beats, tone, tension and release, and genre expectations.
   - `SYN6: Source Conflict` means multiple sources want incompatible structure, pacing, tone, or character behavior. Establish hierarchy before combining them.
   - `SYN7: Synthesis Ready` means functions are mapped, context supports them, conflicts are resolved, and validation passes.
2. Read the available Edda context. Use `project_map` to locate relevant chapters, notes, Story Bible entries, and writing briefs. Use `read_content`, `read_chapter`, `read_story_bible_entry`, `read_entry_section`, or `search_content` for source DNA notes, extracted functions, current canon, target-setting constraints, and draft scenes.
3. Separate source material into a hierarchy:
   - `Must remain`: core emotional experience, central tension structure, essential power dynamics, key relationship functions, structural obligations, and any author-declared non-negotiables.
   - `Must change`: forms that only worked because of the original culture, medium, historical moment, technology, institution, or surface imagery.
   - `May adapt`: authority structures, resources, communication barriers, enforcement mechanisms, geography, pacing devices, tone carriers, and character occupations.
   - `Must discard or transform`: elements that break target-context logic, import harmful assumptions unintentionally, or only exist as recognition markers.
4. Classify each source element by transfer behavior:
   - `Universal elements` can transfer directly because they are grounded in broad human experience, such as belonging, betrayal, justice, competence, fear, obligation, trust, or loyalty.
   - `Setting-dependent elements` need target-context equivalents, such as authority, resources, information flow, enforcement, escape constraints, and institutional pressure.
   - `Culture-specific elements` require replacement rather than direct transfer. Preserve their narrative function, not their surface form.
5. Map functions to target-context forms. For each important source function, record the original form, the function it served, at least two target-context form options, the chosen new form, and what structural or emotional outcomes it enables. Ask: what in this target setting naturally creates power, proximity, secrecy, obligation, vulnerability, inescapability, moral pressure, or consequence?
6. Apply the orthogonality test to every chosen form. A form passes only if it exists for its own reasons in the target context, makes sense to someone who does not know the source, has its own goals and history, and can be described without naming the source. If an element "knows what story it is in," rebuild it.
7. For media or medium changes, transform mechanics instead of transplanting scenes:
   - Film or TV to prose: replace camera-dependent reveals with viewpoint, scene order, interiority, sensory focus, withheld information, and chapter rhythm.
   - Game to prose: replace player agency, loops, fail states, inventory, quests, or level gates with character choices, consequences, information asymmetry, escalation, and reversible or irreversible costs.
   - Prose to visual or serial structure: identify which interior states need externalization, which beats become set pieces, and which exposition becomes action, image, or recurring scene pattern.
   - Episodic to novel or novel to episodic: decide whether case-of-the-week, mobile-base, season arc, ensemble rotation, cliffhanger, or chapter arc functions must remain, compress, expand, or become background structure.
8. When combining sources, choose one primary source before resolving conflicts. Primary source functions take precedence. Secondary sources may add flavor, contrast, texture, or subplots only when they do not break the primary emotional promise. Resolve conflicts by choosing `primary wins`, `blend`, `alternate by subplot`, or `transform the conflict into a feature`.
9. Synthesize tone and voice by function. Keep what the tone does, not the exact phrasing. If a source uses banter to mask pain, find the target-context relationship and speech norms that let characters deflect. If dread, melancholy, awe, competence, or absurdity is essential, state what target-context pressures create that feeling.
10. Validate before drafting. Check function coverage, orthogonality, context coherence, genre alignment, tone alignment, completeness, and cultural sensitivity. Treat failure as a planning issue, not a prose issue.
11. Draft or rewrite adapted Story Text only when the author explicitly asks for applied text and the target passage is clear. Otherwise return diagnosis, mapping, options, or a reusable adaptation brief.

## Function-First Criteria

- A good adaptation preserves what the source does for the audience while replacing forms that would be obvious, forced, or context-breaking in the target work.
- A bad adaptation copies the source's visible inventory: names, roles, locations, costumes, technologies, plot tokens, and beat order without rebuilding why those pieces mattered.
- Fidelity to function outranks fidelity to form. If a copied form no longer creates the same pressure, emotion, relationship, or conflict, it must change.
- New forms must create new implications. If the target setting changes nothing except labels, the adaptation is a reskin.
- The adapted work must stand alone. A reader should understand the stakes, relationships, pressures, and world logic without knowing the source.

## Adaptation Validation

Use this checklist when judging or presenting an adaptation plan:

1. `Function coverage`: every must-remain source function has a target form and a stated outcome.
2. `Orthogonality`: no chosen form exists only because the source had an analogous object, role, or beat.
3. `Context coherence`: power, information, obligation, resources, and escape constraints emerge from the target setting.
4. `Emotional promise`: the adapted work intentionally preserves or deliberately changes the source's emotional experience.
5. `Genre control`: any genre shift is named and accounted for; drift is not accidental.
6. `Source hierarchy`: when sources conflict, the primary source and resolution choice are explicit.
7. `Medium fit`: the plan uses the target medium's native strengths rather than imitating source-medium effects.
8. `Cultural and systemic sensitivity`: replacements do not import assumptions, harms, or institutions without examination.
9. `Completeness`: context mapping, function-to-form mapping, tone synthesis, character synthesis, and readiness are documented.

## Edda Output Handling

- Return short state diagnosis, challenges, and next-step questions in chat.
- Create an Attached Note when the mapping or validation belongs to one chapter, scene, selection, or local adaptation problem.
- Create a Project Note for a reusable adaptation brief, source hierarchy, synthesis worksheet, validation report, or cross-chapter adaptation plan.
- Propose Story Bible changes only when the adaptation would establish durable canon such as setting rules, institutions, character facts, factions, history, names, technology, magic, or timeline logic. Keep proposals separate from confirmed canon until the author accepts them.
- Use Structured Writes only after explicit author intent to draft or rewrite and only against a known target passage or chapter.

When producing an adaptation brief, include these fields when relevant: source material, target context, primary source, secondary sources, must-remain functions, must-change forms, context mapping, function-to-form mapping, tone synthesis, character synthesis, validation results, unresolved questions, and recommended next skill or drafting step.

## Data Files

- `data/form-suggestions.json` — Form and medium transformation prompts for adaptation planning.

Load `data/form-suggestions.json` with `read_skill_file` when the author needs form options for functions such as proximity to power, privileged unverifiable information, structural obligation, inability to escape, entrenched antagonists, innocent bystanders, witness structures, dark mirrors, found-family crews, episodic cases, or mobile home bases. Use the data as prompts, not as a menu of guaranteed answers; each suggestion still needs orthogonality and context-coherence validation.

## Script Compatibility

The source scripts for form options, synthesis worksheets, and synthesis validation are deferred helpers. Their useful logic is represented here as workflow criteria, validation checks, and the lazy-loaded `data/form-suggestions.json` file. In Edda runtime, do not assume script execution; use `skill_script` only if an administrator has approved and enabled a future non-mutating helper for this skill.
