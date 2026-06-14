# Open Edda

Open Edda is a self-hosted writing workspace for long-form fiction projects. It keeps story text, worldbuilding, writing briefs, project notes, revisions, agent activity, prompt records, skills, and script-run artifacts in a local SQLite-backed service with a browser UI.

The project is currently pre-v1. Milestones 1-3.6 are implemented, and Milestone 4 daily-writing polish is in progress.

## Stack

- Backend: Go, `chi`, SQLite, Goose migrations, sqlc-generated queries.
- Frontend: React, TypeScript, Vite, React Router, Redux Toolkit, Tailwind CSS v4, shadcn/ui-style primitives.
- Package/runtime tooling: `mise`, Bun, Go.

## Requirements

The repository includes a `mise.toml` with the expected tools:

```bash
mise install
```

Required versions are:

- Go `1.26.4`
- Node `26`
- Bun

The backend test/build commands use SQLite FTS5, so keep the `sqlite_fts5` build tag when running Go tests.

## Development

Install frontend dependencies:

```bash
cd frontend
bun install
```

Run the frontend dev server:

```bash
cd frontend
bun run dev
```

Run the backend API/server:

```bash
OPEN_EDDA_JWT_SECRET="replace-with-at-least-32-bytes-secret" \
go run -tags sqlite_fts5 .
```

By default the backend listens on `:8080`, uses `edda.db`, runs migrations from `migrations`, and serves the built frontend from `frontend/dist`.

For a production-like local run, build the frontend first:

```bash
cd frontend
bun run build
cd ..
OPEN_EDDA_JWT_SECRET="replace-with-at-least-32-bytes-secret" \
go run -tags sqlite_fts5 .
```

## Configuration

Environment variables:

| Variable | Default | Purpose |
| --- | --- | --- |
| `OPEN_EDDA_ADDR` | `:8080` | HTTP listen address |
| `OPEN_EDDA_DB_PATH` | `edda.db` | SQLite database path |
| `OPEN_EDDA_MIGRATIONS_PATH` | `migrations` | Goose migrations directory |
| `OPEN_EDDA_STATIC_PATH` | `frontend/dist` | Built frontend directory |
| `OPEN_EDDA_JWT_SECRET` | required | JWT signing secret, at least 32 bytes |

Legacy `WRITER_*` equivalents are still accepted for these settings.

## Verification

Backend:

```bash
go test -tags sqlite_fts5 ./...
```

Frontend:

```bash
cd frontend
bun run test
bun run build
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
frontend/    React workspace UI
markdownio/  Elysium Markdown import/export
migrations/  SQLite schema migrations
project/     Story projects, content, revisions, notes
queries/     SQL query sources for sqlc
skill/       Skill import, selection, script runtime, HTTP API
store/       Database opening and generated query models
docs/        Roadmap, specs, implementation plans, skill library docs
```

## Current Product Shape

Open Edda supports:

- Authenticated local author workflow.
- Story projects with chapters, story bible entries, writing briefs, project notes, attached notes, entry sections, relations, and revisions.
- Elysium-style Markdown import/export.
- OpenAI-compatible provider/model configuration.
- Agent sessions for chat, continuation, rewrite, and read/check flows.
- Structured agent tools for project map, content search/read, revisions, writes, selected skills, and skill scripts.
- Skill import, browsing, session selection, and built-in fiction-writing skills.
- Audited skill script runtime with admin approval, safe input envelopes, reviewable outputs, and run history.
- Routed writing workspace with desktop drawers and mobile sheets for editor, assistant, review, skills, model settings, and script runtime surfaces.

See [docs/roadmap.md](docs/roadmap.md) for milestone status and follow-up phases.

## Notes

This is designed for self-hosted/local use first. Multi-user collaboration, sync tooling, and broader deployment hardening are deferred until the single-author workflow is stable.
