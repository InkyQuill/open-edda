---
name: writing-skills
description: Use when creating, auditing, rewriting, or verifying Open Edda skills for fiction-writing agents, especially when source skills mention files, shell tools, vague guidance, or non-Edda workflows.
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
  priority: 100
metadata:
  useCases:
    - Create a new Open Edda skill from scratch.
    - Audit or rewrite a source skill that uses terminal-agent, filesystem, shell, or vague prompt-only assumptions.
    - Verify that an Edda skill exposes enough metadata for selection before loading the full skill body.
  doNotUse:
    - Do not use for drafting or revising Story Text.
    - Do not use for changing Story Bible canon except as an example inside a skill-authoring task.
---

# $writing-skills

You are writing instructions for a future Open Edda agent. Do not write a human essay. Write an operational skill that tells the agent exactly when to activate, what project context to read, what decisions to make, which Edda tools to use, what output to produce, and what not to do.

## Edda Skill Contract

An Open Edda skill is a folder with `SKILL.md` plus optional `templates/`, `references/`, `data/`, and `scripts/`.

`SKILL.md` must use this frontmatter shape:

```yaml
---
name: short-hyphen-name
description: Trigger-focused sentence for agent selection.
route:
  actionKinds:
    - chat
    - read_check
    - rewrite
    - continuation
  contentKinds:
    - chapter
    - story_text
    - story_bible_entry
    - entry_section
    - project_note
    - attached_note
  tags:
    - fiction
    - exact-domain
  priority: 80
metadata:
  useCases:
    - Concrete trigger that should make the agent choose this skill.
  doNotUse:
    - Concrete condition that should make the agent choose another skill.
---
```

Open Edda imports `name`, `description`, `route`/`routing`, `metadata.useCases`, and `metadata.doNotUse` into Skill Core. The model sees these fields in the skill summary before it loads the full skill body.

## Required Skill Body

The body is loaded only after the agent chooses the skill. Do not spend body tokens on selection criteria that belong in metadata.

Every Edda skill body must include these sections unless there is a specific reason not to:

1. `# $skill-name`
2. One-sentence operational purpose.
3. `## Edda Workflow`
4. `## Edda Output Handling`
5. `## Script Compatibility` when source scripts exist or the source assumes executable helpers.

The body must give tasks, not vibes. Each workflow step should be an action the agent can perform inside Edda.

Do not add `## Use When` or `## Do Not Use When` sections to the body. Put those bullets in `metadata.useCases` and `metadata.doNotUse`.

Use separate files for heavy supporting material:

- `references/`: detailed guidance, rubrics, examples, long checklists.
- `templates/`: reusable output formats.
- `data/`: structured tables or JSON.

The main skill body should name these files and state when to call `read_skill_file`; it should not inline large reference content.

## Tool Model

Use Edda database tools, not filesystem assumptions.

Allowed project-context tools:

- `project_map`
- `search_content`
- `read_content`
- `read_chapter`
- `read_story_bible_entry`
- `read_entry_section`
- `list_revisions`
- `skill`
- `read_skill_file`
- `skill_script` only when the helper is selected, approved, enabled, and non-mutating.

Write tools exist only for explicit applied changes:

- `append_to_chapter`
- `insert_into_chapter`
- `replace_selection`
- `update_story_bible_entry`
- `update_entry_section`

Never tell an Edda agent to use `read_file`, shell commands, local Markdown paths, grep, watchers, direct filesystem edits, or terminal-only subagents. If source material says that, translate it into Edda tools or mark it deferred.

## Authoring Workflow

When creating or rewriting a skill:

1. Identify the agent behavior the skill must change.
2. Identify the activation trigger: task type, content kind, user phrasing, and failure symptoms.
3. Identify the context the agent must inspect before acting.
4. Decide the output contract: chat answer, Attached Note, Project Note, Story Bible proposal, or Structured Write.
5. Decide canon policy: confirmed fact, proposal, open question, or no canon effect.
6. Decide script policy: no scripts, inert reference, approved helper, or deferred.
7. Move bulky examples, tables, checklists, and reference material into `references/`, `templates/`, or `data`; tell the agent exactly when to load each file with `read_skill_file`.
8. Write the skill in Edda-native terms.
9. Verify the skill imports and gives a future agent concrete actions.

