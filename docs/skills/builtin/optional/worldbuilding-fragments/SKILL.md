---
name: worldbuilding-fragments
description: Create indirect lore fragments, in-world snippets, and epigraph-style material that imply systems, culture, and history without exposition dumps.
route:
  actionKinds:
    - chat
    - read_check
    - continuation
  contentKinds:
    - chapter
    - story_text
    - story_bible_entry
    - entry_section
    - project_note
    - attached_note
  tags:
    - fiction
    - worldbuilding
    - fragments
    - optional
  priority: 40
metadata:
  useCases:
    - The setting needs texture through snippets, traces, documents, sayings, rumors, labels, epigraphs, or found text rather than direct explanation.
    - A Chapter wants an oblique fragment that resonates with its conflict, system, theme, or hidden cost.
    - The author wants worldbuilding that rewards reader inference, negative space, and implied systems.
    - Existing lore feels like exposition and needs to be converted into indirect evidence inside the world.
  doNotUse:
    - The request is for direct canon definition, encyclopedia-style setting design, or explicit lore explanation.
    - The author needs a continuity audit or factual canon reconciliation rather than fragment drafting.
    - The fragment would need major Story Text insertion or durable canon changes that the author has not requested.
    - The user wants generic flavor text with no relationship to a scene, system, culture, institution, or reader inference goal.
  status: optional
  source:
    - docs > skills > suggested > fiction > worldbuilding > oblique-worldbuilding > SKILL.md
  scriptStatus: no-source-helpers
---

# $worldbuilding-fragments

Create indirect worldbuilding through fragments, traces, and documentary voices whose limitations reveal more than they explain.

## Edda Workflow

1. Read the relevant project context with Edda tools before drafting. Use `read_chapter` or `read_content` for the target passage, `read_story_bible_entry` or `read_entry_section` for established canon, and `search_content` when names, institutions, events, rules, or prior fragments may already exist.
2. Identify the hidden worldbuilding job. Name the system, pressure, institution, belief, taboo, historical wound, material constraint, or human cost the fragment should imply. Do not start by explaining lore; start by deciding what evidence a reader should infer from.
3. Separate established canon from proposal space. Treat any new durable fact about history, names, institutions, laws, magic/technology rules, religion, geography, timeline, or factions as a proposal unless it is already confirmed in project context.
4. Choose the fragment's distance from the scene:
   - First-order distance: same event from another angle. Use sparingly because it can become direct commentary.
   - Second-order distance: related phenomenon in another context. Prefer this for most fragments because it reveals systemic patterns while preserving inference.
   - Third-order distance: thematic or structural echo only. Use when the author wants subtle resonance and the reader can still discover the connection.
5. Choose a fragment source with a limited perspective. For documentary or quoted fragments, define the source's position, need, lens, and blindness:
   - Position: where the source sits relative to power, knowledge, danger, labor, faith, money, law, or the event.
   - Need: what the source must believe to keep status, sanity, identity, employment, innocence, faith, or control.
   - Lens: the framework that turns reality into professional, cultural, personal, legal, commercial, religious, scientific, or bureaucratic terms.
   - Blindness: what the source cannot afford to notice or acknowledge.
6. Pick a fragment form that naturally withholds explanation. Useful forms include epigraphs, marginalia, minutes, ledgers, inspection tags, placards, product copy, school exercises, folk sayings, legal clauses, prayers, recipes, memorial text, transit notices, bureaucratic memos, contracts, field notes, academic abstracts, merchant advertisements, repair logs, diary scraps, trial records, sermons, reviews, and redacted files.
7. Build the fragment from traces and negative space. Include one or two concrete details that only this source would care about, omit one thing the reader expects to be named, and let the omission reveal the culture's assumptions or denial.
8. Use one or more oblique connection types:
   - Systemic echo: the same force appears elsewhere at another scale.
   - Ironic juxtaposition: mundane, technical, celebratory, or procedural language sits beside an implied harm.
   - Thematic rhyme: a different situation repeats the same human pattern.
   - Causal chain: the fragment shows a distant cause, consequence, precursor, or afterimage.
9. Draft the fragment so it works on layered meaning:
   - Surface: it is credible and interesting as an in-world artifact.
   - Contextual: it resonates with the target scene or chapter.
   - Ironic: the source's blind spot creates meaning.
   - Thematic: it enlarges the story's concerns without explaining them.
10. Withhold explanation when explanation would flatten reader inference. Do not state the connection if the reader can infer it from juxtaposition, repeated terms, institutional patterns, cost, contradiction, or omission. Explain only enough for the author to evaluate the draft outside the story.
11. Calibrate after drafting. Aim for "discoverable but not obvious." If the fragment summarizes the scene, move it farther away. If no reader could infer the connection, add a concrete trace, repeated term, consequence, or clearer institutional frame.

## Fragment Methodology

