# Storage and Container Deployment

Date: 2026-10-06. Status: accepted by the user on 2026-10-06; storage foundation and basic new-project web editing implemented; CLI folder import implemented; deployment pending. CWS, flexible project trees, immutable file objects with authoritative SQLite version metadata, and Docker volume/PVC distribution are the agreed direction. The actual Kubernetes storage class remains to be qualified.

## Separate project shape from server persistence

The public project model is a flexible tree of files and directories, primarily serving CWS and CWS-compatible work. Flat organization within logical sections is the preferred starting point, never an enforced rule. Nested directories, multiple books/volumes, arbitrary filenames and unknown formats remain supported storage concepts. Preserve empty directories explicitly if selected for transfer. CWS recognition provides optional views; it cannot decide which files exist.

The flat CWS project format is the structural reference for examples, onboarding and new acceptance fixtures. CWS installation is optional. Prototype conversion endpoints have no migration/compatibility obligation.

A local working copy contains normal author-readable files. That does not require the server to mutate a live author-readable folder on every save. Users manipulate the server project through the API; a mounted data volume is persistence, not a second live editing interface.

## Options

| Option | Benefits | Costs for this project |
| --- | --- | --- |
| Mutable project folders; DB only indexes | Direct server-side inspection/editing; easy file export | Multi-file atomicity, crash recovery, concurrent edits and matching history require a new journal/publication system. Existing DB transactions do not cover filesystem writes. |
| All file bytes and version metadata in SQLite | One transaction domain; simplest coherent backup; reuses the current persistence foundation | Large source PDFs/images and retained versions enlarge database/backup work; streaming file exchange requires BLOB handling. Viable alternative, not inherently incompatible with CWS. |
| Immutable file objects + authoritative SQLite metadata | Stream arbitrary assets as files; reuse transactions for complete version publication; unchanged bytes can be shared between versions | Backup must include DB and referenced objects; ordering/durability must be tested; volume is not a normal writable project checkout. |

**Recommendation: immutable file objects plus SQLite metadata.** This fits both arbitrary source assets and the existing transactional backend without implementing a second mutable-file database. Start with complete version manifests and reuse unchanged objects; defer delta storage, compaction, distributed storage and generic storage-provider frameworks.

CWS itself does not force this choice: all three options can expose a compatible file tree. The deciding factors are transaction/recovery complexity and deployment/asset requirements.

## Authority and minimal data model

- Object store: exact immutable bytes addressed by SHA-256, with size validation. Server object paths derive only from validated digests, never directly from a client-supplied project path.
- SQLite: authoritative project ownership, stable identity, current version, retained version manifests, file/directory identities and paths, parent versions, transaction receipts and operational data.
- Derived: text search and semantic role indexes; these can be rebuilt. Project manifests/history cannot be recovered from opaque objects alone.
- A version manifest maps stable entry IDs and relative paths to object hashes (files) or directory entries. A rename changes the path while retaining identity; a copy creates another identity. New identities are not guessed solely from identical content.
- No imposed chapter/story-bible enum for storage. Existing `content_items` and `project_files` enums cannot be the generic tree model.
- Scope all object access through authorized project/version manifests. A known hash is not permission to retrieve another project's bytes.

Conceptual tables: projects/current head, immutable project versions, version entries, and idempotent operation receipts. These are a design sketch; choose exact schema and migration names in the first storage implementation. There are no old deployed projects (user clarification, 2026-10-06). Do not build a legacy converter or make backward compatibility a delivery gate.

## Save and transfer publication

1. Authenticate and validate ownership, relative paths, entry types, quotas and the requested base version. Assign or validate a stable operation ID.
2. Stream new bytes into staging on the same filesystem as the object store; verify hashes/sizes. Flush file contents, publish immutable objects atomically and persist required directory metadata before a DB commit can reference them. Reuse existing objects only after validating their stored identity/integrity.
3. Start a short SQLite write transaction. Recheck authorization/base head, insert the complete version manifest, advance head conditionally and store the operation receipt together. A reused operation ID with different request content is an error; an identical replay returns its existing receipt.
4. Commit durably, then acknowledge. The client advances its acknowledged base only from the confirmed receipt. If the response was lost, retry discovers the committed receipt.

Readers pin one version for a request/download and use only that manifest. No reader sees partly uploaded files as current. Previous versions stay addressable; restore creates another version, preserving both current history and conflicts.

There is no cross-resource SQL/filesystem transaction. Safety comes from ordering: durable immutable objects first, authoritative DB publication last. A crash before commit can leave unreferenced bytes, but must never expose a committed version whose bytes were not durable. Failed staging does not alter the current head. Missing/corrupt referenced objects cause a clear integrity error, not an invented empty file.

