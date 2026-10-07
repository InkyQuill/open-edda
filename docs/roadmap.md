# Open Edda Roadmap

Updated 2026-10-06 after [implementation audit](audit/2026-10-06-project-direction.md). Product authority: [ADR 0014](adr/0014-portable-projects-and-transactional-sync.md). The [previous milestone tracker](archive/roadmap-2026-07-01.md) is retained as history; its “Implemented” labels are not proof of end-to-end readiness.

## Direction and order

Open Edda is a web writing and translation workspace with versioned, portable projects. The author can work in the browser or take the same project to local Writing Tools and synchronize it back, without adopting git commit workflows or rearranging CWS folders.

CWS is the compatibility authority; Elysium is an obsolete legacy conversion, not a template. Prefer flat organization while permitting arbitrary directory trees. Distribution will be a Docker image with persistent volume/PVC data. [Storage recommendation and deployment contract](adr/0015-storage-and-container-deployment.md) separate the public file tree from physical server storage.

Delivery order: **projects → local synchronization → Pocket Editor → web Writing Tools and skill parity**. Basic usable UI and preservation of tool files belong to the first delivery; broad tool-feature and skill expansion do not.

## Verified baseline

| Existing area | Current status |
| --- | --- |
| Milestone 1: project core | Legacy content CRUD/revisions and Elysium import/export remain; new projects can use generic versioned file trees |
| Milestones 2–3.6: assistant and skills | Substantial implementation exists; not evidence of parity with current local CWS skills |
| Milestone 4: writing UI | Routed UI, editor adapter, assistant and review surfaces exist; baseline build restored; checkpoint labels still wrap item revisions |
| Milestone 5: file-first projects | **Partial foundation**, not completed integration: scanner, IDs, index, local drafts/saves, snapshots, conflicts and CLI state helpers exist |
| Milestone 5: network sync and web integration | **Local delivery implemented**: create/attach/get/send/take/move/history/restore, explicit conflicts and verified container backups; agent adaptation and Pocket provider integration are later stages |
| Go module identity | Renamed to `github.com/InkyQuill/open-edda` in the baseline slice |
| Pocket Editor, Galley Desk, Timeline Helper integration | No implemented Open Edda round-trip or equivalent web feature set established |

## P0 — Restore an honest, reproducible baseline

Build/module/CLI safety slice implemented on 2026-10-06; see [verification](audit/2026-10-06-baseline-repair.md). Revision/checkpoint UI terminology remains a separate follow-up before file-version integration.

- [x] Fix the Go compile failure and make Galley Editor dependency installation/build reproducible.
- [x] Rename the Go module and imports/generated references to `github.com/InkyQuill/open-edda`; verify tests, build and tooling. Preserve legacy configuration aliases intentionally rather than mixing them with module identity.
- [x] Until real transfer exists, CLI commands must not acknowledge an upload or clear its pending state; expose unsupported network operations truthfully.
- [x] New file projects expose project version history/restore directly; the older item-revision UI is outside the file-project flow.

Exit: current-source server/frontend run, checks pass, unreachable-server tests cannot report successful transfer. This gate is part of the project/sync delivery, not a new feature detour.

## P1 — Create and preserve portable projects (implemented)

Development policy (user clarification, 2026-10-06): no installed legacy projects exist. Do not spend delivery work on old-data conversion or backward compatibility. CWS and Writing Tools file interoperability remain required.

Internal storage foundation implemented on 2026-10-06: generic manifests, immutable objects, transactional versions/receipts and restore. See [implementation and verification](architecture/project-versions.md). The first web slice now creates blank file projects, browses folders, edits UTF-8 text and downloads files through the version API. The file UI now also imports selected files/folders, moves/removes entries, previews history and restores saved trees. CLI folder import into an empty project is implemented; see [import contract](architecture/folder-import.md).

- Create a blank project in the browser and register/import an existing folder/archive without forcing the old Edda layout.
- Separate complete file inventory from optional CWS role mapping. Preserve arbitrary nested paths, source languages, frontmatter, binary assets and tool metadata; show exclusions and unsupported entries before acceptance.
- Connect web reads/writes to one server project storage service (recommended: immutable bytes plus authoritative SQLite version manifests) and stable file/project identity. No legacy-project migration or backward-compatibility layer is required: the user confirmed on 2026-10-06 that no old Open Edda projects exist.
- Publish recoverable project transactions with automatic versions; support history, comparison and non-destructive restore.
- Expose a usable projects list and file/directory browser. Unknown formats remain downloadable and transferable even without a specialized editor.