Do not batch-rewrite many skills without checking each one against this workflow.

## Description Rules

The `description` is the shortest discovery hook. `metadata.useCases` and `metadata.doNotUse` carry the detailed selection rules. Do not force the agent to load the body to decide whether the skill applies.

Good:

```yaml
description: Scene and chapter pacing diagnosis for weak escalation, missing aftermath, clean victories, and sequences that do not accumulate pressure.
```

Bad:

```yaml
description: Helps with scenes.
```

Rules:

- Include concrete symptoms and nouns an agent might match.
- Prefer task triggers over marketing language.
- Do not say "I can".
- Do not hide the domain. Use words like `dialogue`, `worldbuilding`, `revision`, `outline`, `canon`, `pacing`, `genre`.
- Keep it short enough to scan.

## Selection Metadata Rules

Use `metadata.useCases` for positive activation cases. Use `metadata.doNotUse` for negative routing cases.

Good:

```yaml
metadata:
  useCases:
    - A chapter has weak escalation, missing aftermath, or too-clean resolution.
    - The author asks why a scene feels slow, flat, rushed, or mechanically connected.
  doNotUse:
    - The author wants sentence-level line editing rather than scene diagnosis.
    - The author wants new lore or canon brainstorming.
```

Rules:

- Make each item a concrete condition.
- Include symptoms an agent can match from the user request.
- Include routing-away cases that prevent false positives.
- Keep metadata concise; put process details in `Edda Workflow`.
- Do not duplicate these lists as body headings.

## Route Rules

Use `route` to make the skill discoverable by action and content type.

Action kinds:

- `chat`: discussion, planning, diagnosis, brainstorming.
- `read_check`: review, audit, diagnosis, continuity check.
- `rewrite`: localized replacement or revision support.
- `continuation`: drafting forward from a chapter or note.
- Use other action names only if the backend/UI explicitly supports them.

Content kinds:

- `chapter` / `story_text`: story prose.
- `story_bible_entry`: durable canon entry.
- `entry_section`: section inside a Story Bible entry.
- `writing_brief`: project or chapter instructions.
- `project_note`: durable planning or analysis note.
- `attached_note`: note tied to a chapter/selection.
- `agent_session`: skill-authoring or process work in chat history.

Tags should be specific enough for filtering: `dialogue`, `pacing`, `canon-safe`, `worldbuilding`, `line-edit`, `outline`, `revision`, `genre`, `skill_authoring`.

## Edda Workflow Rules

Workflow steps must make context requirements explicit.

Good:

```markdown
1. Read the target chapter and any relevant Story Bible entries before diagnosing continuity.
2. Separate established canon from inferred possibilities.
3. Return a ranked list of continuity risks with evidence.
```

Bad:

```markdown
1. Think deeply.
2. Improve the story.
3. Be creative.
```

For fiction skills, specify whether the agent should:

- diagnose only,
- ask questions,
- generate options,
- draft proposal text,
- apply a rewrite,
- create notes,
- propose Story Bible changes.

## Output Handling Rules

Every skill must tell the agent where output belongs.

Use chat when:

- the author is deciding,
- the answer is short-lived,
- the agent is coaching or asking questions.

Use Attached Notes when:

- the result belongs to one chapter, selection, scene, or local issue.

Use Project Notes when:

- the result is a durable plan, checklist, cross-chapter diagnosis, outline, option bank, or rewrite plan.

Use Story Bible proposals when:

- the result changes durable canon, worldbuilding, character facts, factions, history, rules, names, or setting logic.

Use Structured Writes only when:

- the author explicitly asks to apply text,
- the target content and current revision are known,
- the skill is in an action kind where write tools are available.

Quick-action skills must not try to call direct write tools during generation. Quick actions return final text; Edda applies preview/direct-apply afterward.

## Canon Policy

Skills that touch lore, character facts, institutions, timelines, magic, technology, geography, names, or history must state canon policy.

Required distinction:

- `Existing canon`: facts read from Story Bible, Writing Briefs, or Story Text.
- `Inference`: likely implication from context.
- `Proposal`: agent suggestion not yet canon.
- `Confirmed canon`: author-approved durable fact.

