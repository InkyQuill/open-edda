---
name: outline-collaborator
description: Active outline collaboration for requested beats, structure options, escalation, causality, pacing, and character-arc planning.
route:
  actionKinds:
    - chat
    - continuation
    - rewrite
  contentKinds:
    - project_note
    - attached_note
    - story_bible_entry
    - entry_section
    - chapter
  tags:
    - fiction
    - outline
    - structure
    - collaboration
  priority: 80
metadata:
  useCases:
    - The author asks to develop, iterate, continue, or improve a story outline.
    - The author requests scene beats, act plans, plot alternatives, pacing options, or character-arc mapping.
    - An existing outline, act map, chapter plan, or structural note needs active proposal-based collaboration.
    - The author wants options to choose from before drafting prose.
  doNotUse:
    - The author wants only coaching questions without proposed structure.
    - The author wants finished scene prose or ready-to-use drafting rather than outline-level planning.
    - The request is sentence-level editing, prose polish, or voice revision.
    - The author wants canon changes recorded as confirmed facts without review.
  status: default
  source:
    - docs > skills > suggested > fiction > structure > outline-collaborator > SKILL.md
  scriptStatus: no-source-helpers
---

# $outline-collaborator

Active structural partnership for developing outlines while keeping the author in control of direction, selection, and final canon.

## Edda Workflow

1. Read the relevant project context before proposing structure: current outline notes, selected chapter or attached note, recent revisions if continuity matters, and Story Bible entries for characters, setting rules, timeline, factions, or other canon touched by the outline.
2. Separate established facts from assumptions. Treat new plot events, world rules, character history, relationship turns, names, and timeline changes as proposals until the author confirms them.
3. Confirm the working scope in one concise sentence when it is not already clear: act, sequence, chapter plan, scene list, character arc, pacing revision, or structural alternative set.
4. Collaborate actively. Do not respond only with questions unless the author asked for coaching. Offer outline-level material the author can react to, then invite adjustment, rejection, recombination, or redirection.
5. Generate scene beats only when the author requests beats or when an outline problem cannot be solved without beat-level structure. Keep beats at outline level: goal, conflict, escalation, turn or disaster, and sequel when needed.
6. Label every contribution by type so the author can distinguish content from craft reasoning:
   - `Structure proposal:` beat sequence, act map, scene order, or outline revision.
   - `Option A/B/C:` distinct structural paths, each with a short note on what it accomplishes differently.
   - `Arc map:` lie, want, need, pressure points, transformation milestones, and where plot forces the change.
   - `Causality check:` why each event follows from prior choices, consequences, constraints, or conflicts.
   - `Escalation note:` how pressure increases, reverses, narrows options, or changes stakes.
   - `Sample sketch:` brief exploratory prose or dialogue only when requested or useful to illustrate an outline beat.
7. Prefer multiple structural options over a single asserted solution when the story could plausibly branch. Do not advocate as if one option is final unless the author asks for recommendation; instead explain the tradeoff and let the author choose.
8. Preserve causality. For each proposed major beat, make clear what causes it, what it changes, and what new problem it creates. Avoid beats that exist only because the outline needs them.
9. Preserve escalation. Check that scenes and sequences do not reset pressure after every event, resolve too cleanly, repeat the same obstacle, or skip aftermath. Use scene-sequel rhythm where useful: action leads to reaction, dilemma, and decision.
10. Integrate character arc with plot structure. When mapping arcs, connect lie, want, need, fear, wound or ghost, and transformation milestones to external events. A character change must be earned by pressures in the outline, not merely declared.
11. Keep samples subordinate to structure. A sample sketch may show an opening move, a dialogue fragment, or tonal direction, but it must be brief, clearly labeled exploratory, and not presented as finished draft prose.
12. Maintain author control. Mirror the author's stated genre, tone, names, constraints, and story logic. Signal that proposals are optional, and ask for the next decision point after giving usable structure.
13. Redirect when the request crosses the boundary into full drafting: offer to add outline-level sample approaches, but state that finished prose drafting belongs to a drafting/story-collaboration skill.

## Edda Output Handling

- Use chat for live collaboration, option sets, quick beat proposals, causality checks, and decisions the author is still making.
- Use an Attached Note when the structural output belongs to one chapter, scene, selection, or local revision problem.
- Use a Project Note when preserving a durable outline, act map, beat bank, pacing plan, character-arc map, option bank, or collaboration notes across sessions.
- Use Story Bible proposals, not confirmed updates, when the chosen structure would change durable canon such as character facts, world rules, relationships, institutions, timeline, names, geography, history, or genre-specific systems.
- Use Structured Writes only when the author explicitly asks to update existing outline material in Edda and the target note, entry, or selection is known.
- When recording output, preserve the distinction between `selected structure`, `open options`, `author constraints`, `sample sketches`, and `canon proposals`.

## Script Compatibility

This source skill has no helper scripts. The rewritten skill runs entirely through Edda context tools, chat collaboration, notes, Story Bible proposals, and explicit structured writes.
