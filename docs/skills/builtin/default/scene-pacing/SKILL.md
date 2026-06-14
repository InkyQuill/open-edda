---
name: scene-pacing
description: Scene and chapter pacing diagnosis for weak escalation, missing aftermath, clean victories, and sequences that do not accumulate pressure.
route:
  actionKinds:
    - chat
    - read_check
  contentKinds:
    - chapter
    - story_text
    - attached_note
    - project_note
  tags:
    - fiction
    - pacing
    - structure
    - scenes
  priority: 82
metadata:
  status: default
  source:
    - docs > skills > suggested > fiction > structure > scene-sequencing > SKILL.md
  scriptStatus: source-helpers-converted-to-guidance
---

# $scene-pacing

Pacing and sequencing review for scenes that feel slow, exhausting, flat, or mechanically connected instead of dramatically linked.

## Use When

- A scene has no clear goal, weak escalation, or too-clean resolution.
- A chapter feels both slow and exhausting.
- The author wants to inspect scene-to-scene rhythm before rewriting prose.

## Do Not Use When

- The author wants full outline generation rather than scene diagnosis.
- The issue is mainly line style, dialogue polish, or canon building.
- The text is still too early and vague for scene-level diagnosis.

## Edda Workflow

1. Read the target scene or chapter in sequence.
2. Identify the scene goal, conflict escalation, outcome, and aftermath.
3. Diagnose whether the pressure is missing, misplaced, or overextended.
4. Explain what should tighten, expand, or move.
5. Suggest next revision targets without rewriting Story Text by default.

## Edda Output Handling

- Return the pacing diagnosis in chat for immediate use.
- Create an Attached Note when the report belongs to a specific chapter or scene selection.
- Create or update a Project Note when the pacing issue spans several chapters or a whole act.
- Do not propose Story Bible changes unless pacing failure comes from a canon constraint.
- Do not use Structured Writes in this skill unless the author explicitly switches to an applied rewrite workflow.

## Script Compatibility

This rewrite adapts source scene-analysis logic into Edda-native pacing rubrics and reports. Source helpers are not runnable in Milestone 3.5.
