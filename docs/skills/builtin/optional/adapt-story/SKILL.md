---
name: adapt-story
description: Translate story DNA into a new form, setting, or medium while preserving what makes the original work.
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
    - story_bible_entry
    - entry_section
  tags:
    - fiction
    - adaptation
    - synthesis
    - optional
  priority: 48
metadata:
  status: optional
  source:
    - docs > skills > suggested > fiction > application > adaptation-synthesis > SKILL.md
    - docs > skills > suggested > fiction > application > media-adaptation > SKILL.md
  scriptStatus: source-helpers-deferred-data-retained
---

# $adapt-story

Adaptation support for authors moving a story into a new setting, medium logic, or formal container while protecting the underlying functions that matter.

## Use When

- The author wants to adapt a story, influence, or prior draft into a new form.
- A project needs function-first transformation instead of surface reskinning.
- Multiple inspirations need to be merged without losing hierarchy.

## Do Not Use When

- No source material or extracted story DNA exists yet.
- The request is only for direct plot summary.
- The author wants a beat-for-beat copy with cosmetic changes.

## Writer Workflow

1. Read the source description, story DNA notes, and target context the author wants.
2. Identify the non-negotiable functions and the forms that can safely change.
3. Generate setting-native or medium-native options that serve those functions without obvious reskinning.
4. Check for context mismatch, missing functions, and genre drift.
5. Draft or rewrite adapted material only when the author explicitly asks for it.

## Writer Output Handling

- Return the adaptation mapping in chat by default.
- Create an Attached Note when the work belongs to one Chapter, one scene cluster, or one adaptation problem.
- Create or update a Project Note when the author wants a reusable adaptation brief.
- Propose Story Bible updates only if the adaptation establishes durable setting or canon decisions the author confirms.
- Use Structured Writes only when the author explicitly asks to draft or rewrite selected Story Text in the adapted form.

## Script Compatibility

This rewrite preserves built-in adaptation data as Writer-native reference material. Source helper scripts are not runnable in Milestone 3.5, so the skill works through guidance, data, templates, and reviewable adaptation plans only.
