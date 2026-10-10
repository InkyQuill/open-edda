# Open Edda Roadmap

Updated 2026-10-10 with milestone checklists and the author-approved integration backlog; baseline established after [implementation audit](audit/2026-10-06-project-direction.md). Product authority: [ADR 0014](adr/0014-portable-projects-and-transactional-sync.md). The [previous milestone tracker](archive/roadmap-2026-07-01.md) is retained as history; its “Implemented” labels are not proof of end-to-end readiness.

## Direction and order

Open Edda is a web writing and translation workspace with versioned, portable projects. The author can work in the browser or take the same project to local Writing Tools and synchronize it back, without adopting git commit workflows or rearranging CWS folders.

The flat CWS project format is the structural reference. CWS integration remains optional; preserve existing layouts and permit nested directories, including separate book folders within a series. Distribution will be a Docker image with persistent volume/PVC data. [Storage recommendation and deployment contract](adr/0015-storage-and-container-deployment.md) separate the public file tree from physical server storage.

Delivery order: **projects → local synchronization → Pocket Editor → web Writing Tools and skill parity**. Basic usable UI and preservation of tool files belong to the first delivery; broad tool-feature and skill expansion do not.

Integration principle (2026-10-07): Edda, CWS, Hieronymus, Galley Desk, Timeline Helper and Pocket Editor are independently usable writing tools. Enable integrations only for tools installed/configured by the author; no tool or integration skill is mandatory for another. Plain file projects remain fully usable without CWS. See the [integration backlog](backlog.md#writing-tools-integrations-requested-2026-10-07) for scope and acceptance criteria.

Author control (2026-10-10): folder purposes and readiness belong to the author.
KB, drafts, fragments and main text are optional, configurable navigation roles;
no draft-before-ready requirement or automatic lifecycle gates. See [ADR 0016](adr/0016-author-controlled-workspace.md).

Checklist: `[x]` = delivered with recorded verification; `[ ]` = remaining or not yet verified. Checks below summarize existing evidence, not a fresh runtime test run. A completed foundation does not mark later tool integration complete.

## Verified baseline

| Existing area | Current status |
| --- | --- |
| Milestone 1: project core | Legacy content CRUD/revisions and Markdown conversion remain; new projects can use generic versioned file trees |
| Milestones 2–3.6: assistant and skills | Substantial implementation exists; not evidence of parity with current local CWS skills |
| Milestone 4: writing UI | Production projects/file workspace, source comparison, drafts, history, media preview and light/dark/system appearance delivered; live Galley editor and shared themes delivered; broader tool parity remains P4 |
| Milestone 5: file-first projects | Portable versioned file projects, immutable storage, web editing/import and CLI delivered; ZIP import and file-version comparison delivered on 2026-10-08; CWS semantics remain P5 |
| Milestone 5: network sync and web integration | **Linux delivery implemented**: create/attach/get/send/take/move/history/restore, explicit conflicts, container backups and real local-path PVC/restart/restore checks; agent adaptation and Pocket provider integration are later stages |
| Go module identity | Renamed to `github.com/InkyQuill/open-edda` in the baseline slice |
| Pocket Editor, Galley Desk, Timeline Helper integration | Pocket Editor P3 round-trip recorded in its 2026-10-07 audit; Galley live editor/themes now integrated; semantic review and Timeline parity remain |

## P0 — Restore an honest, reproducible baseline

Build/module/CLI safety slice implemented on 2026-10-06; see [verification](audit/2026-10-06-baseline-repair.md). New file projects use project-version history/restore; the older structured item-revision surface remains separate.

- [x] Fix the Go compile failure and make Galley Editor dependency installation/build reproducible.
- [x] Rename the Go module and imports/generated references to `github.com/InkyQuill/open-edda`; verify tests, build and tooling. Preserve legacy configuration aliases intentionally rather than mixing them with module identity.
- [x] Until real transfer exists, CLI commands must not acknowledge an upload or clear its pending state; expose unsupported network operations truthfully.
- [x] New file projects expose project version history/restore directly; the older item-revision UI is outside the file-project flow.

Exit: current-source server/frontend run, checks pass, unreachable-server tests cannot report successful transfer. This gate is part of the project/sync delivery, not a new feature detour.

## P1 — Create and preserve portable projects (core delivered; follow-ups remain)

Development policy (user clarification, 2026-10-06): no installed legacy projects exist. Do not spend delivery work on old-data conversion or backward compatibility. CWS and Writing Tools file interoperability remain required.

Internal storage foundation implemented on 2026-10-06: generic manifests, immutable objects, transactional versions/receipts and restore. See [implementation and verification](architecture/project-versions.md). The first web slice now creates blank file projects, browses folders, edits UTF-8 text and downloads files through the version API. The file UI now also imports selected files/folders, moves/removes entries, previews history and restores saved trees. CLI folder import into an empty project is implemented; see [import contract](architecture/folder-import.md).

- [x] Create blank projects in the browser; import selected files/folders and attach/import existing local folders through the CLI without forcing a new layout.
- [x] Preserve complete file inventory: nested paths, source languages, frontmatter, binary assets and tool metadata, with explicit exclusions/unsupported entries. CLI preserves empty directories; browser folder selection has the documented limitation.
- [x] Connect web reads/writes to immutable bytes plus authoritative SQLite version manifests and stable file/project identity. No legacy-project migration is required.
- [x] Publish recoverable transactions with automatic versions, paginated history, historical preview and non-destructive whole-project restore.
- [x] Deliver the production projects/file browser, text editing with per-file drafts, source/translation comparison, drag/drop and accessible management alternatives, media previews and binary download pages. See [UI acceptance](design/production-acceptance.md).
- [x] Import ZIP archives as complete portable project trees with preview, explicit exclusions, empty directories and one transactional publication. Other archive formats remain binary files. See [2026-10-08 verification](audit/2026-10-08-files-editor-themes.md).
- [x] Compare saved project trees and file text in the current workspace, including moves, additions and removals; download both versions of binary/large files. See [verification](audit/2026-10-08-files-editor-themes.md).
- [ ] Recognize optional CWS directory roles — tracked under P5 / INT-05, separate from complete file preservation.

Exit: browser-created and imported projects can be retrieved with matching paths/hashes; web changes appear in canonical storage; restart/reindex preserve content and history. Test sanitized snapshots shaped like current `alchemist` and `only-sense-online`, not only the old `alchemist-lite` fixture.

## P2 — Local-computer synchronization (implemented on Linux)

First network slice implemented: login/logout, project discovery, verified fresh checkout, offline status and transactional send with durable retry state. Real-server tests cover two independent copies, stale sends, exact binary/text bytes and a lost response followed by later local edits. [CLI contract and limits](architecture/cli-sync.md). Existing-copy `take` now merges independent file changes and records explicit conflict choices, with journaled application and process-exit recovery tests. The remaining local delivery is now implemented: CLI project creation/attachment, explicit stable-ID moves, history/restore, Docker/Compose, a Kubernetes template and verified online backup/fresh-volume restore. P2 core delivery is complete for the tested Linux/single-instance topology, including the real local-path PVC deployment. Other storage drivers, hardware power failure and other operating systems remain unqualified.

- [x] Configure a base URL and credentials, discover/select/create the remote project and attach a local folder.
- [x] Implement actual authenticated `get`/`send` and update retrieval, manifest comparison, staged transfer, receipts and idempotent retry.
- [x] Preserve previous versions automatically; no required checkpoint note or manual commit step.
- [x] Handle concurrent web/local edits, rename/delete conflicts, interrupted publication and edits from local tools during transfer.
- [x] Make status clear: local changes, last confirmed server version, pending operations, offline/error and conflict resolution.
- [x] Package the usable delivery as a reproducible Docker image with volume/PVC examples, single-instance storage requirements and a verified backup/restore procedure.
- [x] Verify the real local-path PVC deployment: two-copy round-trip, pod restart, historical restore and independent backup recovery; verify public HTTPS and deployed frontend. See [deployment evidence](architecture/yggdrasil-deployment.md) and [UI acceptance](design/production-acceptance.md).
- [x] Address the reviewed CLI/conflict/history defects with regression coverage; see [2026-10-07 review disposition](audit/2026-10-07-coderabbit.md).
- [x] Avoid unchanged file transfers in CLI send/take using manifest digests and verified local reuse; show stages, counts/percentages and waiting heartbeats. Root repositories/caches are excluded from recovery snapshots. [Evidence and limits](audit/2026-10-08-cli-sync-performance.md).
- [ ] Qualify Windows and macOS local synchronization **at release**, deferred by the author on 2026-10-08; not a prerequisite to implementing remaining functionality.
- [ ] Qualify hardware power-loss recovery and any additional production storage/CSI driver before claiming support for those environments.

Verification: [P1/P2 acceptance record](audit/2026-10-06-files-sync-acceptance.md), [container/backup acceptance](architecture/deployment-and-backup.md), CLI HTTP and process-exit tests, and desktop/mobile browser tests pass. Remaining platform qualification is stated explicitly; no migration gate exists.

Exit: local → web → second local copy → web → first copy passes real network tests and path/hash comparisons. Timeout-after-commit retry creates no duplicate version; rejected/interrupted transfer leaves both sides recoverable. See [delivery plan and acceptance matrix](plans/2026-10-06-projects-and-sync.md).

## P3 — Pocket Editor synchronization (implemented in Pocket Editor; recorded verification)

The current Pocket Editor checkout includes the provider/binding implementation and `docs/audit/2026-10-07-edda-p3.md` records Android/Edda/CLI/Galley-codec round trips. These are prior recorded checks, not a fresh Android run in this checkout. See [reconciliation](audit/2026-10-08-files-editor-themes.md#p3-reconciliation).

- [x] Provider-independent source/account/project/directory selection in Pocket Editor; retain existing Yandex books and offline cache. One Pocket Editor project is one book; an Edda project may contain a whole series. Adding a book in Pocket Editor must offer selection of an Edda project and then a book folder within it; each Pocket project binds to that pair, rather than only to a global account. Keep several independent bindings, including several folders in one project, with isolated cache/review/sync state ([INT-01](backlog.md#int-01--pocket-editor-projectfolder-integration)).
- [x] Implement Open Edda adapter after the P2 protocol is stable. Keep Seafile and Dropbox as explicit follow-up adapters, not prerequisites to the first Edda round-trip.
- [x] Transfer manifests and review sidecars with source/hash/anchor checks, preserving IDs and conflict states. Chapters remain read-only from Pocket Editor. Preserve the sidecar contract for later semantic review in Edda under P4 ([INT-06](backlog.md#int-06--understand-pocket-editor-sidecars-in-edda)).
- [x] Support several book directories inside one writing/translation project and concurrent review from desktop/mobile.

Exit: Edda book directory → offline Pocket review → Edda → local Galley Desk → Edda/Pocket round-trip retains edits, signals, comments, order and unresolved anchors without modifying source Markdown during review synchronization.

## P4 — Writing Tools in the browser (foundation delivered; integrations remain)

- [x] Deliver the quiet writing workspace and light/dark/system appearance that the live editor and shared themes will extend; verified in [production UI acceptance](design/production-acceptance.md). Markdown now uses Galley live editing; other UTF-8 files retain plain text.
- [x] Integrate the live Markdown editor `@inkyquill/galley-editor` into the current file workspace, preserving its padding, text width and borderless surface, plus existing drafts/save/history/conflict behavior ([INT-02](backlog.md#int-02--live-galley-markdown-editor)).
- [x] Create an Edda theme in `galley-themes` from the current approved light/dark design ([INT-03](backlog.md#int-03--edda-theme-in-galley-themes)).
- [x] Support shared `galley-themes` visual themes across Edda and its editor ([INT-04](backlog.md#int-04--shared-galley-themes-in-edda)).
- [x] Display existing Pocket/Galley reviews in both editing and review modes; navigate highlights, explicitly accept/reject proposals or close signals, preserve drafts and unknown metadata, and publish chapter + sidecar atomically ([INT-06 delivery](audit/2026-10-10-pocket-review.md)).
- [x] Create/edit review signals, proposals and chapter notes; undo/redo individual review decisions without restoring unrelated project files. An ordinary Markdown file can opt into review directly in Edda; compatible manifest/sidecar metadata is created explicitly.
- [x] Let the author assign custom KB/draft/fragment/main-text sections to existing folders, with no mandatory roles or readiness prerequisites ([author-control contract](adr/0016-author-controlled-workspace.md)).
- [x] Manage book/chapter titles and chapter order; retain missing chapters, explicitly add/remove existing files from the spine, edit arbitrary chapter YAML while preserving body bytes and review anchors. No readiness or folder prerequisites. See [delivery](audit/2026-10-10-book-management.md).
- [ ] Galley Desk equivalent reading/editing/review, chapter ordering/management, metadata and compatible review application; continue round-trip with the desktop app. Existing sidecar reading/decisions/application are delivered; annotation creation/editing and per-chapter session undo/redo are now delivered; book/chapter titles, ordered navigation, explicit existing-file addition/removal and YAML metadata editing are delivered locally. Atomic chapter creation and verified adoption of existing sidecars are also delivered; packaged-client qualification remains; reference: `/home/inky/Development/WritingTools/galley-desk/` ([INT-06](backlog.md#int-06--understand-pocket-editor-sidecars-in-edda)).
- [ ] Timeline Helper equivalent event/plotline/temporal editing, shared timeline/calendar projections and compatible file preservation; continue round-trip with the desktop app.
- [ ] Reuse shared Galley themes and domain contracts, not Electron-specific shell/IPC code. The [initial parity matrix and source snapshots](plans/2026-10-08-writing-tools-parity.md) record current delivery and remaining contracts; extend it with the detailed Timeline inventory.

## P5 — Local-agent skill mechanics through Edda (integration/parity deferred)

- [ ] Establish optional, bidirectional CWS integration: Edda understands directory roles from CWS project metadata; CWS understands how to work with Edda projects and synchronization. Preserve arbitrary layouts and standalone use of every tool; integration skills activate only for installed/configured tools ([INT-05](backlog.md#int-05--optional-bidirectional-cws-integration)).
- [ ] Inventory current local CWS skills and source versions; map each operation to Edda storage/tools.
- [ ] Preserve semantics, source authority, supporting assets and outputs within author-selected workflows; never impose source-tool lifecycle gates on ordinary Edda work ([ADR 0016](adr/0016-author-controlled-workspace.md)). Test matching local/web scenarios; explicitly label unsupported helpers.
- [ ] Resolve built-in/global/project scopes and script capabilities as part of this phase.

The remaining P4 integrations and P5 are later work; their relative implementation order can be selected after P3 without delaying projects or synchronization. Multi-author collaboration and a consistency dashboard remain deferred until the single-author round-trip is reliable.

## Completion policy

A helper, route stub, passing unit test or renamed UI label is not a completed workflow. Mark a phase complete only with current-source runtime evidence, adversarial recovery tests appropriate to its data risks, and a real cross-client round-trip. Record environment blockers and unverified behavior separately from confirmed defects.
