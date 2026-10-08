# Open Edda

Open Edda is a self-hosted web workspace and versioned project repository for writing and translation. The target workflow preserves existing CWS-compatible folders: create or import a project, work in the browser or a local working copy, and synchronize changes with previous versions retained automatically.

**Current status:** pre-v1. New browser projects support flexible file/folder trees, versioned text saves and downloads. CLI folder import into an empty project preserves flexible paths and exact bytes. The CLI supports project creation, existing-folder attachment, get/send/take, identity-preserving moves, history/restore and explicit conflicts. The web file UI supports import, rename/delete and version history/restore. Docker and verified online backups are available; see [deployment](docs/architecture/deployment-and-backup.md). Open Edda is in development with no old deployed projects; legacy migration/backward compatibility are not delivery requirements. `get` installs a verified project into a new directory; `send` publishes a staged version with retry receipts. `take` combines independent file changes and retains both sides of conflicts. Earlier prototype snapshot commands do not participate in the network checkout protocol. See [CLI synchronization](docs/architecture/cli-sync.md).

The immediate priority is projects and local-computer synchronization, then Pocket Editor. Web Galley Desk/Timeline Helper capabilities and local-agent skill parity follow later. See the [verified audit](docs/audit/2026-10-06-project-direction.md), [roadmap](docs/roadmap.md) and [delivery plan](docs/plans/2026-10-06-projects-and-sync.md).

The flat CWS project format is the structural reference, with support for nested book folders and preservation of existing directory trees. CWS installation and integration skills are optional. Distribution uses a Docker image with data on a volume or Kubernetes PVC; [ADR 0015](docs/adr/0015-storage-and-container-deployment.md) compares storage options and recommends immutable files plus transactional SQLite version metadata. Container build and volume/PVC instructions are in [deployment and backup](docs/architecture/deployment-and-backup.md).

## File workspace

Markdown files open in Galley live editing; use **Исходный Markdown** to switch
to source text. Drafts stay in the current browser tab until explicitly saved.
Choose a shared Galley theme in **Настройки вида**; light, dark and system modes
remain independent of the theme family.

Use **Импорт ZIP** in the file sidebar to inspect an archive before adding it.
Nested paths and empty directories are kept; existing files are never replaced.
ZIP limits are 64 MiB uploaded, 256 MiB expanded and the configured per-file/tree
limits. Other archive types can be stored as ordinary files. In **История проекта**,
select a saved version and **Сравнить с текущей** to inspect project changes and
text differences. [Implementation and verification](docs/audit/2026-10-08-files-editor-themes.md).

## Stack

- Backend: Go, `chi`, SQLite, Goose migrations, sqlc-generated queries.
- Frontend: React, TypeScript, Vite, React Router, Redux Toolkit, Tailwind CSS v4, shadcn/ui-style primitives.
- Package/runtime tooling: `mise`, Bun, Go.

## Requirements

The repository includes a `mise.toml` with the expected tools:

```bash
mise install
```

`go.mod` currently declares Go `1.26.4`; use the toolchain selected by `mise.toml` for this checkout. Node `26` and Bun are used for the frontend.

The backend test/build commands use SQLite FTS5, so keep the `sqlite_fts5` build tag when running Go tests.

## Development

Install frontend dependencies:

```bash
cd frontend
bun install --frozen-lockfile
```

Run the frontend dev server:

```bash
cd frontend
bun run dev
```

Run the backend API/server:

```bash
OPEN_EDDA_JWT_SECRET="replace-with-at-least-32-bytes-secret" \
OPEN_EDDA_API_KEY_ENCRYPTION_SECRET="replace-with-another-32-bytes-secret" \
OPEN_EDDA_BOOTSTRAP_EMAIL="author@example.com" \
OPEN_EDDA_BOOTSTRAP_PASSWORD="change-this-password" \
go run -tags sqlite_fts5 .
```

By default the backend listens on `:8080`, uses `edda.db`, runs migrations from `migrations`, and serves the built frontend from `frontend/dist`. Open Edda is currently single-user: create the initial login by setting `OPEN_EDDA_BOOTSTRAP_EMAIL` and `OPEN_EDDA_BOOTSTRAP_PASSWORD` on first server start. Existing users are not overwritten by later bootstrap env values.

For a production-like local run, build the frontend first:

```bash
cd frontend
bun run build
cd ..
OPEN_EDDA_JWT_SECRET="replace-with-at-least-32-bytes-secret" \
OPEN_EDDA_API_KEY_ENCRYPTION_SECRET="replace-with-another-32-bytes-secret" \
go run -tags sqlite_fts5 .
```

## CLI: upload a local project

```sh
go build -tags sqlite_fts5 -o edda ./cmd/edda
./edda login
./edda send ./my-book
```

Login asks for the server, email and a hidden password. First send asks whether to create a project or choose an existing one, connects the folder and uploads it. Inside that folder or any subdirectory, use `edda send` to upload edits and `edda take` to receive server changes. They reuse unchanged content by SHA-256 and show stages, counts, bytes and elapsed time; `--quiet` hides progress. Changed files still transfer as whole objects. Commands discover the `.edda` binding automatically. `edda import` also saves this binding after a successful upload. `edda get` downloads an existing project into a new folder through a selection dialogue.

