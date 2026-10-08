# Portable-file follow-ups and live editor

2026-10-08. Work continues on `codex/writing-workspace-sync-galley-017`, preserving
the pre-existing uncommitted roadmap/backlog edits. No deployment or package
publication was performed.

## Delivered in this slice

- ZIP import in the current file workspace: inspect names, omitted local settings,
  select destination, then publish one complete version. Nested paths, empty
  directories, UTF-8 and binary bytes are preserved. Existing paths are not
  overwritten. Retry after a lost publication response retains its operation ID.
- ZIP staging uses the project authorization boundary and immutable object store;
  staging alone does not create a version. No archive path is extracted onto the
  host filesystem. Reject traversal, absolute/Windows paths, links/special files,
  duplicate/case/Unicode collisions, file/directory collisions and corrupt bytes.
  Limits: 64 MiB uploaded ZIP, 256 MiB expanded total (or smaller configured project
  limit), configured per-file and entry limits. ZIP is the supported extraction
  format; other archive types remain ordinary binary files.
- History compares a selected saved project tree with the currently loaded saved
  version, matching stable IDs for moves, displaying additions/removals/edits and
  line differences for UTF-8 files up to 1 MiB. Binary/large files retain download
  actions for both versions. Drafts are excluded explicitly.
- Live Galley Markdown in the actual file workspace, source-mode toggle, plain
  text fallback, preserved draft/save/conflict/history flows and translation pane.
- Shared Galley theme family selection with persisted light/dark/system scheme;
  app, dialogs and editor use the same catalog. Missing family/scheme falls back
  to Edda. `edda-light` and `edda-dark` added in the Galley repository.
  After the shared release, Edda uses published `@inkyquill/galley-editor` and
  `@inkyquill/galley-themes` 0.17.0. The temporary local archive and Docker vendor
  copy have been removed. Edda consumes the shared Edda palettes directly.

## Verification

- Go project/app race tests and vet passed after archive implementation.
- Frontend build/typecheck and 156 unit tests passed.
- All 16 file-workspace desktop/mobile scenarios passed after live editor
  integration, including draft recovery after reauthentication.
- Additional desktop/mobile theme scenarios passed across every catalog family
  and both schemes; checked root/editor palette agreement, system changes and
  persisted settings. Mobile screenshots inspected for comparison and dark live
  editor. Full final checks are recorded below when complete.
- Shared themes build and 11 catalog tests passed in the Galley repository.

## P3 reconciliation

Current Pocket Editor source contains `EddaGateway`, binding identity isolation
and provider/account/project/folder selection. Its checked-in
`docs/audit/2026-10-07-edda-p3.md` records successful Android/Edda and
Edda CLI/Galley codecs/Android round trips. This corrects the stale Edda roadmap
label; Android tests were not rerun in this slice. Galley Desk GUI was not part of
that codec round trip and must not be implied by it.

## Still pending

INT-06 semantic Pocket review decisions/edit application and the broader Galley
Desk parity matrix; Timeline Helper event/calendar editing; INT-05 optional
bidirectional CWS roles/integration; P5 skill inventory/scopes/web parity.
These are separate workflows and are not marked delivered by editor/theme work.
Windows/macOS qualification is deferred to release by the author on 2026-10-08.
Hardware power-loss and additional storage-driver qualification remain separate.

## Final checks

- `go test -race -tags sqlite_fts5 ./...` and `go vet -tags sqlite_fts5 ./...`: passed.
- Additional archive corruption, implicit-directory-count and total expanded-byte
  regression tests passed with race detection.
- `bun run build`, `bun run test`: passed, 156 tests.
- `bunx playwright test --config playwright.files.config.ts`: all 18 desktop/mobile
  scenarios passed against the disposable real server on 2026-10-08.
- `bun install --frozen-lockfile`: passed initially with the shared-theme archive; the registry-based 0.17.0
  replacement is checked below.
- `git diff --check`: passed in both Edda and Galley repositories.
- Clean Docker frontend-stage build: **not verified**. Stopped after more than
  three minutes without progress in dependency installation. A separate Bun
  container request to `https://registry.npmjs.org/react` timed out after 10 seconds.
  This is evidence of a container-network limitation in this run, not a confirmed
  application/build defect. The subsequent 0.17.0 update removes vendoring;
  Docker uses the regular manifest and lockfile again. Retry clean container
  build when registry access is available.

## Published Galley 0.17.0 verification

Both shared packages now resolve to registry version 0.17.0 in `bun.lock`.
`@codemirror/search` is declared explicitly for the editor's new peer contract.
Frozen installation with the CI Bun 1.4.0 version, production build/typecheck,
156 frontend unit tests and all 18 desktop/mobile file-workspace tests passed.
The browser tests include every shared theme family, both schemes, system-theme
changes, Edda fallback, live/source editing and draft/save recovery. Go race
checks and vet also passed. No vendored theme dependency remains.
