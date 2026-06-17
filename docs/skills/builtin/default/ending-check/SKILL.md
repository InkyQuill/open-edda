---
name: ending-check
description: Ending diagnosis for weak payoff, rushed aftermath, predictable resolutions, and climaxes that do not complete the story's promises.
route:
  actionKinds:
    - chat
    - read_check
  contentKinds:
    - chapter
    - story_text
    - project_note
    - attached_note
    - story_bible_entry
  tags:
    - fiction
    - endings
    - payoff
    - revision
  priority: 78
metadata:
  useCases:
    - The ending feels arbitrary, obvious, rushed, overexplained, or emotionally thin.
    - The author wants to know whether the climax and aftermath actually pay off the setup.
    - Beta feedback says the ending does not land.
  doNotUse:
    - The story is not far enough along to judge its ending.
    - The problem is mainly earlier pacing or drafting momentum.
    - The author wants the skill to write the ending for them by default.
  status: default
  source:
    - docs > skills > suggested > fiction > structure > endings > SKILL.md
  scriptStatus: source-helpers-converted-to-guidance
---

# $ending-check

Ending and payoff diagnosis for authors whose story builds energy but loses force, clarity, or satisfaction in the final stretch.

## Edda Workflow

1. Read the target ending, final chapter, or selected ending sequence. If the ending depends on earlier promises, use `project_map`, `search_content`, `read_content`, `read_chapter`, `read_story_bible_entry`, and `read_entry_section` to inspect the setup that feeds the climax.
2. Separate diagnosis from rewrite. By default, identify the ending problem, cite evidence, and recommend revision targets. Draft or apply replacement prose only after the author explicitly asks for rewrite work and the target content is clear.
3. Name the ending shape before judging it: closed, open, ambiguous, twist, circular, book-in-series, series finale, or eucatastrophe. Judge the ending by the promises of that shape, not by a generic demand for total closure.
4. Apply the core test: a strong ending feels inevitable because its seeds were planted, and surprising because the path or cost is not merely obvious. If it is only inevitable, it may be predictable. If it is only surprising, it may be arbitrary.
5. Diagnose the dominant failure state:
   - `E1 Arbitrary`: resolution does not follow from established character, plot, or world logic; readers cannot reread and see the seeds.
   - `E2 Predictable`: the destination and route are both obvious; genre expectations are met too literally with no surprising how, cost, or implication.
   - `E3 Unearned`: luck, coincidence, new power, late ally, or external rescue solves the problem without the protagonist's final choice mattering.
   - `E4 Expanding`: the ending raises major new mysteries, widens scope, or delays the answer to the central dramatic question when the story should contract toward clarity.
   - `E5 Overexplained`: the ending states theme, summarizes everyone's fate, or ties every thread so neatly that image, implication, and reader participation disappear.
   - `E6 Pacing mismatch`: the climax works but aftermath is absent, rushed, endless, or anticlimactic compared with the buildup.
6. Build a setup-payoff inventory for the ending. Track objects, skills, allies or enemies, information, threats, promises, genre promises, and foreshadowing that appear in the first part of the story, then check what pays off in the final stretch. Flag unresolved setups, payoffs without setup, and setups introduced so close to payoff that they cannot carry emotional weight.
7. Test protagonist agency and transformation proof. The climax should force a final choice, and the outcome should require what the protagonist has learned, become, refused, or failed to become. Positive arcs prove a new truth through action; negative arcs complete the fall and show consequences; flat arcs vindicate the character's truth by changing the world.
8. Check emotional resolution and cost. Identify what feeling the story promised, what the ending asks the reader to carry away, and what victory or defeat costs. A perfect resolution with no loss, ambiguity, sacrifice, or permanent consequence usually weakens the ending unless the genre contract specifically calls for comfort.
9. Distinguish open loops from loose ends. Open loops are intentional unanswered questions that fit the story type, theme, or series contract. Loose ends are forgotten promises, unresolved main-plot obligations, or character questions that received enough page space to require closure.
10. Check twist fairness when the ending recontextualizes prior events. A fair twist uses established information, hides meaning rather than facts, changes interpretation on reread, and does not depend on withheld essentials or contradiction of known canon.
11. Evaluate aftermath. After the climax, the reader usually needs some falling action: immediate consequences, character processing, implication, and a glimpse of the new normal. Too little aftermath feels abrupt; too much becomes epilogue dump.
12. Evaluate the final image. Prefer a concrete action, image, choice, or changed relationship that resonates with the opening and theme. Flag endings that close on summary, logistics, theme speech, sequel bait, or explanation unless that mode is clearly intentional and earned.
13. Recommend the smallest structural intervention that addresses the diagnosed failure: plant missing setup earlier, remove orphaned payoff, make help depend on protagonist action, move late mysteries earlier, cut theme explanation, add aftermath, trim denouement, clarify cost, or replace a loose end with an intentional open loop.

## Edda Output Handling

- Return a concise ending diagnosis in chat by default: ending type, dominant failure state, evidence, setup-payoff risks, transformation proof, emotional resolution, final-image assessment, and recommended intervention.
- Create an Attached Note when the report belongs to one chapter, scene, or selected ending sequence.
- Create or update a Project Note when the author wants a durable payoff inventory, open-loop list, loose-end list, or ending repair plan across multiple chapters.
- For rewrite requests, first state the diagnosis and revision target, then provide draft options or a localized Structured Write only when the author has explicitly requested applied prose.
- Propose Story Bible changes only when ending fixes require durable canon changes to character facts, world rules, timeline, history, names, institutions, or lore. Keep these as proposals until the author confirms them.

## Script Compatibility

The source `ending-check.ts` and `setup-payoff.ts` helpers are converted to guidance here: structure detection, ending-type labels, pacing ratios, anti-pattern checks, setup categories, payoff categories, unresolved setup review, orphaned payoff review, and quick-payoff warnings are all manual Edda diagnostic criteria.

Automated helper execution is deferred. Do not treat source scripts as readable reference files or required runtime behavior. If approved `skill_script` support is added later, helpers must remain non-mutating and report-only; until then, perform the checks with Edda project-reading tools and explain confidence limits in the diagnosis.
