---
name: identity-denial
description: Build arcs where a character refuses to admit what they are becoming, and that denial drives the story forward.
route:
  actionKinds:
    - chat
    - read_check
    - rewrite
  contentKinds:
    - chapter
    - story_text
    - attached_note
    - project_note
    - story_bible_entry
  tags:
    - fiction
    - character
    - arc
    - optional
  priority: 46
metadata:
  status: optional
  source:
    - docs > skills > suggested > fiction > structure > identity-denial > SKILL.md
  scriptStatus: no-source-helpers
---

# $identity-denial

Character-arc support for authors exploring self-deception, moral slippage, and the widening gap between what a character says they are and what the story shows.

## Use When

- A protagonist is transforming but refuses to name that transformation.
- The author wants stronger mirror moments, rationalizations, or escalation markers.
- A fall, corruption, addiction, inherited-pattern, or denial arc needs structure.

## Do Not Use When

- The story needs a straightforward growth arc without self-deception.
- The request is mainly about plot logistics or lore.
- The author wants only general feedback instead of a denial-focused lens.

## Edda Workflow

1. Read the relevant Story Text, Chapter scope, and any character notes.
2. Identify the denied identity, what accepting it would cost, and how the character justifies avoidance.
3. Map the escalation, mirror characters, and point of no return.
4. Recommend scene-level or chapter-level changes that sharpen denial on the page.
5. Apply rewrites only if the author explicitly asks for them.

## Edda Output Handling

- Return the denial arc analysis in chat by default.
- Create an Attached Note when the work belongs to one Chapter, one beat, or one scene cluster.
- Create or update a Project Note when the author wants a reusable arc map.
- Propose Story Bible updates only if the author wants durable character facts or history recorded.
- Use Structured Writes only when the author explicitly asks to rewrite selected Story Text.

## Script Compatibility

This source skill has no required helper scripts. The Milestone 3.5 rewrite works through Edda-native arc analysis and revision guidance.
