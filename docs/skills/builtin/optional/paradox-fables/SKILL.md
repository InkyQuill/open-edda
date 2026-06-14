---
name: paradox-fables
description: Draft short fable-like stories that hold a tension or paradox without flattening it into a simple lesson.
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
    - fable
    - theme
    - optional
  priority: 50
metadata:
  status: optional
  source:
    - docs > skills > suggested > fiction > application > paradox-fables > SKILL.md
  scriptStatus: no-source-helpers
---

# $paradox-fables

Paradox-fable support for authors who want compact, memorable story forms that embody tension instead of resolving it into a neat moral.

## Use When

- The author wants a teaching story, inset tale, or standalone short with layered meaning.
- A theme is best approached sideways through image and action.
- The ending should leave productive tension rather than a sermon.

## Do Not Use When

- The author wants a traditional moralistic fable with a single clean lesson.
- The request is mainly for prose polish rather than concept and form.
- The story needs heavy realism instead of stylized compression.

## Edda Workflow

1. Identify the paradox or tension that should be embodied.
2. Find a simple narrative form that naturally expresses that tension.
3. Use archetypal voices, clean imagery, and durable structure without preachiness.
4. Check that the ending preserves pressure instead of explaining it away.
5. Draft or rewrite only within the explicit scope the author chooses.

## Edda Output Handling

- Return the fable draft or concept in chat by default.
- Create an Attached Note when the work belongs to one chapter insert or one small story experiment.
- Create or update a Project Note when the author wants a bank of fable concepts or inset texts.
- Do not propose Story Bible changes unless the author wants the fable treated as in-world canon.
- Use Structured Writes only when the author explicitly asks to insert or rewrite selected Story Text in this form.

## Script Compatibility

This source skill has no required helper scripts. The Milestone 3.5 rewrite works through Edda-native drafting and thematic guidance.
