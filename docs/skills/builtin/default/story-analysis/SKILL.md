---
name: story-analysis
description: Evidence-based post-draft diagnosis for completed chapters or stories, covering structure, character, momentum, payoff, and prioritized revision targets.
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
    - analysis
    - chapter-review
    - revision
  priority: 84
metadata:
  useCases:
    - A chapter or complete story draft is ready for serious evaluation.
    - The author wants more than a quick check and needs a structured report.
    - The author wants strengths, weaknesses, and prioritized revision targets.
  doNotUse:
    - The text is still being actively discovered and the author mainly needs drafting momentum.
    - The request is narrowly about dialogue, prose polish, or pacing.
    - The author wants direct rewriting instead of analysis.
  status: default
  source:
    - docs > skills > suggested > fiction > core > story-analysis > SKILL.md
  scriptStatus: no-source-helpers
---

# $story-analysis

Evidence-based post-draft diagnosis for completed chapters or complete stories, focused on what the draft is trying to do, where it succeeds, where it breaks, and which revisions will matter most.

## Edda Workflow

1. Identify the analysis scope before judging the draft: complete short story, single chapter, chapter range, or project-level pattern. If the user has not supplied enough text, use `project_map`, `read_chapter`, or `read_content` to read the target material before diagnosing it.
2. Read nearby context when it changes the diagnosis. For a chapter, inspect the previous and next chapter summaries or full text when available. For a project-level report, use `project_map`, `search_content`, `read_content`, and relevant `read_story_bible_entry` or `read_entry_section` calls to understand arcs, canon, and revision notes.
3. Separate summary from analysis. First state what happens in the target text, what role it appears to serve, and what promises it makes. Then diagnose whether the text fulfills those functions. Do not substitute plot summary for critique.
4. Ground every major claim in evidence from the draft. Point to specific scenes, beats, choices, omissions, repeated patterns, or chapter positions. Mark inferences as inferences when the text implies but does not state something.
5. Diagnose at the largest useful scale first. Address premise, conflict, role in the larger arc, character stakes, momentum, and payoff before line-level style. Mention prose or dialogue only when it materially affects story function.
6. Evaluate against the draft's apparent goals and genre contract, not against a universal template. Identify the likely genre, mode, and intended reader experience before labeling pacing, ambiguity, exposition, or resolution as a problem.
7. Preserve strengths. Name the working elements that revision should protect before listing problems, especially strong character pressure, vivid turns, clean setups, effective withheld information, or emotionally persuasive beats.
8. Prioritize revision targets by story impact. Distinguish structural blockers, chapter-function problems, continuity/canon issues, scene-level weaknesses, and optional polish. Limit the main action list to the highest-value changes unless the author asks for an exhaustive audit.

### Short Story Diagnostic Dimensions

For a complete standalone story, evaluate these dimensions and report only the ones that matter to the draft:

- Narrative foundation: the central premise or conflict, whether the scope fits the length, how clearly stakes connect to character, and whether the premise's implications are explored in the story itself.
- Character construction: who the story is about, what they want or fear, how characterization is demonstrated through action and choice, whether supporting characters affect the outcome, and whether the protagonist changes or is revealed under pressure.
- Story environment: whether setting, speculative elements, social rules, or worldbuilding are integrated through action; whether details do multiple jobs; and whether the environment constrains or changes character choices.
- Shattering moment: the irreversible turn that changes the meaning, stakes, relationship, self-understanding, or external situation. In speculative fiction, identify whether technology, magic, culture, or world rules cause, enable, prevent recovery from, or are redefined by this turn.
- Scene structure: whether each scene advances character and plot, whether the opening establishes character/situation/conflict, whether conflict stays focused, and whether the ending resolves the immediate pressure while leaving resonance.
- Technical execution: point of view consistency, narrative distance, scene/summary balance, exposition integration, dialogue purpose, tone, subtext, and whether style serves the story's intended effect.
- Emotional architecture: the emotional throughline, tension build and release, reveal pacing, earned payoff, and the gap between what is stated and what the reader is meant to feel.
- Thematic resolution: whether character arc, premise implications, immediate conflict, and lingering larger questions align.

Use the short-story checklist as diagnostic prompts, not mandatory fixes: clear premise/conflict, appropriate scope, character stakes, characterization through action, purposeful environment, irreversible turn, scenes that advance both character and plot, consistent point of view, emotional arc, and satisfying resonance.

### Chapter Diagnostic Dimensions

For a chapter inside a longer work, evaluate the chapter as part of the project rather than as a sealed short story:

- Position in the narrative arc: the chapter's job in the novel, plot progress, character-arc progress, thematic contribution, and relation to surrounding chapters.
- Narrative momentum: connection to the previous chapter's energy, opening hook, answered and raised questions, balance of resolution and forward pull, and whether the ending makes continuation feel necessary.
- Plot threading: which main plot, subplot, relationship, mystery, or world threads are advanced, introduced, resolved, paused, or set up for later.
- Character continuity: how characters have changed since earlier chapters, what new facets appear, how relationships shift, and whether choices remain consistent with established traits or intentionally complicate them.
- Point of view and voice: consistency with established POV patterns, viewpoint transitions, depth of character perspective, internal/external balance, tone, and voice fit with the larger work.
- Information flow: balance of new and established information, reader knowledge versus character knowledge, exposition timing, background integration, and setup for future revelations.
- Pacing elements: speed relative to surrounding chapters, internal rhythm, scene length, time handling, tension development, and release.
- Connection tracking: carryover from the previous chapter, setup for the next chapter, and contribution to broader plot, theme, character, and world development.

When useful, include a compact thread table:

```markdown
| Thread | Status in this chapter | Evidence | Revision implication |
|---|---|---|---|
| Main plot | Advanced / Setup / Resolved / Stalled | ... | ... |
| Character arc | Advanced / Complicated / Repeated / Missing | ... | ... |
| Subplot | Advanced / Setup / Resolved / Dropped | ... | ... |
```

Use the chapter checklist as diagnostic prompts, not mandatory fixes: clear role in the larger work, opening continuity, ending propulsion, advanced plot threads, character development, consistency with established material, future setup, and pacing appropriate to the chapter's position.

### Project-Level Diagnosis

For multi-chapter or project-level requests:

1. Use `project_map` to identify the relevant chapter sequence and durable notes.
2. Use `search_content`, `read_content`, `read_chapter`, `read_story_bible_entry`, and `read_entry_section` for repeated symptoms, major arcs, canon constraints, and previous revision plans.
3. Report patterns across chapters instead of repeating per-chapter notes unless the author requested a chapter-by-chapter audit.
4. Separate local problems from systemic problems. A weak ending in one chapter is local; repeated easy stopping points are a momentum pattern.
5. Recommend a revision pass order, such as structural repair before POV cleanup, continuity repair before line polish, or chapter endings before scene-level tightening.

## Diagnostic Discipline

- Do not deliver a vague "story analysis." Name the exact dimensions examined and why they matter for this draft.
- Do not treat every checklist miss as a defect. Ask whether it harms this particular story, genre contract, or chapter function.
- Do not compare the draft to famous published work as the main standard. Compare it to its own apparent goals and promises.
- Do not bury structural problems under prose nitpicks. Macro diagnosis comes first.
- Do not overwhelm the author with every possible issue. Give a ranked path through revision.
- Do not invent missing canon. If a diagnosis depends on unclear lore, timeline, names, institutions, rules, or history, mark it as an open question or canon proposal.

## Edda Output Handling

- Return concise analysis in chat when the author is deciding what to revise next.
- Create an Attached Note when the report belongs to one chapter, scene, selection, or complete short story attached to a specific text.
- Create or update a Project Note when the report should guide a later revision pass across multiple chapters, the whole manuscript, or recurring project-level weaknesses.
- Use this default report structure unless the author requests another format:

```markdown
## Summary of the Draft
- What happens:
- Apparent narrative purpose:
- Core promises/stakes:

## What Is Working
- Strength:
- Evidence:
- Protect this by:

## Diagnosis
| Dimension | Finding | Evidence | Impact |
|---|---|---|---|
| ... | ... | ... | Structural / chapter-function / continuity / scene-level / polish |

## Highest-Value Revision Targets
1. ...
2. ...
3. ...

## Optional Lower-Priority Notes
- ...

## Open Questions or Canon Proposals
- ...
```

- For chapter reports, include previous-chapter carryover and next-chapter propulsion when that context is available.
- For project-level reports, include recurring patterns and recommended revision-pass order.
- Do not apply direct rewrites through Structured Writes in this skill. If the author asks for implementation, route the findings into the appropriate rewrite, revision, pacing, dialogue, or character skill.
- Do not update Story Bible entries directly from analysis. Present canon-affecting findings as proposals or open questions for author confirmation.

## Script Compatibility

This source skill has no required helper scripts. The Milestone 3.5 rewrite works entirely through Edda-native analysis and note output.
