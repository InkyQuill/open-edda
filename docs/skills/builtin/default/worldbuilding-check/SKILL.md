---
name: worldbuilding-check
description: Diagnostic worldbuilding review for thin settings, weak consequences, shallow institutions, and canon that feels designed instead of lived-in.
route:
  actionKinds:
    - chat
    - read_check
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
    - diagnosis
    - continuity
  priority: 78
metadata:
  status: default
  source:
    - docs > skills > suggested > fiction > worldbuilding > worldbuilding > SKILL.md
  scriptStatus: no-source-helpers
---

# $worldbuilding-check

Diagnostic world review for authors who need to understand why a setting feels thin, inconsistent, or disconnected from the story.

## Use When

- The world feels like backdrop instead of a living system.
- Institutions, economics, beliefs, or consequences do not feel convincing.
- The author wants a worldbuilding diagnosis before expanding or revising canon.

## Do Not Use When

- The author wants exploratory lore creation; use `$worldbuilding-brainstorm`.
- The request is mainly about prose, pacing, or dialogue.
- The author wants immediate canon changes without diagnosis.

## Edda Workflow

1. Read the relevant Story Bible material and the on-page story evidence that relies on it.
2. Identify the weak point: consequence gaps, shallow institutions, monoculture, thin species design, or similar issues.
3. Explain what feels unearned or underdeveloped and why.
4. Suggest targeted areas to strengthen, keeping canon proposals separate from confirmed facts.
5. Hand off to `$worldbuilding-brainstorm` if the author wants to build solutions.

## Edda Output Handling

- Return the diagnostic report in chat by default.
- Create an Attached Note when the diagnosis belongs to one chapter or one local continuity problem.
- Create or update a Project Note when the worldbuilding issues span multiple entries or chapters.
- Propose Story Bible changes only as reviewable recommendations, not confirmed canon.
- Do not use Structured Writes in this skill.

## Script Compatibility

This source skill has no required helper scripts. The Milestone 3.5 rewrite works entirely through Edda-native diagnosis and canon-safe recommendations.
