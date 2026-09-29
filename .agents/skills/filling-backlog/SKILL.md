---
name: filling-backlog
description: Use when working in this repository and noticing relevant out-of-scope work, deferred follow-ups, adjacent frontend/backend tasks, cleanup, risks, missing tests, docs gaps, or afterthoughts that should not be handled in the current task.
---

# Filling Backlog

## Overview

Capture relevant but out-of-scope findings in `docs/backlog.md` instead of dropping them or expanding the current task.

Core principle: if the thought is real project work, but not current-task work, write it down with enough evidence that a future agent can act on it.

## When to Use

Use this whenever you notice:

- A frontend/backend counterpart outside the current task.
- A cleanup, refactor, migration, docs gap, missing test, UX issue, or operational risk.
- A design or implementation thought that is relevant to the project but would widen scope.
- A likely bug or inconsistency found while reading code, but not needed for the requested change.

Do not use this for:

- Work required to complete the current task. Do it now.
- Vague opinions without evidence.
- Personal notes unrelated to the repository.
- Secrets, credentials, or sensitive data.

## Decision Rule

| Situation | Action |
|---|---|
| Needed for current task correctness | Implement or fix now |
| Relevant, concrete, not current scope | Add to `docs/backlog.md` |
| Interesting but unsupported by evidence | Ignore or investigate only if needed |
| Security issue with exploitability | Tell the user directly and add backlog only if deferred |

## Entry Requirements

Every backlog item must include:

- `Title`: concise action-oriented name.
- `Found while`: current task or file/context.
- `Why it matters`: project impact.
- `Evidence`: file paths, commands, observed behavior, or exact code references.
- `Not doing now because`: scope boundary.
- `Suggested next step`: first concrete action.
- `Status`: `open` unless the item is already tracked elsewhere.

If `docs/backlog.md` does not exist, create it before adding the entry.

## Format

Append new items under `## Open Items`:

```markdown
### <title>

- Status: open
- Found while: <current task/context>
- Why it matters: <impact>
- Evidence: <paths, symbols, command output summary, or observed behavior>
- Not doing now because: <scope boundary>
- Suggested next step: <smallest useful follow-up>
```

## Common Mistakes

| Mistake | Correction |
|---|---|
| "I'll remember this" | Add a backlog entry now |
| Turning the aside into extra implementation | Record it and return to the requested task |
| Writing "fix frontend" with no context | Include evidence and a first next step |
| Logging every random thought | Only record relevant, actionable project work |
| Hiding urgent risk in backlog only | Surface urgent risk to the user immediately |

## Completion Check

Before finishing any coding turn, ask: "Did I notice relevant work that I intentionally did not do?" If yes, verify it is in `docs/backlog.md` with evidence.
