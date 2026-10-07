# Baseline Repair — 2026-10-06

Status: first implementation slice complete. Storage migration and actual network synchronization remain unimplemented. This follows the [initial audit](2026-10-06-project-direction.md) and the user's acceptance of [ADR 0015](../adr/0015-storage-and-container-deployment.md).

## Changes

- Go module and all application/test imports now use `github.com/InkyQuill/open-edda`. Historical audit/plan references to the old module remain as evidence, not active import paths. Existing legacy environment aliases are retained.
- The revision-restore HTTP handler now supplies the response writer to the bounded JSON decoder and uses the existing malformed/oversized-body error mapping. A regression test verifies that an oversized restore request returns 413 before reaching the mutation service.
- `edda get`, `send` and `take` return an explicit unsupported-network error and process exit 1. They cannot initialize a misleading downloaded project, remove pending uploads, record a successful send or advance a retrieval timestamp. Help and local-save output disclose the limitation.
- CLI regression tests cover absent/configured server URLs, two queued local versions, repeated send attempts, unchanged local file contents/state, and nonexistent destination folders. Existing local snapshot/history/restore/conflict helpers remain available.
- Frontend dependencies were restored with `bun install --frozen-lockfile`. The declared Galley Editor 0.10.0 package installed successfully; no package or lockfile changes or dependency upgrade were necessary. README now documents the frozen install.
- Existing formatting inconsistencies in the touched `agent/tools.go` file were normalized by gofmt. Build output names are ignored. The user's pre-existing `mise.toml` changes are retained.

## Verification

| Check | Result |
| --- | --- |
| Go formatting for tracked `.go` files | Clean |
| `go test -tags sqlite_fts5 ./...` | All 13 packages pass |
| `go test -tags sqlite_fts5 -race -count=1 ./...` | All 13 packages pass; changed restore-body test also rechecked with race detection after addition |
| `go vet -tags sqlite_fts5 ./...` | Pass |
| Staticcheck | Not installed; skipped |
| Server and CLI builds | Pass, canonical module path |
| Frozen frontend install | Pass without manifest/lock changes |
| Frontend unit tests | 15 files / 129 tests pass |
| Frontend production build | Pass; existing >500 kB chunk-size warning remains |
| Playwright workspace smoke | Desktop and mobile scenarios pass; two counterpart device cases intentionally skip. These tests mock the API. |
| Real HTTP runtime | Fresh temporary DB, migrations, built SPA/assets, bootstrap login, project/chapter creation, edit, revision restore, stale-write 409 and restart persistence all pass. Three revisions retained. |
| Built CLI regression | Against `http://127.0.0.1:1`, get/send/take exit 1 with no stdout success; snapshots, local content and pending state remain byte-identical |
| Documentation/diff checks | Local links and whitespace checked |

Runtime checks used temporary local roots and newly built binaries, not author projects or stale artifacts. Temporary processes, databases and binaries were removed. The live HTTP scenario validates the existing database-backed application; it is not evidence of file-store or local/server synchronization completion.

## Remaining boundaries

- UI review labels still call item revisions checkpoints; correcting those labels remains an explicit P0 follow-up before project-version integration.
- Existing old `lastSentCheckpointId` values cannot prove delivery; this change does not rewrite historical local state or pretend to repair an earlier false acknowledgement.
- No Docker image, PVC deployment, generic CWS tree migration or Pocket Editor adapter is delivered in this slice.
- Next implementation slice: generic file/directory identities, immutable objects and transactional project-version manifests, with isolated publication/recovery tests before web/API migration.
