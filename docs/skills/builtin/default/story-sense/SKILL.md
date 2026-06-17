---
name: story-sense
description: Broad fiction diagnosis and routing for stuck, broken, flat, generic, stalled, or uncertain stories, chapters, concepts, and drafts.
route:
  actionKinds:
    - chat
    - read_check
  contentKinds:
    - chapter
    - story_text
    - project_note
    - attached_note
    - story_bible_entry
    - entry_section
  tags:
    - fiction
    - diagnosis
    - routing
    - revision
  priority: 96
metadata:
  useCases:
    - The author says the story is stuck, broken, flat, generic, stalled, thin, or not working.
    - The author asks what is wrong with a story, chapter, concept, outline, or draft.
    - The session needs diagnosis before choosing drafting, revision, dialogue, structure, worldbuilding, prose, or evaluation work.
    - A story problem is described as a symptom rather than a clear task.
  doNotUse:
    - The author already knows the exact task and wants direct drafting, rewriting, or line editing.
    - The task is a focused dialogue, ending, pacing, prose, or continuity check with a clear scope.
    - The author wants coaching questions only rather than diagnosis and routing.
    - The author wants publishing, marketing, or non-fiction advice.
  status: default
  source:
    - docs > skills > suggested > fiction > core > story-sense > SKILL.md
  scriptStatus: source-helpers-converted-to-guidance-and-data
---

# $story-sense

Diagnose what a story needs in its current state, name the smallest useful intervention, and route the author to the next Edda-native workflow without prematurely drafting.

## Edda Workflow

1. Start from the author's stated symptom. Treat "stuck" as undiagnosed state, not as the diagnosis.
2. Read only the context needed to locate the symptom:
   - Use `read_chapter` or `read_content` for the chapter, selection, outline, project note, or attached note under discussion.
   - Use `project_map` when the relevant chapter or planning note is unclear.
   - Use `read_story_bible_entry` or `read_entry_section` only when the symptom involves canon, world rules, character facts, institutions, timeline, names, or setting logic.
   - Use `list_revisions` only when the symptom is revision drift, regression, or uncertainty about what changed.
3. Separate symptom from cause. Ask one or two clarifying questions when the available context cannot distinguish between adjacent states; otherwise proceed with a provisional diagnosis.
4. Identify the current story state:

| State | Diagnosis | Typical symptoms | Route or intervention |
| --- | --- | --- | --- |
| 0 | No Story | Blank page, no concrete premise yet | Idea generation, genre elements, premise options |
| 1 | Concept Without Foundation | Idea exists but characters, world, or conflict feel thin | Cliche transcendence, systemic worldbuilding, key moments |
| 2 | World Without Life | Setting exists but behaves like backdrop | Belief, economy, governance, ecology, institutions, daily-life worldbuilding |
| 3 | Flat Non-Humans | Aliens, fantasy species, monsters, or constructed cultures feel human in costume | Species, language, biology, culture, and non-human viewpoint work |
| 4 | Characters Without Dimension | Characters serve plot instead of driving it | Character arc, pressure, desire, contradiction, positional revelation |
| 4.5 | Plot Without Pacing | Scenes work alone but do not accumulate pressure | Scene sequencing, escalation, aftermath, reversals |
| 5 | Plot Without Purpose | Events happen but do not accumulate meaning | Moral parallax, thematic pressure, key moments |
| 5.5 | Dialogue Feels Flat | Characters sound alike or conversations lack tension | Dialogue diagnosis and rewrite workflow |
| 5.75 | Ending Does Not Land | Setup works but resolution disappoints | Ending diagnosis, payoff, cost, transformation, final image |
| 5.85 | Draft Not Progressing | Planning exists but drafting does not happen | Drafting workflow, next-scene constraint, momentum plan |
| 5.9 | Prose Feels Flat | Story logic works but sentences feel merely functional | Prose style, voice, image systems, line-level revision |
| 6 | Draft Complete, Needs Revision | A complete draft exists but revision feels overwhelming | Revision map, pass planning, structural triage |
| 7 | Ready for Evaluation | The work exists and the author needs quality, risk, or fit assessment | Story analysis, sensitivity check, targeted reader-response questions |

