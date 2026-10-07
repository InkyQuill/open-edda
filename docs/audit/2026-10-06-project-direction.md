# Project Direction and Implementation Audit — 2026-10-06

Baseline: checkout `46d10d3`; pre-existing local change in `mise.toml` retained. This audit changes documentation only. Sibling repositories and author projects were inspected read-only. Findings describe this checkout, not every branch or deployed instance.

Follow-up: the [baseline repair](2026-10-06-baseline-repair.md) resolves the build/module/false-acknowledgement findings. The observations below preserve the original audit baseline; storage and real synchronization remain outstanding.

## Conclusion

The intended product direction is sound, but the previous roadmap overstates delivery. Open Edda has a substantial database-backed writing/assistant prototype and separate file-project foundations. It does not yet provide the portable, transactionally versioned web/local project repository requested by the author.

The immediate work is **project storage and local synchronization**, then **Pocket Editor**. Full web Galley Desk/Timeline Helper functionality and equivalent local-agent skill mechanics remain later commitments. See the corrected [roadmap](../roadmap.md), [decision](../adr/0014-portable-projects-and-transactional-sync.md), and [delivery plan](../plans/2026-10-06-projects-and-sync.md).

## Requirement-by-requirement evidence

| Requirement | Assessment | Implementation evidence and gap |
| --- | --- | --- |
| 0. Store complete writing/translation projects with history | Partial | `project/service.go` creates database project/content rows; `fileproject/checkpoint.go` implements local snapshots. They are not one server storage workflow. Elysium import transforms recognized Markdown rather than preserving an arbitrary project tree. |
| 1. Any CWS structure | No | `fileproject/layout.go:classify` accepts fixed top-level roots, two guidance names and `.agents/skills`; unknown files are omitted. Current CWS `kb/`, `work/`, `project.md`, translation/source trees and ordinary non-Markdown assets are not covered. |
| 2. Base URL + real CLI get/send, automatic transactional versions | CLI foundation only | `cmd/edda/main.go:runGet` initializes metadata; `runSend` calls `CompletePendingUpload`; `runTake` records a timestamp. No network transport, authentication, remote project discovery, receipt or sync API connects them. |
| 3. Go module identity | Incorrect | `go.mod` and imports still use `git.inkyquill.net/inky/writer`. Required identity: `github.com/InkyQuill/open-edda`. Rename is planned, not performed by this docs audit. |
| 4. Pocket Editor review synchronization and multiple sources | Absent in Edda | Pocket's `sync/SyncEngine.kt` depends directly on `YandexDiskGateway`; library selection is `RoomYandexBookLibraryData`. Existing review/manifest schemas are reusable, but provider/account/project/directory selection and an Edda adapter remain work in both repos. |
| 5. Usable UI in the Galley/Timeline direction | Partial, current runtime unverified | Current source has routed workspace, basic create/import, editor and assistant/review panels. It does not expose arbitrary folders/local sync; source critique below identifies task and layout gaps. |
| 6. Galley Desk and Timeline Helper web features and synchronization | Deferred, not implemented | Galley Editor is a frontend dependency, not Desk's review application. No corresponding Edda timeline/sidecar integration was found in app/project/fileproject/frontend surfaces. Generic transfer must preserve their files before full feature integration. |
| 7. Skills equivalent to local-agent mechanics, using Edda storage | Not established | `docs/skills/open-edda-skill-mechanics.md` explicitly narrows behavior to available DB tools; `agent/tools.go` remains content/revision-oriented. Imported/rewritten skills and an approved script runtime exist; no current CWS versioned parity matrix or local/web scenario equivalence was established. |

## Confirmed blocking findings

### P0: CLI acknowledges a transfer that never happens

An isolated fixture was initialized against `http://127.0.0.1:1`, saved, sent and taken. All commands returned exit 0. `send` printed `Sent checkpoint …` and removed the pending operation; `take` printed `Checked …`. Local state gained `lastSentCheckpointId` and `lastTakeAt` with no network call in either implementation. This is reproducible false acknowledgement, not an intermittent network failure. No author files were changed by this test.

