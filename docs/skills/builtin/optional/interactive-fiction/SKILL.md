---
name: interactive-fiction
description: Design branching fiction that offers meaningful choices without exploding into unmanageable structure.
route:
  actionKinds:
    - chat
    - read_check
    - continuation
    - rewrite
  contentKinds:
    - chapter
    - story_text
    - attached_note
    - project_note
  tags:
    - fiction
    - interactive
    - branching
    - optional
  priority: 52
metadata:
  useCases:
    - The story includes player or reader choices.
    - Branches feel fake, too numerous, or structurally messy.
    - The author wants a cleaner balance between agency and authored payoff.
  doNotUse:
    - The project is a standard linear story.
    - The request is only about prose polish inside one branch.
    - The author wants a full implementation plan for external game tooling rather than story design.
  status: optional
  source:
    - docs > skills > suggested > fiction > application > interactive-fiction > SKILL.md
  scriptStatus: no-source-helpers
---

# $interactive-fiction

Diagnose and improve interactive fiction as a designed possibility space where player agency and authored meaning coexist.

## Edda Workflow

1. Read the relevant chapter, Story Text selection, Attached Note, Project Note, branch outline, choice map, or prior revision history before diagnosing. If the request could affect durable lore, characters, timelines, setting rules, institutions, endings, or canonical outcomes, also read the relevant Story Bible entries and keep changes as proposals until the author confirms them.
2. Identify the interactive-fiction form before prescribing fixes:
   - Parser-based: natural-language commands, high freedom, puzzle-oriented, vulnerable to "guess the verb" friction.
   - Choice-based: selected options, easier to author, vulnerable to false choices and over-frequent menus.
   - Hybrid, visual-novel, or RPG: multiple interaction modes and persistent state, richer but heavier to maintain.
   - Tabletop scenario: facilitator-mediated, dynamic and improvisational, dependent on GM interpretation.
3. Diagnose the dominant IF state and name it in the response:
   - IF1 Meaningless Choices: options converge immediately, players feel nothing matters, or choices are navigation instead of values.
   - IF2 Unmanageable Branching: content grows exponentially, quality drops across routes, or the tree collapses under path count.
   - IF3 False Choice Discovered: players compare paths and realize promised agency was not delivered.
   - IF4 Agency vs. Authored Meaning: freedom creates incoherence, or a fixed story makes interactivity feel pointless.
   - IF5 Story Feels Like Flowchart: decision points interrupt pacing, scenes become menus, or narrative does not breathe.
   - IF6 Multiple Endings, No Satisfaction: endings feel hollow, punitive, ranked, unearned, or invalidated by one "true" ending.
   - IF7 State Management Chaos: flags, variables, inventory, relationship values, or route conditions contradict or proliferate beyond use.
4. For any key choice, apply the Meaningful Choice Test:
   - Distinct options: each option represents a different approach, value, risk, or relationship stance.
   - Perceivable consequences: the player can see results now or later, with reminders when delayed.
   - Irreversibility: the choice cannot be immediately undone without cost.
   - Character expression: the choice reveals protagonist, player intent, or moral priority rather than only optimizing.
5. Analyze branch structure before adding content. Count true branches and show the author the maintenance cost when needed: 3 binary choices create 8 paths, 5 create 32, and 10 create 1,024. Recommend reducing true branches when state flags, texture variation, or reconvergence would preserve agency with less sprawl.
6. Choose a structural pattern that matches the story:
   - Linear with windows: mostly linear arc with occasional choice moments and local variation.
   - Foldback: routes reconverge at key beats; hide the foldback through meaningful texture and delayed consequences.
   - Bottleneck: multiple approaches through fixed story beats, preserving authored climaxes.
   - Branch and bottleneck: early route differences converge around shared endings or major beats.
   - State-based or quality-based variation: accumulated flags, qualities, relationships, resources, or inventory alter scenes without creating parallel universes.
   - Time loop: repeated events with persistent player knowledge, useful for puzzle stories or fixed tragedies where understanding changes.