Explicit arguments skip prompts: `edda send ./my-book --title "My book"` creates a project on first send. `edda COMMAND --help` explains every command. Output is readable by default; `projects`, `create`, `history` and `import` offer `--json` for scripts. See [CLI synchronization](docs/architecture/cli-sync.md) for details.

## Configuration

Environment variables:

| Variable | Default | Purpose |
| --- | --- | --- |
| `OPEN_EDDA_ADDR` | `:8080` | HTTP listen address |
| `OPEN_EDDA_DB_PATH` | `edda.db` | SQLite database path |
| `OPEN_EDDA_DATA_DIR` | Parent directory of the configured database | Immutable file objects under `objects/`; does not relocate the database |
| `OPEN_EDDA_MIGRATIONS_PATH` | `migrations` | Goose migrations directory |
| `OPEN_EDDA_STATIC_PATH` | `frontend/dist` | Built frontend directory |
| `OPEN_EDDA_JWT_SECRET` | required | JWT signing secret, at least 32 bytes |
| `OPEN_EDDA_API_KEY_ENCRYPTION_SECRET` | required | Dedicated provider API-key encryption secret, at least 32 bytes |
| `OPEN_EDDA_BOOTSTRAP_EMAIL` | optional | Initial single-user email; requires `OPEN_EDDA_BOOTSTRAP_PASSWORD` |
| `OPEN_EDDA_BOOTSTRAP_PASSWORD` | optional | Initial single-user password; requires `OPEN_EDDA_BOOTSTRAP_EMAIL` |

Legacy `WRITER_*` equivalents remain accepted for earlier settings; `OPEN_EDDA_DATA_DIR` is new. Persist **both** the database and the objects directory: for a shared data volume set `OPEN_EDDA_DB_PATH=/data/edda.db` and `OPEN_EDDA_DATA_DIR=/data`. Docker Compose and Kubernetes manifests are provided in [deployment and backup](docs/architecture/deployment-and-backup.md).

## Verification

The initial 2026-10-06 build blockers were fixed in the first baseline slice: the restore handler uses the bounded JSON decoder correctly, and frontend dependencies install from the existing lockfile. The Go module is `github.com/InkyQuill/open-edda`. See the [baseline verification record](docs/audit/2026-10-06-baseline-repair.md) for checks and remaining limitations.

Backend:

```bash
go test -tags sqlite_fts5 ./...
```

Frontend:

```bash
cd frontend
bun run test
bun run build
bun run test:smoke
bun run test:files
```

The same Go test command is available through mise:

```bash
mise run test
```

## Repository Layout

```txt
agent/       Agent sessions, tools, prompts, prompt records, activity
app/         HTTP router, SPA serving, API composition
auth/        JWT auth service and middleware
cmd/edda/    CLI connection, import, checkout and transactional send
fileproject/ File inventory, IDs, local saves, snapshots and conflicts
frontend/    React workspace UI
markdownio/  Legacy database-content Markdown import/export
migrations/  SQLite schema migrations
project/     Story projects, content, revisions, notes
queries/     SQL query sources for sqlc
skill/       Skill import, selection, script runtime, HTTP API
store/       Database opening and generated query models
docs/        Roadmap, specs, implementation plans, skill library docs
```

## Current Product Shape

Implemented foundations include authenticated single-author project/content management, database item revisions, Legacy database-content Markdown import/export, OpenAI-compatible assistant workflows, skill import/routing, an approved script runtime, and a routed React writing workspace. Provider/model/skill administration has a settings surface.

The `fileproject` package now supplies a generic [folder inventory and import](docs/architecture/folder-import.md); earlier prototype helpers still supply fixed-layout scanning, stable IDs, indexing, local drafts/saves, snapshots, version checks and conflict helpers. The legacy web editor and agent still use database content/revision APIs. UI labels that say “Checkpoint” currently refer to item revisions, not integrated project snapshots.

The internal `project.VersionStore` now implements generic file/directory manifests, immutable bytes, transactional versions, idempotent publication and restore. Its [contracts and verification](docs/architecture/project-versions.md) are documented; new file projects use it through the web editor and authenticated API. CLI get/send/take use the same API, with explicit file conflict resolution and journaled recovery; see [CLI synchronization](docs/architecture/cli-sync.md).

The target is layout-independent project storage with recoverable transactional versions and real local/web synchronization. [ADR 0014](docs/adr/0014-portable-projects-and-transactional-sync.md) supersedes the old mandatory Edda-layout policy. Supporting an unfamiliar file in storage does not require an editor or agent adapter for that format.

See [agent tools](docs/agent-tools.md) for the existing tool catalog. Skill availability does not imply equivalence with current local CWS workflows. Multi-user collaboration remains deferred until single-author mobility and recovery are reliable.
