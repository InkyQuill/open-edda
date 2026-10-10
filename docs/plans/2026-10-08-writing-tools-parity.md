# Writing Tools parity baseline

Source snapshots inspected on 2026-10-08 (working trees may include separate
uncommitted work; hashes identify the committed baseline, not a release):

| Tool | Baseline |
| --- | --- |
| Galley Desk | `5de5ae75cd6bac8ab3a6a853fc8fa8b3a5d59e31` |
| Timeline Helper | `285b022bf8909d46471cda0fca54d05fbc92bf65` |
| Galley editor/themes | `54808abce1898f5ad74c204f8c3ed84750382116` plus local Edda theme addition |
| Pocket Editor | `2ac9a6ba809647fbe2578c9c9142f611106bbc2f` |

## Current browser parity

| Workflow | Edda status | Reference / remaining contract |
| --- | --- | --- |
| Portable file tree, binaries, empty directories | Delivered | Versioned manifests, CLI round trip |
| Live Markdown, source mode, translation comparison | Delivered | Galley Editor, existing browser drafts and explicit saves |
| Theme families and system scheme | Delivered | Published `galley-themes` 0.17.0 with Edda family |
| Book manifest and chapter order | Delivered locally: book/chapter titles, spine ordering, explicit addition/removal; missing chapters retained | Galley `pocket-format` manifest v1/v2, `docs/user-guide.md` |
| Chapter metadata | Delivered locally: YAML editor, byte-preserved body and guarded review reanchoring | Galley metadata editing must preserve frontmatter/unknown data |
| Review reading: notes, signals, proposed edits | Delivered in workspace; shared panel in editing/review modes | `pocket-format/src/documents.ts`, schemas, `availability.ts` |
| Anchor classification | Delivered through Go adapter, checked against shared fixtures | `anchors.ts`: exact UTF-8 byte offsets, source/selection hashes, unique/contextual match; stale/ambiguous stay visible |
| Explicit edit application | Delivered with expected-version checks and replay receipts | `review-engine/src/apply.ts`: atomically publish chapter + sidecar; remove applied edit; re-anchor only deterministically mapped records; retain unavailable/conflicting records |
| Review decisions/session undo | Create/edit/accept/reject/close and per-chapter session undo/redo delivered; project history persists versions | `signals.ts`, `history.ts`; unknown metadata retained |
| Timeline events/plotlines | Files preserved; semantic editing pending | Timeline Helper domain/filesystem contracts require inventory before porting |
| Calendar and temporal projections | Pending | Preserve native shared calendar/timeline representations |
| CWS directory roles | Files preserved; metadata interpretation pending | Optional integration and arbitrary layouts remain required |
| Agent skill parity/scopes | Pending | Inventory current CWS sources; map operations to versioned file tools before enabling helpers |

## Review slice delivered 2026-10-10

INT-06 now uses a Go server adapter for the Pocket/Galley wire contract. The
existing Node-only shared codec and review engine provide independent golden
fixtures and an optional live codec verification script; they are not browser
runtime dependencies. See [implementation and checks](../audit/2026-10-10-pocket-review.md).

The author can keep the review panel open while editing, save a draft to recheck
anchors, and explicitly accept a proposal. Review mode makes the source read-only;
switching modes preserves the draft. Both modes expose the same review decisions.

Next: detailed Timeline inventory. Book/chapter titles, ordering, existing-file addition/removal and YAML metadata editing are delivered locally; atomic chapter creation and verified adoption of existing sidecars are also delivered; packaged-client qualification remains follow-up work. Packaged
Galley and physical Android requalification remain distinct from codec checks.

Acceptance must include two book folders sharing chapter/book IDs, stale and
ambiguous selections, concurrent chapter/sidecar changes, atomic apply with a lost
response, unknown metadata preservation, and reopening through real Galley/Pocket
codecs. Passive synchronization must never apply proposed edits to Markdown.

All integrations remain optional. Windows/macOS qualification waits until release;
Linux and browser behavior remain part of feature delivery.

## Author control and review composition, 2026-10-10

Galley signal creation, separate proposals, chapter notes and session history are
adapted to Edda's versioned HTTP writes. Undo/redo restore only the affected chapter
and review pair, refuse subsequent conflicting edits, and keep unrelated KB work.
New proposals do not modify canonical source. Current implementation exposes
explicit selection-based proposal creation rather than Galley's derived-text
typing/grouped IME history model.

Optional custom folder sections now provide whole-book navigation without CWS,
mandatory drafts or readiness stages; [ADR 0016](../adr/0016-author-controlled-workspace.md)
is authoritative. Automatic CWS role discovery remains P5.
