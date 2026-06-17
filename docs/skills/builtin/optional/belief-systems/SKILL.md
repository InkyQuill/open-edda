---
name: belief-systems
description: Build religions, spiritual practices, and moral frameworks that feel lived in, internally varied, and historically shaped.
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
    - culture
    - optional
  priority: 44
metadata:
  useCases:
    - A setting needs faith, ritual, or meaning structures with real social consequences.
    - Characters are shaped by belief, heresy, taboo, or spiritual obligation.
    - The author wants internal factions, lived practice, and historical change.
  doNotUse:
    - The request is mostly about government, trade, or settlement planning.
    - The story only needs a quick symbolic detail rather than a full belief framework.
    - The author wants canon fixed before exploration.
  status: optional
  source:
    - docs > skills > suggested > fiction > worldbuilding > belief-systems > SKILL.md
  scriptStatus: no-source-helpers
---

# $belief-systems

Belief-system design for authors who need religion, ritual, doctrine, doubt, and institutional power to function as lived worldbuilding and story pressure rather than decorative lore.

## Edda Workflow

1. Read the relevant Story Bible Entries, Entry Sections, Project Notes, Attached Notes, and chapter evidence before inventing doctrine. Separate confirmed canon from author-facing proposals and open questions.
2. Identify the belief system's story function before expanding lore. Name what the system must do in the draft: explain misfortune, legitimize authority, bind a community, regulate behavior, transmit history, justify taboo, create dissent, motivate a character, or pressure a plot choice.
3. Ground the system in the source principles: experiential foundation, ecological integration, social cohesion, institutional evolution, power legitimation, narrative embedding, ritual entrainment, adaptive reinterpretation, orthopraxy versus orthodoxy, and syncretism versus boundedness. Do not design a faith that is only a list of gods or abstract claims.
4. Build the doctrine as answers to concrete questions: origin and cosmic structure, divine or spiritual agents, afterlife or destiny, moral foundations, explanations for suffering, sources of knowledge, interpretive authority, doubt management, outsider treatment, and the balance between correct belief and correct practice.
5. Design ritual and daily practice together. Specify regular observances, crisis rituals, life-cycle ceremonies, sacred calendar, materials or offerings, emotional tone, identity markers, diet or clothing rules, household practices, work rhythms, and sacred spaces. Show what believers do on an ordinary day, not only at temples or festivals.
6. Design institutions and material incentives. Name leadership structures, specialist roles, training, initiation, property, taxes or donations, services, state support, economic privileges, obligations to the poor or powerful, and the practical benefits of membership. Make clear who gains power, money, legitimacy, protection, status, or absolution.
7. Add internal variety. Include at least two interpretations, factions, regional practices, class differences, reform movements, skeptics, zealots, or syncretic variants. For each heresy or schism, specify whether the break is doctrinal, political, regional, generational, economic, or caused by a crisis.
8. Test for contradiction and adaptation. A strong belief system contains tensions between doctrine and practice, compassion and purity, hierarchy and equality, local custom and central authority, missionary impulse and boundary maintenance, sacred poverty and institutional wealth, or official teaching and survival needs. State how believers reinterpret old claims when history changes.
9. Tie the system to setting genre and material conditions. For fantasy, decide whether divine presence, magic, multiple species, or manifest afterlives are verifiable and how that changes faith. For science fiction, account for AI, alien life, space habitation, virtual afterlives, or engineered transcendence. For post-apocalyptic settings, account for artifacts, collapse theodicy, survival knowledge, and remnant traditions.
10. Apply the source anti-pattern checks before presenting output: avoid monolithic belief, theology without practice, simplistic good or evil dualism, deities used only as power sources, and modern secular values wearing ancient costume. If any appear, revise with concrete factions, rituals, social consequences, meaning-making, and historically shaped values.
11. Connect the belief system to story pressure. Identify scenes, relationships, plot turns, character choices, moral conflicts, institutional threats, taboos, oaths, ceremonies, punishments, or scandals where belief can change what characters are willing or allowed to do.
12. Keep canon-safe proposals explicit. Use labels such as `Confirmed`, `Inferred`, `Proposed`, and `Open question` when the output could change names, gods, history, doctrine, rituals, institutions, sacred law, factions, geography, or timeline.

## Design Criteria

A complete belief-system proposal should cover these criteria when relevant to the author's request:

- Belief system function: the social, emotional, political, and narrative work the faith or moral framework performs.
- Doctrine: cosmology, divine or spiritual order, destiny, ethics, knowledge sources, authority, doubt, and truth claims.
- Ritual: daily, seasonal, crisis, and life-cycle practices with participants, materials, formalization, and emotional tone.
- Institutions: leadership, specialists, initiation, hierarchy, economics, political relationship, property, and enforcement.
- Heresy and schism: internal disagreement with clear stakes, not random flavor factions.
- Daily practice: ordinary behavior shaped by belief, including food, clothing, language, schedule, labor, family, sex, death, charity, purity, and public identity.
- Social power: who is legitimized, excluded, taxed, protected, educated, punished, married, absolved, or made sacred.
- Contradiction: places where believers sincerely hold values that conflict with survival, politics, wealth, family, or compassion.
- Material incentives: concrete benefits and costs that make participation, reform, hypocrisy, or dissent plausible.
- Story pressure: specific ways belief can force choices, expose hypocrisy, create danger, bind alliances, or make a scene harder.
- Canon-safe proposals: durable facts remain proposals until the author confirms them.

## Quality Bar

- Good output gives a lived system with doctrine, ritual, institution, variation, incentives, contradictions, and scene-useful pressure.
- Good output makes opposing believers understandable even when their claims conflict.
- Good output shows how the same tradition changes across region, class, generation, crisis, or contact with other beliefs.
- Weak output is only a pantheon, a moral slogan, a magic-power source, a single uniform culture, or a religion that has no cost in daily life.
- Weak output treats belief as universally sincere or universally cynical; include faith, habit, doubt, ambition, fear, comfort, exploitation, and reform where appropriate.

## Edda Output Handling

- Return short diagnosis, option sets, or canon questions in chat when the author is still deciding.
- Create an Attached Note when the output diagnoses or proposes belief-system pressure for one Chapter, one selection, one scene, one community, or one local conflict.
- Create or update a Project Note when the output is an exploratory framework, faction map, ritual set, schism history, contradiction list, or cross-chapter design plan.
- Propose Story Bible updates only after the author confirms durable names, doctrines, rituals, institutions, sacred history, factions, laws, or setting facts. Keep unconfirmed material in a proposal block or note.
- Use Structured Writes only when the author explicitly asks for an applied rewrite, entry update, or durable note creation.

## Script Compatibility

This source skill has no helper scripts. The source methodology is converted into Edda-native guidance, analysis, and canon-safe proposals.
