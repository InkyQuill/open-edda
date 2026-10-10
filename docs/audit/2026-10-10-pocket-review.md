# Pocket reviews and author editing — 2026-10-10

## Delivered behavior

The file workspace has explicit **Редактирование / Рецензирование** modes. Review
mode makes the source read-only; editing mode permits normal author changes. Both
keep the same review panel with chapter note, proposals, signals and comments.
Cards follow source order, show before/after, and navigate to highlighted text.
On narrow screens the panel sits below the document so both remain accessible.
Mode switches retain drafts. Live/source editor switches preserve original CRLF,
CR and mixed newlines outside the changed span, including opening without edits.

An author explicitly accepts/rejects a proposal or closes a signal. With unsaved
drafts these actions are disabled; “Сохранить и проверить привязки” saves the
current chapter and reloads classification. Other unsaved files must also be
saved first. Unavailable records remain visible and can be adjudicated manually;
stale, ambiguous and overlapping proposals cannot be applied automatically.

## Storage contract

The Go adapter reads the adjacent `.pocket-editor.json` v1/v2 and review v1,
scoped to the selected book folder and stable tree entry ID. It supports either
`chapter.review.json` or `chapter.md.review.json`, rejects ambiguous dual sidecars,
and verifies the sidecar's chapter ID/source filename against the manifest.
Ordinary projects need no manifest and retain their existing file workflow.

Anchor resolution follows Galley: exact UTF-8 byte selection/hash and source hash,
then unique occurrence or unique prefix/suffix context after source changes.
Overlapping proposals are unavailable. Browser navigation maps resolved byte
ranges to CodeMirror UTF-16 positions with normalized newlines.

Accepting a proposal replaces its exact source bytes, removes only that proposal,
and reanchors deterministically mapped active records. Signals consumed by or
partially crossing a replacement retain their original anchors; unavailable
records remain untouched. Unknown top-level, record and anchor metadata remain
raw JSON, including large integers. The manifest and other books are unchanged.

Chapter and sidecar publish through one normal project version with expected-head
checks and a stable operation receipt. A retry after a lost response returns the
original version, including when the head has since advanced. The browser refreshes
the current head after success. Passive synchronization never applies proposals.
Previous source and sidecar states remain available in project history.

Limits: text/manifest/sidecar at most 1 MiB each and at most 1,000 records per kind.
Unsupported/malformed inputs remain accessible as ordinary files; no implicit
migration or data repair occurs.

## Evidence and remaining qualification

- Go integration tests compare accepted source bytes and review JSON with the
  independent Galley `applyEdit` result; use the Android project's real fixture;
  cover duplicate chapter/book IDs across two folders, stale/ambiguous/conflicting
  anchors, unknown metadata, identity mismatch, ownership, concurrent versions,
  note/remove isolation and lost-response replay after head advancement.
- Browser tests use authenticated HTTP and disposable real server/database. They
  exercise review/edit mode switches, rapid Cyrillic typing, draft preservation,
  save/reclassification, explicit acceptance, lost response, remote changes,
  manual rewrites with stale records, and rejection on desktop/mobile viewports.
- The optional `scripts/verify-pocket-roundtrip.ts` reads the actual browser-test
  saved chapter/review through Galley's `decodeReview`, `encodeReview`,
  `classifyReview` and `openChapter`. Fixture provenance is recorded in
  `project/testdata/pocket/README.md`.

Verification completed on Linux:

- `go test -race -tags sqlite_fts5 ./...` and `go vet -tags sqlite_fts5 ./...`: pass.
- Frontend build/typecheck and Vitest: 162 tests pass.
- Full file-workspace Playwright suite: 24 tests pass across desktop/mobile Chromium.
- Actual browser-saved output reopened through the Galley codec/engine: one
  remaining active edit and two active signals; encode/decode round trip passes.
- `git diff --check`: pass. The build retains its existing bundle-size warning.
- `staticcheck` is not installed; it was not run.

The follow-up in the same working branch adds selection-based annotation
composition/editing, chapter notes and per-chapter session undo/redo. Book/chapter
ordering/management and semantic metadata editing were subsequently implemented in the [book management increment](2026-10-10-book-management.md); its remaining boundaries are documented separately.
Physical Android and installed Galley UI requalification have not been performed
for this slice. macOS/Windows qualification remains deferred until release.


## Follow-up: composition, decisions and author-controlled sections

Mechanics are adapted from Galley `addSignal`, `setChapterNote`, `removeRecord`,
`applyEdit` and `InputHistory`, with Edda's server versions taking the role of
in-memory before/after snapshots. Creation/editing preserves record identity,
exact byte anchors and unknown metadata. Changing an unavailable record's comment
or proposed replacement does not silently relocate its anchor. Proposal creation
is an explicit selected-span form; Galley's derived-text typing/IME grouping is
not claimed as implemented.

“Начать рецензию главы” explicitly enrolls an ordinary Markdown file in the adjacent
manifest without changing its source or creating content directories. An existing
manifest keeps its schema version, identity, chapter order and unknown metadata.
An existing unbound sidecar is not adopted by guessing a chapter identity. The
first actual annotation creates the sidecar. No Pocket/Galley installation is
required to start this workflow.

Undo/redo is session-local, per chapter, capped at 100 retained undo entries, and
publishes normal project versions. Each step restores only the changed chapter
and review files. Later conflicting changes to those files are refused; unrelated
KB/project work remains intact. Undoing first annotation creation removes its
new sidecar and redo recreates it. A new decision clears redo. Receipt IDs remain
stable across lost responses and failed refreshes. Text-editor undo remains its
separate normal input history.

Optional book sections assign stable folder IDs, custom labels and explicit
content purposes. No filesystem moves, stage transitions or readiness decisions
are inferred. Ordinary projects and sections containing only “ready” text work
without KB/draft folders. Starting new files/imports in a section defaults to its
folder. See [ADR 0016](../adr/0016-author-controlled-workspace.md).

Added tests cover first-review creation in an ordinary project, comment/proposal
updates through both editor surfaces, individual undo/redo, original CRLF source
preservation, source conflicts during undo, lost responses while creating reviews
and saving sections, author-selected roles without drafts, and folder rename
identity. Generated review/source/manifest files are exported by browser tests for
real Galley codec/engine checks. Physical-device and packaged-app qualification
remain separate from these checks.

Follow-up verification: full Go race/vet and frontend build pass; all 162 unit
tests and 24 desktop/mobile file-workspace tests pass. The manifest, chapter and
review created entirely in Edda decode through Galley and reopen in its review
engine with one active edit and one active signal.
