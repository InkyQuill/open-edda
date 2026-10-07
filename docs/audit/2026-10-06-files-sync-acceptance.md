# File projects and local synchronization acceptance

Scope: Roadmap P1/P2, Linux, one Open Edda instance. Source: base commit `46d10d3b2e7ce0969e753e1a4498f829e7d03c8b` plus the current uncommitted file/synchronization implementation. No release, push, real author-project upload or production deployment was performed.

## Delivered behavior

- Flexible, versioned file/folder manifests and immutable verified bytes, including arbitrary CWS/translation paths, unknown files, sidecars, binary assets and empty directories through the CLI.
- Web blank projects, text editing with draft recovery, selected-file/folder import, move/rename with stable IDs, removal, paginated history, historical preview and non-destructive restore.
- CLI saved login, discovery/create, attach with persistent exclusions, verified get (including historical versions), offline status, staged send, in-place take, explicit conflict choices, journaled recovery, stable-ID moves, history and server restore preserving unsent local files.
- Docker/Compose, single-instance Kubernetes template, consistent online backup, corruption verification and installation into a fresh data root.

## Checks

- `go test -race -tags sqlite_fts5 ./...`: all 14 packages passed.
- `go vet -tags sqlite_fts5 ./...`: passed.
- `bun run build`, `bun run test`: frontend build and 129 tests passed.
- `bun run test:files`: four real-server Playwright scenarios passed across desktop/mobile. Includes Unicode/binary import, identity-preserving move, deletion, historical preview, restore, draft recovery and concurrent-save refusal. Screenshots were inspected; no horizontal overflow. Impeccable mechanical detector reported no findings for the modified file UI.
- Update recovery subprocesses exited after backup, installation, before base commit and after base commit. Retry restored the previous working tree or finalized the committed tree. Separate coverage retained edits made after interruption.
- Real `ENOSPC` on an 8 MiB container tmpfs rejected an upload without advancing the remote head or losing the pending local snapshot. After freeing space, retry and a second download returned the exact bytes.
- Online backup tests verified retained history, all referenced hashes, rejection of corruption and refusal to overwrite an existing restore destination.

## Container evidence

Run `docker build -t open-edda:sync-test .` then `python3 scripts/verify-container.py open-edda:sync-test` to reproduce with disposable resources.

Verified image: `sha256:5e94230950f72391ae959e93d1f68019944d2196ce916da97dc18e439db80d87`.

Latest acceptance project: `project-1791286274629373208-1`.
Original publication: `da411657e80de8376f8a75cac1871ab3`.
Restored-volume history, newest first:

1. `7802215d3273f8364a9856a47c15d3e1` — new version restoring the original tree.
2. `f600875dba6e1512970d4e117cf893df` — second-copy rename/edit retained after abrupt server restart and backup/restore.
3. `da411657e80de8376f8a75cac1871ab3` — original uploaded tree.
4. `version-1791286274629519168-2` — initial empty project.

The check compared Unicode text and binary bytes after download, rename, restart, fresh-volume restore and historical restore. It ran the image as UID/GID 10001. Its containers, volumes and temporary credentials were removed after verification; the local image remains available.

## Deliberate boundaries

No legacy-data conversion is required. No automatic line merge, forced overwrite, history/object garbage collection, device-specific tokens, keyring storage or resumable partial download is claimed. Browser folder selection cannot enumerate empty directories; use CLI attachment/import for exact directory preservation. Local application is recoverable but not atomically visible to external editors; pause writers during take/move. Excluded local files are retained, and removal of their parent is refused.

A real production PVC/CSI driver, hardware power loss, Windows and macOS are not qualified. Kubernetes YAML is a deployment template. Pocket provider integration and the deferred Writing Tools/agent feature sets are distinct roadmap stages; successful local synchronization does not claim those stages are complete.
