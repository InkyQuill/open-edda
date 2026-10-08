# Backlog

Active delivery priorities live in the [roadmap](roadmap.md) and [projects/sync plan](plans/2026-10-06-projects-and-sync.md). The items below must not displace that order.

## CLI synchronization performance (requested 2026-10-08)

Delivered: digest-based object reuse in send/take, transfer deduplication within a version, default stage/byte/count progress with waiting heartbeats and `--quiet`, and omission of reserved root caches/repositories from recovery snapshots. [Measurements, checks and boundaries](audit/2026-10-08-cli-sync-performance.md).

Remaining candidates after measuring real workloads:
- Content-defined chunks for large changed binary files, with chunk hashes and verified final SHA-256; this reduces transfers within a changed file and enables partial resume, but requires new server/storage contracts.
- Reducing repeated local recovery reads/copies while retaining concurrent-edit and crash guarantees, especially for nested ignored caches. Do not substitute mtime alone for content verification.
- Negotiating missing objects across server history for interrupted uploads and copies not present in the acknowledged base.

## Writing Tools integrations (requested 2026-10-07)

Status (2026-10-08): INT-01 implemented in Pocket Editor with its recorded P3 round-trip; INT-02–04 delivered locally with [verification](audit/2026-10-08-files-editor-themes.md); INT-05–06 remain planned. Delivery follows roadmap P3–P5; the editor and theme tasks are independently reviewable parts of P4, not a claim of full Galley Desk parity.

Shared requirement: Open Edda, CWS, Hieronymus, Galley Desk, Timeline Helper and Pocket Editor are optional peers in the author's toolset. No tool, skill pack or integration is a prerequisite for using another. Enable integrations according to what the author has installed/configured; absent tools must not block ordinary project work or trigger their integration skills. Preserve unknown tool files and arbitrary folder layouts.

### INT-01 — Pocket Editor project/folder integration

- Phase: P3.
- Related repository: `/home/inky/Development/WritingTools/pocket-editor/`.
- Scope: one Pocket Editor project represents one book; one Edda project may represent an entire book series. When adding a book in Pocket Editor, the author selects an Edda project and then the folder containing that book. Each Pocket project binds to that pair. The Yandex model of one or two accounts is not the Edda binding model; shared credentials may serve multiple independent book bindings.
- Example: Edda project “Series translation” contains `book-01/` and `book-02/`; add them as two separate Pocket Editor projects by selecting the same Edda project and different folders.
- Acceptance: add two books from different folders of the same Edda series project as separate Pocket projects, and another book from another Edda project; switching between them preserves separate source identities, offline caches, review sidecars and sync/conflict state. Identity includes server/account, project and selected folder. Retain existing Yandex behavior and the P3 offline-review round-trip, including anchor checks; review synchronization must not alter canonical chapter Markdown.

### INT-02 — Live Galley Markdown editor

- Phase: P4.
- Package and repository: `@inkyquill/galley-editor`, `/home/inky/Development/WritingTools/galley-editor/`.
- Scope: integrate the live Markdown editor into the current file editor surface. Match the existing padding, text column width, typography and responsive layout; use a borderless surface without a nested framed editor/card.
- Acceptance: Markdown editing works in the current writing and source/translation workflows with keyboard access and both light/dark appearances. Preserve local drafts, explicit save, version history and conflict handling; switching files must not lose edits. Keep a suitable plain-text path for non-Markdown files. Verify the actual desktop/mobile workspace, not only an isolated editor demo.

### INT-03 — Edda theme in galley-themes

- Phase: P4; basis for INT-04 and visual alignment in INT-02.
- Related source: `galley-themes` in `/home/inky/Development/WritingTools/galley-editor/`.
- Scope: create a reusable theme from Edda's current approved appearance, using [DESIGN.md](../DESIGN.md) and the current frontend styles as the reference. Map semantic colors, typography and editor roles into the shared theme contract.
- Acceptance: the theme has matching light/dark variants, readable text/selection/focus states and consistent editor/background colors. Applying it reproduces Edda's current appearance without introducing a redesign or maintaining a separate copied palette catalog.

### INT-04 — Shared galley-themes in Edda