Do not use the present CLI success output as evidence that a remote backup exists. First implementation gate: explicit unsupported transfer or real acknowledged transfer, with failed operations retained.

### P1: Server and file model are disconnected

`project/http.go:RegisterRoutes` exposes DB content and revision routes plus Elysium import/export. `project/service.go` and `agent/tools.go` use those DB-backed operations. The `fileproject` imports occur in the CLI, not in the web project mutation path. `frontend/src/features/review/ReviewDrawer.tsx` labels item revisions “Checkpoint”; `reviewThunks.ts` restores revision numbers through the old API.

Milestone 5 phase 7's own plan says it adds contracts while preserving DB paths. The previous roadmap treated that foundation as completed migration. Corrected status: contracts implemented; web/agent migration pending.

### P1: Current real layouts are not supported

Read-only `edda status` observations:

- `/home/inky/Yandex.Disk/writing/alchemist`: reports 17 story files and one guidance file; misses the current `kb/`, `work/`, `project.md` and other project material. Its current `project.md` identifies `story/chapters`, `kb/canon`, `kb/characters`, `kb/world` and `work/` roles. The old fixture's shape is no longer representative.
- `/home/inky/translations/only-sense-online`: fails reading `.agents/skills/hieronymus-bootstrap` with `is a directory` (a linked skill directory). Its schema-v2 `project.md` declares a translation series; the project contains `sources/`, `translations/`, `kb/`, `workarea/` and guidance. Even without that scan error, the classifier cannot preserve those trees.

Unknown files must be storage objects even without semantic interpretation. Link handling needs a safe explicit policy, not traversal into installed external skills or silent partial import.

### P1: Snapshots are useful but are not complete project transactions

`CreateCheckpoint` stages and publishes a snapshot directory, hashes files and retains prior snapshots. `RestoreCheckpoint` verifies snapshots before applying them. These are valuable foundations. However, checkpoint inclusion uses stable classified files: `.edda/project.json` is scanned but excluded by `ids.go:isStableIDFile`, and arbitrary project files are absent. Restore removes and replaces working files sequentially, without a recoverable whole-project publication boundary or automatic preservation of unsaved current work. Do not claim full-project recovery/atomic sync from per-file atomic writes alone.

## Verification performed

| Check | Result |
| --- | --- |
| `go test -tags sqlite_fts5 ./...` | Fails to build: `project/http.go:284`, `decodeJSON(r, &input)` lacks the required response-writer argument. Several independent packages, including `cmd/edda` and `fileproject`, pass. |
| Build CLI and isolated offline get/save/send/take experiment | CLI builds; reproduces false transfer success described above. |
| Read-only status on actual writing/translation folders | Partial inventory / linked-directory error, respectively. No initialization or mutation run on those folders. |
| `bun run test` in frontend | 15 files, 129 tests pass. |
| `bun run build` in frontend | Fails TS2307: `@inkyquill/galley-editor` module/type declarations unavailable in this checkout. Package is declared; this result alone does not establish a source API incompatibility. |
| Current-source authenticated browser workflow | Not verified because fresh builds fail. Existing June binary/dist fallback cannot log in (404); do not attribute this stale-artifact behavior to current source. |

No complete backend suite, production deployment, real remote transfer, Pocket device test or cross-tool round-trip passed in this audit. Existing tests validate foundations but do not justify the old completion labels.

## UI direction: Impeccable assessment

Method: dual-agent (A: `/root/ui_design` · B: `/root/ui_evidence`). Assessments were isolated; A completed before detector results entered synthesis. Mode: Operate. Target: `frontend/src/features/projects/ProjectsPage.tsx` with the workspace it opens. Scores are **provisional source judgments**, not measured usability or live visual acceptance.

