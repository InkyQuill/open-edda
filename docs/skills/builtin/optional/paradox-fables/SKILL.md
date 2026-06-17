---
name: paradox-fables
description: Draft short fable-like stories that hold a tension or paradox without flattening it into a simple lesson.
route:
  actionKinds:
    - chat
    - continuation
    - rewrite
  contentKinds:
    - chapter
    - story_text
    - attached_note
    - project_note
  tags:
    - fiction
    - fable
    - theme
    - optional
  priority: 50
metadata:
  useCases:
    - The author wants a teaching story, inset tale, or standalone short with layered meaning.
    - A theme is best approached sideways through image and action.
    - The ending should leave productive tension rather than a sermon.
    - The author asks for a fable, parable, symbolic micro-story, oral-tradition fragment, or narrative embodiment of a paradox.
    - A draft has become didactic, allegorical, over-explained, or too cleanly moralized.
  doNotUse:
    - The author wants a traditional moralistic fable with a single clean lesson.
    - The request is mainly for prose polish rather than concept and form.
    - The story needs heavy realism instead of stylized compression.
    - The author needs cultural research into an existing tradition rather than an original universal symbolic form.
  status: optional
  source:
    - docs > skills > suggested > fiction > application > paradox-fables > SKILL.md
  scriptStatus: no-source-helpers
---

# $paradox-fables

Paradox-fable support for authors who want compact, memorable story forms that embody tension instead of resolving it into a neat moral.

## Edda Workflow

1. Establish the working mode before generating prose:
   - Planning mode: return a paradox engine, fable compression plan, symbolic cast, escalation pattern, ending turn, and quality risks.
   - Drafting mode: produce the fable only after the author asks for a draft, continuation, or rewrite.
   - Revision mode: diagnose the supplied draft against the criteria below before rewriting it.
2. Read the target chapter, selected Story Text, Attached Note, or Project Note when the fable must fit an existing scene, theme, voice, or insertion point. Use `project_map`, `read_chapter`, `read_content`, `read_entry_section`, or `read_story_bible_entry` as needed.
3. Build the paradox engine:
   - Name the irreducible tension in one sentence outside the fable.
   - Identify the two valid forces that cannot both fully win.
   - Convert the tension into action, relationship, natural process, or repeated choice; do not merely state the paradox.
   - Reject tensions that collapse into advice, preference, or a solved problem.
4. Define the moral contradiction and impossible lesson:
   - The moral contradiction is the pair of truths the fable must keep alive.
   - The impossible lesson is what a reader can feel but cannot reduce to a command.
   - Do not put either phrase into the fable as explanation.
5. Compress the fable:
   - Prefer one image system, one central action, and one pressure pattern.
   - Use simple, oral-tradition language with durable concrete nouns.
   - Remove contemporary slang, decorative preciousness, and analytical commentary unless the project voice explicitly requires them.
6. Choose symbolic characters that remain beings:
   - Use animals, objects, natural forces, travelers, artisans, elders, children, or ancient beings when they naturally embody the tension.
   - Give each character a desire, limit, or stake beyond representing an idea.
   - Avoid cute alliteration, label-names that announce the lesson, and characters who exist only to make a point.
7. Let form emerge from the paradox:
   - Use circular form when the contradiction returns the character to the same place changed.
   - Use parallel action when two valid truths mirror each other.
   - Use reversal when pursuit creates its opposite.
   - Use accumulation or reduction when the paradox intensifies through repetition.
   - Add a witness chorus only when multiple observers can reveal facets the protagonist cannot see; no witness should explain the whole meaning.
8. Escalate through choices, not exposition:
   - Each beat should increase the cost of holding one truth while neglecting the other.
   - The trap or wisdom must emerge from action and consequence.
   - If the fable can keep the same plot after removing the paradox, rebuild the engine.
9. Make the ending turn without resolving:
   - End on an image, action, reversal, recognition, or returned motif that keeps pressure alive.
   - The ending may feel inevitable, satisfying, funny, unsettling, or quietly open, but it must not provide a clean answer.
   - Remove explicit morals, summary lessons, and characters explaining what the story means.
10. Check cultural safety before finalizing:
   - Prefer universal observations such as seasons, water, tools, weather, hunger, shelter, craft, growth, and decay.
   - If the image, structure, character, or phrase feels tied to a specific cultural or sacred tradition, name the risk and ask whether to research, credit, replace, or contextualize it.
   - Do not create imitation wisdom stories from specific traditions without the author's explicit direction and context.

## Paradox-Fable Criteria

A strong paradox fable should pass these tests:

- Removing the paradox would break the story's structure or central action.
- At least three valid interpretations can coexist without one becoming the official answer.
- The ending maintains productive tension instead of resolving into advice.
- The story feels compact, memorable, and re-readable rather than like an argument in costume.
- Symbolic characters act from believable desire or limitation, not from assigned thesis roles.
- The form feels discovered from the paradox rather than imposed from a template.
- The prose can be read aloud without sing-song cuteness, contemporary clutter, or ornate self-consciousness.

Repair these failure modes before presenting the work as complete:

- Forced moral: delete lesson statements and end on an action or image.
- Allegory characters: replace labels with beings that want something concrete.
- Imposed structure: return to the paradox engine and choose a form that mirrors the tension.
- Explanation temptation: move analysis outside the fable; the story itself should not define its meaning.
- False resolution: preserve both sides of the contradiction at the close.
- Cultural appropriation risk: replace culturally specific symbols with universal imagery or ask for permission and context.

## Edda Output Handling

- In planning mode, return a compact plan with: paradox engine, moral contradiction, impossible lesson, natural form, symbolic characters, escalation, ending turn, and risks.
- In drafting mode, return the fable draft plus a short out-of-story note naming the embodied paradox and any unresolved risks. Do not append an in-story moral.
- In revision mode, lead with the failure criteria found, then provide a rewrite only when the author asked for one.
- Return the work in chat by default when the author is exploring options or asking for a standalone fable.
- Create an Attached Note when the fable belongs to one chapter, selection, inset tale, or local scene experiment.
- Create or update a Project Note when the author wants a durable bank of fable concepts, teaching stories, or inset texts.
- Treat any in-world fable, myth, custom, proverb, named teller, cultural practice, historical event, or recurring symbol as canon-affecting. Propose Story Bible changes separately and wait for author confirmation before treating them as established.
- Use Structured Writes only when the author explicitly asks to insert, continue, or replace selected Story Text in this form.

## Script Compatibility

This source skill has no required helper scripts. The Milestone 3.5 rewrite works through Edda-native drafting and thematic guidance.
