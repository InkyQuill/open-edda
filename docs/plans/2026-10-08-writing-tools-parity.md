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
| Theme families and system scheme | Delivered locally | Shared `galley-themes`; Edda theme publication pending |
| Book manifest and chapter order | Files preserved; semantic management pending | Galley `pocket-format` manifest v1/v2, `docs/user-guide.md` |
| Chapter metadata | Raw text editing; semantic editing pending | Galley metadata editing must preserve frontmatter/unknown data |
| Review reading: notes, signals, proposed edits | Sidecars preserved; semantic UI pending | `pocket-format/src/documents.ts`, schemas, `availability.ts` |
| Anchor classification | Pending | `anchors.ts`: exact UTF-8 byte offsets, source/selection hashes, unique/contextual match; stale/ambiguous stay visible |
| Explicit edit application | Pending | `review-engine/src/apply.ts`: atomically publish chapter + sidecar; remove applied edit; re-anchor only deterministically mapped records; retain unavailable/conflicting records |
| Review decisions/session undo | Pending | `signals.ts`, `history.ts`; preserve unrelated metadata and cross-client readability |
| Timeline events/plotlines | Files preserved; semantic editing pending | Timeline Helper domain/filesystem contracts require inventory before porting |
| Calendar and temporal projections | Pending | Preserve native shared calendar/timeline representations |
| CWS directory roles | Files preserved; metadata interpretation pending | Optional integration and arbitrary layouts remain required |
| Agent skill parity/scopes | Pending | Inventory current CWS sources; map operations to versioned file tools before enabling helpers |

## Next implementation boundary

Start INT-06 with the shared Pocket/Galley contracts, not an independent lossy
review schema. The current Galley packages use Node `crypto` and filesystem schema
loading, so a browser-safe shared entry point or server adapter is needed. Do not
bundle Electron/Node shell code into Edda or silently weaken anchor checks.

Acceptance must include two book folders sharing chapter/book IDs, stale and
ambiguous selections, concurrent chapter/sidecar changes, atomic apply with a lost
response, unknown metadata preservation, and reopening through real Galley/Pocket
codecs. Passive synchronization must never apply proposed edits to Markdown.

All integrations remain optional. Windows/macOS qualification waits until release;
Linux and browser behavior remain part of feature delivery.
