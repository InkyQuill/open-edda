---
name: avoid-cliches
description: Originality guidance that keeps the story function intact while moving characters, plot moves, and world details away from generic default expressions.
route:
  actionKinds:
    - chat
    - read_check
    - rewrite
  contentKinds:
    - chapter
    - story_text
    - project_note
    - attached_note
    - story_bible_entry
    - entry_section
  tags:
    - fiction
    - originality
    - revision
    - craft
  priority: 76
metadata:
  useCases:
    - Feedback says a scene, character, trope, or reveal feels familiar in a tired way.
    - The first idea works functionally but feels too expected.
    - The author wants fresher choices without making the story strange just for novelty.
  doNotUse:
    - The element is not actually the problem and the draft needs broader diagnosis first.
    - The author wants random weirdness instead of purposeful specificity.
    - The task is mainly copyediting or proofing.
  status: default
  source:
    - docs > skills > suggested > fiction > craft > cliche-transcendence > SKILL.md
    - docs > skills > suggested > fiction > character > statistical-distance > SKILL.md
  scriptStatus: source-helpers-converted-to-guidance
---

# $avoid-cliches

Originality coaching and revision support that preserves a story element's function while moving its expression away from generic defaults and toward project-specific, fertile choices.

## Reference Files

- Load `references/cliche-methodology.md` with `read_skill_file` when the author asks for a systematic originality pass, when several elements feel generic, or when you need the full Orthogonality Principle, CTF process, statistical center/edge method, fertility rubric, and pitfalls.

## Edda Workflow

1. Read the target chapter, selection, Attached Note, Project Note, or Story Bible material before diagnosing the element. Use `project_map`, `read_content`, `read_chapter`, `read_story_bible_entry`, `read_entry_section`, and `search_content` as needed to understand current story function and canon.
2. Identify the element under review: character role, relationship, trope, reveal, obstacle, world detail, institution, scene move, or premise component.
3. Separate the element's functional core from its current form. Name what it must provide: stakes, pressure, sympathy, complexity, expertise, access, urgency, moral weight, information, contrast, or emotional effect.
4. Enumerate the statistical center: list the 3-5 most common expressions of that function in the relevant genre or story type. Make the default visible before proposing alternatives.
5. Apply the Orthogonality Principle. Check four axes:
   - Form: what the element is.
   - Knowledge: what it knows about the central plot.
   - Goal: what it wants independent of the protagonist.
   - Role: what story function it serves.
6. Run the key test: "Does this element know what story it is in?" If it behaves as if it exists to help, block, inform, tempt, or mirror the protagonist, look for a version with its own logic that collides with the main story rather than orbiting it.
7. Push along the same emotional vector toward the statistical edge. Keep the function and emotional effect, but change the expression through adjacent substitution, complication layering, ironic inversion, category jumping, specificity injection, imported domain logic, or perspective inversion.
8. Evaluate fertility before recommending an alternative. Prefer choices that generate subplot possibilities, relationship complications, future evidence, thematic pressure, or downstream consequences. Reject novelty that solves only one scene and then goes inert.
9. Calibrate distance. Aim for 60-80% familiarity: fresh enough to require some explanation, familiar enough to click quickly and still satisfy the story need.
10. Present options without choosing for the author unless asked. State what each option preserves, which axis or vector it moves, what new story possibilities it creates, and what risks it introduces.
11. When the author selects an option, map the downstream consequences across relevant chapters, relationships, and canon entries. Treat new facts as proposals until the author confirms them.

## Edda Output Handling

- Return the originality diagnosis and option set in chat by default.
- Create an Attached Note when the work belongs to one chapter or one local trope problem.
- Create a Project Note when the author wants a larger originality pass, a three-column workspace, or a cross-chapter plan. Track functional need, statistical center options, statistical edge options, selected direction, and downstream consequences.
- Use Structured Writes only when the author explicitly asks to apply one of the proposed revisions.
- Propose Story Bible changes separately if the chosen alternative changes durable canon such as character facts, relationships, factions, institutions, timelines, names, setting logic, or history. Do not turn a brainstormed alternative into confirmed canon without explicit author approval.

## Script Compatibility

The source `cliche-transcendence` skill included an `orthogonality-check.ts` helper. This built-in replaces that helper with Edda-native guidance in `references/cliche-methodology.md`; there are no approved runtime scripts for this skill. Do not ask the author to run source scripts or use filesystem paths. Any future script helper is deferred until it is imported as a skill script, audited, approved, enabled, and callable through `skill_script`.
