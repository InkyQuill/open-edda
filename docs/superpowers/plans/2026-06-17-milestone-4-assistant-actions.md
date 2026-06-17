# Milestone 4 Assistant Actions Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` to execute this plan. Each task must be implemented by a fresh worker subagent, then reviewed by a fresh spec-review subagent, then reviewed by a fresh code-quality subagent before moving to the next task.

**Goal:** Wire the editor-local Generate, Rewrite, and Check controls to the existing agent APIs, show preview results, support accept/reject, and surface revision-safe conflict errors without turning the workspace assistant drawer back into an administration surface.

**Architecture:** Add a small assistant-actions vertical slice that owns quick-action request state, candidate previews, read-and-check reports, accept/reject status, and conflict/error text. Keep editor text/cursor/selection ownership inside `features/editor`, keep provider/model selection in the existing model settings slice, keep skills as optional selected IDs from the skills slice, and call the existing `frontend/src/agentApi.ts` wrappers instead of introducing new API clients.

**Tech Stack:** React 19, Redux Toolkit, React Router 7, TypeScript, Vitest, Bun, Go service tests with SQLite FTS5 tag.

---

## Current Context

Milestone 4 Phase 3.5 moved provider/model/skill administration to `/settings`, kept the workspace assistant drawer chat-only, added project/content creation controls, and redesigned `/projects`. Phase 4 is the next planned Daily Writing Polish phase in `docs/roadmap.md`.

Existing API wrappers already available:

- `frontend/src/agentApi.ts`
  - `runContinuation`
  - `runRewrite`
  - `runReadAndCheck`
  - `acceptCandidate`
  - `rejectCandidate`
- `frontend/src/agentTypes.ts`
  - `ContinuationRequest`
  - `RewriteRequest`
  - `ContinuationResult`
  - `RewriteResult`
  - `ReadAndCheckResult`
  - `AcceptCandidateResult`
  - `GenerationCandidate`
  - `ApplyMode`

Existing editor state already available:

- `frontend/src/features/editor/editorSlice.ts`
  - `contentContext`
  - `cursorByte`
  - `selection`
  - `actionModal`
  - `generateInstructions`
  - save status and errors
- `frontend/src/features/editor/EditorFrame.tsx`
  - `GenerateComposer`
  - `SelectionActionDialog`
  - `onContentSaved(content)` callback for parent list synchronization
- `frontend/src/features/model-settings/modelSettingsSlice.ts`
  - `activeModelVariantId`
- `frontend/src/features/skills/skillsSlice.ts`
  - `selectedSkillIds`

Do not reintroduce provider, model, or skill administration into the workspace drawer. The drawer remains chat-only. Quick actions use the currently active model variant and the current selected skill IDs when present.

## Desired User Flow

1. The author opens a chapter or story-bible entry in the workspace.
2. The author places the cursor and uses the editor-local Generate control.
3. Open Edda calls continuation in preview mode.
4. A preview panel shows the generated candidate and Accept/Reject controls.
5. Accept applies through the revision-safe candidate endpoint and updates the editor with the accepted content returned by the backend.
6. Reject clears the candidate without changing the editor content.
7. The author selects text and opens Rewrite or Check from the selection actions.
8. Rewrite calls the rewrite endpoint in preview mode and shows an accept/reject candidate.
9. Check calls the read-and-check endpoint and shows the assistant report without accept/reject.
10. If the content revision changed before accept, the UI shows a clear conflict/error message near the preview instead of silently overwriting text.

## Task 1: Add Assistant Actions Slice And Request Thunks

**Files**

- Add `frontend/src/features/assistant-actions/assistantActionsSlice.ts`
- Add `frontend/src/features/assistant-actions/assistantActionsSlice.test.ts`
- Edit `frontend/src/app/store/store.ts`
- Edit `frontend/src/test/utils.tsx` if the shared test store helper requires the new reducer

**Implementation**

Create a new Redux vertical slice named `assistantActions`. It owns only quick-action state and must not store provider settings, model catalog state, project lists, or editor body text.

State shape:

