---
name: conlang
description: Design coherent naming languages, invented vocabulary, dialect drift, loanwords, and lightweight language history for fiction worldbuilding.
route:
  actionKinds:
    - chat
    - read_check
    - rewrite
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
  useCases:
    - Character, place, faction, ritual, alien, or culture names need shared phonology, syllable patterns, or naming rules.
    - Existing invented words feel random, hard to pronounce, biologically mismatched, or inconsistent across one culture or species.
    - The author wants a naming language with lexicon domains, registers, dialects, loanwords, sound shifts, or language-family history.
    - A Story Bible language note needs review for phonotactic consistency, diachronic plausibility, contact effects, or canon-safe expansion.
  doNotUse:
    - The author only needs a few standalone names and no language system; use `$character-names` when available.
    - The request is mainly prose voice, dialogue polish, translation style, or line editing rather than language design.
    - The author wants a complete grammar or academic linguistic reconstruction beyond fiction-useful naming and worldbuilding support.
    - The author asks to make new language facts final canon without review.
  status: optional
  source:
    - docs > skills > suggested > fiction > worldbuilding > conlang > SKILL.md
    - docs > skills > suggested > fiction > worldbuilding > language-evolution > SKILL.md
  scriptStatus: source-helpers-deferred-data-retained
---

# $conlang

Construct fiction-useful languages by keeping sounds, word shapes, naming rules, social registers, and language history internally consistent while leaving canon changes reviewable.

## Edda Workflow

1. Read the relevant project context before designing or diagnosing language: target Chapter or selection, existing Story Bible entries for cultures/species/places/history/religion/governance, prior Project Notes about languages or names, and any established invented words the author wants preserved.
2. Separate `existing canon`, `inference`, and `proposal`. Treat current Story Bible facts and already-published names as constraints. Treat new phonemes, glosses, etymologies, dialects, loanwords, and language history as proposals until the author confirms them.
3. Diagnose the current language state:
   - `L1 No Language`: names have no shared sound identity or cultural naming logic.
   - `L2 Relexified English`: words or phrases substitute invented forms into English grammar or concepts without linguistic difference.
   - `L3 Inconsistent Phonology`: names from one culture use incompatible sounds, clusters, or syllable shapes.
   - `L4 Missing Depth`: the language has no formal/informal register, dialect variation, archaic layer, professional jargon, sacred vocabulary, or history.
   - `L5 Biology Mismatch`: a non-human speaker uses sounds, mouth movements, or concepts that do not fit its body or cognition.
4. Choose the needed scope and say which scope you are using:
   - `Flavor`: 10-15 consonants, 3-5 vowels, 2-4 syllable templates, and 10-20 sample names for quick background use.
   - `Naming`: 15-22 consonants, 5-7 vowels, reusable syllable templates, phonotactic limits, and a 50+ word/name bank when the culture recurs.
   - `Full foundation`: 20-35 consonants, 7-12 vowels, phonotactic constraints, stress/register notes, and grammar-ready hooks only when the author asks for close language attention.
5. Build the phonology from a small repeatable sound palette. Include consonants, vowels, excluded sounds, distinctive features, and a reader-accessibility note. Prefer common phonemes unless the story needs a rare feature; choose at most one or two distinctive features such as tone, ejectives, clicks, vowel harmony, or glottal stops.
6. Define syllable and word-shape rules before making names. Specify templates such as `CV`, `CVC`, `CVV`, `CCVC`, maximum onset/coda size, cluster rules, stress pattern, apostrophe meaning if used, and which shapes are reserved for personal names, places, titles, ritual terms, or common words.
7. Create or revise sample words against the rules. For each proposed name or word, provide pronunciation guidance, meaning or use-domain if known, syllable breakdown when helpful, and whether it is a personal name, place name, clan/faction name, ritual term, title, insult, sacred word, or everyday word.
8. Add lexicon and register structure when the setting needs depth. Tie vocabulary domains to world systems: power and governance create honorifics and taboo words; belief creates sacred and liturgical registers; economy creates trade terms; geography creates landscape names; occupations create jargon; intimacy creates private words.
9. Add diachronic logic when the author wants history or related languages. Start from a proto-language or older register, define systematic sound shifts, then apply them consistently to daughter dialects, place names, surnames, sacred survivals, and common words. Track regular correspondences rather than changing each word ad hoc.
10. Model language change through concrete mechanisms: lenition, fortition, vowel shift, palatalization, assimilation, metathesis, grammaticalization, analogical leveling, case simplification, tense/aspect development, lexical borrowing, calques, code-switching, contact-zone convergence, pidgin/creole development, and spelling or ritual forms that lag behind speech.
11. Model dialect drift from geography and society. Use mountains, rivers, migration, conquest, trade routes, isolation, class, profession, age, gender roles, religion, and political borders to explain which speakers preserve conservative forms and which innovate.
12. Handle loanwords as evidence of contact. Name the source language or group, borrowed domain, adaptation into the receiving phonology, social prestige or stigma, and whether the loan is direct borrowing, calque, technical jargon, sacred borrowing, trade pidgin, or conquest residue.
13. Check anti-patterns before returning output: kitchen-sink exoticism, unpronounceable strings, undefined apostrophes, incompatible phonotactics, English grammar with swapped words, perfect regularity with no historical residue, frozen languages across millennia, contact with no borrowing, and monolingual societies with no register or regional variation.
14. If the author asks for applied changes to Story Text, keep the rewrite localized to names, glosses, or brief language-facing phrasing. Do not rewrite unrelated prose. If language rules affect durable canon, present a Story Bible proposal first.

## Data Files

Load supporting data with `read_skill_file` only when the task needs concrete phoneme or syllable reference beyond the workflow:

- `data/phoneme-frequencies.json`: use for realistic phoneme inventory choices, frequency tiers, complexity presets, and family-pattern reminders.
- `data/syllable-templates.json`: use for syllable complexity levels, style presets, cluster constraints, word structures, and stress-pattern guidance.

Do not treat these files as canon. They are construction aids for proposals.

## Edda Output Handling

- Return short diagnostic advice, scope choice, phoneme inventory, syllable rules, name lists, and register ideas in chat when the author is deciding.
- Create an Attached Note when the work belongs to one Chapter, selection, location, culture mention, name list, or local continuity issue.
- Create a Project Note for reusable language briefs, word banks, dialect maps, sound-shift tables, family trees, contact-zone notes, or cross-chapter diagnostics.
- Propose Story Bible updates only after author confirmation for durable language facts: phonology, orthography, names, glossaries, etymologies, registers, dialects, language families, historical sound changes, loanword sources, or sacred/legal terminology.
- Use Structured Writes only when the author explicitly asks to apply a localized name, word, glossary, or language-note change and the target content is known.

## Script Compatibility

The source `phonology.ts` and `words.ts` helpers generated seeded phoneme inventories and word lists from data files. In this optional built-in rewrite, their methodology is preserved as guidance and data loading: choose complexity, select a bounded phoneme inventory, define syllable templates, generate categorized words, record seed-like parameters when useful, and show syllable structure for consistency checks.

The scripts themselves are deferred helpers. Do not ask the author to run them, do not call shell commands, and do not assume script execution is available. Use `skill_script` only if these helpers are later imported, audited, approved, enabled, and the author explicitly wants generated inventory or word-list assistance.
