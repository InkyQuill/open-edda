---
name: settlement-design
description: Design towns, cities, and stations whose layout, history, and social geography support story instead of feeling map-first.
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
    - place-design
    - optional
  priority: 44
metadata:
  useCases:
    - A city, town, station, colony, or village needs believable layout and development.
    - Scene locations should express class, power, infrastructure, or history.
    - The author wants place logic before drafting more scenes there.
  doNotUse:
    - The request is only for a quick atmospheric detail.
    - The work is really about broad government or economy rather than one place.
    - The author wants fixed canon without exploratory place design.
  status: optional
  source:
    - docs > skills > suggested > fiction > worldbuilding > settlement-design > SKILL.md
  scriptStatus: no-source-helpers
---

# $settlement-design

Settlement design for authors who want places to grow from geography, history, labor, and status rather than existing only as convenient sets.

## Edda Workflow

1. Read the relevant Story Bible entries, Project Notes, Writing Briefs, and any Chapters already set in or near the settlement. Separate confirmed facts from inferences and open questions before proposing new geography, history, names, institutions, or districts.
2. Establish site logic. Explain why the settlement exists at this location: water access, food supply, resource proximity, defensive position, trade-route placement, climate suitability, sacred value, orbital/technological constraints, magical constraints, or post-disaster survivability. If the site has no reason to exist, propose one or flag the weakness.
3. Define the settlement's primary function and network role. Classify it as market settlement, administrative center, religious complex, military outpost, production center, agricultural community, transport hub, knowledge center, frontier camp, station, colony, ruin, or mixed form. Tie that function to nearby settlements, roads, ports, rivers, passes, gates, signal systems, routes, or political borders.
4. Choose morphology from cause, not decoration. Use grid, radial, organic, hierarchical, linear, concentric, dispersed, nucleated, or composite patterns only when the geography, founding authority, technology, defense needs, or later growth history justifies them.
5. Build the life-support systems before the scenic districts. Account for water supply, drainage, waste, food storage and distribution, fuel or energy, transportation, communications, markets, warehouses, repair capacity, and emergency access. Let these systems shape gates, bridges, canals, docks, wells, cisterns, granaries, power conduits, airlocks, landing bays, or service corridors.
6. Place governance and defenses in physical form. Identify councils, courts, archives, tax offices, patrol posts, barracks, walls, checkpoints, towers, natural barriers, escape routes, shelters, surveillance, magical wards, security zones, or habitat-control rooms. Show who can enter these spaces and who is kept out.
7. Design districts as social geography. Include elite spaces, common quarters, marginalized zones, transitional spaces, contested territories, ethnoreligious or species-based districts, occupational clusters, commercial areas, industrial areas, religious areas, administrative areas, entertainment areas, and military areas as needed. Use class layout, access, pollution, safety, views, elevation, age, crowding, materials, and distance from power to make differences visible.
8. Layer growth history. Trace the place from foundation through expansion, infill, leap-frog growth, satellite formation, annexation, rebuilding, decline, or repurposing. Add at least one historical scar or adaptation unless canon forbids it: fire, flood, disease, siege, invasion, resource depletion, overcrowding, political collapse, contamination, technological regression, magical accident, or habitat failure.
9. Anchor daily life. Describe how residents get water, food, work, worship, messages, medical help, schooling, entertainment, transport, privacy, safety, and waste removal. Check how daily routines differ for elites, workers, outsiders, children, migrants, soldiers, clergy, criminals, or nonhuman residents.
10. Turn place logic into story pressure. Identify movement constraints, access barriers, jurisdictional boundaries, class friction, supply bottlenecks, environmental hazards, contested spaces, surveillance blind spots, symbolic routes, memory sites, and institutions that can create scenes without making every feature serve the current plot.
11. Check failure modes before presenting the design. Reject designer's-map symmetry, functional perfection, scale mismatch, missing infrastructure, homogeneous neighborhoods, implausible defense, unsupported food/water/energy, isolated settlements with no trade network, districts that exist only for protagonists, and canon changes presented as fact.
12. Produce canon-safe proposals. Mark each new location, district name, institution, historical event, route, resource system, or demographic fact as a proposal until the author confirms it. When revising existing canon, state the conflict and offer compatible alternatives instead of overwriting established facts.

## Design Criteria

- Site logic is strong when geography, water, food, energy, defense, economy, belief, or technology makes the settlement necessary in that exact place.
- Morphology is strong when street patterns, density, boundaries, skyline, open spaces, and district placement reveal foundation, growth, disaster, power, and adaptation.
- Infrastructure is strong when water, waste, food, fuel or power, transport, storage, repair, and communication explain how the population survives each day.
- Social geography is strong when class, status, ethnicity, species, occupation, religion, migration, law, and exclusion are legible in space without turning the settlement into a single-note map.
- Historical layering is strong when old uses remain visible through ruins, reused buildings, widened streets, renamed wards, displaced walls, buried infrastructure, sacred sites, scars, or contested memories.
- Story fit is strong when the settlement creates plausible movement, pressure, access, concealment, obligation, and conflict while still feeling like it would exist without the protagonist.

## Output Shape

When designing or diagnosing a settlement, organize the answer around the parts the author needs, usually:

1. Existing canon and constraints.
2. Site rationale: geography, water, food, energy, resources, climate, defense, sacred or technical constraints.
3. Network role: trade routes, neighboring settlements, political borders, transport, communications, imports, exports, dependencies.
4. Morphology and growth history: original core, expansion pattern, boundaries, density, old and new layers, disaster responses, repurposed spaces.
5. Districts and class layout: elite, common, marginalized, contested, transitional, occupational, religious, administrative, military, industrial, commercial, and entertainment zones.
6. Infrastructure and daily life: survival systems, governance touchpoints, defenses, routines, bottlenecks, maintenance, smells, noise, crowding, hazards.
7. Story pressure: access problems, conflicts, symbolic routes, vulnerable systems, secrets, failure modes, scene opportunities.
8. Canon-safe proposals: proposed names, facts, history, unresolved questions, and alternatives that preserve established canon.

## Edda Output Handling

- Return diagnosis, options, and compact place models in chat when the author is deciding.
- Create an Attached Note when the work supports one Chapter, selection, scene route, district, or immediate location problem.
- Create or update a Project Note for a durable settlement design, cross-chapter location plan, district map in prose, route model, failure-mode list, or unresolved canon questions.
- Propose Story Bible updates only after the author confirms durable facts such as location names, district names, institutions, routes, resources, population scale, founding events, disasters, rulers, defenses, or infrastructure.
- Use Structured Writes only when the author explicitly asks to apply text to a known target and the current revision is available; otherwise provide proposal text for review.
- Never present new worldbuilding as confirmed canon. Label facts as existing canon, inference, proposal, or open question.

## Script Compatibility

This source skill has no required helper scripts. Its useful methodology is converted into Edda-native place design, infrastructure analysis, social-geography checks, failure-mode diagnosis, and canon-safe proposals.
