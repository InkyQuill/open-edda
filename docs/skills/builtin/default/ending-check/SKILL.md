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
  status: default
  source:
    - docs > skills > suggested > fiction > structure > endings > SKILL.md
  scriptStatus: source-helpers-converted-to-guidance
---

# $ending-check

Ending and payoff diagnosis for authors whose story builds energy but loses force, clarity, or satisfaction in the final stretch.

## Use When

- The ending feels arbitrary, obvious, rushed, overexplained, or emotionally thin.
- The author wants to know whether the climax and aftermath actually pay off the setup.
- Beta feedback says the ending does not land.

## Do Not Use When

- The story is not far enough along to judge its ending.
- The problem is mainly earlier pacing or drafting momentum.
- The author wants the skill to write the ending for them by default.

## Edda Workflow

1. Read the ending in the context of the setup that feeds it.
2. Diagnose the likely failure mode: arbitrary, predictable, unearned, expanding, overexplained, or pacing mismatch.
3. Check whether the protagonist's final choice and the story's promises align.
4. Explain what must be planted, tightened, shortened, or left implicit.
5. Keep the result diagnostic unless the author explicitly switches to revision work.

## Edda Output Handling

- Return the ending diagnosis in chat by default.
- Create an Attached Note when the report belongs to one chapter or ending sequence.
- Create or update a Project Note when the author wants a payoff inventory or ending repair plan.
- Do not use Structured Writes in this skill unless the author explicitly switches to an applied rewrite workflow.
- Propose Story Bible updates only when ending fixes depend on confirmed canon changes.

## Script Compatibility

This rewrite adapts source payoff and ending checks into Edda-native reports and revision guidance. Source helpers are not runnable in Milestone 3.5.
