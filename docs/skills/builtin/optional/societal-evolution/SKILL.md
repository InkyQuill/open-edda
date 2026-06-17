---
name: societal-evolution
description: Multi-generational civilization design for compounding adaptations, institutional drift, and long-horizon social conflict.
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
    - evolution
    - optional
  priority: 42
metadata:
  useCases:
    - A society, colony, migration, diaspora, or civilization must change across multiple generations.
    - The author wants environment, technology, scarcity, hazards, or settlement conditions to reshape institutions, values, bodies, identities, or conflict.
    - Present-day world logic feels too close to baseline humanity, contemporary culture, or first-generation assumptions.
    - The task asks for multi-order consequences, winners and losers, unintended consequences, adaptation paths, or feedback loops over time.
  doNotUse:
    - The story only needs present-tense cultural texture.
    - The request is mainly about one religion, economy, faction, or city without long time depth.
    - The author wants quick lore instead of generational causation.
    - The task is line editing, scene pacing, or character voice rather than worldbuilding causality.
  status: optional
  source:
    - docs > skills > suggested > fiction > application > multi-order-evolution > SKILL.md
  scriptStatus: no-source-helpers
---

# $societal-evolution

Generational worldbuilding for making a civilization's present-day values, institutions, adaptations, and conflicts emerge from compounding pressure across time.

## Edda Workflow

1. Read the relevant project context before designing or diagnosing the society:
   - Use `project_map` to locate setting, timeline, faction, technology, ecology, economy, government, migration, and history material.
   - Use `read_story_bible_entry`, `read_entry_section`, `read_chapter`, or `read_content` for the target society and any neighbors, parent cultures, founding conditions, or later-era scenes.
   - Use `search_content` for named hazards, resources, institutions, conflicts, or prior evolution notes when the user references them but does not provide a target.
   - Treat Story Bible facts and established Story Text as existing canon; treat inferred implications as inference; treat all new timeline, institution, identity, technology, and history changes as proposals until the author confirms them.
2. Establish the foundation layer that creates evolutionary pressure:
   - Physical and ecological pressures: resource abundance or scarcity, climate, radiation, gravity, atmosphere, pathogens, predators, geography, closed-system limits, distance, and other survival constraints.
   - Technological pressures: required life-support, transport, communication, modification, automation, agriculture, extraction, medicine, or infrastructure just to remain viable.
   - Initial human conditions: population mix, expertise, demographics, founding cultures, imported beliefs, founding trauma, governance model, economic system, dependencies, and communication with outside powers.
3. Build a causal chain through societal time scales instead of jumping straight to the endpoint:
   - First order, 1-2 generations: immediate adaptations in bodies, infrastructure, resource use, status markers, social hierarchy, customs, and self-description.
   - Second order, 3-5 generations: new pressures created by those adaptations, including governance changes, economic shifts, family and community forms, education, knowledge transmission, vocabulary, taboo, sacred values, and status reconfiguration.
   - Third order, 6+ generations: deep structural change where property, personhood, time, kinship, work, risk, authority, purity, citizenship, or memory may mean something different from the founding culture.
   - Later orders when the setting demands it: mythologized origins, institutional capture, schism, reform, stagnation, collapse, diaspora, recombination, or renewed environmental pressure.
4. Track every stage across the dimensional grid:
   - Physical: bodies, health, built environment, tools, infrastructure, and ecological constraints.
   - Economic: resources, labor, exchange, scarcity, ownership, debt, surplus, and what counts as value.
   - Political: authority, legitimacy, coercion, law, representation, succession, sovereignty, and power projection.
   - Social: family, kinship, class, status, education, community membership, obligation, and exclusion.
   - Cultural: rituals, beliefs, aesthetics, virtues, taboos, holidays, shame, honor, and inherited narratives.
   - Linguistic: new vocabulary, dead metaphors, changed titles, translated concepts, and old words with new functions.
   - Identity: group belonging, personhood, ancestry, citizenship, profession, modification status, purity, exile, and mixed identities.
