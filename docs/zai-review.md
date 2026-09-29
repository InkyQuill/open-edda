# Code Review: Open Edda (zai/glm-5.1)

**Date:** 2026-06-20
**Scope:** Core Go backend (`app/`, `auth/`, `project/`, `agent/`, `skill/`, `store/`, `markdownio/`), SQL schema/queries, and React frontend (`frontend/src/`).

---

## Summary

Open Edda is a well-structured Go+React application with clear package boundaries, consistent error handling, and thoughtful security decisions for a v1 self-hosted product. The codebase is readable and follows idiomatic Go patterns. That said, there are several security gaps (one critical), significant code duplication, and a few design choices that will become painful as the codebase grows.

**Overall impression:** Solid foundation, needs targeted hardening and deduplication before the next milestone.

---

## Security

### CRITICAL: API keys stored in plaintext

`agent/service.go:308-316` — `CreateProviderConfig` stores the user's LLM provider API key directly into `api_key_encrypted` with no encryption. The column name is misleading; the comment acknowledges this ("Encryption is handled in the later auth/security milestone"). Until encryption lands, any SQLite file read (backup, filesystem access, debug tool) exposes all provider API keys in cleartext.

**Recommendation:** Implement envelope encryption before or immediately after shipping this endpoint. At minimum, consider base64 + AES-GCM with a server-managed key derived from the JWT secret. Rotate the column name to `api_key_encrypted` only after encryption is real.

### HIGH: Arbitrary shell command execution via skill scripts

`skill/runtime/runner.go:54-55` — `Run()` passes the admin-approved `RuntimeCommand` directly to `sh -c`. While the approval gate (`skill/script_runtime.go:159`) requires admin enablement, the command string is stored verbatim and never validated or sandboxed beyond the working directory isolation.

Mitigating factors: the runner strips environment variables (`runner.go:62-65`), sets a temp workdir, and uses process group kill on timeout. The approval flow also blocks scripts flagged with `destructive_operations` or `allow_network` (`script_runtime.go:82-84, 172-173`).

**Recommendation:**
- Add an allowlist of permitted interpreter prefixes (e.g., `python3 `, `node `, `bash `) during approval, rejecting raw shell metacharacter commands.
- Consider using `exec.Command(commandName, args...)` directly instead of `sh -c` when the runtime is known, to avoid shell injection within the command string.

### MEDIUM: No request body size limits on JSON endpoints

`project/http.go:397-406`, `agent/http.go:554-563` — `decodeJSON` reads the entire request body into memory with no size cap. Only the Elysium import path uses `MaxBytesReader`. A single oversized request to any JSON endpoint can exhaust server memory.

**Recommendation:** Wrap `r.Body` with `http.MaxBytesReader(w, r.Body, maxJSONBytes)` (e.g., 1 MB) before decoding in a shared middleware or in `decodeJSON`.

### MEDIUM: No rate limiting on login

`auth/http.go:16-17` — `/api/auth/login` has no rate limiting. An attacker can attempt unlimited password guesses.

**Recommendation:** Add per-IP or per-email rate limiting, even a simple token-bucket middleware, before exposing the service on a network.

### LOW: Exported `Secret()` method on auth.Service

`auth/service.go:38-40` — `Secret()` returns the raw JWT signing key as a string. It isn't called anywhere in the current codebase, but being exported makes it available to any package that imports `auth`.

**Recommendation:** Unexport or remove this method. If it's needed for testing, expose it through a test-only helper.

### LOW: Inconsistent ID generation security

`auth/helpers.go:9-14` uses `crypto/rand.Read` for IDs (strong), while `project/service.go:966-968`, `agent/service.go:1992-1994`, and `skill/service.go:531-533` use `time.Now().UnixNano() + atomic counter` (predictable, not cryptographically random).

