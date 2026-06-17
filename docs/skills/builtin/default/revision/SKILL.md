---
name: revision
description: Revision diagnosis and pass planning for completed drafts with overwhelm, blindness, conflicting feedback, wrong-level edits, cutting resistance, or endless tinkering.
route:
  actionKinds:
    - chat
    - read_check
    - rewrite
  contentKinds:
    - chapter
    - story_text
    - project_note
    - attached_note
    - story_bible_entry
  tags:
    - fiction
    - revision
    - editing
    - planning
  priority: 86
metadata:
  useCases:
    - A completed draft, chapter, scene, or feedback packet needs a diagnosis before revision starts.
    - The author feels overwhelmed, blind to problems, unable to stop revising, blocked by contradictory feedback, reluctant to cut material, or caught polishing too early.
    - The author needs an ordered revision pass plan, scene decision audit, feedback synthesis, or definition of done.
    - The author asks for targeted rewrite support after structural, scene, or line-level priorities are clear.
  doNotUse:
    - The author is still drafting forward and needs momentum rather than edit diagnosis.
    - The author wants new worldbuilding, plot invention, or outline design rather than improving existing draft material.
    - The author only wants copyediting or sentence polish before story and scene structure have been checked.
    - The request is a specialized continuity, character-arc, dialogue, prose-style, or ending problem that should be routed to a narrower skill after revision diagnosis.
  status: default
  source:
    - docs > skills > suggested > fiction > craft > revision > SKILL.md
  scriptStatus: source-helper-deferred-policy-guidance
---

# $revision

Revision help for completed draft material: diagnose the revision state first, then work from large-scale structure toward scene, character, dialogue, line, and polish passes without erasing what already works.

## Edda Workflow

1. Confirm the revision scope before advising: whole manuscript, chapter, scene, selection, reader feedback, or an existing revision note. Use `read_content`, `read_chapter`, `read_entry_section`, `read_story_bible_entry`, `search_content`, `project_map`, and `list_revisions` as needed for the scope.
2. Diagnose before rewriting. Name the author's current revision state, the scale of change required, and the next pass. Do not start line edits while structural or scene survival questions are unresolved.
3. Preserve strengths explicitly. Before recommending cuts or rewrites, identify what should be kept: working tension, clear character choices, effective images, strong dialogue turns, fulfilled promises, or scenes that carry necessary plot or emotional weight.
4. Separate change scale:
   - Structural changes alter premise execution, plot logic, sequence, pacing architecture, protagonist agency, arc completion, theme, opening, climax, or ending.
   - Scene changes alter individual scene goal, conflict, outcome, entry, exit, transition, sequel/reaction, or keep/cut/combine/revise decisions.
   - Character changes alter motivation, consistency, voice, arc progress, relationship dynamics, or whether transformation is earned.
   - Dialogue changes alter subtext, tension between speakers, exposition load, character distinctiveness, or whether an exchange advances plot, reveals character, or builds dynamics.
   - Line changes alter sentence clarity, rhythm, verb strength, filter words, redundancy, paragraph flow, description balance, or read-aloud smoothness.
   - Copy/polish changes alter spelling, grammar, punctuation, formatting, name consistency, timeline details, physical descriptions, and style consistency.
5. Work top-down unless the author explicitly narrows the task: developmental structure first, then scene, character, dialogue, prose, and polish. Explain any exception and keep it local.
6. Track proposed changes by status. Use `keep`, `cut`, `combine`, `revise`, or `pending` for scenes; use `accept`, `reject`, `defer`, or `needs author decision` for feedback items; mark canon-impacting items as proposals until confirmed.
7. Ask at most one necessary clarifying question when the next pass cannot be chosen from available context. Otherwise provide the diagnosis and a concrete next action.

## Revision Diagnosis

Use these states to identify the blockage before building a plan:

