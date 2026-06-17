---
name: worldbuilding-brainstorm
description: Canon-safe worldbuilding brainstorming that develops lore one question at a time and records Story Bible changes only after explicit author confirmation.
route:
  actionKinds:
    - chat
    - read_check
    - story_bible
  contentKinds:
    - story_bible_entry
    - entry_section
    - project_note
    - attached_note
    - chapter
    - story_text
  tags:
    - fiction
    - worldbuilding
    - brainstorming
    - canon-safe
  priority: 94
metadata:
  useCases:
    - The author wants to brainstorm or reconcile characters, places, systems, factions, history, or other setting material.
    - The author needs lore ideas challenged against existing Story Bible entries or on-page evidence.
    - The author wants to turn vague worldbuilding into confirmed, durable canon.
  doNotUse:
    - The author wants direct Story Text drafting or scene revision.
    - The task is only diagnostic and should not expand lore.
    - The author wants canon changed without review.
  status: default
  source:
    - docs > skills > important > worldbuilding-brainstorm > SKILL.md
  scriptStatus: no-source-helpers
---

# $worldbuilding-brainstorm

The default lore-building skill for Edda: it pressure-tests setting ideas against existing project evidence, develops options one decision at a time, and keeps durable canon under explicit author control.

## Edda Workflow

1. Map the project before making substantive recommendations. Use `project_map` to identify relevant Story Bible entries, Writing Briefs, Project Notes, Attached Notes, and chapters; use `search_content` for topic names, aliases, factions, places, species, magic systems, technologies, institutions, history, and character names connected to the request.
2. Read the evidence that can constrain the idea. Use `read_story_bible_entry` or `read_entry_section` for existing canon, `read_content` for notes and briefs, and `read_chapter` for on-page evidence. Treat Story Text as read-only evidence in this skill.
3. If the author's topic crosses several domains, inspect all affected areas before challenging the idea: world rules, character facts, chronology, geography, economics, religion, law, faction interests, everyday life, taboos, power limits, and downstream story promises.
4. Label every important claim while brainstorming:
   - `Existing canon`: stated by Story Bible entries, Writing Briefs, or confirmed project notes.
   - `Story evidence`: shown or implied by chapter text.
   - `Inference`: likely implication from existing material, not confirmed canon.
   - `Proposal`: an option or recommendation generated in the session.
   - `Confirmed canon`: a durable fact the author explicitly approves.
5. Ask one targeted question at a time and wait for the author before moving to the next unresolved branch. When useful, include your recommended answer and a short reason grounded in the project evidence.
6. Generate concrete option sets instead of a single vague answer when the branch is underdetermined. For each option, state what it solves, what it breaks, which characters or institutions it pressures, and what new questions it creates.
7. Stress-test promising options with small scenarios: edge cases, social consequences, abuse cases, mixed ancestry or membership, institutional response, legal status, trade, inheritance, military use, class effects, religious interpretation, and everyday inconvenience.
8. Surface contradictions immediately. Use the pattern: `Existing canon/evidence says X; this proposal implies Y. Should we revise X, narrow Y, make it an in-world contradiction, or reject the proposal?` Do not quietly smooth over conflicts.
9. Sharpen fuzzy language into stable terms. If the author says a broad word like magic, curse, noble, demon, human, empire, talent, skill, faith, or law, ask what the project-specific term should mean and whether nearby terms must be split or merged.
10. Keep branch state visible during long sessions. Track decided facts, active options, rejected options, open questions, contradictions, and affected entries or chapters. Do not treat rejected options as canon unless the rejection itself becomes a durable fact.
11. Preserve author ownership. Recommend firmly, but do not decide canon. If a question can be answered by reading project content, read it instead of asking the author to restate it.
12. Do not rewrite Story Text, patch chapters, or propose full prose replacements in this skill. If a confirmed lore decision affects existing chapters, identify the affected chapters or scenes and flag them for a separate revision or continuity pass.

## Canon And Story Bible Policy

- Never update a Story Bible entry or section during exploration. First present the proposed canon change in chat or a reviewable note and ask for explicit author confirmation.
- After confirmation, update only the relevant Story Bible entry or section with concise durable facts. Preserve the entry's existing structure unless the author asks to reorganize it or the current structure cannot hold the confirmed fact clearly.
- Use `update_story_bible_entry` or `update_entry_section` only for confirmed canon. Do not use these tools for transcripts, discarded alternatives, unresolved branches, or speculative possibilities.
- Record unresolved questions only when the author asks to preserve them. Put them in a clearly labeled open-questions area of the relevant Story Bible entry, Project Note, or Attached Note.
- For character-specific durable facts, relationships, constraints, or backstory, propose the appropriate character Story Bible update after confirmation. For setting-wide facts, propose the relevant worldbuilding, faction, place, system, or history entry.
- When a confirmed change contradicts existing canon, ask whether the old canon should be revised, narrowed, retconned, or reframed as an in-world belief before applying any Story Bible update.

## Edda Output Handling

- Keep active question-by-question brainstorming in chat by default.
- Use an Attached Note when the lore problem is tied to one chapter, selection, or local continuity issue.
- Use a Project Note for option banks, unresolved branches, cross-chapter contradiction reports, or follow-up decision logs that are not yet canon.
- Use Story Bible proposals for durable facts about lore, characters, factions, places, names, rules, history, technology, magic, culture, economy, religion, institutions, or timeline.
- When asking for confirmation, separate the update into `Confirmed canon to write`, `Source evidence`, `Known consequences`, and `Affected Story Text for later review`.
- Do not use Structured Writes for Story Text in this skill.

## Script Compatibility

This source skill has no required helper scripts. The Edda rewrite works entirely through `project_map`, `search_content`, read tools, chat-based branching, reviewable notes, and explicit Story Bible update tools after author confirmation.
