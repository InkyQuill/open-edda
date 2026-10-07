# Portable Projects and Transactional Synchronization

Date: 2026-10-06. Status: accepted product direction; implementation pending.

This decision supersedes the mandatory-layout policy in [ADR 0008](0008-elysium-layout-for-markdown-interoperability.md), clarifies the target in [ADR 0001](0001-database-source-of-truth.md), and extends [ADR 0013](0013-linear-checkpoints.md). It does not claim that the current server has migrated from database content to files.

## Product boundary

Open Edda is a self-hosted web workspace and versioned project repository for writing and translation. An author can create a project in the browser, obtain a local working copy, use local Writing Tools or agents, and synchronize the result back. Core storage, editing and synchronization must not require an AI provider.

The first delivery is project creation/import plus reliable local-computer synchronization. Pocket Editor follows. Web equivalents of Galley Desk and Timeline Helper, their deeper integrations, and agent-skill parity remain required later work.

## Preserve the project, interpret it separately

- A project is a portable tree of author-owned files, not a prescribed set of chapter and bible database rows. Existing CWS and CWS-compatible layouts remain in place. Prefer flat organization where useful without imposing depth or folder names; expose the actual directory tree. Elysium is historical legacy, not the compatibility target.
- Storage and versioning preserve selected files byte-for-byte regardless of whether Edda understands their format. Markdown interpretation, search and specialized views are derived capabilities; an unrecognized file must not silently disappear from transfer or history.
- Preserve project guidance, frontmatter, unknown fields, sources, translations, knowledge bases, plans, assets, review sidecars, timeline files and skill assets. Layout adapters may recognize these roles without moving files or rewriting content.
- Initial import/attachment must show the included and excluded inventory. Generated caches, credentials, machine-local settings, installed external skill links and large source assets need an explicit policy. Exclusions must be visible; no blanket hidden-directory exclusion or automatic upload of external symlink targets.
- `.edda/` holds Edda-owned identity/version metadata. Separate portable identity/history from local connection, credentials, queues and transient state. Do not overwrite CWS metadata or another tool's IDs.
- Symlinks and unsupported filesystem entries must be reported safely, without following paths outside the selected root. The initial implementation may reject or explicitly exclude unsupported links; silent partial success is unacceptable.

“Any CWS structure” means layout-independent storage and lossless round-trip of the selected project. It does not mean that every CWS version, plugin or helper can already execute in the browser.

## One save authority

Physical persistence is settled by accepted [ADR 0015](0015-storage-and-container-deployment.md): immutable file objects and authoritative SQLite version metadata. Flexible local files and a web directory tree do not require a mutable server checkout. The earlier DB-as-cache assumption is superseded.

The target server preserves file bytes and immutable project versions. The recommended storage design uses immutable file objects and authoritative transactional SQLite manifests; only derived indexes are rebuildable. Web edits, CLI transfers and later clients use the same project identity and mutation service. User clarification on 2026-10-06: there are no existing Open Edda projects outside development. Legacy data migration and backward compatibility are not requirements; prototype paths can be replaced directly. Preservation of imported CWS/Writing Tools files remains mandatory.

An accepted transaction publishes a coherent new project version with its predecessor retained. Readers see either the old or new version. Implementation may use staged snapshots and an atomic head switch or a recoverable journal; atomic per-file renames alone do not make a multi-file update transactional.

Normal save/send operations create history automatically. Naming a milestone is optional. Authors must not stage files or write commit messages to synchronize. Restoring an old state creates a new recoverable version; it must not erase later history or unsent work.

## Synchronization contract

Proposed protocol requirements (wire schema belongs to the implementation plan):

1. Authenticate a device against a configured base URL; select/create a stable remote project identity. Credentials stay outside portable project data.
2. Compare the client's acknowledged base version, local file manifest and current server version. Preview additions, edits, renames and deletions, including exclusions and unresolved conflicts.
3. Upload/download missing bytes into staging; verify sizes and hashes before publication. Transfer failure leaves the current project usable and pending work intact.
4. Publish with a base-version precondition and a stable operation ID. Repeating an operation after a timeout returns the original receipt rather than duplicating history. Advance acknowledged local state only after verified server acceptance.
5. Preserve base/local/server content for divergent edits, delete-versus-edit and rename collisions. Initially, explicit resolution is sufficient; no last-writer-wins or speculative prose merge.
6. Apply received changes recoverably on the local computer. Detect edits made by local tools during transfer and never overwrite them using a stale scan.

The CLI should offer simple configure/login, project selection, `get`, `status` and `send` workflows. Exact command spelling is not yet an API promise. The existing `save`, `take`, checkpoint and conflict helpers can be reused internally where their guarantees fit; the current command names do not establish a working network protocol.

## Pocket Editor and Writing Tools

Pocket Editor remains an offline reader/editorial overlay. It writes its manifest and review sidecars, not canonical chapter Markdown. Edda must support selecting a source, project and subdirectory; one project can contain multiple books or volumes. Source identity must include provider/account, project or container, and directory to avoid cross-source collisions.

Introduce a provider boundary in Pocket Editor and migrate Yandex Disk behind it, then add Open Edda. Seafile and Dropbox are later adapters to that same boundary, with capabilities and concurrency rules verified per provider. Do not promise identical atomicity or lock behavior across providers.

Preserve existing `.pocket-editor.json` / `*.review.json` schemas, stable IDs, source and selection hashes, UTF-8 byte anchors, comments, signals, edits and unresolved states. Synchronizing a review does not authorize applying all its proposals to prose. A separate review/apply transaction validates the current source and anchors.

Galley Desk and Timeline Helper files must already survive the first generic file round-trip. Later web features reuse their domain contracts: review/revision workflows for Desk; temporal entities, shared axis, plotlines, calendars and projections for Timeline Helper. Sharing the Galley Editor component alone does not implement either application.

## Skill parity, later

Adapt storage/tool access while preserving the source skills' workflow, source hierarchy, review gates, outputs, references and helper semantics. For each supported skill maintain a source-version and capability mapping and parity fixtures. Unsupported steps must be explicitly unavailable/deferred, not silently removed while claiming equivalence. Keep existing script approval/isolation boundaries; equivalent mechanics do not require granting arbitrary host-shell access.
