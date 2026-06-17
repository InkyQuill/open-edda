---
name: emotional-beats
description: Shape scenes and outlines around the emotional moments readers should actually feel, not only plot turns.
route:
  actionKinds:
    - chat
    - read_check
    - rewrite
  contentKinds:
    - chapter
    - story_text
    - attached_note
    - project_note
  tags:
    - fiction
    - structure
    - emotion
    - optional
  priority: 54
metadata:
  useCases:
    - A story has plot but not enough emotional lift.
    - The author wants to design or revise key beats for awe, dread, intimacy, grief, triumph, or similar reader experience.
    - A scene needs stronger emotional escalation or payoff.
  doNotUse:
    - The request is mainly about lore systems or canon organization.
    - The author wants broad developmental diagnosis more than beat work.
    - The target text is not ready for emotional shaping yet.
  status: optional
  source:
    - docs > skills > suggested > fiction > structure > key-moments > SKILL.md
  scriptStatus: no-source-helpers
---

# $emotional-beats

Emotional beat shaping for authors who want scenes, chapters, or outlines to land as earned reader experience instead of mechanical event sequence.

## Edda Workflow

1. Read the target Chapter, Story Text, Attached Note, or Project Note in context. If the beat depends on established characters, relationship history, world rules, factions, timeline, or prior consequences, read the relevant Story Bible entries or earlier context before diagnosing.
2. Separate the plot event from the intended emotional experience. Name what the reader should feel at each important beat: awe, dread, curiosity, relief, grief, shame, intimacy, triumph, betrayal, wonder, pressure, or another precise effect.
3. Identify the key emotional moments that define the scene, chapter, sequence, or outline. Treat a key moment as a vivid, sceneable experience that changes character understanding, reader expectation, relationship state, world understanding, or the pressure on the protagonist.
4. Classify each key moment by its story function, not only by event:
   - Wonder: initial encounter, scale revelation, perspective shift, wonder escalation, transcendent integration.
   - Mystery: question inception, pattern recognition, false resolution, progressive revelation, solution crystallization.
   - Adventure: threshold crossing, capability test, resource depletion, ultimate challenge, return transformation.
   - Horror: wrongness glimpse, safety violation, threat escalation, failed solution, confrontation.
   - Thriller: stakes establishment, deadline imposition, near miss, option elimination, decision under duress.
   - Relationship: significant connection, intimacy deepening, value conflict, relationship crisis, reconciliation or resolution.
   - Drama: internal conflict revelation, external pressure point, failure moment, truth confrontation, character evolution.
   - Issue: perspective challenge, stake personalization, complexity recognition, position testing, perspective integration.
   - Ensemble: group formation, role establishment, group fracture, collective challenge, synergy moment.
5. Test setup and payoff. For every large beat, identify what has been planted before it, what the reader has been made to expect, what rule or vulnerability makes it possible, and what changes afterward. If the beat arrives without preparation, propose earlier setup. If setup exists without payoff, propose a later beat or remove the plant.
6. Check pressure. A key emotional moment should be forced by escalating constraints, not dropped into a neutral scene. Look for time pressure, social pressure, moral pressure, danger, scarcity, secrecy, competing obligations, exposed weakness, or irreversible choice. If pressure is flat, add or sharpen the constraint that makes the emotion necessary.
7. Check vulnerability. The beat should reveal what the character can lose, cannot control, wants too much, fears admitting, or has misunderstood. Emotional intensity is weak when the character is untouched by the event, protected from cost, or allowed to remain composed without a meaningful reason.
8. Check reversals. Strong beats often turn the scene: safety becomes danger, certainty becomes doubt, victory exposes a cost, intimacy creates risk, a clue misleads, a solution fails, a private truth becomes public, or a feared loss becomes a chosen sacrifice. If the beat only confirms the prior state, decide whether it needs a reversal, escalation, or aftermath instead.
9. Check scene placement. Place key moments where they can do structural work: opening beats create promise and questions, midpoint beats change understanding or available options, pre-climax beats strip resources or certainty, climactic beats force the defining choice, and aftermath beats let the consequence register. Move a beat if it currently arrives before the reader understands why it matters or after the pressure has already dissipated.
10. Test whether the beat is earned. A beat is earned when the story has supplied motive, pressure, setup, vulnerability, causality, and consequence. Flag beats that rely on coincidence, sudden confession, unexplained competence, unearned forgiveness, instant grief, arbitrary betrayal, or spectacle without personal stake.
11. Diagnose overplaying and underplaying:
    - Underplayed beats lack sensory focus, interior reaction, changed behavior, pressure, consequence, or enough page space for the reader to feel the turn.
    - Overplayed beats repeat the same emotion, explain what the reader already understands, force tears or declarations too early, or pause the story after the feeling has landed.
    - Recommend more detail, silence, action, restraint, compression, or displacement according to the actual failure.
12. Require aftermath for major beats. Identify what the beat changes in the next scene or chapter: decision, relationship, belief, resource, danger, public status, self-image, available options, or world state. If nothing changes, the beat is probably decorative.
13. Design connective tissue between key moments. Bridge scenes should do more than move characters between events: they should carry consequence, deepen character function, reveal relevant world information, create setup, increase pressure, or complicate the next beat.
14. When generating new beats, start from the desired emotional experience, then work backward to the world condition, character function, setup, vulnerability, pressure, reversal, and aftermath needed to make that experience possible.
15. Rewrite only when the author explicitly asks for applied beat changes. Otherwise return diagnosis, a beat map, alternatives, or a revision plan.

## Beat Map Criteria

For each key emotional moment, include enough of these fields to make the recommendation actionable:

- `Moment`: a short name for the beat.
- `Reader experience`: the specific feeling the beat is meant to create.
- `Story function`: the genre or arc function it serves.
- `Setup`: what must be planted before the beat lands.
- `Pressure`: what forces the beat now.
- `Vulnerability`: what the character risks, exposes, loses, or misunderstands.
- `Reversal`: what changes inside the beat.
- `Scene placement`: where the beat belongs and why.
- `Payoff`: what earlier promise, fear, clue, desire, or rule it answers.
- `Aftermath`: what changes because the beat happened.
- `Earnedness risk`: what currently feels too sudden, too easy, too loud, too quiet, or too disconnected.

## Edda Output Handling

- Return the beat map and recommendations in chat by default.
- Create an Attached Note when the beat work belongs to one Chapter or one selection.
- Create or update a Project Note when the author wants an emotional map across multiple Chapters.
- For cross-chapter planning, include the key emotional moments, setup/payoff chain, bridge scenes, pressure curve, vulnerability progression, reversal points, and aftermath obligations.
- If a beat requires new durable facts about character history, relationship status, institutions, world rules, factions, timeline, names, or prior events, separate those as Story Bible proposals. Do not treat proposed lore or backstory as confirmed canon until the author approves it.
- Use Structured Writes only when the author explicitly asks to rewrite selected Story Text for emotional effect.
- When rewriting, preserve confirmed canon and the existing narrative point of view unless the author explicitly asks to change them. Make the emotional change visible through scene action, perception, dialogue, silence, decision, and consequence rather than explanatory summary alone.

## Script Compatibility

This source skill has no required helper scripts. The rewrite works entirely through Edda-native beat analysis, planning, diagnosis, and author-approved revision.
