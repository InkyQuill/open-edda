---
name: story-dna
description: Extract the functional patterns behind a story, chapter, or outside influence so authors can reuse what works without copying the surface.
route:
  actionKinds:
    - chat
    - read_check
  contentKinds:
    - chapter
    - story_text
    - attached_note
    - project_note
    - story_bible_entry
  tags:
    - fiction
    - analysis
    - adaptation
    - optional
  priority: 38
metadata:
  status: optional
  source:
    - docs > skills > suggested > fiction > application > dna-extraction > SKILL.md
  scriptStatus: source-helpers-deferred-data-retained
---

# $story-dna

Functional analysis for authors who want to understand what a story is doing beneath the surface so they can adapt, remix, or strengthen it on purpose.

## Use When

- The author wants the structural and emotional DNA of a story, chapter, or influence.
- Adaptation, remix, or comparative analysis needs more than summary.
- A story works on instinct and the author wants to know why.

## Do Not Use When

- The author wants immediate drafting help rather than analysis.
- The request is only for marketing positioning.
- The goal is to copy surface features instead of underlying function.

## Edda Workflow

1. Read the target Story Text, Chapter, or outside description the author provides.
2. Separate surface form from structural, emotional, thematic, and relational function.
3. Identify the load-bearing patterns and the optional style layer.
4. Organize the results so the author can reuse function without cloning the original shell.
5. Keep the analysis advisory unless the author asks for a next-step adaptation plan.

## Edda Output Handling

- Return the DNA analysis in chat by default.
- Create an Attached Note when the analysis belongs to one Chapter or one focused excerpt.
- Create or update a Project Note when the author wants a reusable extraction record or adaptation brief.
- Do not propose Story Bible changes unless the author wants analysis conclusions tracked as project guidance.
- Do not use Structured Writes in this skill.

## Bundled Data

This skill includes extraction schemas and taxonomies as Writer-native references:

- `data/extraction-templates.json` — Structured templates for extracting story mechanics from existing works.
- `data/function-categories.json` — Taxonomy of story functions for classification.

The agent should load these through the `skill` tool when an analysis session needs detailed reference material.

## Script Compatibility

This rewrite preserves built-in extraction schemas and function taxonomies as Edda-native references. Source helper scripts are not runnable in Milestone 3.5, so the skill works through analysis, data, and reviewable notes only.
