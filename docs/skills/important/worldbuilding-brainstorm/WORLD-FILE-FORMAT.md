# Worldbuilding File Format

## Structure

```md
# {Title}
---
status: canon
updated: YYYY-MM-DD
related:
  - Other File
---

## {Section}

Concise setting facts.
```

## Rules

- Use the title for the topic, species, institution, force, place, or concept.
- Keep metadata brief. Use `status`, `updated`, `related`, or `notes` when useful.
- Prefer durable facts over discussion notes.
- Keep sections focused. Split only when it helps future lookup.
- Use `## Open Questions` only when the user explicitly wants unresolved questions recorded.
- Avoid writing rejected alternatives unless the rejection itself is important canon.
- If an existing file lacks this structure, normalize only the file being edited and only as much as needed for the current change.

## Concision Standard

Each file should be short enough to scan but complete enough to preserve needed facts. If a topic grows into several unrelated concerns, propose splitting it into separate files before doing so.
