---
name: revision-planner
description: Whole-novel revision planning with ordered passes, cascade tracking, rollback criteria, and risk control.
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
    - revision
    - planning
    - optional
  priority: 56
metadata:
  useCases:
    - A draft exists and the author needs a whole-novel or multi-chapter revision plan before editing.
    - A proposed change may ripple through structure, character arcs, continuity, themes, or later prose.
    - Revisions keep creating new problems, cascade debt, or uncertainty about what pass should happen next.
    - The author wants triage, rollback criteria, monitoring checkpoints, or a pass order for structural, character, continuity, and prose work.
  doNotUse:
    - The author wants immediate sentence-level polishing or a direct prose rewrite rather than a revision plan.
    - The work is still in discovery drafting and no stable draft exists to diagnose.
    - The request is only local scene repair without cross-chapter consequences.
    - The author needs primary story diagnosis before deciding what should be revised.
  status: optional
  source:
    - docs > skills > suggested > fiction > structure > novel-revision > SKILL.md
  scriptStatus: no-source-helpers
---

# $revision-planner

Plan whole-novel revision as multi-level change management: diagnose before editing, protect what already works, order passes by dependency, and prevent one fix from creating untracked new problems.

## Edda Workflow

1. Establish the revision target before prescribing edits. Use `project_map` to identify available chapters, Story Text, Story Bible entries, project notes, and prior revision notes. Use `read_content`, `read_chapter`, `read_story_bible_entry`, `read_entry_section`, and `list_revisions` as needed to understand the draft state, author goal, and already-recorded decisions.

2. Separate diagnosis from implementation. Do not recommend rewriting until you can state:
   - the current problem in story terms,
   - the evidence from the manuscript or notes,
   - the level of change required,
   - the strengths that should be preserved,
   - the risks of acting now versus waiting.

3. Classify each candidate change by level:
   - Conceptual: theme, meaning, core promise, character growth logic, symbolic pattern, emotional journey.
   - Structural: plot beats, causality, scene order, chapter organization, pacing, tension and release cycles, climax or resolution setup.
   - Character: motivation, agency, relationship movement, arc visibility, behavioral consistency, how others respond to the character.
   - Continuity: timeline, facts, names, locations, rules, prior promises, Story Bible alignment, established consequences.
   - Manuscript/prose: actual wording, dialogue, description, voice, style, scene texture, local clarity.

4. Triage before ordering passes. Rank problems by dependency and story damage:
   - Foundation first: concept, premise promise, ending logic, protagonist desire, major causality, and non-negotiable canon conflicts.
   - Architecture second: plot sequence, chapter order, escalation, reveal timing, missing aftermath, and pacing rhythm.
   - Character and relationship passes third unless they are the foundation problem; track where changed motivation alters later choices.
   - Continuity after structural and character decisions are stable enough to check facts against them.
   - Prose and line polish last, after scenes are likely to remain.
   - Defer optional beautification when it risks endless revision or distracts from the pass goal.

5. Build the pass order around dependencies, not around manuscript order alone. A good plan usually moves from diagnosis to structural pass, character pass, continuity pass, scene-level execution pass, then prose polish. Change this order only when evidence shows a different dependency, such as a character motivation issue driving the plot failure.

6. For every significant change, create a cascade map before edits:
   - Immediate consequences: what must change in the next one or two chapters?
   - Medium-term consequences: what changes three to five chapters later?
   - Story-wide consequences: what changes in the ending, major turns, recurring motifs, promises, or canon?
   - Reverse consequences: what earlier setup, foreshadowing, motivation, or world rule must change so the new moment is earned?
   - Preservation targets: which scenes, lines, beats, character qualities, or emotional effects should survive the revision?

7. Define monitoring criteria for each pass. Use concrete warning signs:
   - character behavior no longer matches established motivation or development,
   - other characters fail to respond to changed behavior,
   - events no longer follow logically,
   - tension drops, spikes without setup, or resolves too cleanly,
   - a scene starts dragging because it now carries too much exposition,
   - the revised event undermines the intended theme,
   - a canon fact, timeline, or rule now conflicts with a Story Bible entry,
   - the change creates many secondary tasks but no clear story benefit.

