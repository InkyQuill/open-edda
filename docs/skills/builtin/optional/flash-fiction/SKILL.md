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
  status: optional
  source:
    - docs > skills > suggested > fiction > application > flash-fiction > SKILL.md
  scriptStatus: no-source-helpers
---

# $flash-fiction

Short-form fiction support for authors working in drabble, micro, flash, or sudden-fiction lengths where compression matters as much as the idea.

## Use When

- The author wants to draft or revise fiction under roughly 1,500 words.
- A short piece feels flat, over-explained, or emotionally incomplete.
- The goal is strong compression rather than expansion.

## Do Not Use When

- The story wants room to breathe as a chapter-length draft.
- The author mainly needs worldbuilding support rather than short-form execution.
- The request is for publishing copy instead of fiction.

## Writer Workflow

1. Read the full short piece or the intended prompt and target length.
2. Identify the dominant problem: hook, compression, subtext, ending, image pattern, or logic.
3. Tighten around one emotional movement and one durable image or turn.
4. Preserve what is already sharp instead of over-expanding the piece.
5. Draft or rewrite only within the scope the author explicitly chooses.

## Writer Output Handling

- Return diagnostic notes or draft text in chat by default.
- Create an Attached Note when the work belongs to one short piece or one selection.
- Create or update a Project Note when the author wants a reusable flash-fiction brief or batch plan.
- Do not propose Story Bible changes unless the author wants the short piece folded into project canon.
- Use Structured Writes only when the author explicitly asks to draft, replace, or compress selected Story Text.

## Script Compatibility

This source skill has no required helper scripts. The Milestone 3.5 rewrite works through Writer-native drafting, diagnosis, and revision guidance.
