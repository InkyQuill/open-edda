# Ignore rules, attachment diagnostics and project deletion

Implemented and deployed 2026-10-07.

- Root `.eddaignore`: glob patterns, directory/root matching, comments and `!` re-inclusion. The file travels with the project. Additional defaults exclude `.remember`, `.creative-writing` and root `.agents/skills`; ordinary guidance stays portable. Full syntax and tracked-file behavior: [folder import](../architecture/folder-import.md#eddaignore).
- New patterns never silently delete tracked files. Tests cover send/take, preservation of local ignored data and refusal of incoming ignored paths. Existing snapshot/recovery checks still pass.
- External symlinks anywhere in the tree are automatically excluded without copying targets. `edda rm PATH...` stages untracking while preserving local files; `--undo` reverses its local exclusion. Integration tests cover publication, historical retrieval, disjoint and conflicting remote updates, nested paths and atomic rejection of invalid batches.
- `attach` reports offending paths directly. `import --dry-run` prints counts and problems; `--verbose` lists all paths, `--json` retains the structured report. A bound folder's saved exclusions/tracked files are included in dry-run policy.
- Project-list delete action requires an exact title. The authenticated server independently checks ownership/title and removes dependent database records transactionally. Local folders and existing backups stay intact; immutable object bytes await future garbage collection.

Verification:

- `go test -race -tags sqlite_fts5 ./...` and `go vet -tags sqlite_fts5 ./...`: passed. Initial untagged test invocation failed because SQLite FTS5 requires the documented build tag; corrected run passed. `staticcheck` unavailable on the workstation.
- Frontend typecheck/production build and all 147 unit tests: passed. Existing large-JavaScript-chunk build warning remains.
- Six headless browser integration cases (desktop/mobile): passed, including cancellation, incorrect title, server rejection and retry, file edits, conflicts and history.
- Read-only inventory of Only Sense Online: 4,162 files, 509 directories, 668,270,226 bytes; three excluded subtrees; no problems. The nine external skill symlinks are excluded with their containing directory. No source files or bindings changed; no author content uploaded.
- Full container build and CLI smoke check passed. Installed `~/.local/bin/edda` verified against the same real folder.
- Verified backup, image transfer checksum, k3s rollout and public HTTPS health. Live UI created/deleted only a disposable test project; all four existing project heads remained unchanged. [Deployment evidence and rollback](../architecture/yggdrasil-deployment.md#ignore-rules-and-project-deletion--2026-10-07).
