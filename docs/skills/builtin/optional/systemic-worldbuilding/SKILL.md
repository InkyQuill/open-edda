---
name: systemic-worldbuilding
description: Trace speculative changes through feedback loops, institutions, ecology, economy, power, religion, and lived culture without locking unapproved canon.
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
    - systems
    - canon-safe
    - optional
  priority: 50
metadata:
  useCases:
    - A speculative technology, species, historical divergence, alternate physics, magic rule, social innovation, ecology, or habitat constraint needs believable cascading consequences.
    - The author wants second-order and third-order effects across economy, institutions, power, religion, ecology, daily life, language, resistance, and culture.
    - The setting feels like a cool premise, surface aesthetic, or isolated lore fact instead of a lived system with winners, losers, feedback loops, and contradictions.
    - A closed-loop ship, station, habitat, or recycled-matter society needs metabolic culture logic for identity, boundaries, death, power, and integration.
  doNotUse:
    - The request is only for isolated flavor text, names, or encyclopedia lore without system consequences.
    - The author wants sentence-level prose revision rather than world-logic diagnosis or proposal.
    - The world problem is narrowly about one already chosen subsystem, such as only a currency, only a religion, or only a language, and a specialized skill should handle it.
    - The author wants durable Story Bible canon changed immediately without a reviewable proposal.
  status: optional
  source:
    - docs > skills > suggested > fiction > worldbuilding > systemic-worldbuilding > SKILL.md
    - docs > skills > suggested > fiction > worldbuilding > metabolic-cultures > SKILL.md
  scriptStatus: no-source-helpers
---

# $systemic-worldbuilding

Build speculative settings by tracing how one change alters systems, creates feedback loops, and becomes visible in institutions, culture, conflict, and ordinary life.

## Edda Workflow

1. Establish context before inventing. Use `project_map`, `read_content`, `read_chapter`, `read_story_bible_entry`, `read_entry_section`, and `search_content` as needed to separate confirmed canon from draft notes, open questions, and inferred possibilities.
2. Define the initial divergence or pressure as a concrete cause: speculative technology, alternate history, alternate physics or magic, species biology, disease, social innovation, ecological constraint, economic rule, or closed-loop habitat condition. If the cause is vague, ask one clarifying question or offer 2-3 sharply different formulations.
3. Trace consequences by order, not by topic list alone:
   - First order: immediate applications, constraints, market or social responses, obsolete systems, early unintended effects, and who gains immediate advantage.
   - Second order: institutional adaptations, infrastructure, regulation, enforcement, criminal or underground uses, resistance movements, exploitation patterns, and shifts in wealth or power.
   - Third order: cultural normalization, language and metaphors, education, art, ethics, taboos, religious or philosophical reinterpretation, status symbols, and changed assumptions about identity or personhood.
4. Cover the major domains that the source change plausibly touches: ecology and resource flows, economy and labor, governance and law, power and hierarchy, religion and belief, family and kinship, military or security, infrastructure, education, medicine or biology, geography, and daily behavior. Do not force every domain equally; explain which are central, secondary, or out of scope for the current story.
5. Model feedback loops. Identify reinforcing loops where a consequence strengthens its own cause, balancing loops where institutions or scarcity push back, and delayed loops where effects surface years or generations later. Mark at least one plausible unintended consequence.
6. Use intersections to prevent homogeneous response. Compare effects across class, region, generation, profession, ideology, gender or kinship role when relevant, marginalized groups, and rival states or cultures. For each major benefit, name who pays the cost; for each harm, name who can profit from it.
7. Find contradictions instead of smoothing them away. Good systemic worldbuilding usually has mixed effects: liberation paired with dependency, abundance paired with control, safety paired with surveillance, purity paired with stagnation, or openness paired with exploitation.
8. Convert abstractions into visibility markers: physical markers, rituals, architecture, clothing or tools, changed habits, slang, etiquette, taboos, professions, bureaucratic forms, criminal practices, and other details a viewpoint character can notice without exposition.
9. For story usefulness, identify pressure points where systems collide: regulation versus black markets, faith versus technical reality, family loyalty versus institutional rules, ecological limits versus growth, class mobility versus gatekeeping, or individual desire versus communal survival.
10. Keep canon safe. Present new institutions, terms, histories, rules, conflicts, and cultural practices as proposals unless they are already confirmed. Label contradictions with existing canon and offer retcons or alternatives instead of silently overwriting facts.

