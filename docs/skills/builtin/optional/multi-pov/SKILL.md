---
name: multi-pov
description: Multi-POV structure diagnosis and planning for catalyst environments, knowledge asymmetry, POV function, sequencing, overlap, and distinct voices.
route:
  actionKinds:
    - chat
    - read_check
    - rewrite
    - continuation
  contentKinds:
    - chapter
    - story_text
    - story_bible_entry
    - entry_section
    - attached_note
    - project_note
  tags:
    - fiction
    - structure
    - pov
    - optional
  priority: 46
metadata:
  useCases:
    - A story needs a multi-POV constellation around a shared place, event, institution, crisis, or time-compressed situation.
    - The author wants to decide which characters deserve viewpoint space, how much weight each POV should carry, or whether to add or remove a POV.
    - Existing POV threads repeat the same information, blur together, intersect too conveniently, or fail to reframe one another.
    - A chapter, outline, or project note needs information-control mapping across viewpoint characters.
    - The author asks for sequencing, overlap, contradiction, or reader triangulation across multiple perspectives.
  doNotUse:
    - The story is clearly single-POV and the author is not exploring additional viewpoints.
    - The request is only sentence-level line editing inside one scene.
    - The task is physical setting design without viewpoint, information, or narrative-structure questions.
    - The task is an individual character arc with no multi-POV structure problem.
  status: optional
  source:
    - docs > skills > suggested > fiction > structure > perspectival-constellation > SKILL.md
  scriptStatus: no-source-helpers
---

# $multi-pov

Build, diagnose, or revise multi-POV fiction so every viewpoint has a distinct narrative job, controlled knowledge access, and a structurally organic relationship to the shared pressure.

## Edda Workflow

1. Read the relevant project context before judging the POV structure: use the current Chapter or Story Text, nearby Chapters when sequence matters, relevant Project Notes or Attached Notes, and any Story Bible entries for established characters, timeline, institutions, locations, and canon constraints.
2. Separate confirmed canon from inference. Treat new POV roles, hidden motives, timeline links, institutional facts, relationship changes, or setting rules as proposals until the author confirms them.
3. Identify the shared thread that makes the viewpoints belong in one constellation: place, event, institution, crisis, route, countdown, trial, emergency, ceremony, workplace, queue, journey, or other catalyst environment.
4. Test the catalyst for transformation pressure. A usable catalyst should create at least some of these conditions:
   - forced intimacy between strangers or unlikely combinations,
   - consequential stakes where choices have real cost,
   - temporal intensity that compresses normal behavior,
   - mask-dropping conditions where pretense becomes difficult,
   - high throughput or diverse entry paths,
   - variable exposure time,
   - asymmetric power, agency, or knowledge.
5. Choose the POV constellation model that best fits the material:
   - Iceberg: each POV reveals part of a larger hidden network.
   - Prism: the same facts refract into different genres, tones, values, or emotional textures.
   - Archaeological: each POV becomes a new layer that recontextualizes earlier assumptions.
6. For each candidate POV, map the following before recommending structure:
   - Access path: why this person naturally enters the catalyst.
   - Stakes: what they can gain, lose, protect, expose, or misunderstand.
   - Knowledge: what they know, cannot know, wrongly assume, overhear, conceal, or learn too late.
   - Narrative function: witness, driver, counterpoint, validator, contradiction, emotional center, institutional insider, outsider, antagonist, aftermath carrier, or reframing lens.
   - Transformation pressure: what forces this character out of ordinary patterns.
   - Voice markers: diction, sensory priorities, rhythm, attention habits, metaphors, social assumptions, professional vocabulary, and blind spots.
7. Check distribution of knowledge across POVs. Avoid giving characters information their position could not plausibly provide. Use awareness gradients deliberately: obliviousness, peripheral awareness, active investigation, insider knowledge, and meta-awareness can all coexist if the structure explains them.
8. Check contrast and contradiction. Strong POV sets do not merely repeat events; they create productive differences in values, access, emotional interpretation, genre texture, and reliability. Contradictions should either expose character limitation, reveal institutional asymmetry, or give the reader a solvable tension to triangulate.
9. Design intersections from catalyst logic rather than convenience. Ask who would naturally share a waiting room, process paperwork, control access, compete for resources, overhear fragments, witness consequences, inherit fallout, or misread the same signal.
10. Plan sequencing and overlap:
    - Use simultaneous POVs when different vantage points transform the same moment.
    - Use sequential handoffs when cause and effect should ripple across characters.
    - Use recursive returns when a later POV must reframe an earlier event.
    - Use overlap only when the second pass adds new stakes, contradiction, or reader understanding.
11. Control reader information. Decide what the reader knows before, during, and after each POV segment. Preserve suspense by delaying confirmation, create irony when the reader knows more than a character, and create mystery when each POV holds only a partial truth.
12. Evaluate whether to add, keep, reduce, or remove POVs:
    - Add a POV when it supplies structurally unavailable knowledge, a distinct pressure path, a necessary contradiction, or a story with its own beginning, middle, and end.
    - Keep a POV when it changes the reader's understanding of the catalyst or another viewpoint.
    - Reduce a POV when its function is useful but not dense enough for equal weight.
    - Remove or merge a POV when it repeats information, exists only to deliver plot mechanics, has no independent stakes, or blurs in voice and function with another viewpoint.
13. Diagnose common failure modes explicitly:
    - Forced intersection: connection exists because the plot needs it, not because the catalyst would produce it.
    - Equal weight assumption: all POVs receive the same space despite unequal narrative density.
    - Omniscient fog: characters know too much or too little for their position.
    - Plot-only connection: one POV exists only to service another character's arc.
    - Low-pressure catalyst: the shared setting gathers people but does not change, expose, endanger, or intensify them.
14. When revising prose, preserve the established facts and requested POV boundaries. Rewrite only the selected Story Text or Chapter area the author asked to change, and make voice, access, and information control visible in the prose.

## Edda Output Handling

- Use chat for short diagnosis, brainstorming, decision support, or a compact POV recommendation.
- Use an Attached Note when the analysis belongs to one Chapter, scene, selection, or local chapter cluster.
- Use a Project Note for a durable POV map, constellation plan, sequencing plan, information-control table, or cross-chapter revision checklist.
- Use a Story Bible proposal only when the POV plan would create or alter durable canon: character history, relationship facts, institutions, locations, timeline, names, rules, or hidden truths. Label each item as proposed, not confirmed.
- Use Structured Writes only when the author explicitly asks to draft, continue, or rewrite Story Text. Keep the output within the requested scope.
- A useful output should include the catalyst, POV roster, each POV's narrative function, knowledge distribution, intersection points, sequence or overlap plan, voice-distinction notes, add/remove/reduce recommendations, and canon proposals or open questions when relevant.

## Script Compatibility

This source skill has no helper scripts. Its operational method is converted into Edda-native reading, planning, diagnosis, canon-safe proposal, and optional rewrite guidance.
