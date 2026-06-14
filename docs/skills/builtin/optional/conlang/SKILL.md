---
name: conlang
description: Design naming languages and lightweight language history that keep invented words, place names, and speech patterns coherent.
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
    - language
    - optional
  priority: 46
metadata:
  status: optional
  source:
    - docs > skills > suggested > fiction > worldbuilding > conlang > SKILL.md
    - docs > skills > suggested > fiction > worldbuilding > language-evolution > SKILL.md
  scriptStatus: source-helpers-deferred-data-retained
---

# $conlang

Constructed language support for authors who need coherent naming languages, vocabulary rules, or language history without turning the project into a linguistics dissertation.

## Use When

- Character, place, faction, or ritual names need a shared sound logic.
- The author wants a naming language plus a little history, register, or dialect drift.
- Existing invented words feel random or inconsistent.

## Do Not Use When

- The author only needs a few names and no broader language logic; `$character-names` may be enough.
- The request is mainly about prose style or dialogue polish.
- The author wants final canon inserted into Story Text without review.

## Edda Workflow

1. Read existing names, Story Bible language notes, and any relevant world constraints.
2. Define the language scope: quick naming palette, deeper lexicon rules, or historical divergence.
3. Build a compact phonology, syllable patterns, and naming rules that fit the setting.
4. Generate example words or registers and test them against what is already canon.
5. Keep all new vocabulary and language history reviewable until the author confirms them.

## Edda Output Handling

- Return phonology notes, sample lexicon, and naming guidance in chat by default.
- Create an Attached Note when the work serves one Chapter, one location, or one naming problem.
- Create or update a Project Note when the author wants a reusable language brief.
- Propose Story Bible updates for confirmed language rules, glossaries, or historical notes only after author confirmation.
- Do not use Structured Writes in this skill.

## Bundled Data

This skill includes linguistic reference data as Writer-native references:

- `data/phoneme-frequencies.json` — Cross-linguistic phoneme frequency data.
- `data/syllable-templates.json` — Syllable structure templates for language construction.

The agent should load these through the `skill` tool when a language construction session needs detailed reference material.

## Script Compatibility

This rewrite preserves built-in phoneme and syllable data as Edda-native references. Source helper scripts are not runnable in Milestone 3.5, so the skill works through guidance, data, and reviewable language notes only.
