---
name: outline-collaborator
description: Active outlining help for proposed beats, act structures, scene sequences, and revision-ready structural options.
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
  status: default
  source:
    - docs > skills > suggested > fiction > structure > outline-collaborator > SKILL.md
  scriptStatus: no-source-helpers
---

# $outline-collaborator

An active structural partner that proposes beats, sequences, and outline alternatives while keeping them clearly framed as options for the author.

## Use When

- The author wants act plans, scene beats, or alternate structures proposed directly.
- An outline note exists and needs continuation or rewrite.
- The author wants multiple structural options before drafting.

## Do Not Use When

- The author wants coaching-only questions instead of proposed structure.
- The author wants finished prose drafting instead of planning.
- The author wants canon changes recorded without review.

## Edda Workflow

1. Read the current outline material, relevant Story Bible context, and any chapter evidence.
2. Confirm the target scope: act, sequence, chapter plan, or revision of an existing outline.
3. Generate clearly labeled structural proposals and alternatives.
4. Keep sample text brief and illustrative if used at all; the main output is structure.
5. Apply outline changes only through explicit continuation or rewrite actions requested by the author.

## Edda Output Handling

- Return short structural options in chat while the author is choosing direction.
- Create an Attached Note when the proposal belongs to one chapter or one sequence.
- Create or update a Project Note when the author wants a working outline, act map, or beat bank preserved.
- Use Structured Writes only when the author explicitly asks to update the outline material stored in Edda.
- Propose Story Bible updates separately if structural choices imply new canon.

## Script Compatibility

This source skill has no required helper scripts. The Milestone 3.5 rewrite works entirely through Edda-native structure proposals and explicit note updates.
