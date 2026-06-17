---
name: genre-check
description: Genre-promise and reader-expectation diagnosis for unclear emotional contracts, missing conventions, stale execution, or competing hybrid genres.
route:
  actionKinds:
    - chat
    - read_check
  contentKinds:
    - chapter
    - story_text
    - project_note
    - attached_note
  tags:
    - fiction
    - genre
    - diagnosis
    - expectations
  priority: 74
metadata:
  useCases:
    - The story's opening, outline, chapter, or premise makes an unclear genre promise.
    - Required or expected genre conventions feel missing, misplaced, stale, or insufficiently integrated.
    - Secondary genre material may be competing with the primary emotional experience.
    - The author describes a setting label such as fantasy, science fiction, historical, or contemporary as if it were the genre.
    - A hybrid, ensemble, or multi-POV story needs a hierarchy of primary and secondary reader expectations.
  doNotUse:
    - The author already knows the genre problem and needs focused scene, prose, or ending work.
    - The request is mainly about setting details rather than the emotional experience those details create.
    - The author wants the skill to choose a genre identity by force.
    - The author wants drafting, line editing, or prose generation rather than diagnosis and revision guidance.
  status: default
  source:
    - docs > skills > suggested > fiction > craft > genre-conventions > SKILL.md
  scriptStatus: source-helpers-converted-to-guidance-and-data
---

# $genre-check

Diagnose whether a story makes a clear genre promise, supplies the conventions needed to satisfy reader expectations, and keeps hybrid genres in a useful hierarchy.

## Edda Workflow

1. Read the target context before diagnosing: use `read_chapter` or `read_content` for the relevant prose, `read_content` for project or attached notes, and `project_map` or `search_content` when the author asks about the whole project, a hybrid structure, or cross-chapter expectations.
2. If the author names a genre, translate it into an emotional promise instead of accepting a bookshelf or setting label. Science fiction, fantasy, historical, contemporary, and similar labels describe where a genre lives; they do not by themselves define the reader's promised experience.
3. Identify the intended primary promise by asking what the reader should feel or track first: awe, intellectual fascination, external adventure, dread, curiosity, urgent danger, amusement, relationship investment, internal transformation, an issue under pressure, or group dynamics.
4. When the author's intent is unclear, ask one or two focused questions before diagnosing. Do not assign the story a genre by force; offer likely candidates and the expectation tradeoffs attached to each.
5. Load `data/genre-elements.json` with `read_skill_file` when the diagnosis needs concrete convention checks, missing-element generation, or hybrid-genre comparison.
6. Diagnose the most relevant genre state:
   - `G1 Missing Genre Promise`: the opening or premise does not signal the emotional contract.
   - `G2 Wrong Genre for Story`: the material naturally creates a different experience than the stated genre.
   - `G3 Genre Elements Misplaced`: the right conventions appear, but setup, escalation, clues, obstacles, dread, or payoff arrive in the wrong order.
   - `G4 Secondary Genre Undermining Primary`: a subplot or tonal mode overwhelms the main reader experience.
   - `G5 Genre Without Required Elements`: non-negotiable conventions for the chosen genre are absent or too weak.
   - `G6 Genre Conventions Stale`: the story fulfills the checklist but uses predictable defaults.
   - `G7 Setting Mistaken for Genre`: world, era, or speculative premise substitutes for an emotional throughline.
   - `G8 Ensemble Without Genre Assignment`: multiple POVs or threads each create an experience, but the project lacks an umbrella promise.
7. Audit required and optional conventions separately. Treat `required_elements` in the data file as the baseline contract that must be present or deliberately replaced; treat category examples as optional element banks for diagnosis, fresh alternatives, or scene-level strengthening.
8. Check placement, not just presence. Required conventions should support an expectation curve: opening signal, progressive development, escalation or complication, and payoff. Flag any convention that appears too late, too early, all at once, or without setup.
9. For hybrid genres, establish hierarchy. Name the primary genre, then explain how each secondary genre should serve it: by raising stakes, deepening character pressure, supplying contrast, or occupying contained subplots. If two genres compete for the same story position, identify the mismatch and offer hierarchy options.
10. For stale conventions, preserve the core promise and vary the execution. Recommend inversion, complication, specificity, or an unexpected delivery mechanism. Do not advise subverting every convention; the reader still needs stable ground and a satisfying version of the promised experience.
11. For expectation mismatch, compare what the opening promises against what later chapters, outline beats, or the ending appear to deliver. Mark any bait-and-switch risk and propose whether to revise the opening signal, revise the later payoff, or explicitly frame the story as a hybrid from the start.
12. Recommend concrete interventions tied to the diagnosis: add a genre marker, strengthen a missing required element, move a convention earlier or later, reduce secondary genre page weight, integrate a subplot into the primary stakes, or hand off to a more specific skill.
13. Hand off when the root problem is outside genre fit: use character-arc work when the protagonist does not meet genre needs, scene or pacing work when escalation placement is the issue, cliche-transcendence work when freshness is the main problem, and worldbuilding work when the setting does not serve the promised experience.

## Edda Output Handling

- Return the genre diagnosis in chat by default, especially when the author is deciding which promise to emphasize.
- Structure the answer around evidence from the supplied text: likely genre state, primary promise, secondary genres, required convention gaps, optional convention opportunities, expectation mismatch, and next revision moves.
- Create an Attached Note when the report belongs to one chapter, opening scene, selection, or local promise problem.
- Create or update a Project Note when the author wants a durable genre checklist, hybrid hierarchy, expectation map, or cross-chapter revision plan recorded for later work.
- Do not use Structured Writes in this skill.
- Do not write scenes or prose for the author from this skill. If the author asks for applied drafting after diagnosis, route to a drafting or rewrite skill with the genre findings as constraints.
- Do not propose Story Bible changes unless genre diagnosis reveals a durable constraint the author explicitly wants recorded, such as a recurring mystery rule, horror threat rule, relationship premise, ensemble role assignment, or setting principle that now affects canon. Keep such changes as proposals until the author confirms them.

## Data Files

This skill includes reference data the agent can consult during genre diagnosis:

- `data/genre-elements.json` — Genre element expectations and conventions.

Load this file with `read_skill_file` when genre fit, genre promise, missing conventions, optional convention generation, stale convention handling, or hybrid-genre balance needs a concrete reference. Use its `core_promise` values to check the emotional contract, `required_elements` to test baseline reader expectations, and genre `categories` as optional element banks. Do not treat category examples as mandatory checklists.

## Script Compatibility

This rewrite converts the source diagnostic process, convention tables, element generation, and blend guidance into Edda-native workflow plus `data/genre-elements.json`. Source helper assumptions are not runtime requirements here; use Edda project-reading tools and `read_skill_file` instead of external helper execution.
