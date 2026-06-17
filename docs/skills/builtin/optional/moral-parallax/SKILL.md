---
name: moral-parallax
description: Build stories that collapse the comfortable distance between harm and the people who benefit from it.
route:
  actionKinds:
    - chat
    - read_check
    - continuation
  contentKinds:
    - chapter
    - story_text
    - attached_note
    - project_note
    - story_bible_entry
    - entry_section
  tags:
    - fiction
    - theme
    - systems
    - optional
  priority: 40
metadata:
  useCases:
    - A story is about systemic harm, exported consequences, or complicity.
    - The author wants a cleaner way to design moral distance and then collapse it.
    - The setting needs a mechanism that makes invisible costs personal.
  doNotUse:
    - The story depends on clear moral innocence or simple villainy.
    - The request is mainly about ordinary character denial without systemic distance.
    - The author wants a neat, consequence-free resolution.
  status: optional
  source:
    - docs > skills > suggested > fiction > structure > moral-parallax > SKILL.md
  scriptStatus: no-source-helpers
---

# $moral-parallax

Speculative story design for collapsing the comfortable distance between action and consequence while preserving competing moral frames until the story has earned its thematic argument.

## Edda Workflow

1. Read the target Story Text, attached notes, relevant Project Notes, and any Story Bible entries for world rules, institutions, factions, character histories, or prior thematic decisions. Use `project_map` and `search_content` first when the affected system or prior work is unclear.
2. Separate confirmed canon from inference. Treat new institutions, magic rules, technology, historical causes, faction motives, or character facts as proposals until the author confirms them.
3. Identify the comfortable moral fiction the story currently maintains. Classify the distance being collapsed:
   - Temporal: harm is assigned to the future or buried in the past.
   - Spatial: harm is treated as happening elsewhere.
   - Social: harm is assigned to people unlike the protagonist's group.
   - Causal: the protagonist's choices seem too indirect to count.
   - Informational: ignorance is treated as innocence.
4. Choose the engine that makes the distance concrete:
   - Exchange: every benefit has a cost paid by someone kept out of view.
   - Accumulation: individually tolerable acts combine into systemic damage.
   - Cascade: a bounded action propagates through hidden networks.
   - Inheritance: the present carries debts, tools, offices, or harms created by predecessors.
5. Build conflicting moral frames before judging them. For each major POV or faction, state what they value, what evidence they can see, what harm they minimize, what they fear losing, and why their position would feel justified from inside their frame.
6. Create parallax between POVs. Make the same action, institution, or compromise read differently from different proximities: beneficiary, victim, maintainer, inheritor, bystander, and opponent. Do not collapse the story into one authorial answer before the reader has felt why each frame can hold.
7. Design the parallax event that reveals true proximity. The revelation must be specific and embodied: a named person, relationship, place, debt, wound, memory, record, or consequence that makes abstraction impossible.
8. Map ethical pressure after the revelation. Block clean exits, require costs for every available response, and make opposition justified rather than merely obstructive. Strong options include sacrifice of inner circle for outer circle, complicity accepted to reduce harm, system destruction at personal cost, system maintenance with full knowledge, or a least-harmful path that still leaves damage.
9. Test the thematic argument. The story should argue through costs and changed choices, not through a lecture. Ask what becomes impossible to believe after the collapse, what burden of knowledge remains, and what compromise the protagonist must now carry.
10. For revision requests, mark exact scene-level interventions: where to seed the comfortable fiction, where a POV shift changes moral distance, where pressure escalates, where the opposition's case must strengthen, where the cost lands, and where the story must resist premature certainty.

## Scene Tests

- Distance test: Can the scene name which temporal, spatial, social, causal, or informational distance lets a character feel clean?
- Parallax test: Would the same event look morally different from another POV with different proximity to the cost?
- Complicity test: Is the protagonist or focal group already benefiting, maintaining, inheriting, or avoiding knowledge before the revelation?
- Opposition test: Can the antagonist, institution, or resistant character make a coherent case without sounding like a straw villain?
- Cost test: Does every serious choice preserve or create harm somewhere, even when it is the least harmful option?
- Revelation test: Is the collapse concrete enough that a character cannot retreat to statistics, slogans, or distant sympathy?
- Timing test: Does the story withhold a single authorial answer long enough for the conflicting frames to matter?
- Canon test: Are new world mechanisms or historical claims labeled as proposals unless already established?

## Anti-Patterns

- Morality play collapse: the reveal proves simple good people against evil exploiters. Fix by making ordinary, sympathetic, reasonable choices part of the harm and by implicating the protagonist's comfort.
- Clean resolution: the protagonist exposes, reforms, or escapes the system into innocence. Fix by leaving an ongoing burden, a compromised role, or a pyrrhic improvement.
- Symbolic collapse only: the character understands the issue intellectually. Fix by making the harm named, local, relational, and irreversible enough to change behavior.
- Single-axis collapse: only one distance falls while others still excuse the character. Fix by layering distances, such as spatial proximity revealing temporal debt or social proximity exposing causal responsibility.
- Protagonist exceptionalism: the protagonist is uniquely moral or perceptive. Fix by making discovery circumstantial and their prior blindness ordinary.
- Premature answer: the narration announces the correct moral reading before the competing frames have pressure. Fix by strengthening justified opposition and letting costs reveal the argument.
- Consequence-free canon invention: the agent adds systems, histories, or faction motives as fact. Fix by presenting canon-affecting material as a proposal with evidence and open questions.

## Edda Output Handling

- Return compact diagnosis, design options, or scene tests in chat when the author is deciding.
- Create an Attached Note when the analysis belongs to one Chapter, scene, selection, or local revision problem.
- Create or update a Project Note when the author wants a durable parallax map, thematic argument, POV matrix, opposition map, or revision plan across multiple scenes.
- Propose Story Bible updates only for author-confirmed durable changes to world mechanisms, institutions, history, factions, character facts, or setting rules. Until then, label them `Canon proposal`, `Open question`, or `Non-canon option`.
- Use Structured Writes or direct rewrite tools only when the author explicitly asks to draft, continue, or replace Story Text. Preserve the chosen POV's moral frame in prose rather than explaining the framework by name.

## Script Compatibility

This source skill has no helper scripts. Its useful logic is converted into Edda-native context reading, moral-distance diagnosis, POV-frame mapping, scene tests, anti-pattern checks, and canon-safe proposal handling.
