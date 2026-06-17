---
name: character-arc
description: Character transformation analysis for arcs that feel static, abrupt, hollow, or disconnected from the plot.
route:
  actionKinds:
    - chat
    - read_check
    - rewrite
  contentKinds:
    - chapter
    - story_text
    - story_bible_entry
    - entry_section
    - project_note
    - attached_note
  tags:
    - fiction
    - character
    - arc
    - revision
  priority: 84
metadata:
  useCases:
    - A protagonist or major character feels static or underdeveloped.
    - A transformation feels abrupt, unearned, or emotionally hollow.
    - The author wants to connect internal change to plot pressure.
    - The author needs a positive, negative, or flat arc diagnosis.
    - The author wants to map a character's want, need, lie, wound, flaw, truth, and transformation beats.
  doNotUse:
    - The request is mostly about dialogue line quality or prose style.
    - The author wants the skill to invent a character biography without context.
    - The task is only worldbuilding and not a character journey problem.
    - The author only wants a plot outline with no internal character-change question.
  status: default
  source:
    - docs > skills > suggested > fiction > character > character-arc > SKILL.md
  scriptStatus: no-source-helpers
---

# $character-arc

Character-journey diagnosis and revision support for clarifying what a character believes, wants, needs, resists, and becomes under story pressure.

## Edda Workflow

1. Establish context with Edda tools before diagnosing. Use `project_map` to locate relevant chapters, Story Bible entries, attached notes, and project notes. Use `read_chapter`, `read_content`, `read_story_bible_entry`, `read_entry_section`, and `list_revisions` as needed. Use `search_content` for the character name, aliases, key relationships, stated beliefs, decisions, failures, and climax scenes.
2. Separate evidence from inference. Mark what the text or Story Bible already establishes, what is inferred from scenes, and what remains an open question for the author.
3. Identify the likely arc type:
   - Positive change arc: the character begins from a false belief, pursues a want shaped by that lie, resists the truth, and eventually chooses a truer need.
   - Negative change arc: the character has a possible path toward growth but follows a flaw, temptation, or wound into worse choices, rejects redemption, and pays or causes consequences.
   - Flat arc: the character already carries the central truth; story pressure tests that truth and the character changes other people, institutions, or the world by holding steady.
   - No full arc: the character may be intentionally static, a function character, or an arc candidate whose transformation is not yet supported by scene evidence.
4. Map the internal engine. Name the character's visible want, deeper need, governing lie or truth, wound or ghost that plausibly created the lie, active flaw or coping strategy, and the external plot pressure that makes the old self costly. Do not invent missing biography as canon; offer it as optional proposal only when needed.
5. Run pressure tests against the scenes:
   - Does the story force the character to confront the lie, flaw, or truth through choices rather than explanation?
   - Are want and need in tension, or does the character want exactly what would fulfill them?
   - Does the character resist change, double down, rationalize, regress, or pay for old behavior before transforming?
   - Does the midpoint or equivalent turn expose a glimpse of truth, mirror self-knowledge, temptation, or deeper descent?
   - Does the climax require the character to act from the changed self, rejected truth, or steadfast truth?
   - Does the resolution demonstrate the new self, ruin, or changed world through concrete action?
6. Diagnose common failures with scene evidence:
   - No transformation: ending self is not meaningfully different from beginning self.
   - Unearned transformation: the character changes, but events did not demand or pressure the change.
   - Abrupt change: the character flips without enough struggle, resistance, or prior seed scenes.
   - Unclear lie or truth: the internal conflict cannot be stated in one concrete belief.
   - Want/need collapse: the external goal and internal fulfillment are identical, so there is no inner tension.
   - Missing struggle: the character accepts the truth too easily.
   - Informed arc: narration or other characters claim change that scenes do not prove.
   - Mentor shortcut: a mentor explains the truth and the character changes without earning discovery through action.
   - Trauma equals transformation: painful events occur, but the character's choices do not reveal a changed pattern.
   - Perfect protagonist: the character has no meaningful blind spot, flaw, wound response, or testable truth.
7. Build a transformation beat map tied to evidence. Use the story's actual structure rather than forcing a template, but check for setup, catalyst, first commitment, rising pressure, midpoint mirror or temptation, crisis or dark night, climactic choice, and resolution proof. For each beat, cite the chapter, scene, Story Bible entry, or note that supports it, and flag missing or weak beats.
8. Keep revision advice canon-safe. Recommend targeted additions, removals, or reframings as proposals: a pressure scene, doubled-down choice, cost for the lie, sharper want/need split, clearer wound implication, stronger temptation, or proof-of-change action. Do not update Story Bible entries or durable character facts unless the author explicitly confirms them.
9. For rewrite requests, provide a localized revision plan or replacement text only after reading the current target content and confirming the intended arc function of the passage. Preserve established canon, voice, and continuity; label any new internal facts as proposals.

## Edda Output Handling

- Return arc diagnosis and key questions in chat by default.
- Include the restored arc criteria in the output when useful: arc type, want, need, lie or truth, wound or flaw, plot pressure, resistance pattern, transformation beats, scene evidence, and the final demonstrated state.
- Create an Attached Note when the arc review belongs to one chapter, one selected passage, or one local turning point.
- Create or update a Project Note when the author needs a durable cross-chapter arc map, diagnosis, or revision checklist.
- Propose Story Bible updates only for author-confirmed durable character facts. Keep unconfirmed wounds, lies, flaws, needs, backstory causes, and future transformation beats in a proposal or note until confirmed.
- Use Structured Writes only when the author explicitly asks to apply a specific rewrite to known target text.

## Script Compatibility

This source skill has no required helper scripts. The Milestone 3.5 rewrite works entirely through Edda-native diagnosis and planning.
