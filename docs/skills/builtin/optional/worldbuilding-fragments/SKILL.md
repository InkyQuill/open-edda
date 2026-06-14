---
name: worldbuilding-fragments
description: Create indirect lore fragments, in-world snippets, and epigraph-style material that imply depth without stopping the story cold.
route:
  actionKinds:
    - chat
    - read_check
    - continuation
  contentKinds:
    - chapter
    - story_text
    - story_bible_entry
    - entry_section
    - project_note
    - attached_note
  tags:
    - fiction
    - worldbuilding
    - fragments
    - optional
  priority: 40
metadata:
  status: optional
  source:
    - docs > skills > suggested > fiction > worldbuilding > oblique-worldbuilding > SKILL.md
  scriptStatus: no-source-helpers
---

# $worldbuilding-fragments

Indirect worldbuilding for authors who want fragments, documents, sayings, epigraphs, or local details that imply a wider culture without heavy exposition.

## Use When

- The setting needs texture through snippets rather than direct explanation.
- A Chapter wants an epigraph, quoted text, rumor, or documentary fragment.
- The author wants worldbuilding that rewards inference.

## Do Not Use When

- The request is for direct canon definition rather than indirect texture.
- The fragment would confuse the reader more than enrich the scene.
- The author wants major Story Text changes without explicitly asking for them.

## Writer Workflow

1. Read the target Chapter, Story Text, or Story Bible context.
2. Identify what system, conflict, or cultural blind spot the fragment should reveal.
3. Choose a plausible voice or source inside the world.
4. Draft a fragment that implies more than it explains.
5. Keep canon claims reviewable until the author confirms them.

## Writer Output Handling

- Return fragments in chat by default.
- Create an Attached Note when the fragment belongs to one Chapter or one local scene problem.
- Create or update a Project Note when the author wants a bank of reusable fragments.
- Propose Story Bible updates only if the fragment establishes durable canon the author wants recorded.
- Use Structured Writes only when the author explicitly asks to insert a fragment into Story Text.

## Script Compatibility

This source skill has no required helper scripts. The Milestone 3.5 rewrite works through Writer-native fragment drafting and note output.
