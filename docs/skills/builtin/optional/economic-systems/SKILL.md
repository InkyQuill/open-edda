---
name: economic-systems
description: Build resource, labor, trade, scarcity, class, taxation, market, and informal-economy logic that creates believable plot pressure.
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
    - economics
    - canon-safe
    - optional
  priority: 42
metadata:
  useCases:
    - A story depends on scarcity, debt, trade, class, extraction, taxation, currency, labor, markets, or economic survival.
    - The author wants a fictional economy, currency, trade network, class structure, resource economy, post-scarcity claim, or shadow market designed or checked.
    - Characters' choices should be shaped by what the economic system rewards, punishes, hides, taxes, rations, or makes scarce.
    - Existing worldbuilding has economic contradictions such as frictionless trade, invisible labor, unsupported luxury, convenient currency, or institutions that do not adapt to incentives.
  doNotUse:
    - The request is mostly about religion, cultural values, or political legitimacy without resource, labor, exchange, taxation, class, or market pressure.
    - The author only needs decorative trade terms, coin names, or flavor text with no systemic implications.
    - The author wants prose-level line editing rather than economic design or continuity checking.
    - The author wants final canon answers without exploratory modeling or author confirmation.
  status: optional
  source:
    - docs > skills > suggested > fiction > worldbuilding > economic-systems > SKILL.md
  scriptStatus: no-source-helpers
---

# $economic-systems

Economic worldbuilding for turning resources, labor, exchange, institutions, and scarcity into coherent story pressure rather than decorative backdrop.

## Edda Workflow

1. Read the relevant project context before modeling: use `project_map` to locate economic, geographic, political, technological, class, faction, and species material; use `read_story_bible_entry`, `read_entry_section`, `read_chapter`, or `read_content` for the specific canon and scenes involved. Use `search_content` for prior mentions of currency, debt, trade routes, taxes, rationing, markets, guilds, labor, slavery, automation, magic resources, post-scarcity claims, or black markets.
2. Separate confirmed canon from inference. Treat new currencies, institutions, taxes, class facts, trade routes, resource monopolies, historical crises, and market rules as proposals until the author confirms them.
3. Establish the resource foundation: identify critical resources such as food, water, energy, land, transport capacity, materials, information, magic, artifacts, machine time, or safe habitat. Note whether each resource is concentrated, dispersed, mobile, renewable, depletable, seasonal, artificially rationed, or treated as abundant.
4. Map resource flows from extraction or production to storage, transport, sale, taxation, consumption, waste, and theft. For each major resource, name who controls access, who does the work, who captures surplus, who bears risk, and where leakage, spoilage, corruption, smuggling, or sabotage can occur.
5. Define labor and production. Specify whether work is organized through family labor, caste, guilds, wage labor, debt bondage, slavery, communal obligation, state assignment, religious duty, automation, magic, corporate control, or mixed systems. Account for visible and invisible labor, maintenance, food production, care work, cleaning, transport, enforcement, and record keeping.
6. Define exchange and markets. Decide whether value moves through barter, gift exchange, prestige exchange, commodity money, state currency, bank credit, guild scrip, digital credit, reputation, ration tokens, favors, or direct allocation. Name the issuer or authority, backing or trust basis, units, divisibility, counterfeiting risk, exchange rates, and what happens when authority weakens.
7. Map trade and market friction. Identify local, regional, imperial, interstellar, magical, virtual, or border markets; trade routes; hubs; tolls; tariffs; piracy; banditry; spoilage; preservation; transport cost; regulation; embargo; quarantine; information lag; and who profits from controlling chokepoints.
8. Model distribution, class, and taxation. Identify wealth distribution, inheritance, rent, debt, tribute, tithes, tariffs, guild fees, licensing, rationing, charity, confiscation, theft, and state or temple redistribution. Connect access to class, status, species, citizenship, geography, credential, family, military service, caste, or faction membership.
9. Include informal economies. Every formal economy should have unofficial channels: smuggling, black markets, gray markets, barter networks, household production, patronage, corruption, protection rackets, counterfeit goods, tax evasion, forbidden magic, salvage, or mutual aid. State why people use them and what risks they accept.
10. Trace incentives and institutional adaptation. Ask what each major actor is rewarded for doing: rulers, tax collectors, guilds, merchants, soldiers, landlords, workers, debtors, smugglers, priests, corporations, automated systems, and outsiders. Then show how institutions adapt through regulation, evasion, innovation, collapse, rationing, enforcement, new monopolies, or crisis bargains.
11. Choose or blend the system typology that best fits the setting. Useful starting types include subsistence, pastoral, gift, prestige, ceremonial, palace, feudal, command, war economy, mercantile capitalism, industrial capitalism, financial capitalism, mixed economy, network capitalism, commons-based, cooperative, reputation economy, automated production, salvage economy, and post-scarcity with remaining scarce exceptions.
12. Adapt to genre constraints. For fantasy, account for magical resources, guild knowledge, divine limits, cross-species exchange, artifact markets, and monster-part extraction. For science fiction, account for automation, closed habitats, FTL or communication costs, alien markets, consciousness-as-resource, longevity, and non-replicable scarcity. For post-apocalyptic worlds, account for salvage, enclave production, food and water, lost knowledge, and repair capacity.
13. Convert economy into plot pressure. Identify what the system makes hard, expensive, illegal, shameful, prestigious, rationed, or dangerous. Tie those pressures to character choices, debts, jobs, family obligations, betrayals, migration, crime, faction leverage, class resentment, economic crises, and revelations.
14. Run consistency checks before presenting conclusions:
    - Resource coverage: key resources have sources, constraints, controllers, and users.
    - Labor visibility: basic work exists and explains elite leisure, military capacity, magic use, or technological abundance.
    - Technology match: currency, credit, markets, and record keeping fit transport, communication, law, and enforcement capacity.
    - Trade friction: bulky goods, perishables, borders, distance, risk, and tolls affect what can move.
    - Institutional embeddedness: markets obey cultural, political, religious, legal, or technological constraints.
    - Scarcity logic: even post-scarcity settings preserve scarcity in status, access, authenticity, trust, space, time, attention, energy, rare materials, or political permission.
    - Formal-informal balance: official rules create incentives for unofficial channels.
    - Class consequence: extraction, redistribution, inheritance, debt, taxation, and access rules produce believable winners and losers.
    - Adaptation over time: crises, monopolies, shortages, innovations, corruption, and reforms leave historical traces.
    - Story fit: the level of economic detail supports the author's genre and scene needs instead of overwhelming them.
