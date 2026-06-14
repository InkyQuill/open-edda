---
name: multi-pov
description: Design multi-POV stories where each perspective earns its place and the shared setting or event generates real pressure.
route:
  actionKinds:
    - chat
    - read_check
    - continuation
  contentKinds:
    - chapter
    - story_text
    - attached_note
    - project_note
  tags:
    - fiction
    - structure
    - pov
    - optional
  priority: 46
metadata:
  status: optional
  source:
    - docs > skills > suggested > fiction > structure > perspectival-constellation > SKILL.md
  scriptStatus: no-source-helpers
---

# $multi-pov

Multi-POV planning for authors who want several perspectives to feel necessary, distinct, and structurally connected instead of like camera swaps.

## Use When

- A story needs multiple viewpoint characters with different access, stakes, or knowledge.
- The shared event, place, or institution should generate intersecting stories.
- The author wants help deciding who gets POV space and why.

## Do Not Use When

- One protagonist already carries the story cleanly.
- Additional POVs would only repeat information.
- The request is mainly for line edits inside a single scene.

## Edda Workflow

1. Read the relevant Chapters, notes, or outline material.
2. Identify the catalyst environment or shared pressure linking the perspectives.
3. Define each POV's access path, knowledge gap, and emotional function.
4. Map where the viewpoints intersect, contradict, or reframe each other.
5. Draft or revise POV text only when the author explicitly asks for it.

## Edda Output Handling

- Return the POV structure in chat by default.
- Create an Attached Note when the plan belongs to one Chapter cluster.
- Create or update a Project Note when the author wants a project-level POV map.
- Do not propose Story Bible changes unless POV structure reveals durable canon the author wants tracked.
- Use Structured Writes only when the author explicitly asks to draft or rewrite selected Story Text for POV execution.

## Script Compatibility

This source skill has no required helper scripts. The Milestone 3.5 rewrite works through Edda-native planning, analysis, and optional drafting support.
