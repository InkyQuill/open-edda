# Milestone 5 Phase 2: File Index And Stable IDs

## Goal

Add a rebuildable file index for Edda layout folders and stable item ID mapping through `.edda/ids.json`.

This phase still does not replace the web editor's database-backed content. It creates the bridge: Edda can scan a folder, assign stable IDs to files, write those mappings into `.edda/ids.json`, and mirror the file index into SQLite for later search/project-map migration.

## Current State

- `fileproject.Scan(root)` validates and classifies the `alchemist`-style layout.
- `.edda/project.json` can be initialized and read.
- `cmd/edda init/status` can inspect a folder without touching content.
- SQLite has no file-first index table yet.

## Task 1: Add `.edda/ids.json`

- Add `IDMap` and helpers in `fileproject`:
  - `ReadIDMap(root)`
  - `WriteIDMap(root, map)`
  - `AssignStableIDs(root, layout)`
- Map stable item IDs to relative paths.
- Preserve IDs for unchanged paths.
- Generate IDs for new indexed files.
- Remove mappings for deleted files during rebuild.
- Ignore operational files and keep `state.local.json` out of the map.
- Add tests for stable IDs across repeated scans and new IDs for new files.

## Task 2: Add SQLite File Index Schema

- Add a migration for a `project_files` table:
  - `id`
  - `project_id`
  - `relative_path`
  - `kind`
  - `title`
  - `sha256`
  - `bytes`
  - `updated_at`
  - uniqueness on `(project_id, relative_path)` and `(project_id, id)`
- Add indexes for project/kind/path.
- Keep this separate from `content_items` while migration proceeds.

## Task 3: Add File Index Service

- Add `fileproject.Indexer` or `projectfile` service that:
  - scans a root,
  - assigns stable IDs,
  - upserts file index rows for one project ID,
  - removes index rows for files that disappeared from the folder,
  - returns counts by kind and warnings.
- Use explicit SQL in this package rather than regenerating all existing sqlc output for the first slice.

## Task 4: Extend CLI Status

- `edda status [path]` should report whether `.edda/ids.json` exists.
- Add `edda status --write-ids [path]` to create/update IDs without writing SQLite.
- Do not add sync/checkpoint behavior yet.

## Task 5: Tests And Verification

- Tests:
  - IDs are stable across repeated scans.
  - Deleted files are removed from `ids.json`.
  - SQLite index upsert mirrors current files and removes disappeared files.
  - CLI `status --write-ids` creates `.edda/ids.json`.
- Run:
  - `mise run test`
  - `mise exec -- bun run test`
  - `mise exec -- bun run build`
  - `mise exec -- bun run test:smoke`
  - `git diff --check`

## Self-Review Notes

- Do not migrate agent tools or web editor reads in this phase.
- Stable IDs belong to file-backed items, not every operational `.edda` file.
- The index is rebuildable; the folder plus `.edda/ids.json` remains canonical.