15. When diagnosing an existing draft or canon entry, report contradictions as risks with evidence. Prefer canon-safe repairs that preserve established facts by adding missing constraints, local exceptions, hidden informal systems, institutional incentives, or historical transitions.

## Edda Output Handling

- Use chat for short coaching, exploratory models, option comparison, or questions the author must answer before canon is changed.
- Use an Attached Note when the economic analysis belongs to one chapter, scene, selection, faction conflict, trade incident, debt, tax dispute, or local market.
- Use a Project Note for durable economic models, parameter choices, trade maps in prose, class structures, currency notes, crisis histories, shadow-economy maps, and cross-chapter consistency checks.
- Use a Story Bible proposal when the output would add or change durable canon: currencies, resources, trade routes, taxes, laws, institutions, class facts, labor systems, economic history, market rules, technologies, magical constraints, faction assets, or regional dependencies.
- Do not update Story Bible canon directly unless the author explicitly asks for an applied change. Keep proposed facts labeled as proposals, assumptions, or open questions.
- Do not use Structured Writes in this skill unless a future runtime explicitly provides an economy schema; otherwise return readable prose with clear headings and bullets.

## Good Output Criteria

Good economic output names concrete flows, actors, constraints, incentives, and consequences. It explains who produces necessities, who controls exchange, who pays or avoids taxes, who is excluded, how informal systems arise, and how the economy pressures plot or character decisions.

Weak output only invents coin names, exotic goods, or vague trade flavor; treats one resource as the entire economy; ignores labor and maintenance; gives modern financial complexity to institutions that cannot support it; lets goods move without cost or risk; or confirms new canon without author approval.

## Script Compatibility

This source skill has no required helper scripts. Its useful decision logic has been converted into Edda-native methodology, consistency checks, and reviewable canon-proposal policy.
