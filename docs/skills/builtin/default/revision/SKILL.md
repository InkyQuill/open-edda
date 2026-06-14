---
name: revision
description: Revision planning and applied editing help for completed drafts that need ordered passes, stronger priorities, and controlled changes.
route:
  actionKinds:
    - chat
    - read_check
    - rewrite
  contentKinds:
    - chapter
    - story_text
    - project_note
    - attached_note
    - story_bible_entry
  tags:
    - fiction
    - revision
    - editing
    - planning
  priority: 86
metadata:
  status: default
  source:
    - docs > skills > suggested > fiction > craft > revision > SKILL.md
  scriptStatus: source-helpers-converted-to-guidance
---

# $revision

Revision help for authors who have draft material and need an ordered way to improve it without getting lost in random edits.

## Use When

- A draft exists and the author needs to decide what to fix first.
- Revision feels overwhelming, endless, or scattered.
- The author wants a pass plan or targeted rewrite support after diagnosis.

## Do Not Use When

- The chapter is still being drafted and needs momentum more than editing.
- The main task is worldbuilding or outlining.
- The author wants sentence-level polish before structure is stable.

## Writer Workflow

1. Read the draft scope and identify whether the next pass is developmental, scene-level, dialogue-level, or line-level.
2. Name the current revision problem clearly.
3. Build an ordered pass plan before making broad edits.
4. Apply rewrites only to the scope the author has chosen.
5. Keep large structural recommendations separate from small wording fixes.

## Writer Output Handling

- Return the revision diagnosis and pass order in chat by default.
- Create an Attached Note when the report belongs to one chapter or one revision pass.
- Create or update a Project Note when the author needs a cross-project revision plan or feedback synthesis.
- Use Structured Writes only when the author explicitly asks to apply selected revisions.
- Propose Story Bible changes only when revision uncovers confirmed canon drift that the author wants recorded.

## Script Compatibility

This rewrite adapts source revision-audit logic into Writer-native pass planning and review checklists. Source helpers are not runnable in Milestone 3.5.
