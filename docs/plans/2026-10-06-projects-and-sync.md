# Projects and Synchronization Delivery Plan

Status: baseline build/module/CLI safety slice implemented after the 2026-10-06 audit; object/tree/version storage and basic web file editing are implemented; CLI folder import into empty projects is implemented; the initial CLI download/send cycle is implemented; existing-copy update, explicit conflict resolution and process-exit recovery are implemented. See [verification](../audit/2026-10-06-baseline-repair.md). Scope and order are defined by [roadmap](../roadmap.md) and [ADR 0014](../adr/0014-portable-projects-and-transactional-sync.md).

Storage design: [ADR 0015](../adr/0015-storage-and-container-deployment.md) recommends immutable content objects plus authoritative SQLite manifests. CWS/flexible folder trees are the baseline; Elysium is historical only.

## Work packages

| Order | Concrete output | Main owners in the current tree | Acceptance |
| --- | --- | --- | --- |
| 0 | Reproducible builds, correct module identity, honest unsupported CLI behavior | `go.mod`, Go imports, `project/http.go`, frontend dependency setup, `cmd/edda` | P0 checks and offline negative test |
| 1 | Generic inventory and portable project metadata; explicit exclusion/link policy | `fileproject`, project import API | Arbitrary paths and byte-identical selected files; no silent omissions |
| 2 | Server object store and transactional project manifests, common storage/mutation boundary | `project`, `fileproject`, `store`, migrations, app configuration | Web create/edit/read uses canonical project storage; atomic publication and failure recovery |
| 3 | Recoverable version publication and restore | `fileproject`, server storage service | Coherent project version, previous versions retained, recovery after interruption |
| 4 | Authenticated project discovery and transactional transfer API | `auth`, `project`, `app` | Ownership checks, base-version preconditions, idempotency receipts, verified bytes |
| 5 | CLI configure/select/get/status/send/update and resolution | `cmd/edda`, client transport | Real server + two separate working copies pass round-trip and failure matrix |
| 6 | Project onboarding, tree, history and sync status in web UI | frontend projects/workspace/review | Nontechnical author can create/import, locate a directory and understand saved versus synchronized |
| 6.5 | Reproducible Docker image, data-root configuration, volume/PVC examples and verified backup/restart | image/build configuration, app/store | Single-instance mounted-volume round-trip and isolated restore |
| 7 | Pocket provider boundary, Yandex migration and Open Edda adapter | separate Pocket Editor repository plus Edda transfer contract | Offline sidecar round-trip and existing-book compatibility |

Keep these as reviewable increments. Packages 1–6.5 make up the first usable delivery; package 6 should evolve alongside storage, not become a separate redesign blocker. Do not start rewriting bundled skills or building timeline UI to compensate for missing storage.

Storage foundation progress: [project version storage](../architecture/project-versions.md) implements the internal portions of packages 2–3. A first web/API slice covers blank project creation, file/folder creation, tree browsing, text editing and downloads, plus draft recovery and stale-version rejection. Package 1 now has generic inventory, explicit exclusions and an atomic CLI import into an empty project. The first package 5 slice adds saved login, discovery, fresh checkout, offline status and durable sends. The next package 5 slice adds existing-copy updates, file-level conflict choices and journaled recovery. Packages 5–6.5 now include explicit rename identity, existing-folder attachment, web import/move/remove/history/restore, Docker/Compose and verified online backup/fresh-volume restore. Linux single-instance acceptance passes; other platforms and an actual production PVC remain unqualified. Package 7 is the next provider-integration stage. The user confirmed that there are no old Open Edda projects: legacy migration and backward compatibility are outside scope.

## Decisions to close in implementation specs

- Version manifest/head and receipt wire format; server transaction publication and crash-recovery mechanism.
- Device authentication, revocation and local credential storage using the existing single-user auth boundary.
- Include/exclude defaults and large-file limits, with an inspectable inventory. Preserve necessary CWS/tool metadata; distinguish recoverable journals from rebuildable caches and machine-local secrets.
- Safe link behavior and filename/case/Unicode portability across target systems. Unsupported paths must fail visibly before mutation.
- Existing prototype content/agent code can be replaced directly as features move to file storage; no conversion or legacy support gate is required.
- How web drafts survive refresh/navigation and remain distinct from saved versions; dirty work must survive synchronization and restore.

These are implementation choices still to be specified, not reasons to reopen the accepted product direction.

## Required verification matrix

Use isolated temporary roots and sanitized fixtures; real author folders are read-only audit inputs.

| Scenario | Required result |
| --- | --- |
| Empty web project, first local checkout, first chapter, send | Same stable project; automatic prior version; browser sees chapter |
| CWS writing layout: `project.md`, `story/chapters`, `kb`, `work`, guidance and skills | Retain paths, bytes, frontmatter, references and selected assets |
| Translation series: `sources`, `translations`, `kb`, `workarea`, multiple languages/volumes | No forced flattening or fiction-only schema; unknown files retained |
| Pocket manifest/reviews and Timeline `timeline.yaml` + entity folders | Files survive P2 even before specialized UI exists |
| Unreachable server / expired credentials / wrong project owner | No success receipt, no pending-state loss, no partial published project |
| Timeout after server commit and repeated same operation | One resulting version, same receipt, no lost local queue entry |
| Process stop mid-upload, mid-publication, mid-download/apply | Old or new coherent state; deterministic recovery; no mixed project |
| Two clients edit same file; web edit versus local deletion/rename | Preserve base and both variants; explicit resolution |
| External tool modifies file after inventory | Recheck/precondition rejects stale overwrite; local work retained |
| Binary file, non-ASCII paths, long names, unknown frontmatter | Exact bytes and supported path spelling retained |
| Symlink outside root, traversal path, filename collision | Explicit rejection/exclusion, no escape or silent omission |
| Restore with dirty/unsent changes | Preserve current work and history; restore produces another version |
| Rebuild derived search index / restart server | Authoritative DB manifests and immutable bytes remain intact; indexes rebuild without losing history |
| Restore backup to fresh Docker volume/PVC | DB plus all referenced objects restore; identities, paths, history and bytes verified |
| Pocket offline edits, source changed meanwhile | Review records retained; stale anchors marked, no blind prose mutation |

For each accepted scenario record source revision, command/API flow, resulting version IDs and path/hash comparison. Unit suites complement these checks; they cannot stand in for network or cross-application evidence.

## UI direction for the first delivery

Use Galley Desk's shared theme roles, independent UI/document typography, flat reading surface, contextual tools and keyboard focus conventions. Use Timeline Helper's progressive disclosure: create/open a project before exposing advanced modeling or service details. Match behavior and domain needs rather than copying exact desktop panel widths.

Project entry actions: create project, bring existing work, open recent project. A project contains an ordinary directory tree; optional CWS role views are an aid. Show the selected server and last acknowledged synchronization state near the project, with recoverable errors beside the relevant action. Keep provider/model/skill administration in settings and the assistant optional. The editor must retain useful width when panels collapse on smaller screens.

Reference authorities: Galley Desk `apps/desktop/DESIGN.md` and `docs/design-system-transfer.md`; Timeline Helper `DESIGN.md` and `PRODUCT.md` in the sibling repositories. Their current implementations are reference evidence, not dependencies on a particular machine path or permission to replace Open Edda's framework.
