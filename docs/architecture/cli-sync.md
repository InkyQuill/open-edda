# CLI synchronization

Implemented 2026-10-06. `login`, `logout`, `projects`, fresh-directory `get`, offline `status` and transactional `send` now transfer real project bytes. `take` updates existing copies and `conflicts`/`resolve` manage explicit file choices. The local-computer delivery includes creation, attachment, explicit moves, history and restore; platform and operational limits are listed below. No legacy-data migration or backward-compatibility layer is required.

## Author workflow

Upload an existing local project:

```sh
edda login
edda send ./book
# First send asks whether to create a new project or select an existing one.
# Edit files, then repeat:
edda status ./book
edda send ./book
edda take ./book
```

Download an existing server project:

```sh
edda projects
edda get
# Select the project and enter a new local folder when prompted.
```

The CLI is human-first. In a terminal, omitted required values are requested; connected-project commands discover the nearest `.edda` in the current directory or its parents, and explicit arguments skip the corresponding prompts. Outside a project, folder prompts offer the current directory where appropriate. Optional settings retain defaults. `--json` selects machine output on `projects`, `create`, `history` and `import` and disables prompts for those commands. Without a terminal, required inputs must be supplied as arguments; commands never wait for a dialogue. `edda COMMAND --help` and `edda help COMMAND` explain each operation with examples. Server addresses and credentials are reused from login.