## Metabolic Culture Model

Use this model when the setting involves closed-loop life support, space habitats, generation ships, sealed ecologies, recycling of air/water/biomass, or any culture where matter exchange shapes identity.

1. Start from the premise that matter is social reality, not metaphor. Ask how recycled air, water, food, waste, bodies, and biomass alter kinship, citizenship, taboo, death, property, and belonging.
2. Rate or position the culture on five axes:
   - Integration philosophy: purist tracking, deliberate synthesis, pragmatic management, or deliberate forgetting.
   - Temporal dynamics: how hours, months, years, decades, and post-death recycling change status and rights.
   - Boundary management: airlocks, quarantine, visitor protocols, trade rules, mixed gatherings, and interface technologies.
   - Death and continuity: return obligations, distribution rights, absent-dead problems, preservation taboos, and claims over the deceased's matter.
   - Power structures: generational depth, technical control of life support, economic interface control, religious interpretation, and practical operational authority.
3. Derive practical expressions from those axes: breathing etiquette, meal customs, sleep arrangements, work segregation, birth and coming-of-age rites, partnership norms, visitor rules, contamination response, shortage response, system-failure rituals, language, insults, honors, tools, and social classes.
4. Build at least three internal positions inside the culture: orthodox keepers, practical adapters, and reformist challengers. A culture that agrees with itself is usually underdeveloped.
5. When multiple habitats or cultures interact, compare axes for harmony points, friction points, translation needs, exploitation opportunities, and alliance or merger pressure.

## Pressure Tests

Use these tests before presenting the answer as strong worldbuilding:

- Consequence depth: At least one important chain reaches third order, and the answer does not stop at the first cool application.
- Domain coverage: Economy, institutions, power, belief, ecology or resource flow, and daily life were considered, even if some were intentionally deprioritized.
- Feedback loops: The map includes reinforcing, balancing, or delayed effects rather than a one-way list.
- Mixed incentives: Every major change has beneficiaries, casualties, opportunists, resisters, and people forced into adaptation.
- Intersection variation: Different classes, regions, generations, institutions, or cultures respond differently.
- Visibility: The output includes concrete signs a character can observe on the page.
- Canon safety: New facts are marked as proposals, open questions, or conflicts with existing canon.
- Metabolic specificity, when applicable: Closed-loop culture treats matter cycling as concrete social infrastructure, not just symbolism.

Avoid these failure modes: cool-tech trap, monoconsequence thinking, everyone responding the same way, one isolated domain, no visible daily-life markers, unchanged planetary assumptions in closed-loop habitats, metaphor without consequences, tension-free utopias, uniform populations, and binary integrated/not-integrated culture.

## Edda Output Handling

- Use chat for short diagnosis, option comparison, or brainstorming.
- Create an Attached Note when the analysis belongs to one chapter, scene, selection, or immediate plot problem.
- Create or update a Project Note when the output is a durable consequence map, pressure-test report, option bank, or cross-chapter worldbuilding plan.
- Propose Story Bible updates only after author confirmation or when explicitly asked for a proposal. Keep proposed institutions, terminology, histories, religious interpretations, ecological rules, economic systems, power structures, and metabolic customs separate from confirmed canon.
- For rewrites, change prose only when the user explicitly asks for applied revision. Otherwise return the systemic diagnosis and canon-safe proposals.

## Script Compatibility

The source skills are text-first and have no required runnable helpers. Their useful logic has been converted into Edda-native workflow, pressure tests, and canon-safe output rules; no `skill_script` call is needed.