Use indirect worldbuilding as evidence, not lecture. A strong fragment behaves like a shard of a larger system: it suggests a bureaucracy, economy, religion, technology, law, taboo, class structure, ecological condition, or historical trauma without naming the whole design.

When generating options, vary the axis of limitation rather than just the wording:

- Professional deformation: engineers see failure modes, lawyers see liability, priests see sin, merchants see margin, bureaucrats see compliance, scientists see measurements, teachers see curriculum, soldiers see chain of command.
- Positional necessity: managers must believe the system works, revolutionaries must believe change is possible, survivors must believe survival means something, perpetrators must justify themselves, historians must believe the past can be known, prophets must believe the future can be shaped.
- Cultural assumption: the fragment treats a false category, taboo, hierarchy, moral rule, origin myth, or social value as too obvious to defend.

Use fragments to imply systems through:

- Repeated forms: identical stamps, oaths, tariffs, permits, titles, slogans, product names, or ritual phrasing.
- Material traces: repairs, shortages, scars, substitutions, ration marks, obsolete parts, missing names, altered maps, memorial dates, revised forms.
- Institutional residues: audits, penalties, exemptions, euphemisms, seating charts, banned terms, training language, procurement requests.
- Negative space: the obvious question nobody asks, the victim no document recognizes, the cost treated as normal, the taboo hidden inside grammar.

## Anti-Patterns And Fixes

- Too direct: the fragment summarizes chapter events, names main characters unnecessarily, or explains the theme. Fix by using second-order distance and letting a parallel system carry the meaning.
- Too obscure: the fragment is pure flavor, depends on outside knowledge, or has no discernible resonance. Fix by adding a discoverable connection through shared pressure, consequence, image, vocabulary, or institution.
- Missing perspective: the voice could be anyone. Fix by defining position, need, lens, and blindness before revising the language.
- Exposition dump: the fragment teaches history, rules, or cosmology in neutral prose. Fix by turning information into a partial artifact made for a local purpose inside the world.
- Overexplanation: the fragment explains its own relevance or irony. Fix by removing the interpretive sentence and strengthening the artifact's concrete details.
- Voice inconsistency: the fragment sounds like the author rather than its source. Fix by matching the form's diction, omissions, jargon, incentives, and assumed audience.
- Canon sprawl: the fragment invents names, eras, rules, and institutions faster than the project can absorb them. Fix by marking new facts as proposals and preferring reusable hints over unnecessary proper nouns.

## Quality Checks

Before returning a fragment, test it against these criteria:

1. Perspective test: Can you state the source's position, need, lens, and blindness? Does the blindness create meaning?
2. Relevance test: Does the fragment connect to the chapter, scene, or project concern through system, irony, theme, or causality without summarizing it?
3. Inference test: What should the reader infer? Is that inference discoverable from the fragment's evidence?
4. Standalone test: Is the fragment credible and interesting as an in-world artifact even without the surrounding chapter?
5. Withholding test: Have you left enough unsaid for reader discovery, while avoiding random obscurity?
6. Canon test: Which details are confirmed, which are new proposals, and which should remain intentionally ambiguous?

## Edda Output Handling

- Return fragments in chat by default, with a brief craft note naming the perspective, distance, connection type, and any new canon proposals. Keep the craft note outside the story-facing fragment.
- Create an Attached Note when the fragment belongs to one chapter, scene, selection, or local revision problem.
- Create or update a Project Note when the author asks for a fragment bank, epigraph series, recurring document frame, list of implied systems, or cross-chapter worldbuilding plan.
- For Story Bible impact, propose changes separately from the fragment. Do not silently canonize new history, institutions, terminology, rules, timeline facts, names, geography, faction behavior, technology, magic, religion, or culture.
- Use Structured Writes or direct story insertion only when the author explicitly asks to place the fragment into Story Text. Preserve the surrounding prose unless the author asks for broader revision.
- When the author asks for multiple options, make each option differ by perspective, document form, distance, or implied system. Do not return near-duplicate phrasings of the same artifact.

## Canon-Safe Fragment Handling

Treat fragments as either confirmed, proposed, or intentionally ambiguous:

- Confirmed fragment: uses only established canon and can be inserted or saved without changing durable facts.
- Proposal fragment: introduces a new institution, rule, historical event, term, material constraint, faction practice, or cultural assumption. Label the new canon claims clearly and ask for confirmation before Story Bible updates.
- Ambiguous fragment: uses suggestive traces without forcing a single canon answer. Use this when mystery, myth, propaganda, unreliable records, or reader inference matters.

Prefer proposal language such as "This would imply..." or "New canon proposal:" in author-facing notes. Do not put proposal labels inside the story-facing artifact unless the label is part of the artifact itself.

## Script Compatibility

This source skill has no helper scripts. The useful source behavior is converted into Edda-native guidance for context reading, fragment drafting, output routing, and canon-safe proposal handling.
