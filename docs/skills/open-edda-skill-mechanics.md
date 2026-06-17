# Open Edda Skill Mechanics

This is an agent-facing implementation note. Use it when rewriting or auditing Edda skills.

## Storage

Skill source folders currently live under:

- `docs/skills/important/` for source guidance and important upstream skills.
- `docs/skills/suggested/fiction/` for source material.
- `docs/skills/builtin/default/` for default built-in rewrite staging.
- `docs/skills/builtin/optional/` for optional built-in rewrite staging.
- `docs/skills/builtin/archive-notes/` for reviewed but non-installable/deferred sources.

Installed skills live in SQLite through Skill Core:

- `skills`
- `skill_files`
- `skill_routing_hints`
- `agent_session_skills`
- `skill_script_audits`
- `skill_script_approvals`
- `skill_script_runs`

## Import

Skills import from zip archives or from local directories allowed by `WRITER_SKILL_IMPORT_ROOT`.

Import entry points:

- `POST /api/projects/{projectID}/skills/import`
- `POST /api/projects/{projectID}/skills/import-local`
- `skill.ParseSkillArchive`
- `skill.ParseSkillDirectory`
- `skill.Service.Install`

Required source shape:

```txt
skill-folder/
  SKILL.md
  templates/
  references/
  data/
  scripts/
```

Only `SKILL.md` is required.

## Parsed Frontmatter

The parser currently reads:

- `name`
- `description`
- `route`
- `routing`
- `metadata.useCases`
- `metadata.doNotUse`

Supported route keys:

- `actionKinds`
- `actions`
- `contentKinds`
- `content`
- `tags`
- `priority`

Other frontmatter keys can remain in source files as comments for maintainers, but do not rely on them for runtime behavior. Current parser behavior preserves only the `metadata.useCases` and `metadata.doNotUse` lists from arbitrary metadata.

## File Classification

Open Edda classifies files by path:

- `SKILL.md`: instruction
- `templates/...`: template
- `references/...`: reference
- `data/...`: data
- `scripts/...`: script
- everything else: other

Script files are imported disabled by default. Script execution requires a separate audit and approval path.

## Prompt Exposure

Available skills are shown to the model as summaries:

- ID
- name
- description
- use cases
- do-not-use cases

Selected skills are shown with script status and enabled runtime helpers.

The model loads full skill content by calling:

```json
{ "skillId": "skill-..." }
```

through the `skill` tool.

`skill.RenderForModel` returns:

- skill name,
- `instructions_markdown` from `SKILL.md` body,
- a manifest of supporting files with path, purpose, byte count, and readability,
- disabled-script notices instead of executable script bodies.

The model loads one supporting file on demand by calling:

```json
{ "skillId": "skill-...", "path": "references/example.md" }
```

through the `read_skill_file` tool. This tool returns one non-script file body. It rejects scripts and disabled script files.

## Agent Rules For Rewrites

When rewriting a skill:

1. Preserve only behavior that can run through Edda tools and database-backed content.
2. Replace file/path/shell instructions with Edda tools.
3. Put bulky reference material in `references/`, `templates/`, or `data/` and tell the agent when to load it with `read_skill_file`.
4. Give the future agent concrete context-reading requirements.
5. Give the future agent an output destination.
6. Keep canon changes reviewable and separate from confirmed facts.
7. Treat source scripts as optional helpers, inert references, converted guidance, or deferred work.
8. Verify importability before moving to the next skill.

Use `docs/skills/important/writing-skills/SKILL.md` as the authoring standard.