7. Design choices as dilemmas, not menus. Prefer conflicts between two goods, incompatible values, risk-reward strategies, discovery choices, or authentic expression choices. Avoid binary moral obviousness, correct-path mazes, info-dump menus, and optimization puzzles unless the project intentionally wants game-like solving.
8. Track only state that produces visible effects. Classify state as plot flags, relationship values, resources, qualities, or inventory. Keep state when it gates content, varies a shared scene, or changes consequences. Merge or delete state that never surfaces for the player.
9. Treat false choices carefully. Expression choices can share outcomes if they reveal character, alter tone, or affect later recognition. Repeated hidden convergence that players notice breaks trust; in that case, reduce the number of choices or make fewer choices truly consequential.
10. Reconcile agency and authorship through constrained agency. Define the possibility space, the protagonist's fixed or flexible traits, natural fictional constraints, time pressure, incomplete information, and the themes the author must control. Let the player shape how events unfold while the author controls what matters.
11. Check endings as outcomes of values, not mere scores. Each ending should close its path, feel earned by accumulated choices or state, and be worth experiencing. Avoid one canonical "true ending" invalidating the rest unless the author explicitly wants optimization or secret-ending play.
12. Handle dead ends by distinguishing failure states from wasted paths. A dead end can work when it is short, interesting, thematically meaningful, clearly caused, and teaches or reveals something. Revise dead ends that punish experimentation, hide required information, force replay of long unchanged content, or make the player feel tricked.
13. Preserve narrative flow. Choices should emerge from dramatic moments inside scenes; scenes still need goal, conflict, reversal, consequence, and aftermath. Do not recommend more choice points just to make the work feel interactive.
14. When outlining, produce branch maps, state tables, route summaries, choice tests, bottleneck plans, ending matrices, or revision checklists. When drafting or rewriting, only draft the requested choice text, branch passage, transition, state reminder, or ending beat; do not design the entire branching structure unless the author asks.
15. Route adjacent work when needed: use scene-sequencing for branch scene structure, character-arc for transformation across choices, endings for a deeper ending pass, dialogue for player dialogue choices with subtext, or story-sense when the issue is general pacing rather than interactive design.

## Quality Criteria

- Strong choices express values, character, tradeoffs, strategy, or discovery, and their consequences are perceivable.
- Strong branch design uses authored constraints, bottlenecks, foldbacks, and state variation to avoid exponential sprawl.
- Strong state tracking is small, legible, and visible in gates, variation, or consequences.
- Strong endings are earned, coherent with the route, and differently satisfying rather than simply ranked good or bad.
- Weak IF offers frequent choices that do not matter, hides obvious railroading, treats branches as parallel novels, makes the player optimize instead of roleplay, or uses dead ends as punishment rather than meaning.

## Edda Output Handling

- Return clarifying questions, quick diagnoses, and real-time choice feedback in chat.
- Create an Attached Note for local branch diagnostics, a single scene cluster, a selected choice, a passage rewrite plan, or a nearby dead-end fix.
- Create or update a Project Note for durable IF state diagnosis, branch structure notes, choice-quality assessment, ending matrix, state table, complexity recommendations, or an interactive design brief.
- Use Structured Writes only when the author explicitly asks to draft or rewrite selected branching Story Text. Keep drafts scoped to the selected branch, choice, consequence, transition, reminder, or ending beat.
- Keep canon-affecting outcomes as Story Bible proposals or clearly labeled open questions until the author confirms them. Do not silently canonize route-exclusive facts, "true" endings, world rules, character deaths, relationship states, or timeline changes.
- When asked for an outline, separate confirmed structure from optional proposals, and mark where choices converge, where state persists, and where endings branch.

## Script Compatibility

This source skill has no required helper scripts. The Milestone 3.5 rewrite works through Edda-native analysis, structure guidance, and optional branch drafting.
