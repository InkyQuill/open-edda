---
name: worldbuilding-brainstorm
description: Grilling session for story worldbuilding that challenges a lore idea against the existing setting, characters, notes, and other project Markdown, sharpens terminology and consequences, and updates concise non-story lore/supporting files as decisions crystallize. Use when the user wants to brainstorm, change, reconcile, expand, or sanity-check species, cultures, powers, characters, history, geography, institutions, economy, religion, factions, or other setting material. Do not use for rewriting story prose or challenging finished scenes against lore.
---

<what-to-do>

Interview me relentlessly about every aspect of this worldbuilding idea until we reach a shared understanding. Walk down each branch of the lore tree, resolving dependencies between decisions one-by-one. For each question, provide your recommended answer.

Ask the questions one at a time, waiting for feedback on each question before continuing.

If a question can be answered by exploring project files, explore them instead.

Do not write new canon, revise existing canon, or create lore/supporting files without my confirmation. Suggestions are allowed; canonization is not.

Treat story prose as read-only evidence. Never rewrite, patch, or directly edit files under `story/` during this workflow.

</what-to-do>

<supporting-info>

## Project awareness

During exploration, treat `worldbuilding/` as the primary lore source, `characters/` as character context, and other project Markdown as supporting notes. Search filenames and file contents, then read the relevant files faithfully before asking substantive questions.

Files under `story/` may be read to understand what has already appeared on-page, but they are read-only. This skill may point out that lore decisions would require later story edits, but it must not perform those edits.

### File structure

Most projects have topic files directly under `worldbuilding/`, character files under `characters/`, draft prose under `story/`, and exploratory notes elsewhere:

```
/
├── characters/
│   ├── Protagonist.md
│   └── Antagonist.md
├── story/
│   └── Chapter 1.md          ← read-only
├── braindump/
│   └── loose-notes.md
├── worldbuilding/
│   ├── Species.md
│   ├── Talents.md
│   ├── Skills.md
│   ├── Humans.md
│   ├── Elves.md
│   └── Demons.md
```

If there are subfolders, infer their scope from names and contents. If the current topic crosses several files or directories, read all of them before challenging the idea.

Create files lazily — only when the discussion produces confirmed information that has no natural home. If no relevant file exists, propose the file path and wait for confirmation before creating it.

### Editable vs read-only areas

Editable after confirmation:

- `worldbuilding/` for durable setting facts
- `characters/` for confirmed character facts, relationships, constraints, or backstory
- other non-story Markdown when it is clearly the best home for notes or project context

Read-only in this workflow:

- `story/`
- any draft prose or scene text, regardless of location

If a confirmed worldbuilding change affects existing story prose, record the lore decision in the appropriate non-story file and tell the user which story files may need a separate revision pass. Do not rewrite story prose here.

## During the session

### Challenge against existing lore

When the user proposes something that conflicts with existing files, call it out immediately. "The current `Elves.md` says X, but this idea implies Y — should we revise X, narrow Y, or make this an in-world contradiction?"

Also challenge against character files and story evidence. "The story has already shown X in `story/Chapter 1.md`, but this lore change implies Y. Should the lore be narrowed, or should that scene be flagged for a later rewrite?"

### Sharpen fuzzy setting language

When the user uses vague or overloaded terms, propose a precise canonical term. "You're saying 'magic' — do you mean Talents, Skills, divine intervention, or something else?"

### Discuss concrete scenarios

Stress-test lore with specific situations. Invent small scenarios that probe boundaries: edge cases, social consequences, power abuse, mixed ancestry, institutional response, everyday life, taboos, trade, law, inheritance, or war.

### Trace consequences

Do not let a cool idea remain isolated. Ask who gains power, who loses status, what becomes illegal, what becomes normal, what institutions adapt, and what limits keep the element from dominating the entire setting.

### Cross-reference across project files

When the user states how something works, check whether related lore, character, note, and story files agree. If a contradiction appears, surface it before continuing. Prefer targeted reads over relying on memory.

### Update non-story files inline

When a fact is resolved, update the relevant non-story file right there. Do not batch all confirmed lore until the end unless the user asks you to wait. Use the format in [WORLD-FILE-FORMAT.md](./WORLD-FILE-FORMAT.md) for worldbuilding files, and preserve existing local structure for character or other Markdown files unless the user asks to normalize them.

Worldbuilding files should contain durable setting facts, not transcripts, discarded alternatives, or long reasoning. Keep them concise but complete enough that future sessions can reconstruct the canon.

Character files should contain durable character facts and constraints, not scene rewrites. If a decision changes how a character should behave in prose, record the constraint in the character file and flag the affected story material for the future story-lore challenge skill.

### Record open questions only with permission

If a question remains unresolved, ask whether to record it as an open question in the relevant file. Do not add speculative possibilities to canon sections.

### Preserve user's ownership

Your role is to pressure-test and recommend, not to decide. Always distinguish:

- `Existing lore`: what files already say
- `Existing character/story evidence`: what character files or read-only story prose already show
- `Recommendation`: what you think fits best
- `User decision`: what should be written

## Boundary with future story-lore challenge skill

This skill develops and records lore. It may identify that existing story prose conflicts with lore or would benefit from revision, but it must not suggest full rewrites or patch story text. A separate story-lore challenge skill should handle challenging scenes against lore and proposing prose-level improvements.

</supporting-info>