```ts
import type {
  GenerationCandidate,
  ReadAndCheckResult,
} from "../../agentTypes";

export type AssistantActionKind = "generate" | "rewrite" | "check";

export type AssistantActionStatus = "idle" | "running" | "succeeded" | "failed";

export interface AssistantActionState {
  actionKind: AssistantActionKind | null;
  status: AssistantActionStatus;
  acceptStatus: AssistantActionStatus;
  rejectStatus: AssistantActionStatus;
  projectId: string | null;
  contentId: string | null;
  requestKey: string | null;
  requestToken: string | null;
  candidate: GenerationCandidate | null;
  checkResult: ReadAndCheckResult | null;
  error: string | null;
  acceptError: string | null;
  rejectError: string | null;
}
```

Use a `requestKey` string derived by the component from the current route/editor identity and a unique `requestToken` generated for each dispatch, such as `crypto.randomUUID()`. Pending reducers store both values. Fulfilled and rejected reducers must ignore stale results unless both the slice's `requestKey` and `requestToken` match the thunk arg, so multiple rapid actions on the same content cannot let a slower older response replace a newer preview.

Add thunk args:

```ts
export interface RunGenerateArgs {
  projectId: string;
  contentId: string;
  modelVariantId: string;
  expectedRevision: number;
  cursorByte: number;
  instructions: string;
  skillIds?: string[];
  requestKey: string;
  requestToken: string;
}

export interface RunRewriteArgs {
  projectId: string;
  contentId: string;
  modelVariantId: string;
  expectedRevision: number;
  selectionStartByte: number;
  selectionEndByte: number;
  instructions: string;
  skillIds?: string[];
  requestKey: string;
  requestToken: string;
}

export interface RunCheckArgs {
  projectId: string;
  contentId: string;
  modelVariantId: string;
  expectedRevision: number;
  selectionStartByte: number;
  selectionEndByte: number;
  instructions: string;
  skillIds?: string[];
  requestKey: string;
  requestToken: string;
}
```

Use these endpoint payloads:

```ts
runContinuation(projectId, {
  contentId,
  modelVariantId,
  expectedRevision,
  cursorByte,
  instructions,
  skillIds,
  applyMode: "preview",
  insert: true,
  continuationUnits: "paragraph",
  continuationCount: 1,
});
```

```ts
runRewrite(projectId, {
  contentId,
  modelVariantId,
  expectedRevision,
  selectionStartByte,
  selectionEndByte,
  instructions,
  skillIds,
  applyMode: "preview",
});
```

```ts
runReadAndCheck(projectId, {
  contentId,
  modelVariantId,
  expectedRevision,
  selectionStartByte,
  selectionEndByte,
  instructions,
  skillIds,
  applyMode: "preview",
});
```

Reducer behavior:

- `runGenerate.pending`, `runRewrite.pending`, and `runCheck.pending` clear the previous candidate, previous check result, previous action error, `acceptStatus`, `rejectStatus`, `acceptError`, and `rejectError`.
- Fulfilled generate/rewrite stores `result.candidate`.
- Fulfilled check stores the complete `ReadAndCheckResult`.
- Rejected requests set `status: "failed"` and a user-readable error.
- `clearAssistantActionResult()` returns the slice to idle result state while preserving no stale candidate.

Register the reducer in `frontend/src/app/store/store.ts`.

**Tests**

Mock `frontend/src/agentApi.ts` in `assistantActionsSlice.test.ts`.

Cover:

- Generate dispatch calls `runContinuation` with `applyMode: "preview"`, `insert: true`, `continuationUnits: "paragraph"`, and `continuationCount: 1`.
- Rewrite dispatch calls `runRewrite` with selection bytes and preview apply mode.
- Check dispatch calls `runReadAndCheck` with selection bytes and preview apply mode.
- A stale fulfilled result with an old `requestKey` does not replace the current candidate.
- A rejected run stores an error and leaves candidate/check result empty.

## Task 2: Add Accept And Reject Candidate State

**Files**

- Edit `frontend/src/features/assistant-actions/assistantActionsSlice.ts`
- Edit `frontend/src/features/assistant-actions/assistantActionsSlice.test.ts`

**Implementation**

Add thunks:

```ts
export interface AcceptAssistantCandidateArgs {
  projectId: string;
  candidateId: string;
  requestKey: string;
}

export interface RejectAssistantCandidateArgs {
  projectId: string;
  candidateId: string;
  requestKey: string;
}
```

Behavior:

- `acceptAssistantCandidate` calls `acceptCandidate(projectId, candidateId)` and returns the backend `AcceptCandidateResult`.
- On accept success, clear `candidate`, set `acceptStatus: "succeeded"`, and keep the accepted content in the fulfilled action result for the component to apply to editor state through existing callbacks.
- On accept failure, keep the candidate visible and set `acceptStatus: "failed"` plus `acceptError`.
- If the accept error text indicates a revision conflict or stale candidate, use the display text:

```text
Content changed before this preview was accepted. Review the latest draft, then run the action again.
```

- `rejectAssistantCandidate` calls `rejectCandidate(projectId, candidateId)`.
- On reject success, clear `candidate`, set `rejectStatus: "succeeded"`, and return to idle action state.
- On reject failure, keep the candidate visible and set `rejectStatus: "failed"` plus `rejectError`.
- Stale accept/reject completions whose `requestKey` no longer matches the slice must not mutate the current preview.

Conflict detection helper:

```ts
function isRevisionConflictError(message: string): boolean {
  const normalized = message.toLowerCase();
  return (
    normalized.includes("revision") ||
    normalized.includes("conflict") ||
    normalized.includes("stale")
  );
}
```

**Tests**

Cover:

- Accept calls `acceptCandidate` with project ID and candidate ID.
- Successful accept clears the candidate and exposes the accepted content through the thunk result.
- Conflict-like accept failure keeps the candidate and stores the conflict display text.
- Generic accept failure keeps the original error text.
- Reject calls `rejectCandidate`.
- Successful reject clears the candidate.
- Stale accept/reject completions do not mutate a newer preview.

## Task 3: Add Assistant Action Preview Panel

**Files**

- Add `frontend/src/features/editor/AssistantActionPreview.tsx`
- Edit `frontend/src/features/editor/EditorFrame.tsx`
- Edit `frontend/src/features/editor/editorSlice.ts` only if a small action is needed to clear editor-local status text after action completion

**Implementation**

Create a presentational component that reads from the assistant-actions slice and receives callbacks from `EditorFrame`:

```ts
interface AssistantActionPreviewProps {
  onAccept: () => void;
  onReject: () => void;
  onDismiss: () => void;
}
```

Render states:

- Running generate/rewrite/check: show a compact status row near the editor action controls.
- Candidate preview: show action kind, generated markdown text, Accept and Reject buttons.
- Check result: show assistant message/report text and a Dismiss button.
- Run error: show error text and Dismiss.
- Accept error: show accept error text, keep Accept and Reject available.
- Reject error: show reject error text, keep Accept and Reject available.

Accessibility:

- Use real `button` elements or existing `Button` component.
- Add `aria-busy="true"` to the running container.
- Disable Accept while `acceptStatus === "running"`.
- Disable Reject while `rejectStatus === "running"`.
- Keep labels specific: `Accept preview`, `Reject preview`, `Dismiss check result`.

Placement:

- Render the preview inside `EditorFrame`, close to `GenerateComposer` and `SelectionActionDialog`.
- Do not place this in the workspace assistant drawer.
- Do not nest cards inside cards. Use the editor action surface's existing spacing and borders.

**Tests**

If existing component tests are present for editor components, add focused render tests. If there is no established component-test harness for this area, cover the behavior through slice tests in Tasks 1 and 2 and manually verify in browser during Task 6.

## Task 4: Wire Generate Composer To Continuation

**Files**

- Edit `frontend/src/features/editor/EditorFrame.tsx`
- Edit `frontend/src/features/editor/GenerateComposer.tsx` only if the component API needs a pending/disabled label
- Edit `frontend/src/features/editor/EditorFrame.test.tsx` if an existing test harness is available

**Implementation**

Replace the current generate placeholder behavior with a thunk dispatch.

Required inputs:

- `contentContext.projectId`
- `contentContext.contentId`
- `contentContext.revision`
- `cursorByte`
- `activeModelVariantId`
- `generateInstructions`
- `selectedSkillIds`

Validation:

- If no content is loaded, show `Open a chapter or story-bible entry before generating.`
- If no model variant is active, show `Choose an active model variant in Settings before generating.`
- If cursor byte is missing, show `Place the cursor where generated text should be inserted.`
- If instructions are blank, allow generation with an empty instruction string.

Dispatch:

```ts
dispatch(
  runGenerate({
    projectId,
    contentId,
    modelVariantId: activeModelVariantId,
    expectedRevision: contentContext.revision,
    cursorByte,
    instructions: generateInstructions.trim(),
    skillIds: selectedSkillIds,
    requestKey,
  }),
);
```

Request identity:

- Build `requestKey` from `projectId`, `contentId`, current route location key/pathname, and `contentContext.revision`.
- Clear stale assistant action results when project/content route identity changes.

On accept:

- Dispatch `acceptAssistantCandidate`.
- Use the fulfilled `content` from `AcceptCandidateResult` to call existing editor update flow:
  - update editor content body/revision through existing editor slice actions
  - call `onContentSaved(content)` so the workspace content list stays in sync

On reject:

- Dispatch `rejectAssistantCandidate`.

**Tests**

Cover through component tests if available:

- Generate button dispatches `runGenerate` when content, cursor, and active model exist.
- Missing active model shows the validation message and does not call API.
- Successful accept applies returned content through the existing save/update path.

If component tests are not already practical, add slice tests and include manual browser verification in Task 6.

## Task 5: Wire Rewrite And Check Selection Actions

**Files**

- Edit `frontend/src/features/editor/SelectionActionDialog.tsx`
- Edit `frontend/src/features/editor/EditorFrame.tsx`
- Edit `frontend/src/features/editor/SelectionActionDialog.test.tsx` if an existing component harness is available

**Implementation**

Rewrite:

- Enable the Rewrite submit button when:
  - action modal kind is `rewrite`
  - content context exists
  - selection start/end bytes exist and are not equal
  - active model variant exists
- On submit, dispatch `runRewrite` with preview mode through the assistant-actions slice.
- Use the dialog textarea value as `instructions`.
- Close the modal after dispatch starts, leaving the preview panel visible in the editor.

Check:

- Enable the Check submit button when:
  - action modal kind is `check`
  - content context exists
  - selection start/end bytes exist and are not equal
  - active model variant exists
- On submit, dispatch `runCheck`.
- Close the modal after dispatch starts, leaving the report visible in the preview panel.

Note:

- Keep Note disabled and visibly out of scope for this phase unless an existing note API is already wired in the workspace. The button text may remain `Coming soon`.

Validation messages:

- Missing active model: `Choose an active model variant in Settings before running this action.`
- Missing selection: `Select text before running this action.`
- Missing content: `Open a chapter or story-bible entry before running this action.`

**Tests**

Cover:

- Rewrite dispatches `runRewrite` with selection byte offsets.
- Check dispatches `runCheck` with selection byte offsets.
- Missing active model prevents dispatch and shows validation.
- Empty or collapsed selection prevents dispatch and shows validation.

## Task 6: Revision-Safe UX Hardening And Verification

**Files**

- Edit files touched by Tasks 1-5 as needed
- Edit `docs/roadmap.md`

**Implementation**

Hardening requirements:

- Clear assistant action previews when the loaded content changes.
- Do not clear a visible candidate merely because the generate instructions textarea changes.
- Prevent duplicate accept clicks while accept is running.
- Prevent duplicate reject clicks while reject is running.
- Keep failed accept candidates visible so the author can retry or reject.
- After accepted content is applied, clear selection/action-modal state that points at the old revision.
- Display conflict text from Task 2 near the candidate preview.
- Keep all action controls in the editor area, not in the assistant drawer.

Roadmap update:

- After implementation and verification pass, change the Phase 4 row in `docs/roadmap.md` from `Planned` to `Implemented`.
- Keep Phase 5 as `Planned`.

Verification commands:

```bash
cd frontend
bun --bun run test
bun --bun run build
```

```bash
go test -tags sqlite_fts5 ./...
```

```bash
git diff --check
```

Manual browser verification:

1. Start the app using the repository's documented dev command.
2. Open `/projects`.
3. Open a project and content item in the workspace.
4. Confirm Generate calls the continuation flow and renders a preview.
5. Accept the preview and confirm the editor updates to the returned revision.
6. Run Generate again and reject the preview; confirm editor text does not change.
7. Select text, run Rewrite, and confirm preview plus accept/reject behavior.
8. Select text, run Check, and confirm a report appears without accept/reject controls.
9. Navigate to another content item during an in-flight action and confirm stale results do not appear on the new content.
10. Trigger an accept conflict using a stale candidate or mocked response and confirm the conflict message is visible.

## Final Review

After all tasks pass their worker/spec-review/code-quality loop:

1. Run one final review subagent over the full Phase 4 diff.
2. Fix any validated blocker.
3. Re-run the verification commands.
4. Commit the implemented Phase 4 work with a message that names assistant actions.
5. Present branch completion options using the finishing-branch workflow.
