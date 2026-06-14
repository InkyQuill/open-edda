---
name: cultural-depth
description: Add layered cultural texture through customs, inherited assumptions, mixed influences, and things characters take for granted.
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
    - culture
    - character
    - optional
  priority: 40
metadata:
  status: optional
  source:
    - docs > skills > suggested > fiction > character > memetic-depth > SKILL.md
  scriptStatus: no-source-helpers
---

# $cultural-depth

Cultural texture work for authors who want characters and societies to feel shaped by inherited habits, mixed influences, and unspoken assumptions.

## Use When

- A culture feels thin, too uniform, or too newly invented.
- Characters need markers of upbringing beyond exposition.
- The author wants objects, customs, or references that imply history.

## Do Not Use When

- The request is mainly about political systems or conlang structure.
- The author wants only a fast naming pass.
- The work would be better served by direct lore creation instead of implied depth.

## Writer Workflow

1. Read the relevant Story Bible context and how the culture currently appears in Story Text.
2. Identify the missing layer: custom, artifact, taboo, mixed influence, or inherited assumption.
3. Add a small set of recognizable, inferable, and mysterious cultural signals.
4. Tie those signals to characters, place, and conflict rather than trivia.
5. Keep durable culture claims reviewable until the author confirms them.

## Writer Output Handling

- Return the cultural layer suggestions in chat by default.
- Create an Attached Note when the work supports one Chapter, one community, or one scene cluster.
- Create or update a Project Note for broader cultural exploration.
- Propose Story Bible updates only after the author confirms customs, artifacts, or group assumptions as canon.
- Do not use Structured Writes in this skill.

## Script Compatibility

This source skill has no required helper scripts. The Milestone 3.5 rewrite works through Writer-native cultural analysis and canon-safe proposals.
