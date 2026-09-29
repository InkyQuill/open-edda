# Backlog

This file records review findings that were deliberately deferred from the 2026-06-20 review-hardening pass. Each item includes evidence from the current codebase and the reason it is not being handled in that pass.

## Skill Script Runtime Boundaries And Audit Policy

Sources:

- `docs/google-review.md`: "Critical Risk: Lack of OS-Level Sandboxing in Skill Script Runtime"
- `docs/zai-review.md`: "HIGH: Arbitrary shell command execution via skill scripts"
- `docs/skill-script-brainstorm.md`: dedicated brainstorm for script execution, prompt conversion, and isolation strategy

Evidence:

- `skill/runtime/runner.go` runs approved commands through `sh -c` on Unix or `cmd /C` on Windows.
- `skill/service.go` creates default script audit rows with `DestructiveOperations: 0`, `NetworkAccess: 0`, `FilesystemAccess: "temp_workspace"`, and the note "Imported script is disabled until explicitly approved."
- `skill/script_runtime.go` blocks unsafe approval flags at runtime, but there is no OS-level sandbox or interpreter allowlist yet.

Why deferred:

- Script execution boundaries are a separate product and security design. The project already has `docs/skill-script-brainstorm.md` for deciding which scripts should run as code, which should become prompt rubrics, and what isolation model is acceptable.
- Mixing that work into provider-key and request-hardening changes would make both reviews harder.

Later work:

- Decide the v1 script execution model from `docs/skill-script-brainstorm.md`.
- Classify scripts into removed, prompt-converted, code-executed, and deferred pipeline categories.
- Replace optimistic default script audit values with explicit "manual inspection required" or a real analyzer.
- Consider command parsing, interpreter allowlists, Deno permissions, OS sandboxing, or a no-code-only script policy.
- Add tests proving unsafe stored approvals cannot execute.

## ID Generation Consolidation

Sources:

- `docs/google-review.md`: "Medium Risk: Predictable/Sequential ID Generation"
- `docs/zai-review.md`: "LOW: Inconsistent ID generation security"

Evidence:

- `auth/helpers.go` uses `crypto/rand` for IDs.
- `project/service.go`, `agent/service.go`, and `skill/service.go` generate IDs with timestamp plus an atomic counter.

Why deferred:

- The current app is single-author and all project-scoped resources still require authenticated ownership checks.
- Changing ID format is a cross-package data-shape change and should be done deliberately with tests across auth, project, agent, skill, and generated fixtures.

Later work:

- Introduce a shared internal ID helper using crypto-random identifiers.
- Update project, agent, and skill packages to use it.
- Keep prefixes if they are useful for debugging.
- Add tests that generated IDs are unique, prefixed as expected, and not timestamp-derived.

## Large Agent Service Decomposition

Source:

- `docs/zai-review.md`: "`agent/service.go` is too large"

Evidence:

- `agent/service.go` handles sessions, messages, provider config CRUD, model variants, prompt profiles, chat turns, continuations, rewrites, read-and-check, candidates, activity events, prompt records, provider construction, and helpers.

Why deferred:

- File decomposition does not directly fix a current security or correctness issue.
- Refactoring this file while changing provider-key encryption would increase review surface and conflict risk.

Later work:

- Split provider config and model-variant logic into provider-focused files.
- Split action execution, candidate lifecycle, prompt-record storage, and session/message CRUD into focused files.
- Keep public service methods stable while moving internals.

## Broader HTTP And DB Utility Cleanup

Sources:

- `docs/zai-review.md`: duplicated utility functions across packages
- `docs/google-review.md`: duplicated SQLite error checks

Evidence:

- `writeJSON`, `decodeJSON`, `writeError`, and `authorID` patterns repeat across HTTP packages.
- `isSQLiteUniqueConstraint` and `isSQLiteConstraint` are defined in separate packages.
- Transaction helpers and JSON/default helpers are duplicated in service packages.

Why deferred:

- The review-hardening pass should centralize only the JSON response and decode behavior needed for body limits and safe response writing.
- A broader utility extraction touches many files without changing product behavior.

