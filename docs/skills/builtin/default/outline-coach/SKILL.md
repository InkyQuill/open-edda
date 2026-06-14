---
name: outline-coach
description: Coaching-only outlining help that guides structure through questions and diagnosis without generating outline beats for the author.
route:
  actionKinds:
    - chat
    - read_check
  contentKinds:
    - project_note
    - attached_note
    - story_bible_entry
    - entry_section
    - chapter
  tags:
    - fiction
    - outline
    - coaching
    - structure
  priority: 80
metadata:
  status: default
  source:
    - docs > skills > suggested > fiction > structure > outline-coach > SKILL.md
  scriptStatus: no-source-helpers
---

# $outline-coach

Coaching mode for authors who want help discovering structure without having the skill generate the outline itself.

## Use When

- The author wants questions, diagnosis, or frameworks for outlining.
- A plot, act, or sequence is unclear but the author wants to build it themselves.
- The author wants structural feedback on an existing outline note.

## Do Not Use When

- The author wants proposed beats, act plans, or active structural drafting.
- The main task is prose generation or chapter continuation.
- The request is mainly about canon brainstorming rather than story structure.

## Writer Workflow

1. Read the request and the relevant outline notes, chapters, or Story Bible context.
2. Diagnose the structural problem in plain language.
3. Ask targeted questions that help the author discover the next beat, turn, or act shape.
4. Offer frameworks and options without generating copy-ready outline content.
5. End by returning the author to a concrete outlining move.

## Writer Output Handling

- Return coaching and diagnosis in chat by default.
- Create an Attached Note when the guidance belongs to one chapter or one act problem.
- Create or update a Project Note when the author needs a broader outline checklist or structure map.
- Propose Story Bible changes only when outline questions expose canon issues that need review.
- Do not use Structured Writes for Story Text in this skill.

## Script Compatibility

This source skill has no required helper scripts. The Milestone 3.5 rewrite works entirely through Writer-native coaching and notes.
