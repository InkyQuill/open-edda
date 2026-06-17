---
name: story-dna
description: Extract the functional patterns behind a story, chapter, or outside influence so authors can reuse what works without copying the surface.
route:
  actionKinds:
    - chat
    - read_check
  contentKinds:
    - chapter
    - story_text
    - attached_note
    - project_note
    - story_bible_entry
  tags:
    - fiction
    - analysis
    - adaptation
    - optional
  priority: 38
metadata:
  useCases:
    - The author wants the reusable story DNA of a story, chapter, trope cluster, or outside influence.
    - Adaptation, remix, or comparative analysis needs to separate surface form from underlying function.
    - A story works on instinct and the author wants structural, emotional, thematic, relational, or tonal reasons why.
    - The author needs to know which elements are load-bearing and which can change without breaking the effect.
  doNotUse:
    - The author wants immediate drafting help rather than analysis.
    - The request is only for marketing positioning.
    - The goal is to copy surface features instead of underlying function.
  status: optional
  source:
    - docs > skills > suggested > fiction > application > dna-extraction > SKILL.md
  scriptStatus: source-helpers-deferred-data-retained
---

# $story-dna

Extract reusable story DNA by identifying what story elements do beneath the surface, not merely what they are.

## Edda Workflow

1. Read the target Story Text, Chapter, Story Bible entry, Attached Note, Project Note, or outside-work description the author provides. If the source work, medium, extraction goal, or depth is unclear, ask before doing a full extraction.
2. Diagnose the extraction state:
   - `EX0 No Extraction`: the source is named but not analyzed. Start with source, medium, goal, and emotional core.
   - `EX1 Surface Reading`: the analysis summarizes events or names forms. Reframe every element as function, using "what would break if this disappeared?" and "what does this enable?"
   - `EX2 Single-Axis Extraction`: only plot, character, theme, or style is discussed. Expand across all extraction axes.
   - `EX3 Missing Emotional Core`: functions exist but the audience experience is unclear. Identify primary genre promise, emotional peaks and valleys, and what fans love about the work.
   - `EX4 Structural/Stylistic Conflation`: style, setting, era, or iconography is treated as mandatory. Classify whether the function is structural and whether the current form is only one possible carrier.
   - `EX5 Missing Relationships`: characters or elements are isolated. Map contrasts, obligations, secrets, dependencies, foils, and relationship pressure.
   - `EX6 No Hierarchy`: every element is treated as equally important. Rank functions as primary, reinforcing, or optional.
   - `EX7 Extraction Complete`: emotional core, multi-axis functions, hierarchy, structural/stylistic split, relationship web, and adaptation boundaries are clear enough to support a remix.
3. Separate surface from function for every major element. Surface form is what the element is: prince, ghost, ship, sword fight, school, royal court, banter style. Function is what it does: proximity to power without authority, unverifiable information that creates obligation, mobile home base for episodic missions, physicalized value conflict, constrained social hierarchy, tension release through intimacy.
4. Extract across the functional layers that apply:
   - Structural: plot mechanics, causation, constraints, information control, escalation, setup/payoff.
   - Character: wants, needs, lies, arc pressure, transformation catalysts, mirrors, foils.
   - Emotional: genre promise, peaks, valleys, tension, release, catharsis, dread, wonder, longing, satisfaction.
   - Thematic: questions, values under pressure, symbols, moral ambiguity, consequences of choices.
   - Relational: dynamics between characters or story elements, contrast, dependency, obligation, secrets, rivalry, intimacy.
   - Stylistic/Tonal: sincerity level, humor mode, dialogue density, emotional expression, conflict style, voice patterns, tonal shifts.
5. Use function categories when coverage matters. Call `read_skill_file` for `data/function-categories.json` when the extraction needs a taxonomy of structural, character, emotional, thematic, relational, or tonal functions, or when the analysis is stuck in plot summary.
6. Use extraction templates when the output should become a reusable record. Call `read_skill_file` for `data/extraction-templates.json` when the author asks for a quick, standard, or detailed extraction; a reusable DNA note; a trope cluster; an adaptation brief; or validation criteria.
7. Choose extraction depth deliberately:
   - Quick: emotional core, primary genre, 3-5 central elements, core structural requirements, clearly adaptable forms.
   - Standard: six-axis extraction for major characters/elements, tone profile, plot structures, key relationships, structural requirements, adaptable elements, function hierarchy.
   - Detailed: scene or episode functions, voice and dialogue patterns, tonal shifts, minor character functions, relationship evolution.