Exit: browser-created and imported projects can be retrieved with matching paths/hashes; web changes appear in canonical storage; restart/reindex preserve content and history. Test sanitized snapshots shaped like current `alchemist` and `only-sense-online`, not only the old `alchemist-lite` fixture.

## P2 — Local-computer synchronization (implemented on Linux)

First network slice implemented: login/logout, project discovery, verified fresh checkout, offline status and transactional send with durable retry state. Real-server tests cover two independent copies, stale sends, exact binary/text bytes and a lost response followed by later local edits. [CLI contract and limits](architecture/cli-sync.md). Existing-copy `take` now merges independent file changes and records explicit conflict choices, with journaled application and process-exit recovery tests. The remaining local delivery is now implemented: CLI project creation/attachment, explicit stable-ID moves, history/restore, Docker/Compose, a Kubernetes template and verified online backup/fresh-volume restore. P2 is complete for the tested Linux/single-instance topology; production PVC and other operating systems are not qualified.

- Configure a base URL and credentials, discover/select/create the remote project and attach a local folder.
- Implement actual authenticated `get`/`send` and update retrieval, manifest comparison, staged transfer, receipts and idempotent retry.
- Preserve previous versions automatically; no required checkpoint note or manual commit step.
- Handle concurrent web/local edits, rename/delete conflicts, interrupted publication and edits from local tools during transfer.
- Make status clear: local changes, last confirmed server version, pending operations, offline/error and conflict resolution.

- Package the usable delivery as a reproducible Docker image with volume/PVC examples, single-instance storage requirements and a verified backup/restore procedure.

Verification: [P1/P2 acceptance record](audit/2026-10-06-files-sync-acceptance.md), [container/backup acceptance](architecture/deployment-and-backup.md), CLI HTTP and process-exit tests, and desktop/mobile browser tests pass. Remaining platform qualification is stated explicitly; no migration gate exists.

Exit: local → web → second local copy → web → first copy passes real network tests and path/hash comparisons. Timeout-after-commit retry creates no duplicate version; rejected/interrupted transfer leaves both sides recoverable. See [delivery plan and acceptance matrix](plans/2026-10-06-projects-and-sync.md).

## P3 — Pocket Editor synchronization

- Provider-independent source/account/project/directory selection in Pocket Editor; retain existing Yandex books and offline cache.
- Implement Open Edda adapter after the P2 protocol is stable. Keep Seafile and Dropbox as explicit follow-up adapters, not prerequisites to the first Edda round-trip.
- Transfer manifests and review sidecars with source/hash/anchor checks, preserving IDs and conflict states. Chapters remain read-only from Pocket Editor.
- Support several book directories inside one writing/translation project and concurrent review from desktop/mobile.

Exit: Edda book directory → offline Pocket review → Edda → local Galley Desk → Edda/Pocket round-trip retains edits, signals, comments, order and unresolved anchors without modifying source Markdown during review synchronization.

## P4 — Writing Tools in the browser, deferred

- Galley Desk equivalent reading/editing/review, chapter ordering/management, metadata and compatible review application; continue round-trip with the desktop app.
- Timeline Helper equivalent event/plotline/temporal editing, shared timeline/calendar projections and compatible file preservation; continue round-trip with the desktop app.
- Reuse shared Galley themes and domain contracts, not Electron-specific shell/IPC code. Establish a feature-parity matrix against the tool versions selected when this phase starts.

## P5 — Local-agent skill mechanics through Edda, deferred

- Inventory current local CWS skills and source versions; map each operation to Edda storage/tools.
- Preserve semantics, source authority, review gates, supporting assets and outputs. Test matching local/web scenarios; explicitly label unsupported helpers.
- Resolve built-in/global/project scopes and script capabilities as part of this phase.

P4 and P5 are later work; their relative implementation order can be selected after P3 without delaying projects or synchronization. Multi-author collaboration and a consistency dashboard remain deferred until the single-author round-trip is reliable.

## Completion policy

A helper, route stub, passing unit test or renamed UI label is not a completed workflow. Mark a phase complete only with current-source runtime evidence, adversarial recovery tests appropriate to its data risks, and a real cross-client round-trip. Record environment blockers and unverified behavior separately from confirmed defects.
