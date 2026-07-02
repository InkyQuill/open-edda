# Milestone 5 Phase 3: Draft/Save Model

## Goal

Separate ephemeral editor drafts from canonical saved project files. Browser edits should stay local until an autosave or explicit save, server draft autosaves should live under `.edda/drafts/`, and canonical saves should be modeled as file writes against the stable file index rather than as implicit database-only revisions.

## Context

- Phase 1 defined the Edda project folder shape from the stronger `alchemist` layout.
- Phase 2 added stable file IDs in `.edda/ids.json` and a rebuildable `project_files` index.
- The current web editor already tracks local draft markdown in Redux, but the only backend save path is `PUT /api/projects/{projectID}/content/{contentID}`, which updates `content_items` and creates a revision.
- The database-backed content model remains in place while the file-first workflow is introduced incrementally.

## Implementation Steps

1. Add file draft helpers in `fileproject`.
   - Store server drafts beneath `.edda/drafts/{fileID}.md`.
   - Include a compact JSON sidecar with base hash/path metadata so drafts can be checked before canonical save.
   - Add tests for write/read/delete and path traversal rejection.

2. Add canonical file save helpers in `fileproject`.
   - Resolve a stable file ID through `.edda/ids.json`.
   - Reject stale saves when the caller's expected saved hash does not match the current file hash.
   - Write markdown atomically enough for local project files and return the new hash/bytes without rewriting the stable ID map for unrelated files.

3. Add a CLI surface for local workflow.
   - `edda save <path> --id <file-id> --from-draft` promotes `.edda/drafts/{fileID}.md` to the canonical file.
   - `edda save <path> --id <file-id> --body-file <markdown-file>` saves explicit markdown content.
   - Tests cover successful promotion and stale hash conflicts.

4. Update web save semantics without replacing the whole content API yet.
   - Rename UI copy/state toward explicit saved file semantics where possible.
   - Keep browser draft state local and preserve unsaved text on save failure.
   - Add focused frontend tests documenting the local draft versus saved content boundary.

5. Update the roadmap tracker and run verification.
   - Mark Phase 3 implemented.
   - Run Go tests, frontend unit tests, build, smoke tests, and `git diff --check`.

## Review Notes

- This phase does not migrate every existing `content_items` write into file-backed content. That is deferred to Phase 7, after checkpoints, sync, and conflicts exist.
- Server draft storage is project-local operational state under `.edda/`, not canonical story content.
- Canonical save APIs should be narrow fileproject primitives first so later HTTP/Tauri surfaces can reuse them.
