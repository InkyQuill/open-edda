---
name: open-edda-skill-staging
description: Use when editing Open Edda staging skill Markdown files in this repository, converting external agent skills into Edda-native skills, or auditing built-in default/optional skill folders before database import.
---

# Open Edda Skill Staging

Use this skill for repository-local authoring work on `docs/skills/**`. This is a Codex development skill, not an Open Edda runtime skill. You may inspect and edit local Markdown files, run parser/import tests, and update staging skill folders that will later be imported into the database.

## Scope

Work on local source files:

- `docs/skills/important/**`
- `docs/skills/suggested/fiction/**`
- `docs/skills/builtin/default/**`
- `docs/skills/builtin/optional/**`
- `docs/skills/builtin/archive-notes/**`
- `docs/skills/manifest.md`
- `docs/skills/open-edda-skill-mechanics.md`

Do not confuse this with `$skill-writer`, which is an Open Edda skill loaded inside the product through Skill Core.

## Required Context

Before rewriting a staging skill, read:

1. `docs/skills/open-edda-skill-mechanics.md`
2. `docs/skills/important/writing-skills/SKILL.md`
3. The target staging `SKILL.md`
4. Any referenced source folder or manifest row that explains disposition, merges, or script status

If the target has scripts, data, templates, or references, inspect the file list before rewriting the main skill.

## Edda Runtime Contract

Staged Edda skills must be importable by Skill Core.

Required frontmatter:

```yaml
---
name: short-hyphen-name
description: Concrete selection trigger.
route:
  actionKinds:
    - chat
  contentKinds:
    - project_note
  tags:
    - fiction
  priority: 80
metadata:
  useCases:
    - Positive condition for choosing this skill.
  doNotUse:
    - Negative condition for routing away.
---
```

Skill body is loaded after selection. Do not put selection sections such as `## Use When` or `## Do Not Use When` in the body. Put those in `metadata.useCases` and `metadata.doNotUse`.

Required body shape:

```markdown
# $skill-name

Operational purpose.

## Edda Workflow

## Edda Output Handling

## Script Compatibility
```

Add `## Reference Files`, `## Templates`, or `## Data Files` only when the skill has separate files the runtime agent should load with `read_skill_file`.

## Tool Translation Rules

When converting external skills:

- Replace `read_file`, local Markdown paths, shell reads, grep, file watchers, and terminal agents with Edda tools.
- Use `project_map`, `search_content`, `read_content`, `read_chapter`, `read_story_bible_entry`, `read_entry_section`, and `list_revisions` for project context.
- Use `skill` for main skill instructions.
- Use `read_skill_file` for `references/`, `templates/`, or `data/` only when needed.
- Use `skill_script` only for approved non-mutating helpers.
- Keep direct writes behind explicit author intent and Edda write tools.

## Reference File Policy

Keep the main `SKILL.md` lean. Move bulky material into:

- `references/` for long rubrics, examples, and detailed guidance.
- `templates/` for reusable output structures.
- `data/` for structured tables or JSON.

The main body must tell the runtime agent when to call `read_skill_file` for each file. Do not inline large reference content into `SKILL.md`.

Scripts are not reference files. Script files are disabled by default and handled through `skill_script` only after admin approval.

## Rewrite Workflow

1. Identify source status from `docs/skills/manifest.md`: default, optional, merged, archived, or deferred.
2. Compare the target rewrite against every manifest source. Extract the source's operational method: named principles, state taxonomies, diagnostic criteria, rubrics, checklists, anti-patterns, redirect rules, examples that change behavior, and script logic that can be converted into guidance.
3. Decide the runtime behavior: diagnosis, coaching, brainstorming, drafting, rewriting, routing, note creation, canon proposal, or script helper.
4. Move selection logic into `metadata.useCases` and `metadata.doNotUse`.
5. Rewrite body workflow as concrete Edda-agent actions.
6. Preserve enough source methodology that a future Edda agent can make the same judgment the source skill taught. Do not replace a source method with a generic five-step workflow.
7. State output destination: chat, Attached Note, Project Note, Story Bible proposal, or Structured Write.
8. State canon policy when the skill can affect durable lore, characters, timeline, names, institutions, rules, or history.
9. State script compatibility: no scripts, converted to guidance/data, optional approved helper, or deferred.
10. Move heavy supporting material to separate files and reference it via `read_skill_file`.
11. Verify import and prompt behavior.

## Source Preservation Gate

Before calling a staging skill done, compare it to the source skill and answer yes to each applicable question:

- Does the rewrite preserve the source's named principles, state names, diagnostic dimensions, rubrics, anti-patterns, and concrete tests?
- If source scripts existed, is their useful decision logic converted to Edda guidance, data, references, or an explicit deferred helper policy?
- Can an agent use the rewritten skill to distinguish adjacent cases, or does it only say "diagnose", "improve", "brainstorm", or "revise"?
- Does the body include criteria for what good and bad output look like?
- Are examples retained when they teach behavior, or moved to `references/` when they are long?
- Is the source's output persistence translated into Edda chat, Attached Notes, Project Notes, Story Bible proposals, or Structured Writes?

If the rewrite loses the source methodology, keep rewriting. Valid frontmatter and importability are necessary but not sufficient.

## Verification

After editing any staging skill:

1. Run relevant parser/import tests:

```bash
go test -tags sqlite_fts5 ./skill
```

2. If agent prompt exposure changed, run:

```bash
go test -tags sqlite_fts5 ./agent
```

3. For broad changes, run:

```bash
go test -tags sqlite_fts5 ./...
```

4. Search the edited skill for forbidden runtime assumptions:

```bash
rg -n '## Use When|## Do Not Use When|read_file|grep|shell|local Markdown|watcher|terminal|subagent|run .*script' docs/skills
```

Matches are allowed only when the skill is explicitly forbidding those patterns or documenting a deferred source limitation.

## Completion Standard

A staging skill is not done until:

- It imports through Skill Core.
- Its selection metadata is visible before full body loading.
- Its body tells the future Edda agent exactly what to read and output.
- Its body or lazy-loaded reference files preserve the source skill's operational criteria, not just its topic.
- Reference/template/data files are lazy-loaded through `read_skill_file`.
- Scripts are not treated as normal readable files.
- Canon-changing behavior remains proposal-based until author confirmation.
