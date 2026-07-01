# Milestone 5 Phase 1: File-First Layout Foundation

## Goal

Add the first file-first foundation: an Edda layout scanner that validates and classifies the `alchemist`-style project folder, plus a small `edda` CLI entry point that can report folder status without touching story content.

This phase does not replace database-backed editing yet. It establishes the layout contract that later draft/save/checkpoint/sync work will build on.

## Source Context

The target layout is the stronger `alchemist` shape referenced in the file-first spec and verified from `/Users/inkyquill/Yandex.Disk-dark13th.localized/writing/alchemist`:

- top-level guidance files: `AGENTS.md`, `BOOTSTRAP.md`
- content roots: `story/`, `characters/`, `worldbuilding/`, `storyline/`, `drafts/`
- index files under roots: `_index.md`
- nested worldbuilding categories such as `worldbuilding/magic`, `worldbuilding/places`, `worldbuilding/monsters`, `worldbuilding/culture`
- project-local skills under `.agents/skills/`

Elysium remains a conversion source, not the target. The scanner should reject or warn about arbitrary layouts instead of trying to infer every possible personal structure.

## Task 1: Add `fileproject` Layout Scanner

Create a new Go package, `fileproject`, with typed layout scanning.

- Add types:
  - `ProjectLayout`
  - `LayoutFile`
  - `LayoutKind`
  - `LayoutWarning`
- Scan ordinary Markdown files and classify:
  - `story/*.md` as story prose,
  - `characters/*.md` as characters,
  - `worldbuilding/**/*.md` as worldbuilding,
  - `storyline/*.md` as storyline/planning,
  - `drafts/*.md` as drafts,
  - top-level `AGENTS.md` and `BOOTSTRAP.md` as guidance,
  - `.agents/skills/**` as project skill files.
- Ignore operational noise:
  - `.DS_Store`,
  - `.edda/state.local.json`,
  - `.edda/drafts/**`,
  - unresolved `.edda/conflicts/**`.
- Return warnings for missing recommended roots or missing root `_index.md` files, but do not fail if a smaller project has not filled every root yet.
- Fail only when the root is not a directory or required layout identity cannot be established later by `.edda/project.json`.

## Task 2: Add `.edda/project.json` Metadata Types

Add minimal metadata read/write support.

- Define `ProjectMetadata` with:
  - `schemaVersion`
  - `layoutVersion`
  - `id`
  - `title`
  - optional `serverUrl`
- Add `ReadMetadata(root)` and `InitMetadata(root, input)` helpers.
- Use `.edda/project.json`.
- Keep `state.local.json` out of syncable metadata.
- Add tests for:
  - initialization creates `.edda/project.json`,
  - read rejects malformed JSON,
  - scan includes metadata when present.

## Task 3: Add CLI Skeleton

Add `cmd/edda`.

- `edda status [path]` scans the folder and prints:
  - project title/id if metadata exists,
  - count by layout kind,
  - warnings,
  - whether `.edda/project.json` is missing.
- `edda init [path] --title "Title"` creates `.edda/project.json` for an existing Edda-shaped folder.
- Keep output plain text and writer-facing.
- Do not implement sync/checkpoints in this phase.

## Task 4: Tests And Fixtures

Add small fixtures under `fileproject/testdata`.

- `alchemist-lite/` should mirror the target shape without copying private project text.
- `partial/` should prove smaller projects get warnings but still scan.
- `invalid/` should prove malformed metadata fails.
- Tests should cover classification, warning behavior, metadata init/read, and CLI status/init smoke with temp directories.

## Task 5: Roadmap And Verification

- Add a Milestone 5 phase tracker to `docs/roadmap.md` and mark Phase 1 implemented after code passes.
- Run:
  - `mise run test`
  - `mise exec -- bun run test`
  - `mise exec -- bun run build`
  - `mise exec -- bun run test:smoke`
  - `git diff --check`

## Self-Review Notes

- Keep the package independent from the existing database project service for now.
- Avoid reading or writing arbitrary files outside the provided root.
- Keep IDs stable enough for metadata, but do not implement `.edda/ids.json` in this first slice.
- Do not copy private `alchemist` content into fixtures; recreate only the shape and tiny placeholder text.