8. Classify structural vs. stylistic boundaries. Treat the function as structural when changing it breaks plot logic, character arc, emotional promise, theme, or relationship pressure. Treat the current form as stylistic when another form could carry the same function. Use a mixed classification when the function must remain but the shell can change.
9. Build a hierarchy. Mark each extracted function as primary, reinforcing, or optional. A valid extraction names the minimum viable DNA: the small set of functions without which the new work would not produce the same effect.
10. Guard against copying. Do not recommend transplanting distinctive names, exact scenes, prose, dialogue, sequence order, proprietary world details, or recognizable surface combinations from an outside source. Convert inspiration into abstract functions, alternate forms, and adaptation constraints.
11. Keep analysis separate from synthesis. Extract the DNA first; only provide adaptation, remix, or rewrite options after the author asks for next steps or the extraction has reached `EX7`.

## Extraction Criteria

A good extraction:

- States the emotional core before declaring what must be preserved.
- Names surface forms separately from reusable functions.
- Covers structural, character, emotional, thematic, relational, and tonal layers when depth allows.
- Identifies relationship dynamics, not just isolated character traits.
- Distinguishes primary load-bearing functions from reinforcing or optional flavor.
- Lists adaptable forms without breaking the underlying function.
- Gives the author enough reusable DNA to create a different work with a similar effect.

A weak extraction:

- Reads like plot summary with function labels attached.
- Says an element matters without naming what it enables, prevents, pressures, or reveals.
- Treats favorite details as essential because they are memorable.
- Marks everything essential, leaving no room for adaptation.
- Uses the original work's surface as the recommendation instead of abstracting the function.
- Skips emotional core and genre promise.

## Adaptation and Remix Boundaries

- Preserve functions, not forms. If the source has a prince, preserve "insider near power but unable to command it" only if that is the load-bearing function.
- Replace specific expressions with new carriers. If a ghost delivers unverifiable obligation, alternatives might be a corrupted archive, dying testimony, hacked evidence, or a contradictory witness.
- Keep tone as a separate layer. A structural remix can change sincerity, humor, dialogue density, or conflict style; a tonal homage can borrow an emotional register without borrowing plot architecture.
- Do not collapse analysis into imitation. When a proposed adaptation resembles the source too closely, name the copied surface features and provide two or three functionally equivalent alternatives.
- Treat outside works as methodology and inspiration only. For copyrighted or living-source material, summarize at high level, avoid close paraphrase, and produce new expressions.

## Edda Output Handling

- Return exploratory questions, state diagnosis, and short DNA analysis in chat by default.
- Create an Attached Note when the analysis belongs to one Chapter, scene, excerpt, or selected passage.
- Create or update a Project Note when the author wants a reusable extraction record, trope cluster, emotional beat map, structural/stylistic classification, validation checklist, or adaptation brief.
- Propose Story Bible changes only when the author explicitly wants durable project canon, worldbuilding, character facts, faction logic, history, rules, names, or setting constraints updated from the analysis. Keep such changes as proposals until the author confirms them.
- Do not use Structured Writes in this skill.

## Data Files

This skill includes extraction schemas and taxonomies as Writer-native references:

- `data/extraction-templates.json` - Depth levels, work extraction schema, character/plot/relationship templates, trope cluster template, medium guidance, and extraction checklists.
- `data/function-categories.json` - Function taxonomy for structural, character, emotional, thematic, relational, tonal, and surface-form categories.

Load these with `read_skill_file` only when the session needs structured extraction, validation, or taxonomy support. Do not load them for a short high-level summary.

## Script Compatibility

Source helper scripts are deferred and must not be treated as readable references or runnable tools unless an administrator later audits and enables them through `skill_script`.

- The source `extract-functions.ts` behavior is converted into the workflow: source metadata, emotional core, character functions, structural requirements, adaptable elements, depth-specific checks, and validation issues.
- The source `emotional-beat-map.ts` behavior is converted into guidance: map emotional peaks, valleys, sustain moments, shifts, genre promise, intensity, and pacing notes.
- The source `structural-stylistic.ts` behavior is converted into guidance: classify each element by removability, alternate forms, universality, arc dependence, emotional dependence, thematic dependence, and audience attachment.
- Until scripts are approved, produce analysis, templates, JSON-like records, and reviewable notes through Edda chat and note tools only.
