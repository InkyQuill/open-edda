---
name: revision-planner
description: Plan ordered revision passes across chapters and story-level changes without losing track of ripple effects.
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
    - revision
    - planning
    - optional
  priority: 56
metadata:
  status: optional
  source:
    - docs > skills > suggested > fiction > structure > novel-revision > SKILL.md
  scriptStatus: no-source-helpers
---

# $revision-planner

Revision planning for authors who need to sequence big changes, track ripple effects, and avoid turning one fix into three new problems.

## Use When

- A draft exists and the next revision pass feels large or risky.
- One change will likely affect later chapters, scenes, or canon notes.
- The author wants a pass order before rewriting.

## Do Not Use When

- The author wants immediate prose changes more than planning.
- The work is still in early discovery drafting.
- The request is mainly sentence polish.

## Writer Workflow

1. Read the relevant Chapter, Story Text, and any linked notes or canon context.
2. Identify the change level: concept, structure, scene, or line.
3. Map forward consequences across later chapters, Story Bible dependencies, and revision passes.
4. Turn that map into a practical pass order with warning signs to watch for.
5. Keep canon implications separate until the author confirms them.

## Writer Output Handling

- Return the pass plan in chat for quick discussion.
- Create an Attached Note when the plan belongs to one Chapter or one local change.
- Create or update a Project Note when the plan spans multiple Chapters or the whole story project.
- Propose Story Bible updates only as reviewable follow-ups when revision reveals confirmed canon drift.
- Do not use Structured Writes in this skill.

## Script Compatibility

This source skill has no required helper scripts. The Milestone 3.5 rewrite works entirely through Writer-native revision planning and note output.