Initially do not implement automatic object/history deletion. A future collector must coordinate with in-flight writes, retained versions, export/backup pins and restore; do not race publication by collecting newly written unreferenced objects. SQLite WAL is crash recovery, not user version history.

## Docker and Kubernetes contract

Implemented image layout (see Dockerfile and deployment guide):

```text
/app/                 image: executable, migrations, built frontend
/var/lib/open-edda/   one persistent data root
  edda.db        authoritative database (+ adjacent WAL/SHM)
  objects/           immutable content bytes
  # Temporary object uploads are private files inside objects/.
```

Use a non-root service user with write access to the data root, persistent data outside the image layer, and secrets supplied separately through runtime configuration/Secrets. Do not put bootstrap passwords or provider keys into synced project files. `OPEN_EDDA_DATA_DIR` now controls immutable objects and defaults to the parent of `OPEN_EDDA_DB_PATH`. The database location is unchanged: set both variables explicitly for a shared volume (for example `/data` and `/data/edda.db`). Container packaging and verified online backup/restore are implemented; see [deployment and backup](../architecture/deployment-and-backup.md).

First supported topology: one service process / one pod, one volume, no horizontal replicas. On Kubernetes prefer `ReadWriteOncePod` with a supporting CSI driver. `ReadWriteOnce` alone permits multiple pods on the same node; replica count/rollout policy must prevent competing instances. Use a recreate-style rollout with explicit termination and storage handoff; test migration/restart behavior.

PVC names/access modes do not prove filesystem suitability. SQLite WAL requires same-host shared-memory/locking semantics; do not advertise NFS/SMB shared filesystem mounts as supported for the DB. A remote block device mounted as a local filesystem in one pod is a different case and can qualify after durability/locking tests. Select the actual storage class before deployment acceptance. If a shared network filesystem or multiple active replicas is a hard requirement, reconsider a transactional server database before implementing this design.

Set and verify durability settings (including `synchronous=FULL` for acknowledged versions) explicitly, alongside busy handling and short transactions. `store/open.go` now explicitly sets WAL, FULL synchronization and a busy timeout. Connection tests and publication checks enforce durability mode; deployment hardware/filesystem guarantees remain unqualified.

## Backup and recovery

The database and object store together are required. Rebuilding search is possible; losing the DB is not a supported way to recover project identities, trees or version history. This revises the DB-as-cache promise in ADR 0001 and the earlier storage paragraph in ADR 0014 under this accepted decision.

For the initial release use a documented quiesced backup of the whole data root, with clean DB shutdown and no writers, or a verified SQLite online snapshot followed by copying every object reachable from that snapshot while collection is disabled/pinned. Do not copy only the live main SQLite file. A backup is accepted only after isolated restoration verifies DB integrity, manifests, ownership and referenced byte hashes. Kubernetes volume snapshots alone are not proof of an application-consistent backup.

Keep backup policy separate from user-facing version history. No automatic retention deletion in the first slice. Test disk full, permission failure, corrupted/missing object, kill before/after DB commit, lost response, retry, and restore before release.

Implementation record: [project version storage](../architecture/project-versions.md) documents the API, schema, concrete limits and crash tests.

## First implementation slices

1. Baseline: repair Go/frontend builds, correct module identity, make unimplemented sync commands fail honestly. No storage migration or image release in this slice.
2. Generic project tree + immutable object/version service behind internal APIs, with temporary-volume tests for publication and conflict boundaries. Preserve existing DB behavior until integration is verified.
3. Connect web create/read/save to that service without a legacy migration layer; add minimal folder/file navigation, then the real sync protocol/client.
4. Publish a reproducible Docker image and volume/PVC examples with backup/recovery and restart tests as part of the first usable project/local-sync delivery, not as an unrelated later packaging task.

Pocket Editor and broader Writing Tools/skill features follow the existing roadmap. The next slice is deliberately small; do not combine backend replacement, UI redesign and container release into one change.

## Primary references

- [SQLite WAL](https://sqlite.org/wal.html): concurrency, same-host constraints and durability settings.
- [SQLite atomic commit](https://sqlite.org/atomiccommit.html): durability assumptions and filesystem flush behavior.
- [SQLite backup API](https://sqlite.org/backup.html): consistent database snapshots.
- [Kubernetes persistent volumes](https://kubernetes.io/docs/concepts/storage/persistent-volumes/): PVC access modes and the difference between RWO and RWOP.