8. Use controlled implementation as the planning model. Recommend the smallest viable revision that tests the hypothesis, then checkpoints before expanding it. For large changes, specify rollback points: the last stable draft state, the condition that would mean the change is not working, and the alternate strategy to try next.

9. Manage risk explicitly. Push forward when problems are localized, benefits clearly outweigh complications, cascade tasks are well defined, and core story logic remains intact. Recommend rollback or redesign when multiple warning signs appear within two chapters, fundamental character or plot logic breaks, required cascade tasks become overwhelming, or the change creates more problems than it solves.

10. Protect strengths. Every pass plan must name what not to break: effective scenes, strong emotional turns, distinctive voice, useful ambiguity, working tension, memorable dialogue, reader promises, or character dynamics. Do not flatten the manuscript by making every element serve the current problem equally.

11. Avoid endless revision. Define a pass goal and a stop condition before the pass begins. A pass is complete when its stated problem is resolved to the agreed standard and remaining issues belong to a later pass or author choice. Do not keep adding new goals to an active pass; record them as later tasks or open questions.

12. Keep canon and durable story facts reviewable. When the plan implies changes to names, timeline, world rules, character history, institutions, relationships, or other Story Bible material, label them as proposals until the author confirms them. Do not treat revision discoveries as confirmed canon by default.

## Revision Pass Methods

### Structural Pass

Use when pacing is off, events feel disconnected, the climax lacks impact, or scenes do not build on each other. Create or summarize the current beat timeline, identify broken causality or escalation, then order changes so earlier setup and later payoff stay aligned. Success looks like connected plot points, earned turns, accumulating pressure, and a climax or resolution that uses what the draft prepared.

### Character Pass

Use when a character feels flat, motivation is unclear, agency disappears, relationships jump, or an arc is invisible. Map the character's state across chapters, identify where growth should be visible, and mark which dialogue, choices, and reactions must change later. Success looks like believable behavior, visible development, and other characters responding appropriately to the revised person.

### Continuity Pass

Use when structural or character decisions are stable enough to check facts. Compare revised chapters against Story Bible entries, prior chapters, project notes, timeline, names, rules, and consequences. Success looks like explicit resolution of fact conflicts and a list of proposed canon updates separated from manuscript edits.

### Thematic Pass

Use when theme is heavy-handed, unclear, inconsistent, or contradicted by the ending. Audit recurring choices, images, costs, and resolutions. Prefer subtle integration through character action and consequence over explanatory speeches. Success looks like theme emerging naturally and the ending satisfying the story's emotional argument.

### Prose Pass

Use only after higher-level decisions are stable enough that the text is likely to remain. Focus on voice, dialogue, description, rhythm, clarity, and local texture while preserving prior structural and character decisions. Success looks like cleaner pages without reopening foundation questions.

## Change Record Shape

For significant changes, return or create a compact change record with:

- Revision: brief description.
- Change level: conceptual, structural, character, continuity, thematic, or prose.
- Rationale: why this change is needed and what evidence supports it.
- Preservation targets: what must remain strong.
- Predicted consequences: immediate, medium-term, story-wide, and reverse setup.
- Monitoring criteria: warning signs and success indicators.
- Pass order: where this change belongs relative to other passes.
- Cascade tasks: required follow-up edits, marked pending or done.
- Rollback trigger: what would mean this approach should stop.
- Outcome assessment: to complete after the pass.

## Edda Output Handling

- Use chat for short triage, pass-order discussion, rollback advice, or when the author is still deciding.
- Create an Attached Note when the plan belongs to one chapter, scene, selection, or local change with limited ripple effects.
- Create or update a Project Note for whole-novel revision plans, multi-chapter pass orders, cascade task lists, monitoring logs, rollback points, and cross-session revision records.
- Propose Story Bible updates only as reviewable follow-ups when the plan affects durable canon, character facts, timeline, setting rules, names, institutions, relationships, or history.
- Do not apply manuscript changes unless the author explicitly asks for rewriting after the plan. This skill plans and tracks revision; it does not silently execute the passes.
- Do not use Structured Writes in this skill unless the author explicitly requests a formal note or template output through an Edda write action.

## Script Compatibility

This source skill has no helper scripts. The Edda version runs entirely through project context tools, chat planning, Attached Notes, Project Notes, and reviewable Story Bible proposals.
