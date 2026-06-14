# Built-In Skill Library

This directory is the Milestone 3.5 staging area for Edda-native rewrites of the built-in skill library. The folders here are source buckets for rewrite/import work, not proof that the final built-in library has already been imported or installed.

## Invocation Syntax

- `$` is a Skill Mention.
- `/` is a Slash Command.
- `@` is an Entity Mention.

Use concise, intent-first `$` names in the skill picker. Treat slash commands as commands, not skill aliases.

## Buckets

- `default/`: staging folder for rewrites that are intended to become default built-in skills after rewrite/import integration.
- `optional/`: staging folder for rewrites that are intended to become optional built-in skills after audit, rewrite completion, and import integration.
- `archive-notes/`: documentation-only notes for reviewed skills that are not being imported as Milestone 3.5 built-ins.

After rewrite/import integration, built-in skills live under Skill Core control. Default skills become enabled on first run, optional skills become available disabled by default, and either set can be disabled by admins or authors.

## Milestone 3.5 Rules

- Default Skills are the baseline fiction-writing toolkit targeted for first-run enablement once the rewrites are complete and imported. They should cover common drafting, revision, analysis, outlining, and worldbuilding workflows without requiring scripts.
- Optional Skills are useful specialized tools targeted for later availability as disabled-by-default built-ins once their audit and rewrite status is complete. Do not describe unresolved or conditional entries as already installed.
- Archived Skills are not installed skills. Keep them as notes only so future workers can see they were reviewed intentionally and understand why they were deferred.
- Genre-specific skills are a future expansion track. Milestone 3.5 seeds that direction with `$children-stories` without turning the built-in library into a full genre-pack catalog.

## Script Runtime Boundary

Milestone 3.5 rewrites the library shape and manifest, not the script runtime.

- If a source skill has helper scripts that are optional, keep the rewrite usable without them.
- If a source skill depends on scripts for its core value, defer it until Milestone 3.6 and record the reason in `archive-notes/` or the manifest.
- Do not treat source scripts as normal in-app behavior until Skill Script Runtime exists.

## Rewrite Status

Milestone 3.5 populates `default/` and `optional/` with Edda-native rewrites. Follow-up integration work can connect these staged rewrites to Skill Core import/runtime behavior. Archived material should stay as documentation rather than installed skills.