5. Diagnose by function and form before prescribing a fix:
   - Function is the job a story element performs: pressure, revelation, obstacle, witness, temptation, cost, contrast, shelter, proof, rupture, or choice.
   - Form is the concrete expression of that job: a character, scene, institution, magic rule, clue, conversation, setting detail, object, or sentence-level pattern.
   - If the form exists but the function is weak, recommend changing what the element does.
   - If the function is right but the form is generic, recommend changing the concrete expression.
   - If both are missing, route to ideation or foundational design instead of drafting prose.
6. Route with the decision tree:
   - Nothing concrete exists -> idea generation or genre element exploration.
   - The premise feels generic -> cliche transcendence or form replacement.
   - The world feels thin -> systemic worldbuilding.
   - Non-humans feel fake -> species, language, culture, and viewpoint work.
   - Characters feel flat -> character arc or desire-pressure diagnosis.
   - Pacing feels off -> scene sequencing.
   - Dialogue feels wooden -> dialogue workflow.
   - Ending feels weak -> ending workflow.
   - Meaning is unclear -> moral parallax or thematic pressure.
   - Drafting has stalled after planning -> drafting workflow.
   - Prose feels flat -> prose style workflow.
   - Complete draft feels overwhelming -> revision workflow.
7. Recommend one primary intervention. Add a secondary route only if the diagnosis genuinely depends on two layers, such as "character desire is unclear, which is why the pacing stalls."
8. Preserve the author's energy. If two diagnoses are plausible, name the one that gives the author the most immediate traction and mark the other as a reassessment point.
9. Avoid premature drafting:
   - Do not write scenes, paragraphs, dialogue, lore entries, or revised prose unless the author explicitly asks to switch from diagnosis to generation or rewrite.
   - Do not solve every possible weakness at once.
   - Do not convert a diagnostic answer into confirmed canon.
   - Do not create a large revision plan until the diagnosed state is stable enough to justify one.
10. Reassess after the intervention. Tell the author what signal would show that the route worked and what symptom would mean the diagnosis should change.

## What To Read

- For a blank-page or concept problem, read the author's prompt and any project note that contains premise, genre, audience, or constraints.
- For a chapter-level problem, read the target chapter or selection plus any attached note that states the author's concern.
- For a cross-draft problem, use `project_map` to identify the relevant chapters or notes, then read a representative small set before diagnosing.
- For character, world, timeline, institution, rule, or lore symptoms, read the relevant Story Bible entries before proposing any canon-facing route.
- For revision uncertainty, read the current text and use `list_revisions` when the problem depends on comparing prior decisions.
- Do not load supporting data by default. Use `read_skill_file` only when the main diagnosis needs a specific structured reference.

## Routing To Other Skills

- Route to a more specific skill once the state is clear; `$story-sense` should not keep control after a better specialized workflow is available.
- If no specific skill is available in the current runtime, provide the intervention as a concise Edda-native action plan in chat.
- Ask the author before switching into drafting, rewriting, or durable note creation.
- Treat canon-changing work as a proposal until the author confirms it.

## Edda Output Handling

- Return a short diagnosis in chat by default: symptom, likely state, reason, primary route, first next step, and reassessment signal.
- Create an Attached Note only when the author wants the diagnosis saved against a specific chapter, scene, or selection.
- Create a Project Note only when the author wants a durable cross-story diagnosis, route map, or revision triage plan.
- Propose Story Bible updates only when the diagnosis exposes a canon gap or contradiction; do not mark proposed lore as confirmed.
- Do not use Structured Writes in this skill. Switch to the appropriate drafting or rewrite workflow if the author asks for generated text.

## Data Files

This skill has optional structured references converted from the source helper material:

- `data/functions-forms.json` contains abstract story functions and setting-specific forms. Load it with `read_skill_file` when a problem appears to be "the right role exists, but its concrete form is generic or mismatched," or when the author needs role/form alternatives after diagnosis.
- `data/genre-elements.json` contains genre-specific pressure elements. Load it with `read_skill_file` when a diagnosis depends on mystery, thriller, horror, romance, science fiction, or fantasy expectations.

Load at most one data file before the initial diagnosis unless the author explicitly asks for a broader option bank.

## Script Compatibility

The source skill included optional `entropy.ts` and `functions.ts` helpers. This built-in rewrite converts their usable logic into Edda-native diagnostic guidance and data files. Do not call source scripts or assume script execution is available.
