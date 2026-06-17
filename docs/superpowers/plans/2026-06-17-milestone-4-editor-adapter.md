# Milestone 4 Editor Adapter Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the Milestone 4 read-only textarea assumption with a Galley-backed editor adapter that owns editable draft state, UTF-8 cursor/selection snapshots, revision context, and revision-safe save wiring.

**Architecture:** Keep Galley Editor behind a small adapter boundary instead of coupling workspace actions directly to CodeMirror or DOM offsets. The implementation uses Galley's controlled React component and selection callbacks, while all editor actions read from adapter snapshots and `editorSlice` context.

**Tech Stack:** React, Redux Toolkit, TypeScript, Vitest, `@inky/galley-editor`, existing project content API.

---

## File Structure

- Create: `frontend/src/features/editor/editorAdapter.ts`
- Test: `frontend/src/features/editor/editorAdapter.test.ts`
- Modify: `frontend/src/features/editor/editorSlice.ts`
- Test: `frontend/src/features/editor/editorSlice.test.ts`
- Modify: `frontend/src/features/editor/EditorFrame.tsx`
- Modify: `frontend/src/features/editor/GenerateComposer.tsx`
- Modify: `frontend/src/features/workspace/WorkspacePage.tsx`
- Modify: `frontend/src/features/workspace/WorkspaceShell.tsx`
- Modify: `frontend/src/api.ts`
- Modify: `frontend/package.json`
- Modify: `frontend/bun.lock`
- Create: `frontend/.npmrc`
- Modify: `docs/roadmap.md`

## Task 1: Add Editor Adapter Snapshot Contract

- [x] **Step 1: Write adapter tests**

Create tests that assert UTF-8 byte offsets for cursor-only and selected ranges, including multibyte prose.

- [x] **Step 2: Implement adapter helpers**

Create a small adapter module that converts a text value and UTF-16 DOM selection positions into an editor snapshot with UTF-8 byte offsets.

- [x] **Step 3: Run focused adapter tests**

Run `cd frontend && bun run test -- editorAdapter`.

## Task 2: Store Editor Context And Draft State

- [x] **Step 1: Extend reducer tests**

Pin hydration, dirty tracking, save success, save failure, and reset behavior.

- [x] **Step 2: Extend `editorSlice`**

Store content identity, expected revision, persisted markdown, draft markdown, dirty flag, save status, and save error alongside cursor/selection state.

- [x] **Step 3: Run focused editor tests**

Run `cd frontend && bun run test -- editorSlice editorAdapter`.

## Task 3: Wire Revision-Safe Save API

- [x] **Step 1: Add frontend API wrapper**

Add `updateContent(projectId, contentId, input)` for the existing backend `PUT /api/projects/{projectID}/content/{contentID}` route.

- [x] **Step 2: Update workspace state after save**

Let `WorkspacePage` replace the saved content item in its local list without a full route reload.

## Task 4: Replace Read-Only Editor Surface

- [x] **Step 1: Make `EditorFrame` editable**

Use `editorSlice.draftMarkdown` as the displayed value, render `GalleyEditor`, update adapter snapshots from Galley selection events, and expose save status.

- [x] **Step 2: Validate Generate against cursor and revision**

Generate remains a staged action in this phase, but it must validate that content context and cursor/revision are present before reporting readiness.

- [x] **Step 3: Preserve selection actions**

Rewrite, Check, and Note continue to open from selection snapshots with UTF-8 byte offsets.

## Task 5: Verify And Update Roadmap

- [x] **Step 1: Run frontend tests and build**

Run `cd frontend && bun run test && bun run build`.

- [x] **Step 2: Run diff hygiene**

Run `git diff --check`.

- [x] **Step 3: Update tracking docs**

Mark Phase 2 implemented if still stale and mark Phase 3 implemented after verification.

## Consumer Feedback To Track

- Backend content save exists, but the frontend had no wrapper before this phase.
- Galley local source verified with `npm test` and `npm run build:lib`; both pass, but `npm ci` fails because `package-lock.json` is stale against `package.json`.
- Open Edda uses the published registry package via `frontend/.npmrc` and `@inky/galley-editor@0.9.1`, not a local `file:` dependency.
- The production bundle now warns about a >500 kB JS chunk after adding Galley; lazy-loading the editor should be a follow-up task.
- Generate/Rewrite/Check execution is still Phase 4; this phase only gives those actions reliable cursor, selection, and revision inputs.
