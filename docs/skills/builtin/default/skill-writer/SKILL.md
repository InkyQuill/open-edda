---
name: skill-writer
description: Open Edda skill-authoring and conversion workflow for creating, auditing, rewriting, and verifying agent-readable skills with routing metadata, lazy reference files, Edda tool usage, output handling, canon policy, and script boundaries.
route:
  actionKinds:
    - chat
    - read_check
    - rewrite
  contentKinds:
    - project_note
    - attached_note
    - agent_session
  tags:
    - skill_authoring
    - open-edda
    - agent-instructions
    - skills
    - process
  priority: 100
metadata:
  useCases:
    - Create a new Open Edda skill from an author request, project note, attached note, or source skill text.
    - Audit or rewrite an Open Edda skill whose trigger metadata, Edda tools, output destination, reference files, canon policy, or script policy are incomplete.
    - Convert external agent skills from OpenClaw, skills.sh, Claude, Codex, or similar systems into Edda-native runtime skills.
    - Split bulky examples, rubrics, templates, or data out of a skill body so the agent can load them later with read_skill_file.
  doNotUse:
    - Do not use for drafting, revising, or checking Story Text.
    - Do not use for changing durable Story Bible canon except as an example inside a skill-authoring task.
    - Do not use when the author is asking how to install or administer the Skill Core backend rather than author skill content.
    - Do not use when the source workflow depends on unavailable runtime capabilities and the correct result is to archive or defer it.
  status: default
  source:
    - docs > skills > important > writing-skills > SKILL.md
  scriptStatus: authoring-helpers-deferred-admin-controlled
---

# $skill-writer

Write Open Edda skills as executable instructions for future Edda agents. Do not write human-facing essays, marketing copy, or loose advice.

## Edda Workflow

1. Identify the requested operation: create a new skill, audit an existing skill, rewrite an Edda skill, or convert an external skill.
2. Read the available source through Edda context tools:
   - Use `read_content` for project notes, attached notes, imported source drafts, or skill-design documents.
   - Use `search_content` and `project_map` when the author references related notes, drafts, or prior skill decisions without naming the exact content.
   - Use `skill` to load the current skill body when rewriting an installed Edda skill.
   - Use `read_skill_file` when this skill or the target skill lists a relevant `references/`, `templates/`, or `data/` file.
3. Separate durable requirements from examples, style preferences, unsupported source assumptions, and open questions.
4. Extract the source methodology before rewriting: named principles, state taxonomies, diagnostic criteria, checklists, rubrics, anti-patterns, redirect rules, examples that teach behavior, and script logic that can become guidance or data.
5. Define the future agent behavior the skill must cause: diagnosis, coaching, brainstorming, drafting, rewriting, routing, note creation, canon proposal, or approved helper use.
6. Write trigger metadata before body content:
   - `description` is a concise discovery hook.
   - `metadata.useCases` lists positive activation conditions.
   - `metadata.doNotUse` lists routing-away conditions.
   - `route.actionKinds`, `route.contentKinds`, `route.tags`, and `route.priority` make the skill discoverable by task and content.
7. Write the body as operational steps and judgment criteria the future Edda agent can execute with Edda tools. Do not compress a source method into a generic workflow that loses its tests, categories, or anti-patterns.
8. Define where output belongs: chat, Attached Note, Project Note, Story Bible proposal, Structured Write, or an applied edit only when the author explicitly asked for one.
9. Define canon policy whenever the skill can affect worldbuilding, character facts, timeline, institutions, names, rules, or history.
10. Define reference-file policy:
   - Keep the main `SKILL.md` lean.
   - Move large examples, rubrics, conversion tables, templates, and structured data into `references/`, `templates/`, or `data/`.
   - Tell the future agent exactly when to call `read_skill_file` for each file.
11. Define script policy:
   - Assume source scripts are unavailable unless Skill Script Runtime marks them enabled and approved.
   - Convert script behavior into instructions, data, templates, or an explicit deferred helper.
12. Verify the result against the quality gate before returning it.

## Required Skill Shape

Every Edda skill draft must have this contract:

```yaml
---
name: short-hyphen-name
description: Concrete trigger sentence for agent selection.
route:
  actionKinds:
    - chat
  contentKinds:
    - project_note
  tags:
    - exact-domain
  priority: 80
metadata:
  useCases:
    - Concrete condition that should make the agent choose this skill.
  doNotUse:
    - Concrete condition that should make the agent route away.
---
```

Every main body should normally contain:

```markdown
# $skill-name

One-sentence operational purpose.

## Edda Workflow

## Edda Output Handling

## Script Compatibility
```

Add `## Reference Files`, `## Templates`, or `## Data Files` only when separate files exist and the agent needs routing instructions for loading them.

## Conversion Rules

When converting from OpenClaw, skills.sh, Claude, Codex, or another agent-skill format:

1. Preserve the source's useful agent behavior, not its platform mechanics.
2. Move activation rules out of body sections such as `Use When`, `When to Use`, `Don't Use`, or similar headings into `metadata.useCases` and `metadata.doNotUse`.
3. Translate filesystem, terminal, shell, direct path, file watcher, grep, and subagent assumptions into Edda tools or mark the behavior deferred.
4. Replace direct knowledge dumps with context-reading requirements. The future agent should decide which chapters, Story Bible entries, notes, and skill reference files matter.
5. Keep examples only when they change agent behavior. Put long examples in `references/` and tell the agent when to load them.
6. Preserve useful templates as `templates/` files and mention them in `## Templates`.
7. Preserve structured lists, rubrics, and lookup tables as `data/` or `references/` files when they are too large for the main body.
8. Do not convert external scripts into runnable Edda behavior unless the author explicitly approves a Skill Script Runtime helper.
9. Treat source methodology as load-bearing. Preserve named frameworks, state labels, diagnostic dimensions, score/rubric criteria, anti-patterns, failure modes, and "what not to do" corrections unless they are platform-specific or contradicted by Edda.

For detailed external-format mapping, call `read_skill_file` for `references/external-skill-conversion.md` on this selected skill.

## Reference Files

Use `read_skill_file` for this skill's bundled files only when the task needs them:

- `references/external-skill-conversion.md`: load when converting a source skill from OpenClaw, skills.sh, Claude, Codex, or another platform.
- `templates/edda-skill.md`: load when producing a complete new `SKILL.md` draft or replacing a vague draft with a full Edda-native structure.

Do not load these files for a short audit if the main rules are enough.

## Edda Output Handling

- Return short audits, routing decisions, and rewrite notes in chat.
- Produce a full replacement `SKILL.md` when the author asks to rewrite a skill or when the existing draft is not salvageable.
- Create or update an Attached Note when the feedback belongs to one selected draft or one imported source skill.
- Create or update a Project Note when preserving a conversion plan, routing matrix, audit report, or multi-skill migration plan.
- Use Story Bible proposals only for examples inside a skill-authoring task; this skill must not change story canon.
- Use Structured Write or applied edits only when the author explicitly asks to apply the skill draft to stored content.

## Quality Gate

Before finishing, verify:

1. The skill can be selected from metadata alone.
2. The body does not duplicate `metadata.useCases` or `metadata.doNotUse` as selection sections.
3. The body names concrete Edda tools and concrete context the future agent should inspect.
4. The source method's concrete criteria survived: named principles, state taxonomy, diagnostic dimensions, rubrics, anti-patterns, redirect rules, or equivalent Edda-native replacements.
5. The skill can distinguish adjacent cases. A workflow that only says "read context, diagnose, suggest improvements" fails this gate.
6. The output destination is explicit.
7. Canon-changing behavior is proposal-based until author confirmation.
8. Supporting files are lazy-loaded through `read_skill_file`; large reference content is not inlined into the main body.
9. Scripts are absent, converted to guidance/data, or clearly marked as deferred/approved helper behavior.
10. The draft contains no operational dependency on `read_file`, shell commands, local paths, grep, file watchers, terminal-only agents, or unapproved script execution.
11. A misuse case from `metadata.doNotUse` would route away from this skill.

## Script Compatibility

This skill does not require executable helpers. External authoring scripts may inform conversion decisions only if their behavior is represented as Edda-native guidance, data, templates, or an explicitly deferred Skill Script Runtime helper.