For a single-author self-hosted app this is acceptable, but the inconsistency suggests different authors wrote these at different times. If IDs ever need to be unguessable (e.g., to prevent enumeration of other projects' resources), the timestamp+counter scheme is insufficient.

**Recommendation:** Consolidate to a single `newID` implementation using `crypto/rand`, or explicitly document the tradeoff and when to use which.

### GOOD: Prompt record redaction

`agent/service.go:1885-1931` — `scrubJSON` recursively removes `api_key`, `apiKey`, and `authorization` keys before storing prompt records. This is a good practice.

**Caveat:** Only known key names are scrubbed. If a provider returns the key under an unexpected field name, it leaks into the database. Consider a deny-by-default approach: only include whitelisted top-level keys in prompt records.

---

## Code Quality

### Duplicated utility functions across packages

The following functions are copy-pasted across 2-3 packages with identical or slightly different implementations:

| Function | Locations |
|---|---|
| `defaultJSON` | `project/service.go`, `agent/service.go`, `skill/service.go` |
| `emptyDefault` | `project/service.go`, `agent/service.go`, `skill/service.go` |
| `boolInt` | `agent/service.go`, `skill/service.go` |
| `newID` | `auth/helpers.go` (crypto/rand), `project/service.go` (ts+counter), `agent/service.go` (ts+counter), `skill/service.go` (ts+counter) |
| `nowString` | `auth/helpers.go` (RFC3339), `project/service.go` (RFC3339Nano), `agent/service.go` (monotonic RFC3339Nano), `skill/service.go` (RFC3339Nano) |
| `nullString` | `project/service.go`, `agent/service.go` |
| `nullInt64` | `project/service.go` (value!=0), `agent/service.go` (value, valid bool) |
| `isSQLiteConstraint` | `project/service.go`, `agent/service.go` |
| `inTx` | `project/service.go`, `skill/service.go` |
| `writeJSON` | `app/app.go`, `auth/http.go`, `project/http.go`, `agent/http.go` |
| `writeError` | `project/http.go`, `agent/http.go` |
| `authorID` | `project/http.go`, `agent/http.go` |
| `decodeJSON` | `project/http.go`, `agent/http.go` |

**Recommendation:** Create an `internal/httputil` package for HTTP helpers and an `internal/dbutil` package for database helpers. Move `newID`, `nowString`, `nullString`, `inTx`, etc. into shared packages. The `nullInt64` signature discrepancy (value-based vs. explicit-validity) should be resolved in favor of the explicit-validity version.

### `agent/service.go` is too large (2007 lines)

This single file handles: session CRUD, message CRUD, provider config CRUD, model variant CRUD, prompt profile CRUD, chat turns, continuations, rewrites, read-and-check, generation candidates, activity events, prompt records, prompt building, and dozens of helper functions.

**Recommendation:** Split into focused files or sub-packages:
- `agent/session.go` — session + message operations
- `agent/provider.go` — provider config + model variant (already partially done)
- `agent/prompt.go` — prompt building (already extracted)
- `agent/action.go` — chat turn, continuation, rewrite, read-and-check
- `agent/candidate.go` — generation candidate lifecycle
- `agent/helpers.go` — shared utilities (or move to `internal/`)

### Magic strings for candidate status values

`agent/service.go` uses string literals `"pending"`, `"applying"`, `"accepted"`, `"rejected"`, `"conflict"` throughout. The `GenerationCandidate.Status` field is just a `string`.

**Recommendation:** Define a `CandidateStatus` type with constants, similar to `ActionKind` and `ApplyMode`.

### Missing input validation

- `project/service.go:46` — `CreateProject` doesn't validate that `Title` is non-empty.
- `project/service.go:190` — `CreateContent` doesn't validate that `Title` is non-empty or that `MetadataJSON` is valid JSON.
- `agent/http.go:197-228` — `createModelVariant` doesn't validate that `Temperature` is in a reasonable range (0–2), or that `MaxOutputTokens` > 0.

**Recommendation:** Add minimum validation for required string fields and numeric ranges in service-layer methods.

### `SearchContentCandidates` query has no LIMIT

`queries/project_core.sql:120-125` — This query returns all FTS-matching rows without a limit, relying on Go-side post-filtering for metadata/tags. For a project with many content items, this could return very large result sets.

**Recommendation:** Add a generous safety limit (e.g., 500 or 1000) to the query, or at minimum document the tradeoff.

### `ProviderConfig` API key in `completionProvider`

`agent/service.go:1262` — `NewOpenAICompatibleClient(providerConfig.BaseUrl, providerConfig.ApiKeyEncrypted, modelVariant)` passes the stored "encrypted" key directly to the HTTP client. This is correct functionally but reinforces that the key is stored in plaintext.

---

## Readability

### Transaction pattern is verbose and repeated

Every transaction follows this pattern:

```go
tx, err := s.db.BeginTx(ctx, nil)
if err != nil { return fmt.Errorf("begin transaction: %w", err) }
committed := false
defer func() {
    if !committed { _ = tx.Rollback() }
}()
// ... work ...
if err := tx.Commit(); err != nil { return fmt.Errorf("commit transaction: %w", err) }
committed = true
```

This appears in `project/service.go`, `skill/service.go`, and `agent/service.go` (multiple times).

**Recommendation:** Extract the `inTx` helper (which already exists in `project/service.go` and `skill/service.go`) into `internal/dbutil` and use it everywhere, including the inline transactions in `agent/service.go`.

### JSON tag naming inconsistency

Store models use `json:"project_id"` (snake_case), while domain types use `json:"projectId"` (camelCase). The conversion happens in `*FromStore` functions. This is intentional (database convention vs. API convention) but creates a class of bugs where a developer might accidentally serialize a store model to the API.

**Recommendation:** This is acceptable if documented. Consider adding a linter rule or code review checklist item that store types should never appear in HTTP handler signatures.

### Long method chains in `RunChatTurn`

`agent/service.go:519-717` — This method is ~200 lines with a tool-call loop, nested error handling, and prompt record storage. The control flow is hard to follow.

**Recommendation:** Extract:
- The tool-call loop into `runToolCallLoop(ctx, provider, request, projectID, sessionID) (CompletionResponse, []CompletionRequest, []CompletionResponse, Usage, error)`.
- The prompt record storage into a separate method (partially done with `storePromptRecord`).

### Mixed receiver styles in HTTP handlers

`project/http.go` uses value receivers `(h httpHandler)`, while `agent/http.go` also uses value receivers but both reference `h.service`. This is consistent but worth noting that both are fine for this pattern since `httpHandler` only holds a pointer.

---

## Frontend

### Token stored in localStorage

`frontend/src/authApi.ts:28` — `localStorage.setItem("open_edda_token", data.token)` stores the JWT in localStorage, which is accessible to any JavaScript running on the same origin (XSS).

**Recommendation:** For a self-hosted single-author app this is acceptable for v1. For future versions, consider `httpOnly` cookies or `sessionStorage` with a shorter token lifetime.

### Missing error handling for token expiry

`frontend/src/api.ts:4-12` and `frontend/src/agentApi.ts:25-46` — `authFetch` / `requestJSON` attach the token but don't handle 401 responses by redirecting to login or refreshing the token. A stale token results in generic error messages.

**Recommendation:** Add a 401 interceptor that clears the token and redirects to login.

### Duplicate `authFetch` implementation

`api.ts:4-12` and `agentApi.ts:25-46` both implement token-injection fetch wrappers with slightly different error handling.

**Recommendation:** Consolidate into a single shared `fetchClient.ts` module.

---

## Database

### GOOD: Schema design

- Foreign keys with `ON DELETE CASCADE` are consistently applied.
- CHECK constraints on `kind`, `created_by`, `source` enforce valid enums at the DB level.
- FTS5 virtual table with proper triggers for content search.
- Proper indexing for common query patterns.

### GOOD: Optimistic concurrency

`UpdateContentItemBody` and `UpdateEntrySectionBody` use `current_revision = expected_revision` as an optimistic lock, returning 0 affected rows on conflict. This is well-implemented.

### Minor: `entry_sections` migration down path doesn't preserve `current_revision`

`migrations/00004_entry_section_revisions.sql:7-21` — The down migration recreates `entry_sections` without the `current_revision` column, losing revision tracking data on rollback.

**Recommendation:** Acceptable for a v1 early migration, but note that rolling back this migration is lossy.

---

## Positive Observations

- **Consistent error wrapping**: All service methods use `fmt.Errorf("operation: %w", err)`, making error tracing reliable.
- **Optimistic concurrency**: Revision-based conflict detection is well-designed and consistently applied.
- **Tool result bounding**: The `boundModelVisible` / `boundDirectRead` / `boundListResult` system for limiting what the model sees is sophisticated and well-implemented.
- **Security-conscious defaults**: Scripts disabled by default, network/filesystem access blocked, process group kill on timeout.
- **Zip safety**: The Elysium import validates paths with `filepath.IsLocal`, limits file count, limits uncompressed size, and caps individual entry reads.
- **Trailing JSON rejection**: `decodeJSON` rejects requests with extra data after the first JSON object, preventing parameter-smuggling attacks.
- **Clean test structure**: Tests use real databases with migrations, proper cleanup, and focused assertions.