5. Apply the source principles while designing consequences:
   - Compounding divergence: each generation should inherit earlier adaptations and create non-linear downstream effects.
   - Environmental causation: physical and technological realities should shape social structures, which then shape values.
   - Functional drift: preserve some old terms, rituals, offices, or institutions while changing what they do.
   - Identity reformation: show how new challenges make old affiliations less central or transform their meaning.
6. Stress-test feedback loops, winners, losers, and unintended consequences:
   - Identify who gains status, safety, wealth, political leverage, mobility, reproduction, knowledge access, or moral authority from each adaptation.
   - Identify who loses those things, who is excluded, and which groups become interpreters, outcasts, brokers, rebels, priestly operators, regulators, or inherited underclasses.
   - Trace feedback loops where institutions intensify the original pressure, reduce it, exploit it, hide it, ritualize it, or create dependency on the adaptation.
   - Include unintended consequences: ecological damage, resource traps, institutional corruption, ideological hardening, medical side effects, succession crises, language barriers, intergenerational resentment, or brittle infrastructure.
7. Generate conflict and interaction only after the evolution path is coherent:
   - Compare the evolved society with parent cultures, neighbors, rivals, colonies, diasporas, or hybrids.
   - Mark harmony points where similar adaptations support alliance, merger, trade, or shared interpretation.
   - Mark friction points where values, resources, conceptual frameworks, or power projection methods clash.
   - Mark translation needs where different but compatible systems require interpreters, brokers, specialists, rituals, or legal fictions.
   - Derive story seeds from generational conflict, inter-civilizational collision, internal subgroup divergence, hybrid identity, regression movements, reform attempts, exile, or contact between first-order and third-order populations.
8. Keep the proposal canon-safe:
   - Separate confirmed facts, evidence from text, inference, and proposal labels in the response.
   - Do not overwrite established canon to make the evolution cleaner; instead, name the tension and offer canon-compatible options.
   - When a proposal would change durable lore, timeline, institutions, names, identities, political history, ecology, economics, or technology, present it as a Story Bible proposal and ask for author confirmation before treating it as confirmed canon.
   - When the author asks for applied writing, provide concise proposal text or a localized rewrite only after the target content and current revision are known.

## Quality Criteria

- Each stage must visibly build from the foundation layer and prior stage, not from arbitrary aesthetic preference.
- The present-day society should differ from baseline assumptions in institutions, economy, politics, culture, language, identity, and conflict logic, not only in surface customs.
- Adaptations should create new pressures; solved problems should still leave costs, dependencies, losers, or distorted incentives.
- Institutions and inherited terms should show functional drift where useful: an old office, ritual, family term, currency, law, or title may survive with a changed role.
- External conflicts should come from evolved incompatibilities, resource needs, translation failures, or power projection friction rather than generic hostility.
- Canon-safe outputs should preserve existing facts, state assumptions explicitly, and offer options when the source material leaves time scales or causation open.

## Edda Output Handling

- Return the evolution model, diagnosis, or option set in chat when the author is deciding.
- Create an Attached Note when the work serves one chapter, selected passage, scene, era snapshot, or local continuity issue.
- Create a Project Note for a durable civilization model, multi-era timeline, dimensional grid, conflict map, or interaction-zone plan.
- Use a Story Bible proposal for durable changes to societies, institutions, ecological conditions, economies, political systems, cultural values, language, identities, technology, history, or timelines. Do not present proposals as confirmed canon until the author approves them.
- Use Structured Writes only when the author explicitly asks to apply text, the target content and current revision are known, and the requested action is a localized rewrite or entry update.
- For compact outputs, include only the needed sections. For full designs, use this order: foundation layer, staged evolution, dimensional grid, feedback loops and winners/losers, interaction zones, story seeds, canon proposals or open questions.

## Script Compatibility

The source skill has no helper scripts. Its useful logic has been converted into Edda-native workflow, quality criteria, and output policy; no `skill_script` call is required.
