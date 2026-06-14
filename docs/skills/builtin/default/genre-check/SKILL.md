---
name: genre-check
description: Genre-fit diagnosis for stories whose promise, conventions, or hybrid balance feel unclear or mismatched.
route:
  actionKinds:
    - chat
    - read_check
  contentKinds:
    - chapter
    - story_text
    - project_note
    - attached_note
  tags:
    - fiction
    - genre
    - diagnosis
    - expectations
  priority: 74
metadata:
  status: default
  source:
    - docs > skills > suggested > fiction > craft > genre-conventions > SKILL.md
  scriptStatus: source-helpers-converted-to-guidance-and-data
---

# $genre-check

Genre diagnosis for authors who need to confirm what emotional promise the story is making and whether the current draft actually delivers it.

## Use When

- The story's genre promise feels unclear or mixed up.
- Required genre elements feel missing, misplaced, or stale.
- Secondary genre material may be competing with the primary experience.

## Do Not Use When

- The author already knows the genre problem and needs focused scene, prose, or ending work.
- The request is mainly about setting details rather than reader expectation.
- The author wants the skill to choose a genre identity by force.

## Edda Workflow

1. Read the relevant chapter, note, or Writing Brief.
2. Identify the likely primary emotional promise and any meaningful secondary genres.
3. Check whether the opening, middle, and current problem area support that promise.
4. Call out missing conventions, misplaced emphasis, or stale defaults.
5. Recommend next steps or hand off to a more specific skill when needed.

## Edda Output Handling

- Return the genre diagnosis in chat by default.
- Create an Attached Note when the report belongs to one chapter or one local promise problem.
- Create or update a Project Note when the author wants a genre checklist or hybrid hierarchy recorded for later work.
- Do not use Structured Writes in this skill.
- Do not propose Story Bible changes unless genre diagnosis reveals a durable constraint the author wants recorded.

## Bundled Data

This skill includes reference data the agent can consult during genre diagnosis:

- `data/genre-elements.json` — Genre element expectations and conventions.

The agent should load this through the `skill` tool when diagnosing genre fit.

## Script Compatibility

This rewrite adapts source genre checks and reference tables into Edda-native diagnosis and built-in genre guidance. Source helpers are not runnable in Milestone 3.5.
