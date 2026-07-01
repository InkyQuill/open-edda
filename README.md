# Open Edda

Open Edda is a self-hosted writing workspace for long-form fiction projects. Its target model is file-first: story text, storyline/planning material, characters, worldbuilding, drafts, project guidance, and project-local skills live as ordinary Markdown files in one defined Edda project layout, while SQLite backs indexes, search, assistant context, prompt records, activity, settings, and other operational data for the browser UI.

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
| `OPEN_EDDA_BOOTSTRAP_EMAIL` | optional | Initial single-user email; requires `OPEN_EDDA_BOOTSTRAP_PASSWORD` |
| `OPEN_EDDA_BOOTSTRAP_PASSWORD` | optional | Initial single-user password; requires `OPEN_EDDA_BOOTSTRAP_EMAIL` |

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
- Story projects with chapters, story bible entries, writing briefs, project notes, attached notes, entry sections, relations, and current legacy per-item revisions.
- Edda-layout Markdown import/export, moving toward file-first project folders and `.edda/` metadata. Elysium is treated as an older conversion source, not the target structure.
- OpenAI-compatible provider/model configuration.
- Agent sessions for chat, continuation, rewrite, and read/check flows.
- Structured agent tools for project map, content search/read, legacy revisions, writes, selected skills, and skill scripts.
- Skill import, browsing, session selection, and built-in fiction-writing skills.
- Audited skill script runtime with admin approval, safe input envelopes, reviewable outputs, and run history.
- Routed writing workspace with desktop drawers and mobile sheets for editor, assistant, and review surfaces. System settings for providers, models, skills, and script runtime administration are planned as the next Milestone 4 IA correction.

See [docs/roadmap.md](docs/roadmap.md) for milestone status and follow-up phases.
See [docs/agent-tools.md](docs/agent-tools.md) for the current agent tool catalog, invocation flow, prompt guidance, and known tool gaps.

## Notes

This is designed for self-hosted/local use first. Multi-user collaboration and broader deployment hardening are deferred until the single-author file-first workflow, lightweight checkpoints, and local/server mobility are stable.
