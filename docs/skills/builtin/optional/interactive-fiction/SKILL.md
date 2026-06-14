---
name: interactive-fiction
description: Design branching fiction that offers meaningful choices without exploding into unmanageable structure.
route:
  actionKinds:
    - chat
    - read_check
    - continuation
    - rewrite
  contentKinds:
    - chapter
    - story_text
    - attached_note
    - project_note
  tags:
    - fiction
    - interactive
    - branching
    - optional
  priority: 52
metadata:
  status: optional
  source:
    - docs > skills > suggested > fiction > application > interactive-fiction > SKILL.md
  scriptStatus: no-source-helpers
---

# $interactive-fiction

Branching-fiction support for authors who need meaningful choices, manageable state, and narrative coherence across interactive story paths.

## Use When

- The story includes player or reader choices.
- Branches feel fake, too numerous, or structurally messy.
- The author wants a cleaner balance between agency and authored payoff.

## Do Not Use When

- The project is a standard linear story.
- The request is only about prose polish inside one branch.
- The author wants a full implementation plan for external game tooling rather than story design.

## Writer Workflow

1. Read the target branch, choice map, or outline note.
2. Identify whether the problem is weak choices, branch sprawl, state chaos, or unsatisfying endings.
3. Recommend a structure that fits the story: foldback, bottleneck, branch-and-bottleneck, or state-based variation.
4. Make each choice express values, tradeoffs, or character, not just menu navigation.
5. Draft or revise branch text only when the author explicitly asks for it.

## Writer Output Handling

- Return the branching diagnosis or proposed structure in chat by default.
- Create an Attached Note when the work belongs to one branch or one scene cluster.
- Create or update a Project Note when the author wants a reusable interactive design brief.
- Do not propose Story Bible changes unless the branching design establishes durable canon rules the author wants tracked.
- Use Structured Writes only when the author explicitly asks to draft or rewrite selected branching Story Text.

## Script Compatibility

This source skill has no required helper scripts. The Milestone 3.5 rewrite works through Writer-native analysis, structure guidance, and optional branch drafting.
