---
name: sleep-story
description: Draft calming fiction built for gentle read-aloud or bedtime listening rather than tension and payoff.
route:
  actionKinds:
    - chat
    - continuation
    - rewrite
  contentKinds:
    - chapter
    - story_text
    - attached_note
    - project_note
  tags:
    - fiction
    - bedtime
    - calming
    - optional
  priority: 54
metadata:
  useCases:
    - The author wants a read-aloud or listen-to-sleep story.
    - Pacing should be gentle, descriptive, and safe.
    - The prose needs calming rhythm, soft imagery, and no cliffhanger pressure.
  doNotUse:
    - The story should build urgency, suspense, or mystery.
    - The request is mainly about meditation instruction rather than fiction.
    - The author wants a conventional dramatic ending.
  status: optional
  source:
    - docs > skills > suggested > fiction > application > sleep-story > SKILL.md
  scriptStatus: no-source-helpers
---

# $sleep-story

Bedtime-story support for creating fiction that occupies attention gently, lowers cognitive load, and lets the listener drift away without needing dramatic payoff.

## Edda Workflow

1. Determine whether the author wants prose now or planning support. If the request asks for a draft, continuation, rewrite, bedtime podcast text, or read-aloud passage, produce story prose. If the request asks for an idea, format, diagnosis, or reusable approach, provide a compact plan or checklist instead of drafting.
2. Read the supplied chapter, selected Story Text, Attached Note, or Project Note before continuing or rewriting. If the sleep story must fit an existing project, read relevant Story Bible entries for established setting, character, timeline, and tone; treat any new lore as a proposal until the author confirms it.
3. Choose a soothing narrative shape:
   - Wandering observer: a character enters a peaceful place, notices details, performs small satisfying actions, and gradually settles.
   - Gentle routine: a character completes a calming ritual such as making tea, tending plants, arranging books, preparing a room, or closing a shop.
   - Infinite journey: a train, boat, path, garden walk, or other slow passage continues through restful scenery without destination pressure.
4. Keep the conflict low-stakes and already-safe. Use tiny frictions such as choosing a cup, finding the softer blanket, waiting for rain to ease, or deciding which path to stroll. Resolve or dissolve them quietly. Do not introduce danger, pursuit, secrets, ticking clocks, injury, betrayal, intense grief, or unresolved mystery.
5. Apply the gentle cognitive load principle. The story should be interesting enough to steady wandering thoughts but not so compelling that the listener tries to stay awake for the ending. Prefer description over action, familiar patterns over surprise, and moment-by-moment presence over plot mechanics.
6. Build safety into the setting and prose. Make the place legible, sheltered, and kind: doors open easily, weather is soft or protective, companions are trustworthy, solitude is comfortable, and every sensory detail supports rest.
7. Layer sensory softness:
   - Visual: warm lamps, muted colors, dusk light, moonlit paths, drifting curtains, slow shadows.
   - Sound: rain, waves, distant trains, ticking clocks, muffled voices, leaves, quiet birds, soft footfalls.
   - Touch: wool, linen, polished wood, smooth stone, warm mugs, cushions, blankets, cool sheets.
   - Scent: lavender, bread, rain, clean linen, tea, old books, wood, earth.
   Use taste only when it remains mild and comforting.
8. Use repetition as comfort rather than filler. Return to sensory anchors, count or list similar objects with slight variation, repeat route markers, echo phrases lightly, and make the pattern predictable enough that missing a sentence does not matter.
9. Downshift pacing across the piece:
   - Opening: mildly inviting setup with no urgency.
   - Middle: gentle exploration, routine, or travel with self-contained micro-scenes.
   - Closing: lower complexity, longer breaths, fewer decisions, and a resting image.
   Use medium to long sentences, soft transitions, 3-5 sentence paragraphs, and a read-aloud pace around 120-140 words per minute.
10. Prefer present tense and, when appropriate, second person for immersion. First or third person is acceptable when the project voice requires it, but keep the prose immediate, unhurried, and easy to follow.
11. Soften language at the word level. Prefer gentle, slowly, quietly, softly, drifting, floating, wandering, meandering, warm, smooth, peaceful, and calm. Replace harsh or arousing words with quieter equivalents: crash becomes settle, bright becomes gentle glow, quick becomes unhurried, excited becomes content.
12. Avoid cliffhangers. End with gentle closure: rest, a completed small task, a safe return, a lamp dimming, a door left peacefully ajar, or a journey continuing without a question that demands an answer. The listener should feel they can stop anywhere without frustration.
13. Review the draft against these criteria before returning it:
   - No urgency, suspense, mystery hook, sudden sound, sharp turn, or dramatic reveal.
   - Description and sensory rhythm carry more weight than plot.
   - Every paragraph can be missed without breaking comprehension.
   - Repetition creates familiarity without sounding mechanical.
   - The close is restful and non-demanding.
   - The prose is suitable for read-aloud sleep content, not conventional dramatic fiction.

## Anti-Patterns

- Do not optimize for entertainment, twist, climax, puzzle, surprise, or payoff.
- Do not create unresolved questions, cliffhangers, ominous foreshadowing, hidden threats, or "just one more chapter" hooks.
- Do not use danger-coded verbs, abrupt sentence fragments, hard sound effects, chase motion, emergency timing, arguments, or high emotional volatility.
- Do not overload the listener with names, rules, exposition, lore, tactical decisions, or intricate worldbuilding.
- Do not turn the request into meditation instruction, breath coaching, affirmations, or ambient soundscape unless the author explicitly asks for that instead of fiction.
- Do not write a dramatic ending. Give a soft landing or let the peaceful pattern continue beyond the frame.

## Edda Output Handling

- Return draft prose, rewrite suggestions, or a short plan in chat by default.
- When drafting, include only the story text unless the author requested notes. If useful, add a brief header naming the chosen shape and setting; do not interrupt a sleep-story draft with analysis.
- Create an Attached Note when the result belongs to one chapter, selected passage, or local story attempt.
- Create or update a Project Note when the author wants a repeatable bedtime format, episode plan, setting bank, sensory palette, or series brief.
- Use Structured Writes only when the author explicitly asks to draft, continue, or rewrite selected Story Text into sleep-story form.
- Do not update Story Bible canon directly for new places, characters, routines, rules, or history. Present those as Story Bible proposals or open questions unless the author explicitly confirms them.
- If the author asks for critique, return a concise diagnosis organized by sleep-story criteria: cognitive load, stakes, safety, sensory softness, repetition, pacing downshift, and closure.

## Script Compatibility

This source skill has no required helper scripts. The Milestone 3.5 rewrite works through Edda-native drafting and revision guidance.
