---
name: drafting
description: Momentum-first drafting help for blank pages, stalled chapters, rough continuations, next beats, and first-draft blocks caused by overplanning or premature polish.
route:
  actionKinds:
    - chat
    - continuation
    - rewrite
  contentKinds:
    - chapter
    - story_text
    - project_note
    - attached_note
  tags:
    - fiction
    - drafting
    - momentum
    - continuation
  priority: 86
metadata:
  useCases:
    - The author is staring at a blank page, stalled chapter, or unwritten scene and needs words to exist before they are polished.
    - The author asks for a rough continuation, next beat, scene entry, zero draft, dialogue-only start, or constraint-based way into a scene.
    - The author is overplanning, waiting to feel ready, rereading, sentence-polishing, or treating the first draft like final prose.
    - A scene or chapter has a known direction but the author cannot get started, continue, finish, or move through the middle.
  doNotUse:
    - The author needs earlier-stage story discovery because there is no outline, premise, character, setting, or clear enough story direction to draft from.
    - The author wants developmental diagnosis, scene sequencing, character-arc repair, or structural redesign before writing more prose.
    - The draft is complete and the task is revision, line editing, prose polish, critique, or style refinement.
    - The author only wants process coaching and explicitly does not want draft text, rough options, placeholders, or continuation material.
  status: default
  source:
    - docs > skills > suggested > fiction > craft > drafting > SKILL.md
  scriptStatus: no-source-helpers
---

# $drafting

Drafting support that turns hesitation into rough Story Text, using the principle that drafting is discovery, not transcription.

## Edda Workflow

1. Establish the working edge before giving advice. Use `read_chapter` or `read_content` for the current chapter or selection, `project_map` for nearby chapters or notes when needed, and `read_story_bible_entry` or `read_entry_section` only for constraints that affect the immediate scene. Do not perform broad research when the author needs momentum.
2. Separate drafting from editing. Treat the current task as generative unless the author explicitly asks for a localized rewrite. Keep the internal editor off: no sentence polishing, no style critique, no backward revision beyond enough context to continue.
3. Identify the drafting state:
   - `D1: Can't Start` means the outline or idea exists but the page is still blank. Look for fear of the blank page, high first-draft standards, vague entry conditions, or endless preparation.
   - `D2: Starts But Can't Continue` means there are first pages or scenes, then the draft stalls. Look for rereading, polishing, dead-end stopping points, or unclear current-scene action.
   - `D3: Middle Stall` means the beginning exists but the middle has lost pressure or direction. Check whether the ending is known, whether the next connective beat is missing, and whether the draft has revealed a structural gap.
   - `D4: Can't Finish` means the draft is near the end but circling. Look for fear of judgment, unclear ending, adding instead of concluding, or reluctance to write an imperfect ending.
   - `D5: Draft Taking Too Long` means the draft is moving too slowly because revision is disguised as drafting. Check whether the author is rereading too much, revising each sentence, or using goals too small to build momentum.
   - `D6: Draft Reveals Story Problem` means forward motion has exposed a real uncertainty. Distinguish prose dissatisfaction from story structure: prose problems get marked and pushed through; fundamental structure may get one brief re-outline pass before drafting resumes.
4. Decide whether to draft, ask, or route away:
   - Draft now when the author has enough premise, character, setting, and immediate direction to produce the next beat, even if the exact prose will be bad.
   - Ask one or two sharp questions only when a missing fact blocks the next paragraph, such as who enters, what the character wants, what changes by scene end, or which constraint cannot be violated.
   - Route away when the missing issue is not draft momentum: no story direction, unresolved structure, undefined character motivation, continuity risk, or requested polish.
5. For a blank page, give a low-friction entry instead of more planning. Use one of these starts:
   - Bad draft permission: state that the next output is a zero draft whose job is to exist, not to be good.
   - Scene entry: start at the first visible action, spoken line, interruption, physical detail, decision point, or conflict cue.
   - Constraint-based start: draft under a simple constraint such as dialogue only, 100 bad words, ten-minute sprint, one sensory anchor, no backtracking, or `[PLACEHOLDER]` for unknown details.
   - Summary start: write "What happens in this scene is..." as rough narrative scaffolding, then convert the first usable piece into prose.