Specificity: the present entry screen is a generic project CRUD launcher. The strongest product identity would be the author's recognizable project structure, dependable version history and clear local synchronization. Existing editor-centered routing is a useful base; AI configuration should not dominate initial project use.

| Heuristic | Score / 4 | Main source observation |
| --- | --- | --- |
| System status | 2 | Basic load/save states; no acknowledged project sync state |
| Real-world match | 1 | Language codes/content kinds/revision terminology |
| Control and freedom | 2 | Modes exist; no direct Projects return in workspace chrome |
| Consistency | 2 | Shared primitives; duplicate mode/context navigation |
| Error prevention | 1 | Draft replacement risk on navigation; runtime reproduction pending |
| Recognition | 2 | Labels exist; custom project tree absent |
| Efficiency | 1 | Large project cards, no project search; fixed panels |
| Minimalism | 2 | Quiet palette but nested permanent frames and duplicate controls |
| Error recovery | 2 | Error states/retry exist; limited task-specific recovery |
| Help | 1 | No clear existing-folder/sync model |
| **Total** | **16 / 40** | **Poor under the skill rubric; provisional, not a release score** |

Priority changes for the projects/sync delivery:

1. **P1 — Entry workflow:** `ProjectsPage.tsx:139–195` offers title/language and Elysium ZIP, not existing CWS project attachment. Create/open/connect should lead to a preserved file inventory and first-sync outcome. Suggested Impeccable workflow: `onboard`.
2. **P1 — Project navigation:** `WorkspacePage.tsx` and `ContextDrawer.tsx` impose four content kinds and additional World/Notes placeholders. Use the actual directory tree with optional semantic views. Suggested workflow: `clarify`.
3. **P1 — Draft safety:** `EditorFrame.tsx:128–144` hydrates selected content and `editorSlice.ts:65–83` resets dirty state. Preserve drafts across navigation/refresh and distinguish saved from synchronized. This is a source-backed risk, not reproduced data loss. Suggested workflow: `harden`.
4. **P1 — Prose width:** `workspaceSlice.ts:34–41` defaults to Assistant with 304/380px panels; `WorkspaceShell.tsx` shows them at `md`. Their nominal width plus rail leaves almost no center at 768px. Validate actual rendering after build repair; collapse tools before prose. Suggested workflow: `layout`.
5. **P2 — Reference design transfer:** shared semantic themes, independent UI/book typography and flat prose surfaces should replace nested permanent frames and competing typography rules. Use Galley Desk's actual transfer guide; avoid copying exact Electron dimensions. Suggested workflow: `distill`.

Strengths: short creation form with pending/error states; explicit loading/empty/retry handling; shared UI primitives and a distinct central editor. These support incremental correction, not a framework replacement.

Cognitive load: creation is moderate; workspace has competing mode controls, context tabs, filters and assistant setup. The concern is simultaneous decisions, not claiming each small list exceeds four choices. Emotional gap: creating a project should end with recognizable work safely available, rather than empty categories and assistant setup.

Persona checks: first-time authors lack an explanation of language/setup; experienced translators cannot recognize their existing folder structure; selected mode/filter state needs accessible semantics and keyboard verification. Contrast, focus restoration, mobile rendering and draft outcomes remain unverified.

The detector ran once over `frontend/src`: zero errors and one `overused-font` warning at `styles.css:10` (`Inter`). Later rules use Geist; the warning does not prove rendered typography or a usability defect. It does not detect the storage/onboarding problems. No live overlay is claimed.

Reference evidence: Galley Desk `apps/desktop/DESIGN.md`, `docs/design-system-transfer.md`; Timeline Helper `DESIGN.md`, `PRODUCT.md`. Transfer their quiet, content-first surfaces and progressive disclosure. Do not copy Timeline's product-specific exclusion of sync status: synchronization is central to Open Edda.

Questions skipped: the user already selected the product direction and requested documentation planning, not an additional design interview. The implementation brief should answer how the author sees that their folder is preserved, identifies originals versus translations, and knows that synchronization has been acknowledged.