- `R1 Overwhelmed`: The draft is complete, too many problems are visible, priorities are unclear, or the author is fixing randomly. Start with a structural pass, establish priorities, and focus on one problem type per pass.
- `R2 Blind`: The author has reread too often and cannot see problems. Recommend distance, format change, read-aloud review, or external readers before more editing.
- `R3 Endless`: Revision never stops, each pass creates more tinkering, or the author cannot declare done. Define pass goals, set revision-round limits, distinguish real problems from preference, and stop when returns diminish.
- `R4 Conflicted`: Feedback contradicts itself. Gather all feedback first, look for repeated patterns, separate problem from preference, identify the underlying issue behind contradictory comments, and protect authorial vision.
- `R5 Delete-Phobic`: The author resists cutting material that weakens the manuscript. Preserve the useful purpose of cut material in a note if needed, then ask whether removal strengthens the remaining story.
- `R6 Wrong Level`: The author is polishing prose before structure or scene necessity is settled. Stop bottom-up editing, verify structure and scene survival, then return to line work later.

## Pass Order And Criteria

Use focused passes instead of trying to fix everything at once:

1. `Structural`: Check the dramatic question, plot holes, story logic, protagonist goal and agency, arc completion, pacing, escalation toward climax, climax as highest-tension point, and whether the ending satisfies and emerges from the story.
2. `Scene`: For each scene, identify POV goal, conflict, disaster or outcome, plot/character advancement, entry and exit timing, transition clarity, scene-sequel rhythm, and whether the scene should be kept, cut, combined, or revised.
3. `Character`: Check the protagonist's starting lie or false belief, final truth, transformation moments, major-character motivations, motivation consistency, earned changes, distinct voice, and visible arc progress in choices or behavior.
4. `Dialogue`: Check subtext, tension between speakers, whether characters say exactly what they mean when they should not, exposition dumps, exchange function, voice distinctiveness, speech-pattern consistency, and unobtrusive tags.
5. `Prose`: Check intentional passive voice, weak verbs, filter words, adverb overuse, redundant phrasing, clear pronoun references, sentence and paragraph variation, flow between paragraphs, specific detail, and description integrated with action.
6. `Polish`: Check spelling, grammar, punctuation, formatting, character and place name consistency, timeline consistency, physical descriptions, world-rule consistency, and a final read-aloud or changed-format review.

Do not move to a smaller pass until the larger pass has an explicit enough `done` condition. If a smaller pass reveals a larger flaw, return to the larger pass and protect any good smaller-scale work that can survive the change.

## Feedback And Decision Handling

- Treat repeated feedback about the same location or effect as a likely problem, even if readers describe it differently.
- Treat one-off stylistic reactions as preferences unless they expose a concrete craft failure.
- Reconcile contradictions by asking what underlying issue both readers may have sensed: uneven pacing can produce both "too fast" and "too slow" comments in different sections.
- Prioritize changes by story impact, not by order received or emotional force.
- Preserve the author's intended effect when proposing alternatives. Do not flatten a distinctive choice merely because one reader disliked it.

## Applied Revision Boundaries

- For diagnosis tasks, produce a pass plan or change log rather than rewriting.
- For rewrite tasks, apply only the chosen scope and pass. Keep structural suggestions separate from line-level replacements.
- When rewriting a scene or selection, state which strengths were preserved and which pass the rewrite serves.
- For canon-affecting changes to durable facts, names, timeline, world rules, character history, institutions, or setting logic, propose Story Bible updates separately and wait for author confirmation before treating them as confirmed canon.
- Do not encourage endless tinkering. A pass is done when its defined criteria are met and the next pass would produce lower-impact preference changes rather than meaningful fixes.

## Edda Output Handling

- Return short diagnoses, next-pass recommendations, and clarifying questions in chat.
- Create an Attached Note when the diagnosis, scene audit, change log, or rewrite plan belongs to one chapter, scene, or selection.
- Create a Project Note for whole-draft pass plans, feedback synthesis, revision limits, definitions of done, or cross-chapter change tracking.
- Use Structured Writes only when the author explicitly asks to apply selected changes and the target content is known.
- Use Story Bible proposals only for revision discoveries that should become durable canon after author confirmation.

## Script Compatibility

The source `revision-audit.ts` script is treated as deferred helper policy, not normal runtime behavior. Its useful logic has been converted into guidance here: scene splitting and word counts become manual scene audits; pass checklists become the six ordered pass criteria; scene decisions use `keep`, `cut`, `combine`, `revise`, and `pending`; full checklist output becomes an Attached Note or Project Note. Do not ask the author to run the script manually. Use `skill_script` only if a future audited and enabled helper is available, and only for non-mutating audit reports.
