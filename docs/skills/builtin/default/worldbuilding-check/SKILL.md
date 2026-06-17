---
name: worldbuilding-check
description: Diagnostic worldbuilding review for backdrop settings, weak consequences, shallow institutions, implausible economies, thin cultures, and canon that feels designed instead of lived-in.
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
  useCases:
    - The world feels like backdrop instead of a living system.
    - Technology, magic, history, institutions, economics, beliefs, species, or culture do not create convincing consequences.
    - Setting details feel surface-level, generic, inconsistent, over-explained, or disconnected from plot and character choices.
    - The author wants a worldbuilding diagnosis before expanding or revising canon.
  doNotUse:
    - The author wants exploratory lore creation; use `$worldbuilding-brainstorm`.
    - The request is mainly about prose, pacing, dialogue, character arc, or scene sequencing.
    - The author wants immediate canon changes without diagnosis.
  status: default
  source:
    - docs > skills > suggested > fiction > worldbuilding > worldbuilding > SKILL.md
  scriptStatus: no-source-helpers
---

# $worldbuilding-check

Diagnostic world review for authors who need to understand why a setting feels thin, inconsistent, or disconnected from the story.

## Edda Workflow

1. Establish evidence before diagnosing:
   - Use `project_map` to identify relevant chapters, Story Bible entries, project notes, and attached notes.
   - Use `read_chapter`, `read_content`, `read_story_bible_entry`, and `read_entry_section` for the specific setting material under review.
   - Use `search_content` when a rule, place, faction, technology, magic system, species, term, or institution appears in multiple places.
   - Use `list_revisions` only when the author asks whether a worldbuilding problem appeared or changed across revisions.
2. Separate evidence into `Existing canon`, `Story-text evidence`, `Inference`, and `Open question`. Do not treat inferred explanations as confirmed lore.
3. Check integration with plot and character:
   - Identify which world elements affect character goals, available choices, social pressure, conflict, risk, and aftermath.
   - Flag elements that are decorative backdrop because they do not change decisions, costs, opportunities, taboos, or power.
   - Flag plot conveniences where technology, magic, law, travel, medicine, surveillance, logistics, or institutions appear only when the scene needs them.
4. Diagnose surface versus systemic worldbuilding:
   - Surface worldbuilding is names, aesthetics, customs, maps, species traits, rituals, titles, foods, or slogans that do not imply behavior.
   - Systemic worldbuilding shows how an element changes economy, authority, religion, class, family structure, language, infrastructure, warfare, crime, education, or daily routines.
   - For each central speculative or historical element, trace at least first-order practical effects, second-order institutional adaptations, and third-order cultural normalization.
5. Classify the main failure state using concrete source criteria:
   - `Backdrop world`: the setting exists but has no independent logic beyond the current scene.
   - `World without consequences`: technology, magic, historical divergence, geography, or species biology has not transformed society.
   - `Institutions without history`: organizations feel invented for the plot, with no founding pressures, crises survived, reforms, contradictions, corruption, or internal factions.
   - `Economy that does not make sense`: trade, prices, labor, scarcity, infrastructure, taxation, supply chains, or underground markets are missing or arbitrary.
   - `Shallow belief system`: religion, ideology, taboo, law, or philosophy is flavor without daily decisions, schisms, doctrine, rituals, moral conflicts, or political effects.
   - `Culture without depth`: traditions, customs, clothing, food, names, and manners feel random or monocultural instead of shaped by region, class, history, environment, and conflict.
   - `Flat non-humans`: species, aliens, monsters, or supernatural groups are humans in costume because biology, sensory experience, lifespan, reproduction, cognition, or ecology does not affect culture.
   - `Generic language texture`: names and terms lack linguistic pattern, history, regional variation, or social meaning.
6. Check consistency and canon safety:
   - Compare rules, timelines, geography, capabilities, limitations, names, ranks, faction motives, and public knowledge across all read evidence.
   - Distinguish contradiction from deliberate mystery, unreliable narration, local ignorance, propaganda, or changed canon.
   - When evidence conflicts, report the conflict with citations to the relevant content rather than choosing a new canon answer.
7. Check lived-in detail:
   - Look for ordinary people, jobs, maintenance, bureaucracy, failures, slang, jokes, workarounds, class differences, regional variation, dissent, black markets, and habits formed by the world.
   - Flag worlds that only show palaces, battlefields, rituals, and named elites when the story needs a broader social texture.
   - Do not demand depth for every background element; go deep when the element is central to plot, repeated, examined closely, or creates ongoing conflict.
8. Check exposition problems:
   - Flag lore dumps where explanation arrives before the reader needs it, repeats known information, or pauses character pressure.
   - Flag missing exposition when the reader cannot understand stakes, constraints, social meaning, or consequences.
   - Recommend moving world information into action, choice, conflict, cost, environment, overheard norms, institutional procedure, or character misunderstanding when that would solve the problem.
9. Produce a ranked diagnosis:
   - Name the strongest worldbuilding issue first.
   - For each issue, give the evidence, why it weakens story credibility, what consequence chain is missing, and what local or canon-level intervention would address it.
   - Keep solutions targeted. Do not generate a new world from scratch unless the author asks to continue with `$worldbuilding-brainstorm` or another creation skill.

## Edda Output Handling

- Return the diagnostic report in chat by default.
- Create an Attached Note when the diagnosis belongs to one chapter or one local continuity problem.
- Create or update a Project Note when the worldbuilding issues span multiple entries or chapters.
- Propose Story Bible changes only as reviewable recommendations, not confirmed canon. Label each recommendation as `Proposal` until the author explicitly accepts it.
- If the author asks for applied changes, first provide the diagnosis and ask whether to move into a rewrite, Story Bible update, or brainstorming workflow.
- Do not use Structured Writes in this skill.

## Script Compatibility

This source skill has no required helper scripts. The rewrite works entirely through Edda-native context tools, diagnosis, and canon-safe recommendations.
