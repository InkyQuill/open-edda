# CodeRabbit review disposition — 2026-10-07

Scope: all 22 comments on [PR #3](https://github.com/InkyQuill/open-edda/pull/3), collected at head `4ec428c`: 19 from July and 3 from October. The old prototype remains distinct from the network checkout implementation. No migration or new product feature is introduced.

## Dispositions

| Comment | Disposition |
| --- | --- |
| [3512433746: `cmd/edda/main_test.go`](https://github.com/InkyQuill/open-edda/pull/3#discussion_r3512433746) | Fixed: regression covers a new target before and after flags, plus status of the empty initialized project. |
| [3512433754: `cmd/edda/main.go`](https://github.com/InkyQuill/open-edda/pull/3#discussion_r3512433754) | Fixed: init extracts a leading path without requiring it to exist; metadata alone identifies an empty initialized project. |
| [3512433763: `cmd/edda/main.go`](https://github.com/InkyQuill/open-edda/pull/3#discussion_r3512433763) | Fixed: CLI acknowledges the saved canonical file while returning the distinct cleanup error. The nonzero exit still signals incomplete cleanup. |
| [3512433782: `cmd/edda/main.go`](https://github.com/InkyQuill/open-edda/pull/3#discussion_r3512433782) | Fixed: expected-hash conflicts no longer mention a draft base for --body-file. |
| [3512433797: `fileproject/checkpoint.go`](https://github.com/InkyQuill/open-edda/pull/3#discussion_r3512433797) | Closed: reused the outer error variable; the prior code already checked the inner error. |
| [3512433807: `fileproject/conflict.go`](https://github.com/InkyQuill/open-edda/pull/3#discussion_r3512433807) | Fixed: preservation and resolution share a per-root cross-process file lock; an unresolved record cannot be replaced. Concurrent preservation regression verifies one complete winning set. |
| [3512433818: `fileproject/indexer.go`](https://github.com/InkyQuill/open-edda/pull/3#discussion_r3512433818) | Fixed: iteration errors are checked before deleting stale index rows. |
| [3512433824: `fileproject/metadata.go`](https://github.com/InkyQuill/open-edda/pull/3#discussion_r3512433824) | Fixed: fully written temporary metadata is published with a non-replacing hard link; concurrent initialization has exactly one winner. Requires the already supported Linux filesystem semantics. |
| [3512433840: `fileproject/save_test.go`](https://github.com/InkyQuill/open-edda/pull/3#discussion_r3512433840) | Fixed: the permission-based cleanup test skips root, which bypasses the intended restriction. |
| [3512433843: `fileproject/save.go`](https://github.com/InkyQuill/open-edda/pull/3#discussion_r3512433843) | Fixed with the save reporting issue: ErrDraftCleanup distinguishes successful canonical save from cleanup failure. |
| [3512433846: `fileproject/sync_test.go`](https://github.com/InkyQuill/open-edda/pull/3#discussion_r3512433846) | Fixed: concurrency regression asserts all ten distinct queued checkpoints are retained. |
| [3512433855: `fileproject/sync.go`](https://github.com/InkyQuill/open-edda/pull/3#discussion_r3512433855) | Fixed: removed the global mutex; per-root flock remains the serialization boundary. |
| [3512433861: `fileproject/sync.go`](https://github.com/InkyQuill/open-edda/pull/3#discussion_r3512433861) | Fixed: removed the discarded attempts increment. |
| [3512433864: `fileproject/version.go`](https://github.com/InkyQuill/open-edda/pull/3#discussion_r3512433864) | Closed without code change: all bounds are enforced; distinct wording is optional diagnostic polish, not a correctness defect. |
| [3512433866: `frontend/playwright.config.ts`](https://github.com/InkyQuill/open-edda/pull/3#discussion_r3512433866) | Partially adopted: forbidOnly is enabled in CI for both configs. Automatic retries were not added without evidence of flakiness; file tests already use one worker. |
| [3512433872: `frontend/src/api.ts`](https://github.com/InkyQuill/open-edda/pull/3#discussion_r3512433872) | Fixed defensively: unstructured responses and 5xx details do not enter the visible error message; expected structured client errors are length-bounded. Raw body remains available on ApiError for explicit diagnostics. |
| [3512433876: `frontend/src/features/review/ReviewDrawer.test.tsx`](https://github.com/InkyQuill/open-edda/pull/3#discussion_r3512433876) | Fixed: automatic revision fetch is bounded per project/content/revision target. Browser regressions with a stale response run on desktop and mobile. |
| [3512433879: `frontend/src/features/review/ReviewDrawer.tsx`](https://github.com/InkyQuill/open-edda/pull/3#discussion_r3512433879) | Fixed: LCS matrix is bounded to 250,000 cells; large comparisons use a linear whole-region fallback preserving common edges and both input texts. |
| [3512433883: `frontend/src/features/review/reviewSlice.ts`](https://github.com/InkyQuill/open-edda/pull/3#discussion_r3512433883) | Partially valid: an explicit history reload already existed. Pending now preserves the prior selection; successful restore clears the selection until the existing reload supplies the new list. Reducer regression covers failed restore selection. |
| [4206155447: `frontend/src/features/projects/ProjectsPage.tsx`](https://github.com/InkyQuill/open-edda/pull/3#discussion_r4206155447) | Fixed: the explicit empty-language choice is sent unchanged; API-backed browser tests verify persisted empty language. |
| [4206155471: `README.md`](https://github.com/InkyQuill/open-edda/pull/3#discussion_r4206155471) | Fixed: README, CONTEXT and project-version status describe delivered packaging, synchronization, production UI, per-file drafts and history. Physical power-loss/CSI and multi-instance limits remain explicit. |
| [4206155475: `scripts/verify-live-ui.mjs`](https://github.com/InkyQuill/open-edda/pull/3#discussion_r4206155475) | Fixed: live UI checker uses Russian accessible names. The actual script passed desktop/mobile against an isolated server with temporary credentials and data. |

## Validation

- Go race tests across all 14 packages and go vet passed; gofmt and diff whitespace checks passed.
- 147 frontend unit tests, TypeScript check and production build passed. The existing bundle-size warning remains.
- Four structured-workspace browser cases passed, including stale-response regressions; two device-inapplicable cases are intentionally skipped.
- Four real-server file-workspace browser cases passed, including the empty-language choice.
- Full isolated interactive CLI acceptance passed, including first send, two-copy synchronization, conflicts, history and backup/restore.
- The live UI verification script passed against an isolated server on desktop/mobile. Production was not modified.

The initial new language regression used an overly strict label lookup and timed out; switching to the accessible combobox role fixed the test, after which all four cases passed. Root execution, physical power loss and CSI faults were not tested here.

The generic docstring-coverage recommendation is not a runtime defect and does not warrant generated comments solely to meet an external percentage. The archive-traversal heuristic attached to PreserveConflict was a false positive: the three version filenames are constants and FileID is validated. Historical CodeRabbit autofix/publication failures do not establish code correctness.
