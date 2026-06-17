---
name: governance-systems
description: Design fictional polities, institutions, succession, laws, factions, corruption, authority limits, and crisis behavior.
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
    - governance
    - optional
  priority: 42
metadata:
  useCases:
    - A story depends on legitimacy, institutions, succession, law, bureaucracy, enforcement, border tension, or elite conflict.
    - A kingdom, empire, republic, federation, corporate authority, religious domain, guild power, city-state, fleet, or council needs internal logic.
    - A polity feels like one villain, one uniform culture, a decorative title hierarchy, or an administration with impossible reach.
    - The author wants canon-safe political proposals, faction maps, crisis behavior, or institutional failure modes.
  doNotUse:
    - The request is mainly about prose, pacing, or local scene revision.
    - The request is mainly about economics, belief systems, military tactics, or general geography rather than governance behavior.
    - The author only needs a name or honorific, not institutional design.
    - The author wants existing canon treated as fixed and complete with no new political proposals.
  status: optional
  source:
    - docs > skills > suggested > fiction > worldbuilding > governance-systems > SKILL.md
  scriptStatus: no-source-helpers
---

# $governance-systems

Political worldbuilding for authors who need authority, institutions, factions, law, enforcement, succession, and administrative limits to behave like a working system rather than a decorative hierarchy.

## Edda Workflow

1. Read the relevant Story Bible entries, Writing Briefs, Project Notes, and any Chapters or selected Story Text where the polity, institution, law, ruler, faction, border, rebellion, crisis, or office already appears. Separate existing canon from inference before proposing anything new.
2. Establish the polity's environmental and historical foundation: territory, resources, communication barriers, threat environment, formation pathway, critical junctures, inherited institutions, and current governance cycle stage such as formation, expansion, stability, contraction, or transformation.
3. Identify the polity type and avoid pure models unless the story has a reason for them. Use hybrids when appropriate: unitary state, federation, confederacy, constitutional or absolute monarchy, military regime, single-party state, theocracy, city-state, tribal confederation, corporate authority, religious domain, criminal territory, autonomous zone, mercenary control, nomadic federation, trade league, technocratic enclave, hegemonic empire, interspecies concordat, protectorate, security alliance, or aristocratic house system.
4. Diagnose legitimacy. State which sources justify rule and who accepts or rejects each source: traditional legitimacy, charismatic loyalty, legal-rational procedure, performance benefits, ideology, divine mandate, conquest, contract, emergency necessity, or external recognition.
5. Map real power sources, not just formal offices. Include control over land, money, food, trade routes, archives, courts, temples, media, technical expertise, magic or technology, military force, police, intelligence, patronage, marriage ties, hostages, foreign backing, or popular mobilization.
6. Define bureaucracy and administrative reach. Specify who records decisions, collects revenue, appoints officials, communicates orders, resolves petitions, audits provinces, and enforces standards. Match reach to transportation, communication, literacy, surveillance, and institutional capacity; distant or low-tech regions should have negotiated, layered, or nominal control rather than precise central micromanagement.
7. Define law and enforcement as separate systems. State what counts as law, who interprets it, who can ignore it, how courts or councils work, what punishments exist, who controls coercive forces, and where customary, religious, military, guild, noble, or local law competes with central law.
8. Design succession and power transition. Specify the normal transition pattern, the emergency fallback, the contested edge case, and the groups that can veto or exploit succession: dynastic inheritance, elite selection, election, appointment, acclamation, prophecy, seniority, examination, coup, external imposition, or temporary regency.
9. Map internal cohesion and factions. Include at least three meaningful divisions for major polities: regional blocs, class interests, bureaucratic ministries, religious currents, military commands, merchant houses, clans, parties, reform movements, old-regime loyalists, colonized groups, youth cohorts, or ideological wings. Give each faction a legitimate goal, resource base, fear, and compromise it might accept.
10. Distinguish local and central authority. Identify where central authority is effective, negotiated, layered, contested, or merely nominal. For each region or institution, state what the center can command, what it must bargain for, what local actors can refuse, and what happens when orders travel slowly or arrive during a crisis.
11. Add corruption and informal governance. Name which offices are vulnerable to bribery, patronage, nepotism, smuggling, regulatory capture, embezzlement, protection rackets, selective enforcement, or loyalty trading. Treat corruption as an alternate power network with beneficiaries, costs, limits, and story pressure, not just moral decoration.
12. Test crisis response. For war, famine, plague, assassination, succession dispute, scandal, religious schism, debt shock, disaster, migration, rebellion, invasion, or technological disruption, state who acts first, who hesitates, who profits, which law is suspended, which faction gains leverage, and which local actors stop obeying.
13. Identify institutional failure modes. Check for overcentralization, brittle succession, legitimacy mismatch, underfunded enforcement, professional bureaucracy captured by elites, military autonomy, regional secession pressure, contradictory legal systems, extractive taxation, information bottlenecks, performative councils, unmanaged diversity, or a ruler whose authority exceeds the system's capacity.
14. Apply plot pressure. Tie the governance system to character choices and scenes: which office blocks the protagonist, which faction offers a dangerous bargain, which law creates the moral trap, which succession rule raises stakes, which corrupt shortcut tempts someone, and which crisis exposes the gap between official ideology and actual power.
15. Check source anti-patterns before finalizing: no "evil empire" without internal benefits and believers; no planet or species of hats without regional, class, generational, or ideological variation; no static thousand-year politics without transformations and crises; no binary world politics without neutrals, opportunists, and split loyalties; no administrative reach beyond technology and infrastructure.
16. Present canon-safe proposals. Label each new office, faction, law, historical event, border claim, succession rule, institution, or crisis response as a proposal until the author confirms it.

## Edda Output Handling

- Return concise governance diagnosis, options, or a proposed model in chat when the author is exploring or deciding.
- Create an Attached Note when the work is tied to one Chapter, selected passage, scene problem, local conflict, or continuity concern.
- Create or update a Project Note for durable political exploration, polity matrices, faction maps, crisis-response plans, or cross-chapter governance diagnostics.
- Propose Story Bible changes for durable facts about institutions, laws, offices, factions, names, borders, legitimacy claims, succession rules, historical junctures, and political relationships. Do not treat those facts as confirmed canon until the author approves them.
- Use Structured Writes only when the author explicitly asks to apply text to a known target. Otherwise provide proposed wording, note text, or Story Bible update text for review.
- Mark output sections as needed: `Existing canon`, `Inferences`, `Proposals`, `Open questions`, and `Story pressure`.

## Script Compatibility

This source skill has no helper scripts. All source methodology is converted into Edda-native reading, diagnosis, proposal, and note-output guidance.
