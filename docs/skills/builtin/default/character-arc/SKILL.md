---
name: character-arc
description: Character transformation analysis for arcs that feel static, abrupt, hollow, or disconnected from the plot.
route:
  actionKinds:
    - chat
    - read_check
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
  status: default
  source:
    - docs > skills > suggested > fiction > character > character-arc > SKILL.md
  scriptStatus: no-source-helpers
---

# $character-arc

Character-journey analysis for authors who need to clarify a lie, want, need, transformation path, or arc completion.

## Use When

- A protagonist or major character feels static or underdeveloped.
- A transformation feels abrupt, unearned, or emotionally hollow.
- The author wants to connect internal change to plot pressure.

## Do Not Use When

- The request is mostly about dialogue line quality or prose style.
- The author wants the skill to invent a character biography without context.
- The task is only worldbuilding and not a character journey problem.

## Edda Workflow

1. Read the relevant Story Text, Story Bible material, and notes for the character.
2. Identify the likely arc type and the core internal tension.
3. Trace how the plot pressures the character's lie, want, and need.
4. Point out missing beats, weak resistance, or hollow payoffs.
5. Suggest revision targets that the author can use in later drafting or revision passes.

## Edda Output Handling

- Return arc diagnosis and key questions in chat by default.
- Create an Attached Note when the arc review belongs to one chapter or one local turning point.
- Create or update a Project Note when the author needs a project-wide arc map.
- Propose Story Bible updates only for durable character facts the author confirms.
- Do not use Structured Writes in this skill.

## Script Compatibility

This source skill has no required helper scripts. The Milestone 3.5 rewrite works entirely through Edda-native diagnosis and planning.
