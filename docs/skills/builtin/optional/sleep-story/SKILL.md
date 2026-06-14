---
name: sleep-story
description: Draft calming fiction built for gentle read-aloud or bedtime listening rather than tension and payoff.
route:
  actionKinds:
    - chat
    - continuation
    - rewrite
  contentKinds:
    - chapter
    - story_text
    - attached_note
    - project_note
  tags:
    - fiction
    - bedtime
    - calming
    - optional
  priority: 54
metadata:
  status: optional
  source:
    - docs > skills > suggested > fiction > application > sleep-story > SKILL.md
  scriptStatus: no-source-helpers
---

# $sleep-story

Bedtime-story support for authors creating calm, low-stakes fiction meant to soothe rather than excite.

## Use When

- The author wants a read-aloud or listen-to-sleep story.
- Pacing should be gentle, descriptive, and safe.
- The prose needs calming rhythm, soft imagery, and no cliffhanger pressure.

## Do Not Use When

- The story should build urgency, suspense, or mystery.
- The request is mainly about meditation instruction rather than fiction.
- The author wants a conventional dramatic ending.

## Writer Workflow

1. Identify the intended delivery mode, perspective, and calming setting.
2. Build around gentle movement, sensory repetition, and low cognitive load.
3. Remove hidden urgency, sharp surprises, and resolution compulsion.
4. Keep the emotional palette safe, warm, and easy to drift away from.
5. Draft or rewrite only within the explicit scope the author requests.

## Writer Output Handling

- Return the sleep-story draft or guidance in chat by default.
- Create an Attached Note when the draft belongs to one story attempt or one scene.
- Create or update a Project Note when the author wants a repeatable bedtime format or series brief.
- Do not propose Story Bible changes unless the author wants the bedtime setting integrated into project canon.
- Use Structured Writes only when the author explicitly asks to draft or rewrite selected Story Text into sleep-story form.

## Script Compatibility

This source skill has no required helper scripts. The Milestone 3.5 rewrite works through Writer-native drafting and revision guidance.