For scripts, the first upload can be `edda send ./book --title "My book"` or `edda send ./book --project PROJECT_ID`; optional `--server URL` and repeated `--exclude relative/path` apply only to first attachment. Ordinary sends use the persisted connection and exact exclusions plus the current root `.eddaignore`; see [ignore rules](folder-import.md#eddaignore). The first-send workflow validates the folder before creating a project, attaches it, then runs the regular transactional send. If upload fails after attachment, repeat `edda send ./book`; existing recovery rules apply. Existing `.edda` metadata is never replaced. Cancellation before choosing/creating a project has no remote or local side effects.

`edda import ./book --dry-run` previews the folder without uploading; `--json` returns its inventory. A real `import` uploads into an empty project and saves `.edda/checkout.json` containing the server, project ID, acknowledged imported version and exclusions. Its JSON output is the publication receipt. Afterwards, run `send`, `take`, `status` or `history` from any subdirectory without supplying a folder or project ID. The binding is relative to the containing folder, so moving the whole folder preserves it. Existing metadata is never replaced; use `send` for an already connected project. An older imported folder can be connected with `edda attach FOLDER`, which verifies its files against the selected project without reuploading.

`edda login` prompts for the server URL (offering `OPEN_EDDA_URL` or the saved server as a default), email and a hidden password. Press Enter after each answer; Ctrl+C cancels. `--server` and `--email` skip their respective prompts. For scripts, `login --email EMAIL --password-stdin` reads a piped password to EOF and requires a server from `--server`, `OPEN_EDDA_URL` or the saved configuration. Using `--password-stdin` directly in a terminal explains how to switch to interactive login. Passwords are never stored. A saved bearer token and server address live in `$XDG_CONFIG_HOME/open-edda/client.json` (otherwise the OS user configuration directory), with mode 0600. This is a private credential file, not an OS keyring. `logout` removes the saved token locally; it does not revoke JWTs server-side or unset shell environment variables. Existing token expiration still applies.

`edda projects` displays project titles and IDs in a readable table, or a helpful message when there are no projects. Use `edda projects --json` for the API JSON array in scripts.

`OPEN_EDDA_URL` and `OPEN_EDDA_TOKEN` override the connection for discovery/import/get. A checkout binds sends to its recorded server/project; a global URL change cannot redirect its content. A saved token is reused only for the same server URL. An explicitly supplied environment token takes precedence. Redirects are refused. Use HTTPS outside local development.

`get` requires a new destination under an existing parent. It downloads a pinned manifest and verifies every file's byte length and SHA-256 in a private sibling directory. Paths are validated before file creation. Files, directories and `.edda/checkout.json` are synced to disk before a no-replace directory rename installs the checkout. This installation currently uses Linux `renameat2`; other operating systems/filesystems are not qualified. Failed downloads expose no partial destination. A competing destination, including an empty folder or symlink, is never overwritten.

The nearest `.edda` is a discovery boundary: nested or invalid metadata never causes the CLI to silently choose a parent project. File paths for `move --from/--to` and `resolve --path` remain relative to that project root. Import saves the exact publication receipt as the synchronization base, including on retry after a lost response; a newer server head is reconciled by `take`.

## Existing folders and identity-preserving moves

`edda send FOLDER` guides first attachment and upload. For separate steps, create a project with `edda create`, then `edda attach FOLDER`; both request missing values. Explicit `--title TITLE` and `--project ID` skip those prompts. Attachment only installs private `.edda` metadata. An empty remote accepts any valid local inventory; a nonempty remote must match local paths and bytes exactly, otherwise attach refuses to invent a common ancestor. For divergent unrelated copies, get a separate checkout and reconcile explicitly. Repeated `--exclude relative/path` flags on attach persist exact local exclusions, useful for machine-local CWS skill symlinks. Status/send/take use those exclusions. A remote path overlapping one is rejected instead of silently deleted.

`edda move CHECKOUT --from old/path --to new/path` moves a file or whole directory, keeping IDs for its descendants and moving matching explicit exclusions. Destination parents must already exist. The intent is written before the filesystem move; if interrupted, repeat `edda move CHECKOUT` to finish metadata recovery. Other synchronization commands refuse an unfinished move. `send` publishes the move normally. External filesystem renames remain delete/add because guessing identity by content hash is unsafe.

`edda history CHECKOUT [--cursor ID]` lists immutable version summaries in pages. `edda get NEW_DIRECTORY --project ID --version VERSION_ID` checks out a historical version without changing server history. `edda restore CHECKOUT --version VERSION_ID` publishes that version as a new server version against the checkout base; retry of the same base/target returns the same receipt. Local files remain untouched. Follow it with `take`, which preserves local changes through the normal merge/conflict and recovery flow.

## Stop tracking without deleting local files

Inside a connected folder (including any subdirectory):

```sh
edda rm debug.log .creative-writing
edda status
edda send
# Undo a local tracking exclusion, before or after send:
edda rm --undo debug.log
```

`rm` changes only `.edda` tracking metadata. Files stay on this computer; directories include their whole subtree. A subsequent `send` removes the paths from the current server tree while retaining prior versions. Other clients receive that deletion through ordinary synchronization. Removed paths remain local-only on this checkout instead of being automatically re-added. `--undo` removes the exclusion recorded by `rm`; other `.eddaignore` and explicit rules still apply.

Paths are relative to the current directory; external paths and `.edda` are rejected. Multiple paths are recorded atomically. Pending send/move/update operations must finish first. Disjoint remote updates can merge before sending removals; concurrent changes to locally untracked paths require `rm --undo`, `take` and reconciliation before another removal. Moving a tracked parent carries its local-only child exclusions with it.

## Sending and recovery

`.edda/checkout.json` stores the server, project, acknowledged base manifest and optional pending operation, explicit exclusions and local move identities. It contains no credentials. CLI operations use a nonblocking per-checkout lock. File inventory has the same visible default exclusions and limits as [folder import](folder-import.md). Remote paths reserved for local state/cache or excluded environment files are rejected at checkout rather than being silently dropped on the next send.

A new send pins/checks the remote head against the local base, inventories local files and prepares a checked snapshot of missing contents under `.edda/pending-*`. Contents already referenced by the acknowledged base are reused by SHA-256 and byte length: unchanged files, copies and renames do not upload their bytes. Identical new contents are uploaded once even when used at several paths. The complete manifest is still published, including deletions and directories. Unchanged paths retain their IDs; new paths receive random IDs. Snapshot bytes, their directory and the pending request are synced before upload. Publication replaces the complete tree with the expected base and recorded operation ID.

The receipt must match project, parent, operation and entries before the local base advances and pending state is cleared. On any failure the prepared snapshot remains available. A retry checks the recorded operation receipt before uploading, so a committed operation with a lost response is acknowledged exactly once. Edits made after staging remain in working files and are shown as unsent against the newly acknowledged base; a subsequent send publishes them separately. Files are never rewritten by send. Interrupted preparation before recording state may leave unused private staging directories; they are excluded from inventory and not treated as pending operations.

Offline `status` compares working paths/hashes against the acknowledged manifest and reports pending sends. It does not claim to know the current remote head. `send` verifies the head even if local files are unchanged. Stale bases fail; remote and local work are both retained.

## Updating an existing copy

`take CHECKOUT` compares the acknowledged base, current local files and a pinned remote version. Independent changes to different files merge automatically. Concurrent changes to the same file, including delete-versus-edit, require a choice; there is no line-level merge. Local changes retained by the merge remain unsent until `send`.

```sh
edda take ./book
edda conflicts ./book
edda resolve ./book --path chapter.md --use local
# Or --use remote; choosing a deleted side removes that portable path.
edda take ./book
edda send ./book
```

Preparation saves `base/`, `local/` and `remote/` snapshots under `.edda/update-*`. Remote and base contents are found by SHA-256 in the verified private local snapshot and already prepared remote files before making any content request; only missing contents are downloaded. Every reused or downloaded file is checked against the manifest digest and byte length. These are private copies, not hardlinks to mutable working files. Working files remain untouched while conflicts await decisions. `resolve` only records a choice; `take` applies it. A structural conflict is reported as `.` and requires an explicit whole-tree choice. This can arise from incompatible file/directory changes or a remote identity-preserving rename concurrent with a local edit. A whole-tree choice supersedes previous individual choices. Choosing a directory selects its subtree.

For manual merging, read the three snapshots, edit the working file, run `take --restart`, and choose the new local version. Restart archives the old prepared plan and re-reads both sides; it does not delete its snapshots. Without restart, changed working files invalidate the prepared plan. A prepared update pins its remote version: later server changes require another take, and send still checks the current head.

Before applying, stop editors/watchers that write to the checkout. CLI operations lock each other, but cannot lock arbitrary Writing Tools processes. Application keeps the checkout root and `.edda` directory in place, journals top-level moves, retains original entries in `backup/`, and checks moved content before installing replacements. Files and directories are synced before acknowledging the new base. The filesystem tree is not atomically visible to external readers during this short application phase. Detected concurrent writes stop application for recovery; an editor holding an old file handle can write into the retained backup.

After interruption, run `take` again. Before the base commit, it restores touched original entries and preserves any displaced current entries in `displaced/`. After the base commit, it completes finalization. Recovery runs locally and returns before attempting another update; inspect the reported recovery directory, then run take again if needed. `send` is blocked while an update remains active.

Local snapshots preserve excluded files and symlinks inside portable top-level directories without following links. Reserved root entries (`.edda`, `.git`, `node_modules`, `__pycache__`, `.DS_Store`, `.env` and `.env.*`) cannot be synchronization units and stay entirely untouched: they are neither read nor copied into snapshots. Nested excluded data remains in snapshots because applying a directory update must preserve its descendants. Removing a portable directory containing excluded local data fails rather than deleting that data. Special local files such as sockets/FIFOs inside snapshot scope block update preparation. Nested caches may still consume substantial snapshot space. They are private recovery data, may contain local secrets, consume additional disk space and are currently retained without automatic cleanup.

For a pending send, take checks the operation receipt first. A committed operation is acknowledged before further reconciliation. If there is no receipt and the remote base has not changed, retry send first. If the remote head has advanced, the old publication cannot pass its base precondition: take reconciles current working files and retains the old pending snapshot and request in recovery data. Successful application clears that obsolete pending operation; failed/rolled-back application preserves it.

## Progress and performance

`send` and `take` show stages, current paths, processed bytes, elapsed stage time and completed/total file counts (with percentages when the count is known). Slow network/server requests continue emitting elapsed-time heartbeats about once per second. Transfer byte counts advance during individual files. Percentages describe the current stage, not a guessed overall completion time; a scanning stage counts entries without inventing a total. `--quiet` hides progress while retaining results and errors.

The optimization is at **whole-file content-object granularity**. A changed file still transfers its full contents; an unchanged digest transfers no contents. Manifest comparison does not load remote bodies. Local SHA-256 computation and concurrency/integrity checks still read files; take still creates verified private recovery copies, and publication still verifies server objects. No persistent mtime-only digest cache or block-delta protocol is implemented. Retries can resend objects uploaded before an unacknowledged publication, unless the publication receipt is already present.

See the [measurement and implementation record](../audit/2026-10-08-cli-sync-performance.md). macOS/Windows qualification remains deferred until release.

## Deliberate remaining work

- `get` creates a new directory; use `take` for existing checkouts. No force-overwrite mode or automatic line merge exists.
- Use `move` for stable identities. External renames are deletions plus additions; no identity guessing by hash.
- OS-keyring integration, separate device tokens/revocation, cross-history object negotiation, block deltas and resumed partial downloads are later optimizations. Authentication uses the existing expiring author token.
- Prototype commands for old local snapshots are not part of the network protocol. Network checkouts use the history/restore commands described above. No legacy-data conversion is required.
- Linux is the currently exercised platform. Hardware power-loss/PVC qualification and automatic recovery-data cleanup are outside the qualified delivery. [Docker and verified backup/restore](deployment-and-backup.md) are implemented.

## Verification

`cmd/edda/update_test.go` covers two-copy update/send round-trips, independent changes, local/remote conflict choices, delete-versus-edit, stale prepared plans, retained excluded files/symlinks, pending-send reconciliation and preservation of post-crash edits. Separate test processes exit at backup, installation, before-base-commit and after-base-commit boundaries; subsequent take verifies rollback or finalization from the actual persisted journal. These tests simulate process failure, not hardware power loss.

`cmd/edda/network_test.go` runs against actual authenticated HTTP handlers with temporary SQLite/object storage. Coverage includes two independent copies, downloaded text/binary bytes, stable identity for same-path edits, deletion/empty directories, stale-send refusal, corrupt-download rejection, reserved-state paths, lost publication response, durable pending retry, later local edits, saved login/project discovery/logout, private credential permissions and server scoping. Import tests separately exercise upload failure and redirect credential isolation. An actual full-filesystem upload/retry passed in the container check. Hardware power-loss/PVC qualification and Windows/macOS support remain unverified.

A separate runtime check built the actual server and CLI executables and ran them as separate processes with private temporary configuration/data. Saved login, discovery, fresh checkout, text/binary round-trip, a second-copy send, stale-copy rejection, server restart persistence and local logout all passed. The update slice also passed a separate built-executable run with independent local/remote edits, explicit conflict choice, send/take in both directions and persistence after server restart. Temporary author data and credentials were removed after the check. All Go packages passed race-enabled tests and `go vet`. No real author folders were modified.

Attachment/move tests also cover identity across rename/edit conflicts, interrupted moves, persistent symlink exclusions and dirty local work during server restore. The container acceptance script verifies the shipped executables and restoration onto a fresh volume.

The human-first CLI acceptance check is `python3 scripts/check-cli-interactive.py CLI_BINARY SERVER_BINARY` after building both executables with `-tags sqlite_fts5`. It uses a Linux pseudo-terminal, a real localhost server and temporary credentials/data to check prompts, first-send upload, project selection, synchronization/conflict recovery, history/restore, human/JSON output, cancellation, backup/restore and local prototype commands. It does not use production projects.

## Deleting a server project

The project list has a separate delete action. Its dialog explains that server files and version history become unavailable and requires the exact project title before enabling confirmation. Cancel is the default escape route; failures keep the dialog open for retry. Local folders remain unchanged and their old binding cannot synchronize after deletion.

`DELETE /api/projects/{id}` requires authentication, ownership and JSON `{ "confirmationTitle": "Exact project title" }`. A mismatched title returns 409; an empty confirmation returns 400. Project rows and dependent history/content records are removed in one transaction. Immutable object bytes are retained pending a future garbage collector, as with unreferenced upload objects; this operation does not promise immediate disk-space recovery. Existing backups remain separate.
