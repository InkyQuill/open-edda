# Milestone 4 Phase 5: Review Surfaces

## Goal

Implement review-mode drawer surfaces for the selected content item: checkpoint/history review, revision diff/restore, attached Read and Check notes, activity, and prompt records.

This phase should make Review mode useful without changing the editor's primary writing surface. The right drawer becomes the place where an author can inspect what happened, compare versions, restore an older checkpoint through the normal revision pipeline, and open recent reports/prompt records.

## Current State

- `ReviewDrawer` loads project activity and prompt records, but its reports and revisions tabs are placeholders.
- The project service already exposes `ListRevisions` through `GET /api/projects/{projectID}/content/{contentID}/revisions`.
- The service can create attached notes, and Read and Check returns a note, but there is no frontend list endpoint for attached notes.
- The service can update content with `CreatedBy: restore`, but there is no explicit restore endpoint.
- `WorkspaceShell` passes only `projectId` into `ReviewDrawer`; the drawer needs the selected content item and a saved-content callback.

## Design Constraints

- Keep Review mode drawer-first. Do not move diff/restore into the editor body.
- Use the existing content revision model. Restoring a checkpoint creates a new revision; it must not overwrite history.
- Use expected-revision conflict handling when restoring. If content changed since the drawer loaded, surface a clear reload/review message.
- Keep prompt records and activity project-scoped. Keep revisions and notes content-scoped.
- Keep the implementation compatible with the file-first checkpoint direction: user-facing language should prefer "checkpoint" where appropriate while using existing `revision` API names internally.

## Task 1: Project API Support

Add frontend and backend API support needed by the review drawer.

- Add a backend route:
  - `POST /api/projects/{projectID}/content/{contentID}/revisions/{revisionNumber}/restore`
  - body: `{ "expectedRevision": number, "reason": string }`, where the frontend sends `restore checkpoint {revisionNumber}` and the service applies a revision-specific default if the field is blank.
  - returns updated `ContentItem`
- Add `Service.RestoreRevision(ctx, input)` that:
  - verifies the target content belongs to the project,
  - finds the requested historical revision,
  - verifies `current_revision == expectedRevision`,
  - updates content body/metadata from the selected revision,
  - creates a new revision with the author restore actor and a default reason such as `restore revision {revisionNumber}`.
- Add tests for successful restore and stale expected-revision conflict.
- Extend `frontend/src/types.ts` with `Revision`.
- Add `listRevisions()` and `restoreRevision()` in `frontend/src/api.ts`.
- Add API unit tests for encoded content/revision paths.

## Task 2: Review State

Extend `reviewSlice` and `reviewThunks` with content-scoped revision loading and restore state.

- Add state:
  - `contentId`
  - `revisions`
  - `revisionsStatus`
  - `revisionsRequestId`
  - `selectedRevisionNumber`
  - `restoreStatus`
  - `restoreRequestId`
  - `restoreError`
- Add thunks:
  - `loadContentRevisions({ projectId, contentId })`
  - `restoreContentRevision({ projectId, contentId, revisionNumber, expectedRevision })`
- Ignore stale loads/restores by project, content, and request ID.
- Clear content-scoped revision state when selected content changes.
- On restore success, refresh revision state around the returned content and clear restore status.
- Add reducer tests for load, stale load, content switch, restore success, and conflict failure.

## Task 3: Drawer Data Flow

Pass selected content into the review drawer and wire refresh/restore.

- Update `WorkspaceShell` to render:
  - `<ReviewDrawer projectId={projectId} content={selectedContent} onContentSaved={onContentSaved} />`
  - same props in the mobile review sheet.
- In `ReviewDrawer`, load revisions when selected content exists.
- On restore success:
  - call `onContentSaved(updatedContent)`,
  - leave the editor content update to existing selected-content propagation,
  - show a short restored checkpoint state in the revisions tab.
- If no content is selected, render an empty state instead of revision controls.

## Task 4: Review Drawer UI

Replace placeholders with working review surfaces.

- Reports tab:
  - Show the transient last-run Read and Check output already held in `assistantActions.checkResult` when it belongs to the selected content.
  - Show an empty state that points the author to the editor Check action when no in-memory check result exists.
  - Keep buttons for Read report / Check draft only if they perform existing actions or are clearly disabled; avoid dead primary controls.
- Revisions tab:
  - Show current checkpoint number from selected content.
  - List revisions/checkpoints with date, reason, createdBy, action kind/model/skill metadata when present.
  - Select a revision to show a compact diff against the current body.
  - Provide Restore for non-current selected revisions.
  - Disable restore while a restore is pending or when the selected revision is current.
  - Surface conflicts as "Content changed before this checkpoint was restored. Review the latest draft, then try again."
- Activity tab:
  - Keep activity and prompt records.
  - When a prompt record is selected, show parsed request/response JSON in a compact details panel instead of selection with no visible effect.

## Task 5: Component Tests

Add focused component tests for `ReviewDrawer`.

- No selected content renders content-scoped empty state.
- Loading revisions shows checkpoint items after fulfilled state.
- Selecting a prompt record reveals request/response details.
- Restore button dispatches restore with selected content revision and current expected revision.
- Restore conflict message appears from rejected state.

## Task 6: Verification

Run:

- `mise exec -- bun run test`
- `mise exec -- bun run build`
- `mise run test`
- `git diff --check`

## Self-Review Notes

- Restore endpoint is intentionally small and uses the existing revision table rather than introducing a separate checkpoint table.
- Attached notes do not yet have a list endpoint. This plan keeps report surfacing limited to the current in-memory Read and Check result plus existing activity/prompt data; a broader notes index can be a later phase if needed.
- The frontend should use "checkpoint" in labels where authors read it, while keeping `Revision` type and route names aligned with existing code.
