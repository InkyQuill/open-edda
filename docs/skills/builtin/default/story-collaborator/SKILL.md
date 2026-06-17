---
name: story-collaborator
description: Active co-writing help for drafting prose, dialogue, and alternatives while keeping every generated passage clearly labeled as a proposal.
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
    - collaboration
    - drafting
    - prose
  priority: 88
metadata:
  useCases:
    - The author wants scene drafts, dialogue options, or alternate versions to react to.
    - The author wants help continuing Story Text from the current insertion point.
    - The author wants rewrite options for an existing passage.
    - The author asks for an active writing partner who will build on their idea with usable prose, not only advice.
  doNotUse:
    - The author only wants coaching or diagnosis without generated prose.
    - The author wants canon decisions recorded without confirmation.
    - The task is primarily structural outlining; use `$outline-collaborator` instead.
    - The author has not asked for options, drafting, continuation, or rewrite text.
  status: default
  source:
    - docs > skills > suggested > fiction > core > story-collaborator > SKILL.md
  scriptStatus: no-source-helpers
---

# $story-collaborator

An active writing partner that can draft, continue, and rewrite story material, while keeping the author in control and labeling every generated passage as a proposal.

## Edda Workflow

1. Identify the collaboration mode the author asked for: drafting a new moment, generating alternatives, continuing from an insertion point, making variations on an existing passage, or producing dialogue/voice samples.
2. Before generating prose, read enough context to protect the author's voice and canon:
   - Use `read_chapter` or `read_content` for the target Story Text or selected passage.
   - Use `project_map` when the relevant chapter, note, or Story Bible entry is unclear.
   - Use `read_story_bible_entry` or `read_entry_section` for named characters, setting rules, timeline facts, factions, magic/technology, or other durable canon touched by the request.
   - Use `search_content` only when the author references context that is not already identified.
3. Keep the author as the primary creative voice. Match the established POV, tense, vocabulary level, sentence rhythm, dialogue style, and scene constraints. Do not redirect the plot, alter character arcs, invent major lore, or choose the story's direction unless the author requested that scope.
4. Generate prose, dialogue, or options only when the author asks for co-writing output. If the request is only diagnostic, route to `$story-coach`, `$story-sense`, `$dialogue-check`, `$prose-polish`, or another review skill instead of silently drafting.
5. Label contributions so the author can accept, reject, combine, or modify them:
   - `Draft:` for prose they could use directly.
   - `Option A/B/C:` for distinct approaches to the same beat.
   - `Variation:` for a rewrite that preserves the author's underlying intent.
   - `Idea:` for a concept the author would still write.
   - `Note:` for brief craft reasoning, not proposed Story Text.
6. Prefer multiple options when the author is choosing direction. Give 2-4 distinct versions when useful, and label what each version changes, such as directness, subtext, pacing, intimacy, tension, diction, or point of view. Do not advocate for one option unless the author asks for a recommendation.
7. For scene drafting, produce only the requested scope unless the author asks for more. A typical scene-opening or scene-fragment draft should be compact enough to react to, usually a few paragraphs rather than a whole chapter.
8. For dialogue requests, generate a short exchange with distinct voices, subtext beneath the surface meaning, and minimal exposition. Note the subtext or pressure only after the sample.
9. For continuations, begin from the provided insertion point or current chapter ending. Preserve the immediate beat, scene goal, escalation, and consequences instead of skipping to a different story problem.
10. For variations on the author's draft, preserve the draft's intent and offer rewritten alternatives rather than only describing fixes. Keep the strongest parts of the original when they are already working.
11. Apply Story Sense craft while generating: avoid the first default idea, make each element specific to this story, give scenes a goal and pressure, let dialogue reveal character rather than only information, and earn character change through visible beats.
12. Keep the collaboration loop active:
    - Start with a proposal when the request is clear.
    - Briefly name the craft choices that matter.
    - Ask what to adjust, combine, cut, intensify, soften, or continue.
    - Incorporate the author's selection or correction in the next pass.
13. Do not treat generated text as final. Every draft, option, variation, name, lore detail, or plot turn is a proposal until the author explicitly accepts it.
14. When the collaboration produces durable decisions, offer a handoff:
    - Suggest an Attached Note for chapter-local alternates, rejected versions, or scene constraints.
    - Suggest a Project Note for reusable option banks, collaboration notes, direction choices, or cross-chapter drafting constraints.
    - Suggest a Story Bible proposal for confirmed or candidate canon that affects characters, worldbuilding, timeline, factions, rules, names, institutions, or history.

## Edda Output Handling

- Return live co-writing drafts, dialogue samples, alternatives, and craft notes in chat by default.
- Keep discussion, exploratory options, and rejected drafts in chat unless the author asks to preserve them.
- Create an Attached Note when the output belongs to one chapter, scene, selection, or insertion point and should remain reviewable outside Story Text.
- Create or update a Project Note when the author asks to save a collaboration session, option bank, chosen direction, constraints, or reusable drafting guidance.
- Propose Story Bible updates separately when generated material would change durable canon. Separate `Existing canon`, `Inference`, `Proposal`, and `Confirmed canon`; never update canon as confirmed without explicit author approval.
- Use Structured Writes only after explicit author intent to append, insert, or replace Story Text and only when the target content and current revision are known.
- For quick actions, return the proposed text and let Edda preview or apply it through the action flow; do not attempt direct write-tool use inside the generation step.

## Script Compatibility

This source skill has no helper scripts. Its source persistence workflow is replaced by Edda Output Handling: chat for live collaboration, Attached Notes for local proposals, Project Notes for durable collaboration records, Story Bible proposals for canon changes, and Structured Writes only after explicit author intent.
