---
name: skill-writer
description: Skill-authoring help for drafting or revising Edda skills, with strong preference for instructions, templates, references, and data over executable helpers.
route:
  actionKinds:
    - chat
    - skill_authoring
  contentKinds:
    - project_note
    - attached_note
    - agent_session
  tags:
    - authoring-tool
    - skills
    - documentation
    - process
  priority: 20
metadata:
  status: default
  source:
    - docs > skills > important > writing-skills > SKILL.md
  scriptStatus: authoring-helpers-deferred-admin-controlled
---

# $skill-writer

An authoring tool for writing or revising Edda skills, optimized for clear instructions, routing metadata, templates, references, and safe review workflows.

## Use When

- The author wants to create, rewrite, or refine a Edda skill.
- A skill needs clearer routing, better guidance, or safer output handling.
- The task is skill documentation and behavior design rather than fiction drafting.

## Do Not Use When

- The author is working on Story Text, Story Bible canon, or chapter revision instead of a skill.
- The task depends on executable helpers as the primary solution.
- The author wants to skip review and ship a skill based only on intuition.

## Edda Workflow

1. Read the current skill draft, intent, and any related notes from the active agent session.
2. Clarify the skill's purpose, routing, guardrails, and output behavior.
3. Prefer instructions, templates, built-in data, and references over executable helpers.
4. Draft or revise the skill in a form that is reviewable inside Edda.
5. Keep any claimed behavior aligned with Milestone 3.5 constraints and explicit author approval.

## Edda Output Handling

- Return short skill guidance and review notes in chat by default.
- Create an Attached Note when the author wants a session-scoped review memo or checklist.
- Create or update a Project Note when the author wants a working skill draft, routing matrix, or rewrite plan preserved.
- Do not propose Story Bible changes in this skill.
- Use Structured Writes only if Edda later exposes an explicit skill-document editing workflow and the author asks for it.

## Script Compatibility

This rewrite intentionally avoids depending on source authoring helpers. In Milestone 3.5, executable helper behavior is deferred and should be treated as admin-controlled later work, not normal skill behavior.
