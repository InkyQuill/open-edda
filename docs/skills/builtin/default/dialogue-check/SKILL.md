---
name: dialogue-check
description: Dialogue diagnosis for flat conversations, identical voices, weak subtext, and chapter exchanges that are functional but dramatically inert.
route:
  actionKinds:
    - chat
    - read_check
  contentKinds:
    - chapter
    - story_text
    - attached_note
    - project_note
    - story_bible_entry
  tags:
    - fiction
    - dialogue
    - voice
    - subtext
  priority: 82
metadata:
  status: default
  source:
    - docs > skills > suggested > fiction > character > dialogue > SKILL.md
  scriptStatus: source-helpers-converted-to-guidance
---

# $dialogue-check

Focused dialogue review for authors who need to know why a conversation feels flat, indistinct, or too literal.

## Use When

- Characters sound too similar.
- Conversations explain information but do not create tension, subtext, or relationship movement.
- The author wants chapter-level dialogue diagnosis before rewriting lines.

## Do Not Use When

- The author wants the skill to draft replacement dialogue by default.
- The main problem is scene structure or chapter pacing rather than dialogue.
- The author is still in early brainstorming and has no text to inspect.

## Writer Workflow

1. Read the target exchange in chapter context.
2. Check voice distinction, subtext, dramatic function, and pacing.
3. Name the core problem clearly: same-voice dialogue, exposition, no subtext, single-function exchange, or pacing mismatch.
4. Give revision directions and targeted questions instead of supplying replacement lines by default.
5. If the author later asks for applied rewriting, hand off to a rewrite workflow or `$story-collaborator`.

## Writer Output Handling

- Return the diagnosis in chat for quick review.
- Create an Attached Note when the report belongs to a chapter or selected exchange.
- Create or update a Project Note when repeated dialogue patterns should guide later revision.
- Do not propose Story Bible changes unless the dialogue problem comes from canon inconsistency.
- Do not use Structured Writes in this skill unless the author explicitly switches to an applied rewrite workflow.

## Script Compatibility

This rewrite adapts source dialogue audits into Writer-native checklists and reports. Source helpers are not runnable in Milestone 3.5.
