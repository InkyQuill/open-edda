# Milestone 4 Phase 6: Mobile and Browser Smoke Hardening

## Goal

Harden the Milestone 4 writing workspace across desktop presets and mobile sheet workflows with focused automated coverage and a small browser smoke path.

This phase should not redesign the workspace. It should prove that the already-built shell, drawers, sheets, editor controls, persisted workspace state, and review/assistant surfaces remain usable across the core modes.

## Current State

- Workspace mode, drawer state, active drawer tabs, selected content fallback, and mobile sheet state live in Redux.
- Project-scoped workspace state persists to local storage, excluding transient `mobileSheet`.
- The shell renders desktop drawers and mobile sheets, but there are no component tests for shell rendering.
- There is no Playwright/e2e setup in the repo yet.
- Existing tests are Vitest SSR/state tests plus Go backend tests.

## Design Constraints

- Keep the app editor-first on mobile. Sheets overlay the editor and must close without losing project/workspace state.
- Do not add broad e2e coverage that depends on external model providers.
- Browser smoke tests should use local fake API responses or existing seeded backend paths. They should not require real OpenAI-compatible credentials.
- Preserve the current Redux/local-storage ownership: URL route params own project/content identity; Redux owns workspace UI mode and sheets.

## Task 1: Workspace State Hardening

Add or extend reducer/persistence tests for expected mobile and desktop behavior.

- Verify mode presets:
  - Draft closes the right drawer.
  - Assistant opens the assistant drawer.
  - Review opens the right drawer on tools/review.
- Verify mobile sheet state:
  - all authoring sheets are accepted,
  - closing a sheet returns to `null`,
  - hydrating project state preserves current transient `mobileSheet`.
- Verify persisted project state excludes `mobileSheet`.

## Task 2: Workspace Shell Component Tests

Add SSR component tests for `WorkspaceShell`.

- Desktop Assistant mode renders assistant drawer content and editor content.
- Desktop Review mode renders review drawer content with selected content.
- Mobile sheet markup includes Files, Assistant, Review, and World/Notes controls.
- Rendering with no selected content shows editor empty state and review empty state without throwing.

Use existing Redux configured test stores and `renderToStaticMarkup`; do not add a DOM/browser dependency for these checks.

## Task 3: Mobile Instruction Preservation

Verify selected action modal/sheet instruction state.

- Add tests around `editorSlice` to show opening/closing selection actions preserves typed instructions unless content context resets.
- If the current implementation drops instructions when temporarily dismissed, fix the reducer/component boundary so mobile dismissal preserves draft instructions for the active content context.
- Keep reset behavior on content switch.

## Task 4: Browser Smoke Setup

Introduce the smallest browser smoke setup that can run locally.

- Add Playwright only if the repository does not already have an equivalent browser runner.
- Add a smoke test script such as `test:smoke`.
- Mock or seed API responses so tests do not require external providers.
- Cover:
  - desktop Assistant/Draft/Review mode switching,
  - mobile bottom sheet opening/closing for Files, Assistant, Review, and World/Notes,
  - editor frame renders nonblank with selected content,
  - no console errors during the smoke path.

## Task 5: Verification

Run:

- `mise exec -- bun run test`
- `mise exec -- bun run build`
- browser smoke command added in Task 4
- `mise run test`
- `git diff --check`

## Self-Review Notes

- This phase should harden behavior, not introduce new product surfaces.
- If Playwright installation is blocked by dependency/network policy, keep Tasks 1-3 implemented and record the browser setup blocker explicitly in the plan and final status rather than silently dropping browser coverage.
- Avoid testing implementation-only CSS details. Test user-visible panels, mode/sheet availability, persistence semantics, and absence of obvious browser runtime errors.
