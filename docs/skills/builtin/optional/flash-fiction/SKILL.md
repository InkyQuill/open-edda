---
name: flash-fiction
description: Draft, evaluate, or tighten very short fiction so every sentence carries weight and the ending still lands.
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
  tags:
    - fiction
    - short-form
    - revision
    - optional
  priority: 58
metadata:
  useCases:
    - The author wants to draft or revise fiction under roughly 1,500 words.
    - A short piece feels flat, over-explained, or emotionally incomplete.
    - The goal is strong compression rather than expansion.
  doNotUse:
    - The story wants room to breathe as a chapter-length draft.
    - The author mainly needs worldbuilding support rather than short-form execution.
    - The request is for publishing copy instead of fiction.
  status: optional
  source:
    - docs > skills > suggested > fiction > application > flash-fiction > SKILL.md
  scriptStatus: no-source-helpers
---

# $flash-fiction

Short-form fiction support for authors working in drabble, micro, flash, or sudden-fiction lengths where compression, implication, and final resonance matter as much as plot.

## Edda Workflow

1. Establish the task before acting: diagnosis, coaching, brainstorming, new drafting, continuation, or localized rewrite. If the author asks "what is wrong" or "why is this not landing," diagnose first instead of silently drafting a replacement.
2. Read the full short piece, selected passage, attached note, or prompt. If it belongs to a larger project, use `project_map`, `read_content`, `read_chapter`, or targeted Story Bible reads only for context needed to preserve voice, canon, and continuity.
3. Identify the target length band and its craft priority:
   - Under 100 words: one image, one moment, maximum compression.
   - 100-500 words: one scene, one shift, implication over statement.
   - 500-1,000 words: small arc, one or two scenes, strong iceberg effect.
   - 1,000-1,500 words: sudden-fiction scale, with room for limited scene movement and more character pressure.
   - 1,500-2,500 words: short-short territory; warn if the premise wants chapter or short-story breadth instead of flash compression.
4. Find the single pressure the piece can hold. Name the one emotional movement, decision, contradiction, or irreversible turn. Cut or defer subplots, explained histories, extra characters, and premise machinery that compete with that pressure.
5. Diagnose against the flash-fiction states:
   - `FF1 Structure and pacing`: weak first sentence, sagging middle, abrupt or dragged ending, poor word-count distribution, incomplete arc. Test whether the first sentence creates immediate engagement, whether there is a clear turn, whether each paragraph does more than one job, and whether scope fits the length.
   - `FF2 Character compression`: generic characters, dumped backstory, unearned change, explained relationships. Test whether the first action reveals character, history can be inferred, objects or mannerisms imply life beyond the page, and change appears through shifted decisions.
   - `FF3 Beginning/ending frame`: opening and closing disconnect, ending fails the opening's promise, first and last images do not echo, transform, or contrast. Test whether the ending feels surprising but inevitable and leaves a specific after-emotion.
   - `FF4 Subtext`: everything is stated, readers have no gaps to cross, backstory is explained rather than implied. Test what can be omitted, displaced into action, or left as purposeful ambiguity.
   - `FF5 Imagery and figurative language`: flat prose, cliche comparison, no image pattern, heavy-handed symbol. Test whether images reveal character, theme, and emotion at once.
   - `FF6 Setting and sensory detail`: generic location, mostly visual description, atmosphere stated rather than embodied. Test whether time/place orient quickly and whether sound, smell, touch, or taste can carry pressure.
   - `FF7 Theme`: absent, preachy, imposed, or oversimplified. Test whether theme emerges from action and object rather than explanation, and whether the close deepens rather than moralizes.
   - `FF8 Language precision and rhythm`: weak verbs, vague nouns, decorative modifiers, monotone sentence shape. Test whether every word earns its place through precision, rhythm, information, or pressure.
   - `FF9 Logical consistency`: physical impossibilities, timeline contradictions, impossible knowledge, broken rules. Test movement, chronology, cause and effect, and character knowledge boundaries.
6. Apply word economy before adding material. Prefer verbs over adverbs, concrete nouns over explanation, meaningful objects over abstract reflection, and sentence cuts over summary. A sentence should carry at least two functions whenever possible: plot plus character, image plus theme, setting plus emotion, action plus backstory, or rhythm plus turn.
7. Use omitted context deliberately. Preserve the iceberg effect by leaving backstory, relationship history, and world explanation offstage when the visible detail lets the reader infer it. Do not omit information needed for basic orientation, cause, or emotional logic.
8. Build or repair the image/ending relationship. Choose a durable image, object, gesture, line of dialogue, or sensory detail that can return changed at the end. The last sentence should close the current movement while leaving resonance, not merely explain the lesson or announce the twist.
9. Decide whether the title is doing enough work. In flash, the title may provide missing context, frame irony, establish a time/place, name the pressure, redirect the ending, or add a final layer. Do not use a title that merely repeats the obvious subject.
10. Watch for anti-patterns:
   - Miniature novel: the piece summarizes a larger plot instead of creating a complete compressed experience.
   - Twist dependency: the story relies on surprise alone and the preceding sentences do not matter after the reveal.
   - Vignette trap: polished mood or description with no changed pressure by the end.
   - Explanation drift: the piece keeps explaining what an image, gesture, or silence already implies.
   - Multi-pressure sprawl: too many wounds, revelations, settings, or symbolic systems compete inside the word count.
11. For drafting, generate around one pressure, one turn, one title function, and one ending image. Keep context implied unless the author explicitly asks for a more expansive draft.
12. For revision, preserve the sharpest existing sentence, image, or turn where possible. Cut around it rather than flattening the piece into a generic polished version.

## Edda Output Handling

- Return short diagnostics, craft coaching, option lists, or draft text in chat by default.
- When diagnosing, name the dominant `FF` state first, give evidence from the piece, then give the smallest useful intervention. Separate diagnosis from replacement prose unless the author asked for both.
- When drafting, provide the flash text itself and, only if useful, a brief note naming the chosen pressure, title function, and ending/image strategy.
- Create an Attached Note when the diagnosis, cut list, or rewrite plan belongs to one short piece, chapter excerpt, or selection.
- Create or update a Project Note when the author wants a reusable flash-fiction brief, batch plan, title bank, image bank, or revision checklist.
- Do not change Story Bible canon unless the author explicitly wants the short piece folded into project continuity. Treat new lore, names, timeline facts, durable character history, institutions, and world rules as proposals until confirmed.
- Use Structured Writes only when the author explicitly asks to draft, replace, continue, or compress selected Story Text.

## Script Compatibility

This source skill has no required helper scripts. The Milestone 3.5 rewrite works through Edda-native drafting, diagnosis, and revision guidance.