- Phase: P4; depends on INT-03 for the default Edda appearance.
- Related source: `galley-themes` in `/home/inky/Development/WritingTools/galley-editor/`.
- Scope: let authors select shared visual themes in Edda. Apply the same theme consistently to the application and Galley Editor, including menus, dialogs and supporting panes.
- Acceptance: persist theme selection; preserve light/dark/system mode and respond to system changes. Use the Edda theme as the default and a safe fallback for unavailable themes. Verify legibility, focus/selection states and desktop/mobile presentation across the shared theme choices.

### INT-05 — Optional bidirectional CWS integration

- Phase: P5; directory-role recognition is a distinct deliverable from full agent-skill parity.
- Related repository: `/home/inky/Development/WritingTools/creative-writing-skills/`.
- Scope: Edda reads CWS project metadata to understand directory purposes (for example, prose, sources, reference material and plans) without imposing folder names or moving files. CWS gains guidance/capabilities for working with Edda-backed projects, including local copies, synchronization, versions and conflicts.
- Acceptance: verify a CWS project with a custom layout and an ordinary project without CWS. Known directory roles are recognized when metadata is available; unknown directories remain accessible and preserved. CWS can identify an Edda-connected project and use supported workflows without overwriting ownership metadata or bypassing sync/conflict rules.
- Optionality: Edda must work without CWS or its skills; CWS must work without Edda. The same rule applies to Hieronymus, Galley Desk and Timeline Helper: use only integrations available in the author's setup, with no mandatory all-tools installation or skill dependency chain. Check both integration-present and integration-absent scenarios.

### INT-06 — Understand Pocket Editor sidecars in Edda

- Phase: P4, as part of reproducing Galley Desk review functionality in Edda. P3 must already preserve and synchronize these files losslessly; semantic review support is a future deliverable.
- Related repositories: `/home/inky/Development/WritingTools/galley-desk/` and `/home/inky/Development/WritingTools/pocket-editor/`.
- Scope: Edda must interpret Pocket Editor review sidecars, not only store/transfer them. Inventory the current sidecar contract and Galley Desk behavior before implementation; reproduce compatible reading, display, review decisions and application of proposed edits in the Edda workspace.
- Acceptance: use real compatible fixtures produced by Pocket Editor and consumed by Galley Desk. Verify chapter/source identity, comments, signals, proposed edits, source/selection hashes, anchors and review state. Stale or unresolved anchors remain visible for adjudication; they must not silently apply edits to changed text. Preserve unapplied entries and unrelated metadata when recording review decisions.
- Cross-client verification: review a book folder from an Edda series project in Pocket Editor, synchronize it, inspect and process its sidecars in Edda, then reopen in Galley Desk/Pocket Editor. Confirm compatible state without lost annotations, duplicate application or leakage into another book folder. Passive synchronization never applies proposed edits to canonical Markdown; applying an edit in Edda is an explicit review action with normal version/conflict safeguards. Galley Desk is the implementation reference, not a required installed app for authors using Edda review.

## Deferred Items

### Add layered skill scopes and settings controls

- Status: deferred to roadmap P5 (skill parity)
- Found while: Milestone 4 information architecture/settings work and follow-up planning for skill loading.
- Why it matters: Open Edda currently treats project skill loading as the main settings concern, but authors will need three distinct skill scopes: built-in skills shipped with the app, global user-written skills available across projects, and project-local skills. System settings should manage global skill sources, while project settings should control which global and local skills are enabled for each project. Without this split, skill availability and enablement state will become ambiguous as soon as users install reusable personal skills.
- Evidence: `frontend/src/features/settings/SettingsPage.tsx` now hosts provider/model, skills, and script runtime administration; `frontend/src/features/skills/skillsThunks.ts` currently exposes project/session loading paths only; `docs/roadmap.md` Milestone 4 Phase 3.5 moved skill administration into settings, and Phase 4 is already scoped to assistant actions rather than skill-scope design.
- Not doing now because: The accepted 2026-10-06 priority is project storage and local synchronization, then Pocket Editor. Skill scopes belong with later skill parity rather than expanding the first delivery.
- Suggested next step: After projects/local sync and Pocket Editor, add a dedicated settings/project-settings plan that defines built-in, global user, and project-local skill sources; global enabled/disabled defaults; per-project enablement overrides; migration behavior for existing project skills; and UI/API changes for managing those scopes.
