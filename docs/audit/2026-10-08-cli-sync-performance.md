# CLI send/take performance — 2026-10-08

## Diagnosis

The initial implementation already compared SHA-256 manifests, but did not use
those digests to avoid transferring content:

- Send hashed the local inventory, copied and hashed every file into pending
  staging, rescanned the tree, synced staged files, then uploaded every file.
- Take downloaded every file in both the remote version and the common base,
  even when identical verified bytes were available locally. Recovery also copied
  and repeatedly hashed the whole local tree, including root `.git` and
  `node_modules` that can never be synchronization units.
- Result preparation rewrote remote-selected files even when their contents
  matched the already copied local snapshot. These writes each required fsync.
- There was no phase or waiting output, so local work, HTTP waits and server
  publication verification were indistinguishable to the author.

This is a source and isolated runtime diagnosis, not a profile of the author's
production projects or server. The new stage timing exposes those remaining
machine-specific costs without sending author content to a profiler.

## Implemented

- Send stages/uploads only SHA-256 + length pairs absent from the acknowledged
  base; identical missing objects are uploaded once. The full manifest still
  carries IDs, renames, directories and deletions.
- Take reuses verified contents from its private local snapshot, prepared remote
  snapshot and files already downloaded in the same version. Reused files undergo
  the same length/digest check as network downloads. Mutable working files are
  not hardlinked into recovery data.
- Reserved root state/cache/environment entries stay untouched and outside
  recovery snapshots. Nested excluded data remains protected when its containing
  portable directory is swapped. Identical result files are not rewritten again.
- Default progress includes stage, path, elapsed time, counts, known stage
  percentages and bytes. Upload/download byte counts advance within a file;
  blocked requests emit waiting heartbeats. `--quiet` retains normal results and
  errors without progress. First-send scans also show progress, paused for prompts.
- Pending publication receipts, frozen changed bytes, full-tree concurrency
  verification and journaled rollback remain in force.

## Reproducible checks

`go test -tags sqlite_fts5 ./cmd/edda -run TestSyncTransferVolume -v`
uses two checkouts and an authenticated localhost HTTP server, four distinct
1 MiB files and an 11-byte chapter edit:

| Traffic | Before | After |
| --- | ---: | ---: |
| Send object bodies | 4,194,315 bytes | 11 bytes |
| Take file-content requests | 10 | 1 |
| Copy + rename + deletion of known contents | not optimized | 0 upload bytes / 0 content downloads |
| Two paths with identical new 9-byte content | not optimized | 9 upload bytes / 1 content download |

Illustrative cached localhost timings were about 13.8 → 6.5 ms for send and
29.4 → 23.1 ms for take; these are single observations, **not production speed
claims**. Byte counts and request counts are regression assertions.

Additional regressions cover retry after upload failure with unchanged files
absent from staging, later working edits remaining unsent, corrupt reused bytes,
correct deduplicated progress totals, waiting heartbeats, escaped paths, quiet
output, and preservation of nested secrets without copying root caches.

Passed:

- `go test -race -tags sqlite_fts5 ./...` (including existing lost-receipt,
  conflict, concurrent-edit and subprocess crash-recovery cases).
- `go vet -tags sqlite_fts5 ./...`.
- Built CLI/server Linux PTY acceptance via
  `scripts/check-cli-interactive.py`: first send, get/send/take, conflicts,
  history/restore, login/logout, create/attach/import, prompts/cancellation,
  exclusions, backup/restore and command help.

macOS/Windows checks are deferred until release, as requested. Production author
folders were not used for tests.

## Remaining costs and alternatives

This is file-object synchronization, not a block-delta protocol. Changed files
still transfer in full. Computing local SHA-256, validating concurrent changes,
private recovery copies and server object-integrity verification still require
local I/O; take can still be disk-bound on large tracked trees or nested caches.
No persistent metadata-only digest cache is trusted. Recovery snapshots remain
retained without automatic cleanup. A previously prepared plan containing root
cache snapshots may require `take --restart` before application.

For frequently changing large binaries, the next suitable extension is
content-defined chunks with independently hashed objects and a final verified
whole-file digest. It would reduce changed-file traffic and allow partial resume,
but requires an explicit server/storage/retention protocol. For interrupted sends,
a server missing-object query could additionally reuse uploads outside the base
manifest. Neither extension is claimed as implemented here.

## Local installation

The checked CLI was installed atomically at `~/.local/bin/edda` and its
SHA-256 matched the tested executable. The previous executable is retained at
`~/.local/state/open-edda/cli-backups/`.
Installed send/take help confirms the new flags. No server redeployment was needed
for the existing manifest/object API optimization.
