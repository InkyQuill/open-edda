# Author-controlled book management and chapter metadata

Implemented locally on `codex/pocket-review-workspace`; not deployed. Reference:
Galley Desk `packages/book-storage/src/book.ts`, `session.ts`, and desktop
`shared/metadata.ts`. This continues [review delivery](2026-10-10-pocket-review.md)
and [ADR 0016](../adr/0016-author-controlled-workspace.md).

## Delivered behavior

- Book selector for multiple adjacent Pocket manifests in a project; ordered
  chapter navigation remains alongside the complete file tree and optional KB/
  draft/fragment/main-text sections. Missing chapters stay explicitly visible.
- Explicit book title, chapter labels, ordering, addition and removal from the
  spine. Removal leaves prose and review files intact. New chapters can be
  created in the same transaction as their manifest entry; filename/title are
  author choices. No draft, readiness status, KB or predecessor is required.
- Existing Markdown with a sidecar can be added only after validating its source
  path and chapter UUID. Dual sidecars, conflicting UUIDs, filename collisions
  and malformed reviews are refused. Review bytes are unchanged on adoption.
- Server-side manifest mutation preserves the original schema version and raw
  unknown fields, including integers beyond JavaScript's exact range. New
  chapters receive stable file identities for operation replay.
- Separate arbitrary YAML metadata editor for Markdown, including unenrolled
  ordinary files. Metadata never auto-enrolls a file in a book/review. Frontmatter
  title appears above the current document; all metadata properties are optional.
- Metadata writes retain BOM and body bytes, use the existing frontmatter EOL,
  and update resolved review anchors atomically with the source. Replacement is
  minimized to avoid invalidating anchors borrowing an unchanged delimiter.
  A review intersecting changed metadata is refused; stale/ambiguous records
  remain untouched. Unclosed frontmatter must be repaired explicitly in source.
- Metadata saves start a fresh browser review-decision history, as in Galley;
  durable project versions remain available. Metadata and book operations use
  expected versions and retryable receipts. Unsaved source/review drafts block
  metadata writes; unrelated file drafts survive book changes.

## Verification

- Go tests cover exact CRLF/CR/mixed-line-ending and BOM body preservation,
  metadata overlap rejection, YAML errors, active review reanchoring, source/pair
  undo, receipt replay, and ordinary Markdown without workflow files.
- Book API tests cover unknown numeric precision, missing chapter ordering,
  unchanged prose, duplicate/stale rejection, non-destructive removal, atomic
  creation and identity-checked review adoption.
- Frontend: 174 Vitest tests; production build passes (existing bundle-size
  warning). Browser tests use a disposable real Go server/database, desktop and
  mobile Chromium. Full 26-test workspace suite passed; focused book/review
  scenarios rerun after extending creation/adoption: all 6 passed.
- Browser book scenario changes title/order, adds existing and reviewed chapters,
  creates a chapter, retries a lost save response, edits YAML and compares exact
  downloaded body bytes. Mobile layout screenshot inspected; modal scrolls.
- Full Go race suite and vet passed; project race checks rerun after book changes.

## Remaining boundaries

Packaged Galley and physical Android cross-client qualification are still pending;
macOS/Windows qualification stays deferred until release. The manifest retains
Pocket's direct-child book format; arbitrary project folders remain available
through the file tree and author-defined sections. File relocation/renaming is
still the generic file workflow and is not an automatic manifest rewrite.
Removing from the spine is distinct from deleting files. Dedicated Timeline
inventory and implementation are the next roadmap increment.

## PR #7 review follow-up

All three actionable CodeRabbit findings were reproduced or confirmed against the
reviewed head and fixed:

- Record a confirmed review decision in session undo/redo before refreshing the
  workspace. Retain its operation receipt until refresh succeeds and record the
  resulting version only once, so both manual refresh and mutation retry recover
  without losing or duplicating history entries.
- Build each 128-code-point anchor context by bounded UTF-8 scanning instead of
  allocating rune arrays for the entire surrounding chapter. Unicode/CRLF window
  and wire-field equivalence tests pass. On a 1,050,000-byte chapter, the local
  benchmark changed from 2,868,180 B/op and 1.50 ms/op to 896 B/op and 0.414 ms/op;
  this measures anchor creation, not complete synchronization.
- Validate the existing delimited block as an empty block or a valid YAML mapping
  before metadata replacement. Scalars, lists, invalid YAML and duplicate keys
  are rejected without publishing a version. UI saving is blocked until the
  original source is repaired, protecting prose between Markdown scene breaks.

Validation: production build, 174 frontend unit tests and all 8 focused desktop/
mobile book/review scenarios passed. The browser regressions cover confirmed
write + failed refresh for decisions, undo and redo, plus receipt retry without
extra history entries and the scene-break metadata guard.
