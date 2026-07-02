# Milestone 5 Phase 6: Conflict Preservation And Resolution

## Goal

Preserve divergent saved-file edits in `.edda/conflicts/` and provide a local resolution path that restores a normal canonical file once the author chooses the resolved content.

## Context

- Phase 5 added local sync state and pending upload metadata.
- The file-first design says conflicts preserve base, local, and server versions before resolving back into normal files.
- The initial implementation does not need a remote transport; it needs reliable file storage and CLI behavior that later sync can call when divergence is detected.

## Implementation Steps

1. Add conflict primitives in `fileproject`.
   - Store conflict records beneath `.edda/conflicts/{fileID}/`.
   - Preserve `base.md`, `local.md`, `server.md`, and `metadata.json`.
   - Include file ID, path, hashes, detected time, and resolution status.

2. Add list and resolve helpers.
   - List unresolved conflicts in stable order.
   - Resolve from `local`, `server`, or an explicit body file.
   - Save the resolved body into the canonical project file and mark the conflict resolved.

3. Add CLI commands.
   - `edda conflicts [path]` lists preserved conflicts.
   - `edda resolve [path] --id FILE_ID (--use local|server|--body-file file.md)` resolves one conflict.

4. Add tests.
   - Preserve writes all three versions and metadata.
   - Layout scans ignore unresolved conflict working files.
   - Resolve writes canonical content and records resolution metadata.
   - CLI list/resolve happy paths are covered.

5. Update roadmap and verify.
   - Mark Phase 6 implemented.
   - Run Go tests, frontend unit tests, build, smoke tests, and `git diff --check`.

## Review Notes

- This phase does not perform server comparison itself; it provides the storage and resolution primitives that sync will call when it sees divergent saved hashes.
- User-facing output should avoid branch/merge language.
