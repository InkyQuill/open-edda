---
name: scene-pacing
description: Scene and chapter pacing diagnosis for weak escalation, missing aftermath, clean victories, and sequences that do not accumulate pressure.
route:
  actionKinds:
    - chat
    - read_check
  contentKinds:
    - chapter
    - story_text
    - attached_note
    - project_note
  tags:
    - fiction
    - pacing
    - structure
    - scenes
  priority: 82
metadata:
  useCases:
    - A scene has no clear goal, weak escalation, or too-clean resolution.
    - A chapter feels both slow and exhausting.
    - The author wants to inspect scene-to-scene rhythm before rewriting prose.
  doNotUse:
    - The author wants full outline generation rather than scene diagnosis.
    - The issue is mainly line style, dialogue polish, or canon building.
    - The text is still too early and vague for scene-level diagnosis.
  status: default
  source:
    - docs > skills > suggested > fiction > structure > scene-sequencing > SKILL.md
  scriptStatus: source-helpers-converted-to-guidance
---

# $scene-pacing

Pacing and sequencing review for scenes that feel slow, exhausting, flat, or mechanically connected instead of dramatically linked.

## Edda Workflow

1. Read the target scene, chapter, attached note, or project note in sequence. If the author asks about multiple scenes, preserve their order and inspect how one scene hands pressure to the next.
2. Name the scene function before diagnosing craft: setup, confrontation, reveal, reversal, recovery, decision, bridge, payoff, or transition. If the function is unclear, treat that as a pacing fault.
3. For each scene or major beat, identify the concrete scene goal: what the POV character wants now, whether the reader can infer it early, and how it connects to the larger story pressure.
4. Identify the conflict opposing that goal. Check whether the opposition escalates beat by beat or merely repeats at the same pressure level.
5. Identify the turn or outcome. Prefer outcomes that change the situation: "yes, but", "no, and furthermore", or a costly partial answer. Flag too-clean victories when the character gets the goal without cost, complication, new information, or narrowed options.
6. Check aftermath or sequel handling after the turn: reaction, dilemma, and decision. The aftermath may be one paragraph or several pages, but it should let the reader process the hit and should move toward a choice.
7. Trace the causal chain: the previous turn should create this scene's goal, and this scene's decision or new pressure should create the next scene's goal. Flag mechanical transitions where scenes are merely adjacent.
8. Diagnose pacing by locating missing pressure, misplaced pressure, or overextended pressure:
   - Missing pressure: unclear goal, weak opposition, no ticking constraint, no cost, or no real dilemma.
   - Misplaced pressure: high-action prose during aftermath, introspection inside a conflict beat that needs motion, or exposition before the reader knows why it matters.
   - Overextended pressure: repeated conflict beats, endless reaction without decision, or relentless action without recovery.
9. Compare scene and aftermath balance to the author's intent. Fast pacing usually compresses aftermath; reflective pacing expands it. Do not prescribe fixed scene lengths.
10. Return practical revision targets: establish the goal earlier, sharpen the obstacle, add escalation, complicate a clean outcome, insert a brief aftermath, compress a wallowing sequel, or move a beat so the causal chain becomes legible.

## Diagnostic Criteria

Use these criteria in the report when they are relevant:

- **Scene goal:** The POV character has a specific, local want that can be pursued inside the scene.
- **Conflict:** Something actively opposes the goal, and later beats make success harder than earlier beats did.
- **Turn:** The scene changes the story state through success with cost, failure with consequence, discovery, reversal, reveal, or commitment.
- **Aftermath/sequel:** The character reacts, faces a real dilemma, and makes or approaches a decision that can drive the next scene.
- **Escalation:** Pressure accumulates instead of resetting, repeating, or resolving without residue.
- **Causal chain:** Each scene's outcome creates the next scene's goal, constraint, danger, or emotional burden.
- **Pacing diagnosis:** Explain whether the scene feels slow, rushed, exhausting, flat, or mechanically connected because of structural rhythm, not only prose style.
- **Too-clean victories:** Flag unqualified wins that lower tension unless the author is intentionally creating relief, closure, or contrast.
- **Missing pressure:** Flag unclear stakes, easy access to answers, absent consequences, weak opposition, and choices with no cost.
- **Sequence function:** State what the scene contributes to the chapter or sequence and whether the current length and emphasis match that function.

## Edda Output Handling

- Return the pacing diagnosis in chat for immediate use.
- Create an Attached Note when the report belongs to a specific chapter or scene selection.
- Create or update a Project Note when the pacing issue spans several chapters or a whole act.
- Keep canon separate from craft diagnosis. If pacing failure depends on a durable fact, rule, timeline, location constraint, or character capability, frame any change as a Story Bible proposal that needs author confirmation.
- Do not rewrite Story Text by default. If the author explicitly asks for applied revision, provide a small targeted replacement or revision plan using the diagnosed structural issue as the constraint.
- Keep output proportional: short chat diagnosis for one scene, Attached Note for a chapter-local breakdown, Project Note for cross-chapter sequencing.

## Deferred Helper Policy

The source `analyze-scene.ts` helper is methodology only for this built-in rewrite. Convert its checks into agent reasoning rather than calling or exposing a script:

- Look for goal, conflict, turn/disaster, reaction, dilemma, and decision signals.
- Estimate whether the passage is action-heavy, balanced, or reflective from the relative amount of conflict prose versus processing prose.
- Treat keyword-style detections as prompts for manual judgment, not proof that a scene is working.
- Do not ask the author to run Deno, shell commands, local files, or source helper scripts.

## Script Compatibility

Source helper scripts have been converted to guidance. No script is available to the runtime agent unless a future admin-approved, non-mutating `skill_script` helper is added and explicitly enabled.
