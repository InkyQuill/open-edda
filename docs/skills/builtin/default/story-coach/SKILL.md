---
name: story-coach
description: Coaching-only help for story problems, stuck scenes, and revision decisions without generating story prose, dialogue, or canon for the author.
route:
  actionKinds:
    - chat
    - read_check
  contentKinds:
    - chapter
    - story_text
    - project_note
    - attached_note
    - story_bible_entry
    - entry_section
  tags:
    - fiction
    - coaching
    - diagnosis
    - process
  priority: 92
metadata:
  status: default
  source:
    - docs > skills > suggested > fiction > core > story-coach > SKILL.md
  scriptStatus: no-source-helpers
---

# $story-coach

Coaching mode for authors who want help thinking, diagnosing, and deciding what to write without having the skill write the story for them.

## Use When

- The author wants questions, diagnosis, or frameworks instead of drafted prose.
- A chapter, scene, outline beat, or revision problem feels stuck but the author wants to solve it themselves.
- The author wants feedback on what is not working and what to try next.

## Do Not Use When

- The author wants the skill to draft story text, dialogue, outline beats, or lore proposals.
- The author wants direct rewriting or continuation of Story Text.
- The task is mainly canon building; use `$worldbuilding-brainstorm` or `$worldbuilding-check` instead.

## Writer Workflow

1. Read the author request and any linked `@` chapters, notes, or Story Bible material.
2. Diagnose the current problem in plain language.
3. Ask targeted questions that help the author discover the answer.
4. Offer frameworks, options, and next-step prompts without supplying story prose, dialogue, or canon text.
5. End by returning the author to a concrete writing or revision move.

## Writer Output Handling

- Return coaching, diagnosis, and next-step prompts in chat by default.
- Create an Attached Note when the guidance belongs to one chapter or text selection.
- Create or update a Project Note when the guidance becomes a broader plan, pattern list, or revision checklist.
- Propose Story Bible changes only as questions or flagged continuity issues; do not create canon here.
- Do not use Structured Writes for Story Text in this skill.

## Script Compatibility

This source skill has no required helper scripts. The Milestone 3.5 rewrite works entirely through Writer-native guidance and reports.
