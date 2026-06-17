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
  useCases:
    - The author needs names for new characters, factions, or families.
    - Existing names feel interchangeable, collision-prone, or tonally wrong.
    - Character naming needs to stay consistent with Story Bible culture notes or phonology rules.
  doNotUse:
    - A name is already canon and the author is not reconsidering it.
    - The request is really about language design; use `$conlang`.
    - The author wants final canon changes without reviewing options.
  status: optional
  source:
    - docs > skills > suggested > fiction > character > character-naming > SKILL.md
  scriptStatus: source-helpers-deferred-data-retained
---

# $character-names

Name generation for authors who want names that fit setting, culture, tone, and cast distinctiveness without falling into repetitive defaults.

Core principle: language models drift toward statistical median names. When a request says "diverse" or "authentic" without constraints, the model will tend to produce familiar defaults such as Chen, Patel, Maya, Marcus, Sofia, or Aiden, then median-hop to another obvious default when corrected. Break that pattern by loading the built-in data pools with `read_skill_file` and drawing candidates from those pools instead of inventing names from model memory.

## Edda Workflow

1. Read the relevant Story Bible entries, writing brief, chapter selection, project notes, and existing cast notes before generating names. Separate confirmed canon from inferred setting logic and open author preferences.
2. Diagnose the naming state:
   - `CN1: No Context` - the author asks for names without setting, period, culture, genre, or cast context. Ask for the missing context before generating unless the author explicitly wants rough placeholders.
   - `CN2: Chen Proliferation` - names cluster around median defaults such as Chen, Kim, Patel, Garcia, Maya, Marcus, Sofia, or Aiden. Use data-pool candidates and avoid replacing one median with another.
   - `CN3: Cultural Incoherence` - invented or speculative names in the same culture do not sound related. Define a phoneme preset, syllable shape, and naming convention before offering candidates.
   - `CN4: Cast Collision` - names are too visually or sonically similar, such as Sarah/Sara, Mike/Mark/Michael, Lee/Leigh, repeated initials, same rhythm, or same surname without a story reason.
   - `CN5: Character Mismatch` - a name conflicts with the character's origin, generation, class, family logic, period, role, or intended associations.
   - `CN6: Mixed Setting` - a contemporary or historical setting includes multiple real-world cultural groups and needs distribution logic rather than a one-of-each checklist.
3. Establish naming constraints before proposing candidates: place, period, cultures or fictional cultures present, character background, generation or family naming logic, gendered or neutral name needs, role in cast, names already locked, sounds to avoid, and whether the name was given by family, chosen by the character, or assigned by an institution.
4. Select the generation mode:
   - Real-world contemporary or historical character: load a culture pool from `data/cultures/` for the relevant background. Use complete cultural packages when possible, pairing given-name and surname pools that belong together unless mixed heritage or adoption/family history justifies a blend.
   - Mixed contemporary setting: define the setting distribution first, then load `data/mixed-pools/contemporary-american.json` only when that pool matches the story. For any other mixed setting, combine specific culture pools according to the story's location and community logic.
   - Fantasy, science fiction, or invented culture: load one `data/phoneme-presets/` file and generate candidates that stay inside its consonants, vowels, and syllable templates. Use `elvish-like` for flowing vowel-heavy cultures, `harsh-fantasy` for hard consonants and stops, and `neutral` for pronounceable low-aesthetic-bias names. If the culture needs a full language system, route to `$conlang`.
5. Apply the external entropy principle inside Edda: after loading the data file, choose from different positions in the list and vary initials, syllable counts, and rhythms. Do not default to the first names, the most famous names, or model-invented alternatives. When possible, provide more candidates than needed and let filtering produce the shortlist.
6. Validate each candidate against the cast and context:
   - Cultural coherence: given name, surname, family generation, region, and period do not contradict established Story Bible logic.
   - Cast collision: initials, first sounds, endings, syllable counts, visual shape, nicknames, and surnames are distinct enough for readers.
   - Character fit: the name supports or intentionally subverts role, class/status signal, age cohort, family history, and story associations.
   - Mixed-setting fit: the cast distribution follows setting logic and community makeup rather than token coverage.
   - Functional fit: the name is pronounceable, spellable enough for the prose style, readable in dialogue, and not accidentally comic unless intended.
7. Present a shortlist with rationale and warnings. Mark each name as candidate, shortlisted, confirmed, or retired. Do not treat a selected name, convention, or family relationship as canon until the author confirms it.

## Edda Output Handling

- Return shortlists and collision notes in chat by default.
- Create an Attached Note when name work belongs to one Chapter, scene, or local cast problem.
- Create or update a Project Note when the author wants a reusable cast list or naming brief.
- Propose Story Bible updates for confirmed names, family naming logic, cultural naming conventions, pronunciation, aliases, or cast-tracker fields only after author confirmation.
- Do not use Structured Writes in this skill.

## Data Files

This skill includes naming datasets and a template as Writer-native references:

- `data/_meta.json` — Available cultures, pools, and phoneme presets.
- `data/cultures/{culture}-given.json`, `{culture}-given-male.json`, `{culture}-given-female.json`, and `{culture}-surnames.json` — Culture-specific naming pools.
- `data/phoneme-presets/{preset}.json` — Phoneme profiles for fantasy or speculative names.
- `data/mixed-pools/contemporary-american.json` — Contemporary American mixed naming pool.

Data-file loading protocol:

1. Load `data/_meta.json` with `read_skill_file` before choosing pools unless the author already named an exact supported culture or preset.
2. Load only the specific pool files required for the current request.
3. For full real-world names, load both the relevant given-name file and surname file. Use gendered given-name files only when the story context needs that signal; otherwise prefer the combined given-name file.
4. For mixed settings, read the mixed pool only after confirming it matches the setting; otherwise load the individual culture pools that match the story.
5. For invented names, load one phoneme preset and use its consonants, vowels, and syllable templates as constraints. Do not mix presets inside one culture unless the story has multiple languages, regions, or classes.
6. Treat loaded data as entropy and constraint material, not as automatic canon. A candidate becomes canon only through author confirmation.

## Templates

- `templates/cast-name-tracker.md` — Template for tracking character names and narrative fit.

Load this template with `read_skill_file` only when the author asks for a reusable cast tracker or naming brief.

## Script Compatibility

This rewrite preserves built-in naming datasets and the cast-tracker template as Edda-native references. The source `character-name.ts` and `cast-tracker.ts` helpers remain deferred because Edda imports scripts disabled by default and `read_skill_file` cannot expose script bodies. Convert their useful behavior into this workflow: data-pool generation, phoneme-preset generation, cast collision checks, and distribution review.

Use `skill_script` only if a future audited, approved, enabled non-mutating helper is available through Skill Core. Otherwise, do not ask the author to execute helper files and do not depend on script output.
