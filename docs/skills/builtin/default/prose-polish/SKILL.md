---
name: prose-polish
description: Sentence-level polish for chapters whose structure is already stable and now need stronger rhythm, clarity, precision, or voice consistency.
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
  tags:
    - fiction
    - prose
    - line-edit
    - voice
  priority: 70
metadata:
  status: default
  source:
    - docs > skills > suggested > fiction > craft > prose-style > SKILL.md
  scriptStatus: source-helpers-converted-to-guidance
---

# $prose-polish

Line-level prose help for chapters that already work structurally and now need cleaner, sharper, or more distinctive expression.

## Use When

- Structure, scene order, and character intent are already stable.
- The author wants help with rhythm, clarity, diction, voice, or sentence-level drag.
- A chapter needs a focused polish pass rather than developmental revision.

## Do Not Use When

- The story still has major structural, pacing, or arc problems.
- The author wants broad story diagnosis instead of line work.
- The author wants lore building or continuity planning.

## Edda Workflow

1. Confirm that the passage is ready for line work, not developmental repair.
2. Read the target passage with the Writing Brief and local chapter context.
3. Diagnose the main prose issue: flatness, monotony, confusion, excess, passivity, or voice drift.
4. Offer concise guidance first, then rewrite only if the author asks for applied line edits.
5. Keep any edits tightly scoped to the selected passage.

## Edda Output Handling

- Return short diagnosis and examples in chat by default.
- Create an Attached Note when the polish report belongs to one chapter or selection.
- Create or update a Project Note when the author wants a reusable prose checklist or voice guardrail list.
- Use Structured Writes only when the author explicitly asks to apply a passage rewrite.
- Do not propose Story Bible changes in this skill unless a wording problem reveals a canon contradiction.

## Script Compatibility

This rewrite adapts source prose-check logic into Edda-native line-edit guidance and revision reports. Source helpers are not runnable in Milestone 3.5.