6. For a stalled continuation, find the next beat and draft through it. The next beat should answer at least one of: what changes now, what pressure arrives, what choice appears, what information lands, what the character does because they cannot keep waiting, or what line of dialogue forces a response.
7. Use placeholders deliberately. If a name, transition, description, motive, or connective action is blocking momentum, mark it with a bracketed placeholder and keep drafting. Do not let a missing detail become a reason to stop.
8. Avoid overplanning. If the author keeps requesting more outline detail while no prose exists, set a concrete drafting threshold first: a zero-draft paragraph, a dialogue-only exchange, a rough scene card expanded into prose, or 100-500 intentionally imperfect words.
9. When producing draft text, keep it rough and usable. Prefer short continuations, scene openings, dialogue runs, beat-to-prose conversions, and alternate entry points. Do not present draft text as final copy; label it as rough draft material.
10. End with a continuation handle. Leave the author with the next sentence, next action, next beat, or a small drafting assignment that makes returning easy. When appropriate, advise stopping mid-scene or mid-movement rather than at a clean endpoint.

## Drafting Interventions

- For `D1: Can't Start`, offer a zero draft, dialogue-only start, "What happens is..." summary, first visible action, or 100 bad words. If the author is waiting to feel ready, ask for the smallest concrete scene constraint and draft immediately.
- For `D2: Starts But Can't Continue`, skip the stuck connective tissue, write the scene that can be written, use `[SOMETHING HAPPENS HERE]`, lower the session goal, or resume from the last unfinished action instead of rereading whole chapters.
- For `D3: Middle Stall`, reconnect the current scene to the ending, ask what the worst credible complication is now, draft an exciting later scene out of order, or create one bridge beat instead of redesigning the whole middle.
- For `D4: Can't Finish`, write the ending now in bad form, close the current conflict even if the prose is inadequate, and remind the author that a finished bad draft can be revised while an unfinished one cannot.
- For `D5: Draft Taking Too Long`, switch to time-boxed drafting, a daily word target, no-reread rules beyond the last paragraph, or speed over quality until the current unit exists.
- For `D6: Draft Reveals Story Problem`, name the problem and choose one path: push through and flag for revision, pause briefly to adjust the outline, follow the draft's discovery, or write two versions for later comparison.

## Quality Criteria

Good drafting output:

- Produces new rough material or a specific next beat rather than more abstract preparation.
- Makes the first move easy: a line of dialogue, visible action, pressure event, decision, sensory anchor, or placeholder.
- Protects forward momentum by lowering standards, narrowing scope, and separating first-draft generation from later revision.
- Uses enough project context to honor established constraints without turning the session into continuity research.
- Leaves the author with a clear next drafting action.

Bad drafting output:

- Critiques or polishes prose before the draft exists.
- Adds outline complexity when the author needs words on the page.
- Asks broad worldbuilding or theme questions that do not block the next paragraph.
- Treats placeholders, rough prose, or bad first attempts as failures.
- Routes to revision, prose style, or developmental critique before the draft or current scene has been carried forward.

## Edda Output Handling

- Use chat for immediate momentum coaching, one or two blocking questions, rough options, next beats, short zero-draft passages, and constraint-based starts.
- Create an Attached Note when the result belongs beside a specific chapter, selection, scene start, alternate continuation, or block diagnosis that should not yet enter Story Text.
- Create or update a Project Note when the author asks for a durable drafting plan, recurring block diagnosis, progress strategy, fallback beat list, or session routine.
- Use Structured Writes such as `append_to_chapter`, `insert_into_chapter`, or `replace_selection` only when the author explicitly asks to apply a continuation or localized draft to Story Text.
- Keep canon changes out of confirmed Story Bible entries. If rough drafting invents durable names, facts, institutions, world rules, timelines, or character history, flag them as provisional and propose a Story Bible update only after author confirmation.

## Script Compatibility

This source skill has no required helper scripts. Its decision logic is converted into Edda-native diagnosis, drafting interventions, output handling, and explicit text actions.
