# Folder inventory and first import

Implemented 2026-10-06. Import uploads into an empty file project and records a local `.edda` binding for subsequent `send/take` synchronization. There is no migration of old Open Edda data: the user confirmed the application is still in development and there are no old projects to support.

## Usage

For ongoing synchronization of a local project, use `edda send FOLDER`. For initial import into an existing empty project, create an empty project with `edda create` or in the browser; omit `--project` to select it interactively. Build the helper with `go build -tags sqlite_fts5 -o edda ./cmd/edda`.

```sh
edda import /path/to/book --dry-run
edda import /path/to/book --dry-run --exclude .agents/skills/local-link
edda import /path/to/book --server https://edda.example --project PROJECT_ID --exclude .agents/skills/local-link
```

The upload command uses saved login or an explicit bearer token in `OPEN_EDDA_TOKEN`; `OPEN_EDDA_URL` can supply the base URL. Tokens are obtained from the existing `/api/auth/login` API. `edda login` now stores the connection in private user configuration; import reuses it. See [CLI synchronization](cli-sync.md). Credentials are not accepted as command-line flags or written into author folders. Use HTTPS for a remote server; HTTP also works for local development. Redirects are rejected.

`--dry-run` prints a readable inventory without network access or writes; add `--json` for structured inventory. Normal import displays a readable summary before sending bytes, or a JSON publication receipt when `--json` is supplied. Inspect a dry run first. `--exclude` accepts repeatable exact relative paths (a directory excludes its subtree), not globs; every exclusion is listed. The source root must be a directory.

Default exclusions, at any depth: `.git`, `.edda`, `node_modules`, `__pycache__`, `.DS_Store`, `.env` and `.env.*`. This is an explicit initial policy, not a secret scanner. Other hidden files and unknown formats are included, including CWS guidance, `project.md`, `kb`, `work`, translation sources, Pocket sidecars and `timeline.yaml`. Excluded directories are reported as whole subtrees rather than listing each child.

Symlinks and special files block the import unless explicitly excluded. Links are not flattened or followed. Case/NFC collisions and nonportable names are checked against the same validator as server publication. Empty directories, empty files, binary bytes, Unicode names, frontmatter and line endings are preserved; permissions, ownership and filesystem timestamps are not transported.

Limits match the initial server: 10,000 entries, 64 MiB per file, 1 GiB logical bytes, 8 MiB publication JSON. Inventory hashes files in bounded streams, and import stages up to the project limit in a private temporary directory outside the author folder. It never runs CWS migration, `edda init`, ID synchronization or layout normalization.

## Transaction and recovery

1. Inventory and validate locally. Problems prevent all network mutation.
2. Look up the deterministic import receipt; an already committed import returns its original version, even when head has advanced. The receipt's entries must match.
3. Pin the empty remote head; reject an occupied project.
4. Copy selected files into temporary staging, checking size/hash against inventory. Rescan to detect observed additions, removals, replacements and modifications before uploading.
5. Upload only staged bytes; the server verifies them. Publish the entire manifest with expected head and deterministic operation ID.
6. Verify the returned receipt and save the local binding before reporting success. Remove staging on normal completion, handled errors and interrupt cancellation.

The inventory is a frozen import snapshot. Changes made after staging remain local and are not included. Rescanning detects observed inconsistencies; it is not an atomic snapshot of a live filesystem and does not lock other tools. Keep source files quiescent while preparing an import. A hard process kill can leave a private temporary staging directory; the next invocation rescans and retries instead of trusting it.

Failed/interrupted uploads can leave unreferenced server objects; no partial tree becomes visible. A lost publication response is an error on that invocation; rerunning the same command checks the receipt and creates no duplicate version. Concurrent web changes cause a conflict, never an overwrite. After verifying the receipt, import atomically installs `.edda/checkout.json` with that exact base version, server, project and exclusions. It never replaces existing metadata or marks later local edits as synchronized. If binding installation fails after publication, retry import to recover the receipt, or attach the matching folder. Subsequent changes use `send/take`; no project ID is needed inside the folder or its subdirectories. Stable import entry IDs derive from their initial relative paths; future synchronization must retain those IDs across renames rather than recomputing them.

## Evidence

`fileproject/inventory_test.go` verifies arbitrary CWS/translation paths, binary and empty content, hidden tool metadata, directory preservation, exclusions, links/FIFOs, case collisions, cancellation and changes during staging. `cmd/edda/import_test.go` uses a real authenticated HTTP server and temporary SQLite/object storage to verify exact download bytes, partial upload failure, retry after commit with a dropped response and subsequent head advance, occupied-project refusal, authentication failure and redirect credential isolation.

A read-only inventory of the actual `only-sense-online` folder found 7,957 entries / 875,201,741 bytes, four cache exclusions and nine local Hieronymus skill symlinks requiring explicit exclusion. It performed no upload and wrote no project files. The actual `alchemist` folder also passed a read-only inventory: 472 entries / 13,680,019 bytes, no exclusions or problems. These counts describe the observed development snapshots, not permanent project inventories.
