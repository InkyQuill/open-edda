---
name: character-names
description: Generate culturally grounded, collision-aware character names using built-in reference pools instead of generic model defaults.
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
    - character
    - naming
    - optional
  priority: 48
metadata:
  status: optional
  source:
    - docs > skills > suggested > fiction > character > character-naming > SKILL.md
  scriptStatus: source-helpers-deferred-data-retained
---

# $character-names

Name generation for authors who want names that fit setting, culture, tone, and cast distinctiveness without falling into repetitive defaults.

## Use When

- The author needs names for new characters, factions, or families.
- Existing names feel interchangeable, collision-prone, or tonally wrong.
- Character naming needs to stay consistent with Story Bible culture notes or phonology rules.

## Do Not Use When

- A name is already canon and the author is not reconsidering it.
- The request is really about language design; use `$conlang`.
- The author wants final canon changes without reviewing options.

## Edda Workflow

1. Read the relevant Story Bible material, cast notes, and any names already on the page.
2. Determine the naming context: real-world culture, mixed setting, or invented naming logic.
3. Generate a shortlist with collision checks, tonal fit, and quick rationale.
4. Flag lookalike names, repeated initials, or mismatched cultural cues.
5. Treat the chosen name as canon only after the author confirms it.

## Edda Output Handling

- Return shortlists and collision notes in chat by default.
- Create an Attached Note when name work belongs to one Chapter, scene, or local cast problem.
- Create or update a Project Note when the author wants a reusable cast list or naming brief.
- Propose Story Bible updates for confirmed names, naming conventions, or cast-tracker fields only after author confirmation.
- Do not use Structured Writes in this skill.

## Bundled Data

This skill includes naming datasets and a template as Writer-native references:

- `data/cultures/` — 52 culture-specific naming dataset files (given names, surnames, gender variants).
- `data/phoneme-presets/` — 3 phoneme profiles for constructing fantasy or speculative names.
- `data/mixed-pools/` — Contemporary American mixed naming pool.
- `templates/cast-name-tracker.md` — Template for tracking character names and their narrative fit.

The agent should load these through the `skill` tool when a naming session needs detailed reference material.

## Script Compatibility

This rewrite preserves built-in naming datasets and the cast-tracker template as Edda-native references. Source helper scripts are not runnable in Milestone 3.5, so the skill works through guidance, built-in data, and reviewable name proposals only.
