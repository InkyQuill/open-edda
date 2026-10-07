# Project Version Storage

Status: storage foundation and first web/API slice implemented on 2026-10-06 under [ADR 0015](../adr/0015-storage-and-container-deployment.md). New file projects use this store; existing content and agent workflows retain the legacy service. CLI import, verified checkout/send, existing-copy updates and explicit file conflict resolution are implemented; see [CLI synchronization](cli-sync.md) for recovery behavior and remaining limits. The user confirmed there are no old deployed projects; legacy migration/backward compatibility are not required.

## Ownership and API

The feature lives in `project/`: `versions.go`, `version_tree.go`, `version_objects.go`. `VersionStore` takes a migrated SQLite database, an existing private data directory and configurable limits. The caller owns the DB; closing the store releases its filesystem handle. Existing project ownership in `story_projects` is reused.

| Operation | Contract |
| --- | --- |
| `UploadObject` | Stream bytes with expected hash/size; verify, flush and publish without replacing an existing object; register a reference for an authorized project |
| `Publish` | Submit a complete tree with operation ID and expected parent version; atomically publish or return a conflict |
| `Version` | Read a retained version or pin the current head; return its complete manifest |
| `OpenVersionFile` | Authorize through a committed project/version/entry, verify bytes before returning the reader |
| `Restore` | Publish a retained tree as another version, keeping its current predecessor and all old versions |

A first publication uses an empty expected version. A stale base returns `ErrVersionConflict`. An exact retry returns its original version even when newer versions exist. Reusing an operation ID with a changed base, message, tree or restore source returns `ErrOperationConflict`. Input entry order is immaterial. Canonical request fingerprints include the restore source to distinguish restore from an ordinary publication of identical bytes.

## Flexible trees

- Every entry has an explicit stable ID, relative path and kind: `file` or `directory`. There is no required CWS layout, category or file extension. Root-level files and nested/empty directories work equally.
- Directories are explicit; a child's parent must appear in the same manifest. The conceptual root is implicit. A complete empty tree is valid.
- Retain an ID for rename/move; assign a new ID for a copy. Kind is immutable for an existing ID. No heuristic identity matching from hashes.
- File contents, unknown frontmatter, binary bytes and path spelling are preserved. Case/NFC-equivalent path collisions are rejected using lowercase NFC comparison without rewriting accepted paths.
- Reject absolute/traversal paths, backslashes, control characters, common Windows reserved names/characters, trailing dot/space, paths over 4096 UTF-8 bytes and components over 255 bytes. Reject unsupported kinds including symlinks. This is an explicit portability floor, not certification of every filesystem's name equivalence rules.
- Default configurable limits: 64 MiB per object, 1 GiB of logical file bytes per version, 10,000 manifest entries. Copies count separately toward logical project bytes even when stored bytes deduplicate. Limits are constructor options; the HTTP API applies these defaults and an 8 MiB manifest body limit. Configuration UI and account-wide quotas remain future work.

These rules govern manifests. Generic inventory/staged CLI import now use `fileproject.ScanInventory` and `StageInventory`; the old fixed-layout `fileproject.Scan` is still used by prototype local commands and is not the import authority.

## Publication and persistence

The initial object directory is flat: `DATA/objects/<sha256>`. Upload staging uses random `.upload-*` files in that same directory. This implements the same-filesystem requirement without a separate staging mount. The data root must be service-owned and must not be edited by external tools.

Objects are size/hash-checked, synced, linked into their final names without overwrite, and the directory is synced before a DB reference is recorded. A deduplication hit is verified rather than trusted. `os.Root` confines operations; symbolic object directories and non-regular object files are rejected.

Migration `00007_project_versions.sql` adds:

- `project_objects`: per-project references to verified bytes;
- `project_versions`: immutable version metadata plus operation receipt/fingerprint;
- `project_version_heads`: currently published version;
- `project_tree_ids`: persistent entry identities/kinds;
- `project_version_entries`: complete tree for each version.

Foreign keys keep entries, objects and versions within their project. There is no public raw-hash reader: possessing another project's digest does not authorize reading or referencing it.

Publication verifies objects before opening a short SQL write transaction. Its first statement acquires the SQLite writer lock before reading/rechecking head and operation receipt. It inserts manifest/identities, advances head and records the receipt in one transaction. Readers pin one version before reading its entries. Objects are immutable and not collected, so verification outside the transaction remains valid under the service-owned-volume contract.

