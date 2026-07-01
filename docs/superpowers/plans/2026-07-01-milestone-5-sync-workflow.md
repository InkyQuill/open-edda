# Milestone 5 Phase 5: CLI Sync Workflow

## Goal

Add the local sync workflow skeleton around `.edda/state.local.json`: server connection metadata, device-local state, pending upload records, and retry-oriented `get`, `save`, `send`, and `take` commands.

## Context

- `.edda/project.json` already stores `serverUrl`.
- `.edda/state.local.json` is ignored by layout scans and is intended for device-local sync cursors and pending upload state.
- Checkpoints are now project-wide snapshots under `.edda/checkpoints/`.
- A full remote HTTP exchange contract is not defined yet; this phase should make the local workflow deterministic without inventing an unstable network API.

## Implementation Steps

1. Add sync state primitives in `fileproject`.
   - Read/write `.edda/state.local.json`.
   - Generate a stable local device ID.
   - Track `lastSentCheckpointId`, `lastTakeAt`, and a single pending upload with checkpoint ID, attempts, and last error.

2. Add CLI sync commands.
   - `edda get URL [path] --title "Title"` initializes a local Edda folder connected to a server URL.
   - `edda save [path] "Note"` creates a named checkpoint and records it as pending upload.
   - `edda send [path]` retries/completes pending upload state when a server URL exists.
   - `edda take [path]` records a server take attempt in local sync state.

3. Preserve file-save compatibility.
   - Keep `edda save [path] --id ...` for Phase 3 canonical file writes.
   - Route `edda save [path] "Note"` to checkpoint-plus-pending-upload behavior only when no file-save flags are present.

4. Add tests.
   - Sync state round-trips, initializes missing state, and records pending upload retry attempts.
   - CLI `get/save/send/take` exercises the local workflow.
   - Commands fail clearly when server metadata is missing.

5. Update roadmap and verify.
   - Mark Phase 5 implemented.
   - Run Go tests, frontend unit tests, build, smoke tests, and `git diff --check`.

## Review Notes

- This phase intentionally does not define the remote HTTP protocol. It gives the CLI, metadata, and retry semantics a stable local shape so the later server implementation can fill in transport without changing the author-facing workflow.
- Conflict detection remains Phase 6.
