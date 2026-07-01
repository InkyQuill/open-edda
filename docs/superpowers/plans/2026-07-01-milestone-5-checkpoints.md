# Milestone 5 Phase 4: Linear Checkpoints

## Goal

Add rebuildable project-wide checkpoints under `.edda/checkpoints/` so file-first projects have linear history, diff, and restore primitives independent of database revisions.

## Context

- Phase 1 scans the canonical Edda folder layout.
- Phase 2 assigns stable file IDs in `.edda/ids.json` and indexes files into SQLite.
- Phase 3 separates browser/server drafts from canonical file saves.
- Review surfaces still use database revisions today; Phase 7 will migrate those surfaces toward saved file hashes and checkpoints.

## Implementation Steps

1. Add checkpoint storage primitives in `fileproject`.
   - Store checkpoint manifests at `.edda/checkpoints/{checkpointID}/manifest.json`.
   - Copy canonical stable-ID files into `.edda/checkpoints/{checkpointID}/files/{relativePath}`.
   - Include project file ID, path, kind, title, hash, size, and timestamp metadata.

2. Add checkpoint list, diff, and restore helpers.
   - List checkpoints in creation order.
   - Diff two checkpoints or checkpoint-to-working-tree by comparing stable file IDs and hashes.
   - Restore a checkpoint by replacing canonical files from the checkpoint snapshot and refreshing `.edda/ids.json`.

3. Add CLI workflow.
   - `edda history [path]` lists checkpoints.
   - `edda checkpoint [path] --message "..."` creates a checkpoint.
   - `edda diff [path] --from CHECKPOINT [--to CHECKPOINT]` reports changed/added/deleted files.
   - `edda restore [path] --checkpoint CHECKPOINT` restores canonical files.

4. Add focused tests.
   - Checkpoint creation snapshots canonical files and ignores `.edda/drafts`.
   - Diff reports working-tree changes.
   - Restore rolls files back and refreshes stable IDs.
   - CLI tests cover checkpoint/history/diff/restore happy paths.

5. Update roadmap and verify.
   - Mark Phase 4 implemented.
   - Run Go tests, frontend unit tests, build, smoke tests, and `git diff --check`.

## Review Notes

- Checkpoints are linear local history, not sync history. Server reconciliation and conflicts remain Phase 5/6.
- Restore is intentionally project-wide at this phase; file-scoped restore can be layered into review surfaces later.
