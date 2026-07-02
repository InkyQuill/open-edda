# Review Hardening Design

Date: 2026-06-20

## Scope

This pass addresses confirmed findings from `docs/zai-review.md` and `docs/google-review.md` that are not part of the dedicated skill-script runtime pass.

Included:

- Encrypt provider API keys stored in `provider_configs.api_key_encrypted`.
- Add JSON request body limits to ordinary JSON endpoints.
- Add simple login rate limiting.
- Remove the exported `auth.Service.Secret()` accessor.
- Add a SQL-side safety limit for `SearchContentCandidates`.
- Centralize JSON response helpers enough to fix duplicated status-before-encode behavior.
- Record deferred review findings in `data/backlog.md`.

Excluded:

- Script command boundaries, runtime sandboxing, and script audit policy. Those belong to the later pass driven by `docs/skill-script-brainstorm.md`.
- Broad package decomposition, large utility refactors, and frontend authentication redesign.

## Provider API Key Encryption

Provider keys are currently written directly to `api_key_encrypted` in `agent/service.go`. The field name claims encryption, but the value is plaintext. This pass will introduce a small encryption helper used by the agent service.

The helper will use AES-GCM (AES-256) with a per-value random 12-byte nonce (using `crypto/rand`). Stored ciphertext will be version-prefixed as `edda:v1:<base64 nonce+ciphertext>` using standard Base64 encoding (`base64.StdEncoding`). 

To avoid constructor signature churn in over 120+ unit tests across the codebase, `agent.Service` will not accept the secret in `NewService`. Instead, it will use a setter pattern:

```go
func (s *Service) SetEncryptionSecret(jwtSecret string)
```

Inside this setter, the 32-byte key material will be derived from the existing auth/JWT secret using SHA-256 with a domain-separated label: `SHA-256("open-edda-api-key-encryption" + jwtSecret)`. If the setter is not called or receives an empty string (e.g. in unit tests), the encryption helper will fall back to a static dev/test key (`"test-encryption-key-for-edda-dev-only-32bytes"`) so tests can run without failure.

Create and update provider-config paths will encrypt the supplied API key before storing it. The provider construction path will decrypt before creating the OpenAI-compatible client. List/get API responses already omit the raw key, so their public shape should not change.

For compatibility, unprefixed values will be treated as legacy plaintext during reads. This avoids breaking existing local databases. A later migration can rewrite legacy plaintext rows once there is an explicit operational migration story.

Error handling: encryption/decryption failures should return wrapped internal errors. Provider config creation/update should not store a value if encryption fails. Completion-provider construction should fail before making any network request if decryption fails.

## JSON Request Body Limits

The JSON decoders in auth, project, agent, and skill endpoints currently decode from `r.Body` directly. This pass will introduce a centralized JSON decoder `httputil.DecodeJSON(w http.ResponseWriter, r *http.Request, value any, limit int64) error` located in a new `internal/httputil` utility package.

The decoder will wrap `r.Body` with `http.MaxBytesReader` to limit size, with a default `1 MiB` limit for ordinary JSON requests. It will also preserve the current trailing-JSON rejection behavior by attempting to decode into a dummy struct and ensuring `io.EOF` is returned.

Oversized requests triggering `http.MaxBytesError` will respond with `http.StatusRequestEntityTooLarge` (413). Other parsing/trailing errors will consistently return `http.StatusBadRequest` (400) through existing handler error paths. The auth package (`/api/auth/login`) will also be updated to use this helper. Archive and zip imports already have separate larger caps and should retain their distinct implementations.

## Login Rate Limiting

The login route currently allows unlimited attempts. This pass will add process-local rate limiting around `/api/auth/login`.

The limiter will be intentionally simple for self-hosted v1:

- Key attempts by remote IP address. To prevent connection-specific bypasses due to randomized ports in `r.RemoteAddr`, the IP will be extracted using `net.SplitHostPort`. Use `X-Forwarded-For` only if a trusted proxy configuration is implemented in the future.
- Use a small in-memory token bucket implementation backed by a mutex-guarded map (e.g. limit of 5 burst tokens, refilling at 1 token every 2 seconds).
- Return HTTP `429 Too Many Requests` when the limit is exceeded.
- Keep the state bounded and prevent memory leaks via active/opportunistic pruning (e.g., evicting stale/fully-replenished entries once the map exceeds a threshold size like `1000` entries).

This is not a distributed or persistent brute-force defense. It is a practical default until deployment topology and proxy trust are designed.

## Auth Secret Accessor

`auth.Service.Secret()` exports the raw JWT signing secret and is not needed by production callers. This pass will remove it. Tests that need tokens should use the known test secret directly.

## Search Candidate Limit

`SearchContentCandidates` currently returns all matching FTS rows. This pass will add a generous fixed SQL-side limit of `1000` in the sqlc query file `queries/project_core.sql` and regenerate the store queries.

The goal is to prevent accidental large result sets before Go-side candidate filtering. Existing ranked ordering should remain unchanged.

## HTTP JSON Helpers

`writeJSON` is duplicated across packages and writes the status before encoding. This can commit a success status before a serialization failure. This pass will add a centralized helper package `internal/httputil` containing `WriteJSON(w http.ResponseWriter, status int, value any)` that marshals the JSON body and checks for errors *before* writing headers.

The helper package will be imported by app/auth/project/agent/skill without introducing any cyclic import dependencies. Handler behavior will remain the same for normal serializable responses, but serialization failures will now correctly trigger a `500 Internal Server Error`.

This is not a full HTTP abstraction rewrite. Existing package-specific `writeError` functions can remain where they encode domain-specific errors.

## Backlog

Deferred items from the review docs will be recorded in `data/backlog.md` with evidence, reason for deferral, and later acceptance notes. This makes it explicit that omitted review findings were triaged, not lost.

## Testing

Add focused tests where practical:

- Provider key encryption round trip, including legacy plaintext compatibility.
- Provider completion setup decrypts before use.
- Oversized JSON requests are rejected.
- Login rate limiter returns `429` after repeated attempts.
- `SearchContentCandidates` is capped.
- Shared JSON helper writes successful responses normally.

The current local environment may still fail full migration-backed tests with `no such module: fts5` if SQLite was built without FTS5. If that remains true, report it and run the narrow tests that can execute.

