# Milestone 5 Phase 7: Agent/Write Migration

## Goal

Move structured write and review primitives toward file-backed saved hashes and checkpoints while preserving the existing database-backed editor and assistant paths until the web content model is fully migrated.

## Context

- Current assistant actions and review surfaces use database content IDs, expected revisions, and revision history.
- File-first projects now have stable file IDs, saved hashes, drafts, checkpoints, sync state, and conflicts.
- The first migration should add shared file-version contracts and checkpoint-aware review data without breaking the current UI.

## Implementation Steps

1. Add file version contracts in `fileproject`.
   - Represent a file-backed write target as file ID, path, kind, saved hash, and optional byte range.
   - Validate that an operation still targets the current saved hash before writing.
   - Add tests for valid and stale hash checks.

2. Add checkpoint review summaries.
   - Provide a compact project history summary from checkpoint manifests.
   - Provide file-level history entries by stable file ID.
   - Add tests for project and file history ordering.

3. Add CLI inspection surfaces.
   - `edda files [path]` lists stable file IDs, paths, kinds, and saved hashes for agent/tool use.
   - `edda history [path]` lists project checkpoint history.
   - `edda history [path] --id FILE_ID` lists file-scoped checkpoint history for one stable file.

4. Update assistant/review wording where the web UI still uses database revisions.
   - Avoid presenting database revisions as the final file-first model.
   - Keep current revision endpoints intact until the web content API migration has its own phase.

5. Update roadmap and verify.
   - Mark Phase 7 implemented.
   - Run Go tests, frontend unit tests, build, smoke tests, and `git diff --check`.

## Review Notes

- This phase does not remove database revisions. It creates the file-hash contract that assistant actions and review surfaces can move onto incrementally.
- Direct assistant mutation of project files must continue to pass through explicit saved-hash checks or draft/save workflows.
