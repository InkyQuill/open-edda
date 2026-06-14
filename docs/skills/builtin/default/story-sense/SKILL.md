---
name: story-sense
description: Broad fiction diagnosis for authors who need to identify what a story, chapter, or draft is missing and which next skill or workflow fits best.
route:
  actionKinds:
    - chat
    - read_check
  contentKinds:
    - chapter
    - story_text
    - project_note
    - attached_note
    - story_bible_entry
    - entry_section
  tags:
    - fiction
    - diagnosis
    - routing
    - revision
  priority: 96
metadata:
  status: default
  source:
    - docs > skills > suggested > fiction > core > story-sense > SKILL.md
  scriptStatus: source-helpers-converted-to-guidance-and-data
---

# $story-sense

The broad diagnostic skill for figuring out what is wrong, what stage the work is in, and which Writer-native move should happen next.

## Use When

- The author says the story is not working but cannot name the problem.
- A chapter, draft, or concept feels flat, broken, generic, or stalled.
- The session needs routing to drafting, revision, dialogue, structure, worldbuilding, or prose work.

## Do Not Use When

- The author already knows the exact task and wants direct drafting or rewriting.
- The task is a focused line edit, dialogue audit, or ending check with a clear scope.
- The author wants canon brainstorming more than diagnosis.

## Writer Workflow

1. Read the request and the most relevant `@` chapters, notes, or Story Bible material.
2. Identify the current story state and name the likely failure mode.
3. Explain the diagnosis in plain language.
4. Recommend the next Writer-native skill or action, with one or two concrete next steps.
5. Stay diagnostic unless the author explicitly switches to a writing or revision workflow.

## Writer Output Handling

- Return the diagnosis and routing advice in chat by default.
- Create an Attached Note when the diagnosis belongs to a specific chapter or selection.
- Create or update a Project Note when the diagnosis becomes a broader story plan or revision map.
- Propose Story Bible review only when continuity or canon gaps are part of the diagnosis.
- Do not use Structured Writes in this skill.

## Script Compatibility

This rewrite adapts source helper logic into Writer-native diagnosis, genre tables, and function references. Source helpers are not runnable in Milestone 3.5.