Never turn a brainstormed idea into confirmed canon without explicit author confirmation.

## Script Policy

Scripts are not normal skill behavior.

If source skill has scripts:

1. Decide whether the core skill works without them.
2. Convert script logic into instructions, templates, `data/`, or `references/` when possible.
3. Keep scripts only as optional helpers if they produce reports, proposals, drafts, or generated data.
4. Mark script-dependent workflows as deferred if they need file watchers, batch pipelines, project filesystem access, network access, or direct mutation.
5. State the runtime boundary in `## Script Compatibility`.

Do not instruct the agent to ask the author to run scripts manually. In Edda, scripts run only through `skill_script` after admin approval.

## Good Edda Skill Skeleton

```markdown
---
name: scene-pacing
description: Scene and chapter pacing diagnosis for escalation, aftermath, clean victories, and pressure flow.
route:
  actionKinds:
    - chat
    - read_check
  contentKinds:
    - chapter
    - story_text
    - attached_note
    - project_note
  tags:
    - fiction
    - pacing
    - structure
  priority: 82
metadata:
  useCases:
    - A scene has no clear goal, weak escalation, or too-clean resolution.
    - A chapter feels slow, exhausting, rushed, or rhythmically flat.
  doNotUse:
    - The author wants sentence-level line editing rather than scene diagnosis.
    - The text is too early to support scene-level review.
---

# $scene-pacing

Pacing and sequencing review for scenes that feel slow, flat, rushed, or mechanically connected.

## Edda Workflow

1. Read the target chapter or selected scene in context.
2. Identify goal, conflict, escalation, outcome, and aftermath.
3. Diagnose whether pressure is missing, misplaced, repeated, or overextended.
4. Suggest specific revision targets without rewriting by default.

## Edda Output Handling

- Return the diagnosis in chat by default.
- Create an Attached Note for one chapter or selection.
- Create a Project Note for cross-chapter pacing patterns.
- Use Structured Writes only after explicit author request.

## Script Compatibility

No script is required. Any source analyzer is optional reference logic, not runtime behavior.
```

## Rewrite Checklist

Before finishing a skill, verify:

- [ ] Frontmatter has `name`, `description`, and useful `route`.
- [ ] Description has concrete triggers.
- [ ] `metadata.useCases` and `metadata.doNotUse` handle skill selection without loading the body.
- [ ] Body tells the agent what to read before acting.
- [ ] Heavy references live in separate files and are loaded through `read_skill_file` only when needed.
- [ ] Body tells the agent what output to produce and where it belongs.
- [ ] Canon policy is explicit when canon can change.
- [ ] Script policy is explicit when source scripts exist.
- [ ] No terminal paths, shell commands, `read_file`, local Markdown assumptions, or manual script instructions remain.
- [ ] No vague verbs stand alone: `improve`, `enhance`, `consider`, `think deeply`, `make better`.
- [ ] The skill can work through Edda tools and database content.
- [ ] The body does not contain `Use When` / `Do Not Use When` sections duplicated from metadata.

## Anti-Patterns

Reject or rewrite these:

- "Read the file at `world/...`" -> use Story Bible/content tools.
- "Run `node scripts/...`" -> use `skill_script` only if approved, otherwise defer.
- "Use grep/search the repo" -> use `search_content`.
- "Update canon" without confirmation -> propose Story Bible changes.
- "Improve prose" without scope -> define line edit, rhythm, voice, clarity, or diction task.
- "Analyze story" without output shape -> define report, checklist, ranked issues, or next action.
- "Ask many questions" -> ask one targeted question at a time or provide a short decision set.

## Verification

For each rewritten skill:

1. Parse/import it through Skill Core or run the relevant parser test fixture.
2. Load it with the `skill` tool or `RenderForModel` path and inspect model-visible output.
3. Simulate the target agent task: would the agent know which Edda tools to call, what context to read, and what output to return?
4. Simulate a misuse case from `metadata.doNotUse`: would the agent route away?
5. Check that no source-only assumptions remain.

If the skill fails any verification step, revise it before moving to the next skill.
