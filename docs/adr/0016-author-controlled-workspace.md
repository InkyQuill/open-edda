# ADR 0016: Author-controlled book workspace

- Status: accepted from the author's instruction, 2026-10-10
- Extends: ADR 0014; takes precedence over any interpretation that imposes a writing lifecycle

## Decision

Edda is a workspace for the whole book: prose, KB, research, drafts, fragments and
other author material. The author chooses the folder purposes and names. No
folder, role, stage or companion tool is mandatory. Multiple custom sections may
refer to the same folder; several book folders can coexist in one project.

“Draft”, “main text”, “ready” and “WIP” are author descriptions, not permission
states. A book can remain WIP indefinitely. The author may start in a folder
called “ready”, rewrite it later, omit drafts entirely, or use their own layout.
Edda must never require a draft before a finished text, force a stage transition,
move files based on inferred readiness, or prevent writing/saving/exporting because
some preparation/review/KB milestone was not recorded. The author decides when
work is ready. Edda does not infer that decision from labels or review counts.

CWS integration may offer optional interpretation of configured directory roles.
It must not install CWS's workflow assumptions as application-wide gates. Any
future explicitly selected workflow is an author choice and must expose its
requirements; ordinary project use remains available independently. This clarifies
the earlier skill-parity references to source “review gates”: compatibility is not
permission to impose a lifecycle on every Edda project.

## Current implementation

`edda.workspace.json` is an optional ordinary versioned file at the project root:

```json
{
  "schemaVersion": 1,
  "sections": [
    { "id": "custom-section-id", "label": "Моя база знаний", "purpose": "knowledge", "folderId": "stable-tree-folder-id" }
  ]
}
```

`purpose` is optional (`knowledge`, `drafts`, `fragments`, `manuscript`, `custom`),
chosen explicitly and independent of the visible label. It describes content, not
readiness. Unknown purpose strings remain preserved; omitted values are custom.

The author assigns existing folders through “Настроить разделы книги”. The suggested
labels KB/drafts/fragments/main text can be renamed, omitted or extended. Saving
creates no content directories. Shortcuts filter the file tree to the assigned
folder; new files/import dialogs default to it. “Все файлы” always remains
available. A blank folder ID means the whole project. Folder IDs survive Edda
moves/renames; deleted folders remain explicitly unbound until the author chooses
a replacement. Importing into a different project with different IDs may require
reassignment. Unknown JSON fields are retained when editing supported settings;
unsupported schemas are not silently overwritten. The settings file participates
in normal synchronization, versions and conflict handling.

This is manual author configuration, not automatic CWS integration or a readiness
state machine. CWS role import and book/chapter metadata management remain separate
roadmap items. No global or project writing policy is inferred from this file.
