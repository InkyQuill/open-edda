---
name: worldbuilding-brainstorm
description: Canon-safe worldbuilding brainstorming that develops lore one question at a time and records Story Bible changes only after explicit author confirmation.
route:
  actionKinds:
    - chat
    - read_check
    - story_bible
  contentKinds:
    - story_bible_entry
    - entry_section
    - project_note
    - attached_note
    - chapter
    - story_text
  tags:
    - fiction
    - worldbuilding
    - brainstorming
    - canon-safe
  priority: 94
metadata:
  status: default
  source:
    - docs > skills > important > worldbuilding-brainstorm > SKILL.md
  scriptStatus: no-source-helpers
---

# $worldbuilding-brainstorm

The default lore-building skill for Edda: it pressure-tests setting ideas, asks one question at a time, and keeps canon under explicit author control.

## Use When

- The author wants to brainstorm or reconcile characters, places, systems, factions, history, or other setting material.
- The author needs lore ideas challenged against existing Story Bible entries or on-page evidence.
- The author wants to turn vague worldbuilding into confirmed, durable canon.

## Do Not Use When

- The author wants direct Story Text drafting or scene revision.
- The task is only diagnostic and should not expand lore.
- The author wants canon changed without review.

## Edda Workflow

1. Read the relevant Story Bible entries, notes, and any on-page chapter evidence.
2. Ask one targeted question at a time, with a clear recommendation when helpful.
3. Surface contradictions, pressure points, and downstream consequences before canon changes are made.
4. Distinguish existing canon, recommendations, and open questions.
5. Record canon only after the author confirms it.

## Edda Output Handling

- Keep active brainstorming in chat by default.
- Create an Attached Note when a lore discussion is tied to one chapter or one selection.
- Create or update a Project Note for unresolved branches, option sets, or follow-up questions.
- Propose Story Bible entry or section changes only after explicit confirmation from the author.
- Do not use Structured Writes for Story Text in this skill.

## Script Compatibility

This source skill has no required helper scripts. The Milestone 3.5 rewrite works entirely through Edda-native questioning, Story Bible proposals, and reviewable notes.