`store.Open` explicitly requests SQLite `synchronous=FULL` on connections. Publication also checks the actual transaction connection and refuses weaker modes. Interrupted uploads/publications can leave unreferenced objects or crash-left staging files; they cannot expose a partial committed tree. Automatic collection/retention is deliberately absent until pins and in-flight operations have a defined policy.

SQLite and referenced objects are both authoritative. Search may be rebuilt; version manifests cannot be reconstructed from bytes alone. Full backup/restore tooling and actual volume/PVC durability qualification remain required before deployment.

## Verification and limits

`project/versions_test.go` exercises real temporary files and SQLite databases, including:

- exact Markdown/binary/metadata bytes, flat/nested/empty trees, rename, delete, retained versions, restore and reopened handles;
- exact retries, different-payload operation reuse, stale bases and concurrent writers using separate DB handles;
- readers observing only whole pinned trees while publications proceed;
- unauthorized access and attempts to reference another project's known digest;
- bad paths, identity/type collisions, Unicode/case collisions, object/entry/project limits and zero-byte files;
- interrupted/invalid uploads, unavailable storage, cancellation, missing/corrupt objects and unsafe links;
- SQL failure mid-manifest rolling back versions, identities and head together;
- child-process exit inside the publication transaction and immediately after commit, followed by recovery/retry;
- weak durability rejection, concurrent object deduplication, and schema upgrade preserving existing content and revision history.

All Go packages pass the race-enabled suite and `go vet`; sqlc models were regenerated. Follow-up cases were checked separately after addition. Process-exit tests demonstrate process-crash recovery, **not** a physical power-loss simulation. Actual disk-full/CSI fault injection, performance at production scale, backup/restore tooling, deployment, multi-instance coordination and web/local round-trips remain unverified or unimplemented. No author folders were modified.

## First web/API slice

Migration 00008 gives projects a `storage_mode`: existing rows default to `legacy`; new browser projects explicitly request `files`. Creation commits a project and its empty version together. Legacy content creation is rejected for file projects; file routes reject legacy projects. There is no automatic conversion or inferred folder structure.

All routes below require authentication and project ownership:

| Request | Purpose |
| --- | --- |
| `POST /api/projects` with `storageMode: "files"` | Blank file project and initial manifest |
| `GET /api/projects/{id}/files/operations/{operationId}` | Read the immutable receipt after a lost response, even if head advanced |
| `GET /api/projects/{id}/files/versions/current` | Pin the current manifest |
| `GET /api/projects/{id}/files/versions/{version}` | Read an immutable manifest |
| `PUT /api/projects/{id}/files/objects/{sha256}` | Stream bytes with required Content-Length; verify size/hash |
| `POST /api/projects/{id}/files/versions` | Publish complete entries with expectedVersion and operationId |
| `GET /api/projects/{id}/files/versions/{version}/entries/{entry}` | Download verified bytes, attachment/no-sniff, no raw-hash reads |

The browser offers folder/file creation with explicit paths, tree navigation and a plain UTF-8 text editor (opening limit 1 MiB). Larger/binary files remain downloadable. Parents must exist. Specialized Markdown/timeline/review editing, arbitrary file uploads/import, rename/delete and history/restore controls are follow-ups. Versions are retained server-side but not yet browsable in UI.

A single unsaved draft per project is retained in the current tab's session storage, including its pinned version and retry operation ID. Navigation never silently switches an unsaved file. Reload restores the draft after authenticated project access; conflicts leave text intact with download and explicit discard/reload actions. This is tab recovery, not a durable backup; quota errors are shown and tab closure may lose unsaved text. A lost save response can be retried with the same operation ID without duplicating a version. Editing again creates a new operation and still requires the original base version.

`OPEN_EDDA_DATA_DIR` defaults to the parent of the configured DB path. It controls only immutable objects; set both paths when mounting a volume. Neither existing DB files nor legacy content move automatically.

Verification: `app/files_test.go` covers authenticated HTTP publication/download, ownership, limits, stale versions, retries, legacy separation and transactional project-creation rollback. `bun run test:files` builds a disposable real server with temporary SQLite/object storage and exercises the production frontend on desktop/mobile, including draft recovery, competing writes and a lost response. No writer folders are used.
