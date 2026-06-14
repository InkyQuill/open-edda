---
name: societal-evolution
description: Track how a society changes across generations so later-world institutions, values, and conflicts feel earned.
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
    - evolution
    - optional
  priority: 42
metadata:
  status: optional
  source:
    - docs > skills > suggested > fiction > application > multi-order-evolution > SKILL.md
  scriptStatus: no-source-helpers
---

# $societal-evolution

Generational worldbuilding for authors who want a civilization's present-day values, institutions, and conflicts to grow from long-term pressure instead of arriving fully formed.

## Use When

- The setting spans generations, migrations, colonies, or civilizational drift.
- The author wants to know how one environment reshapes people over time.
- Present-day world logic feels too close to baseline assumptions.

## Do Not Use When

- The story only needs present-tense cultural texture.
- The request is mainly about one religion, economy, or city without long time depth.
- The author wants quick lore instead of generational causation.

## Edda Workflow

1. Read the relevant Story Bible material and identify the foundational pressures.
2. Map first-order, second-order, and later-order changes across bodies, institutions, values, and identity.
3. Track which inherited terms or structures persist while their meaning changes.
4. Use those shifts to generate present-day tensions and cross-society misunderstandings.
5. Keep durable canon changes reviewable until the author confirms them.

## Edda Output Handling

- Return the generational model in chat by default.
- Create an Attached Note when the work serves one Chapter, one era, or one society in focus.
- Create or update a Project Note for broader exploration.
- Propose Story Bible updates only after the author confirms timeline facts, institutional changes, or evolved identities.
- Do not use Structured Writes in this skill.

## Script Compatibility

This source skill has no required helper scripts. The Milestone 3.5 rewrite works through Edda-native worldbuilding guidance and canon-safe proposals.
