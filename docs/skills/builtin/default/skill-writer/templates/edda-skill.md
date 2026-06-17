# Edda Skill Template

Use this template when producing a complete `SKILL.md` draft.

```markdown
---
name: short-hyphen-name
description: Concrete trigger sentence for agent selection.
route:
  actionKinds:
    - chat
  contentKinds:
    - project_note
  tags:
    - fiction
    - exact-domain
  priority: 80
metadata:
  useCases:
    - Concrete condition that should make the agent choose this skill.
  doNotUse:
    - Concrete condition that should make the agent route away.
---

# $short-hyphen-name

One-sentence operational purpose for a future Open Edda agent.

## Edda Workflow

1. Read the target content and relevant project context with Edda tools.
2. Separate confirmed canon, inferred context, author preference, and open questions.
3. Perform the skill-specific diagnosis, drafting, rewrite, routing, or proposal task.
4. Produce the required output in the destination named below.

## Reference Files

- `references/example.md`: load with `read_skill_file` only when ...

Delete this section if the skill has no reference files.

## Templates

- `templates/example.md`: load with `read_skill_file` only when ...

Delete this section if the skill has no templates.

## Edda Output Handling

- Use chat for short-lived decisions and questions.
- Use Attached Notes for chapter-local or selection-local findings.
- Use Project Notes for durable plans, audits, option banks, and cross-chapter work.
- Use Story Bible proposals for canon-affecting changes until the author confirms them.
- Use applied edits only when the author explicitly requests a write operation.

## Script Compatibility

State one:

- This skill does not require executable helpers.
- Source scripts were converted to guidance, references, templates, or data.
- Optional helper behavior is deferred until Skill Script Runtime approval.
- Approved helper `name` may be called with `skill_script` only for ...
```