Later work:

- Create focused internal packages for HTTP and DB helpers.
- Move SQLite constraint matching to one place.
- Move transaction helpers where they reduce real duplication.
- Avoid changing domain error mapping unless tests cover the behavior.

## Project Access Middleware Route Param Cleanup

Source:

- `docs/google-review.md`: "Middleware Access Check Robustness"

Evidence:

- `app/app.go` uses `projectIDFromPath(r.URL.Path)` to infer project ownership checks for paths under `/api/projects/{projectID}`.

Why deferred:

- The current logic is small and route-specific.
- Switching to chi route params may require changing route nesting so the middleware runs after route matching, which is broader than this pass.

Later work:

- Revisit protected project route structure.
- Prefer `chi.URLParam(r, "projectID")` where the route context reliably contains the value.
- Keep behavior for non-project protected routes unchanged.

## Frontend Token Storage And Expiry Handling

Source:

- `docs/zai-review.md`: frontend findings for localStorage token storage, 401 handling, and duplicate fetch wrappers

Evidence:

- `frontend/src/authApi.ts` stores the bearer token in `localStorage`.
- Frontend request helpers attach tokens but do not consistently clear state or redirect on `401`.
- Token-injection wrappers are split across frontend API modules.

Why deferred:

- This pass is backend-focused and will add login throttling and provider-key storage hardening.
- Moving to httpOnly cookies or a richer auth lifecycle affects frontend routing, deployment assumptions, and CSRF/session design.

Later work:

- Centralize frontend fetch behavior.
- Clear auth state and route to login on `401`.
- Decide whether to stay with bearer tokens, use sessionStorage, or move to httpOnly cookies.
- If cookies are used, design CSRF and same-site behavior explicitly.

## YAML Frontmatter Parser Replacement

Source:

- `docs/google-review.md`: "Fragile Custom YAML Frontmatter Parser"

Evidence:

- `markdownio/elysium.go` parses frontmatter with custom string splitting and line-by-line matching.

Why deferred:

- The parser is isolated to Elysium import/export behavior.
- Replacing it adds a dependency and requires compatibility tests for existing Elysium documents.

Later work:

- Evaluate `gopkg.in/yaml.v3`.
- Add fixtures for nested maps, lists, multiline strings, and values containing colons.
- Preserve current accepted frontmatter behavior unless intentionally changing the format.

## Prompt Record Redaction Hardening

Source:

- `docs/zai-review.md`: prompt record redaction caveat

Evidence:

- `agent/service.go` scrubs known keys such as `api_key`, `apiKey`, and `authorization` before storing prompt records.

Why deferred:

- Current redaction is already a positive control and this pass focuses on provider API key storage.
- A deny-by-default prompt-record schema requires deciding which provider request/response fields are valuable for debugging.

Later work:

- Design a prompt-record whitelist.
- Preserve enough data for debugging model behavior.
- Add tests for unexpected secret-like provider fields.

## Model Variant And Project Input Validation

Source:

- `docs/zai-review.md`: missing input validation

Evidence:

- Some service methods trust caller-provided fields such as titles, model temperature, max output tokens, and metadata JSON more than they should.

Why deferred:

- This is worthwhile correctness work but not tightly connected to provider-key encryption, body limits, or login throttling.
- Validation changes can alter API behavior and should be grouped with user-facing error handling tests.

Later work:

- Validate required titles at the service layer.
- Validate model temperature and token bounds.
- Validate metadata JSON where callers can submit it.
- Add HTTP tests for the resulting `400` behavior.

## Candidate Status Type Constants

Source:

- `docs/zai-review.md`: "Magic strings for candidate status values"

Evidence:

- Candidate statuses such as `pending`, `applying`, `accepted`, `rejected`, and `conflict` appear as string literals.

Why deferred:

- This is a maintainability improvement with low immediate risk.

Later work:

- Introduce a `CandidateStatus` type and constants.
- Replace string literals in service logic and tests.
- Keep database CHECK constraints aligned with constants.

