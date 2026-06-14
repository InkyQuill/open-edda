# Edda Skills Documentation

This directory tracks Edda skill source material, rewrite staging, and policy notes for the Built-In Skill Library.

## Milestone 3.5 Scope

Milestone 3.5 follows the accepted library rewrite plan in [docs/superpowers/plans/2026-06-14-writer-skill-library-rewrite.md](/home/inky/Development/writer/.worktrees/milestone-3-5-skill-library/docs/superpowers/plans/2026-06-14-writer-skill-library-rewrite.md). The goal is to curate a Edda-native built-in skill shelf, not to import every copied prompt as an installed skill.

## Mention And Command Policy

- `$name` is a Skill Mention used to select or discover a skill.
- `/command` is a Slash Command used for explicit application actions.
- `@name` is an Entity Mention used for story project content.

Built-in skill names should stay clear in the `$` picker. Prefer concise, intent-first names such as `$scene-pacing`, `$revision-planner`, or `$children-stories`.

## Built-In Library Policy

- `default/` contains Default Skills: built-in skills intended to be enabled on first run because they support common daily fiction-writing work.
- `optional/` contains Optional Skills: built-in skills intended to ship disabled by default because they serve specialized genres, advanced workflows, or publishing-adjacent work.
- `builtin/archive-notes/` contains Archived Skill and deferred-skill review notes only. Archive notes are not install records and do not imply the skill ships with Edda.

Genre-specific skills remain a future expansion track. Milestone 3.5 seeds that direction with `$children-stories`, but it does not try to ship a full genre-pack catalog yet.

## Script Runtime Boundary

Milestone 3.5 curates and rewrites skills. Skill script execution belongs to Milestone 3.6: Skill Script Runtime.

- Script-bearing skills can remain in scope for 3.5 only when the core skill still works through Edda-native instructions, templates, references, or data.
- If a skill's core value depends on running helper scripts, the skill should be deferred and documented as such.
