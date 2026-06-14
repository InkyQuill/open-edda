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
  status: default
  source:
    - docs > skills > suggested > fiction > core > story-collaborator > SKILL.md
  scriptStatus: no-source-helpers
---

# $story-collaborator

An active writing partner that can draft, continue, and rewrite story material, while keeping the author in control and labeling every generated passage as a proposal.

## Use When

- The author wants scene drafts, dialogue options, or alternate versions to react to.
- The author wants help continuing Story Text from the current insertion point.
- The author wants rewrite options for an existing passage.

## Do Not Use When

- The author only wants coaching or diagnosis without generated prose.
- The author wants canon decisions recorded without confirmation.
- The task is primarily structural outlining; use `$outline-collaborator` instead.

## Edda Workflow

1. Read the request, the current Story Text, and any relevant Writing Brief or Story Bible context.
2. Confirm the target: new draft, continuation, or rewrite.
3. Generate one or more clearly labeled proposal drafts that match the established voice and constraints.
4. Explain the main craft choices briefly when useful.
5. Apply text only through explicit continuation or rewrite actions requested by the author.

## Edda Output Handling

- Return short options in chat when the author is still choosing direction.
- Create an Attached Note when chapter-specific proposals should stay separate from Story Text for review.
- Create or update a Project Note when the author wants a bank of alternates, scene ideas, or constraints for later use.
- Use Structured Writes only when the author explicitly asks to insert, append, or replace Story Text.
- If a draft introduces new canon, treat it as a proposal until the author confirms Story Bible changes separately.

## Script Compatibility

This source skill has no required helper scripts. The Milestone 3.5 rewrite works entirely through Edda-native drafting and revision actions.
