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

The helper will use AES-GCM with a per-value random nonce. Key material will be derived from the existing auth/JWT secret so this pass does not add a second required deployment secret. Stored ciphertext will be version-prefixed, for example `edda:v1:<base64 nonce+ciphertext>`, so the read path can distinguish encrypted values from legacy plaintext rows.

Create and update provider-config paths will encrypt the supplied API key before storing it. The provider construction path will decrypt before creating the OpenAI-compatible client. List/get API responses already omit the raw key, so their public shape should not change.

For compatibility, unprefixed values will be treated as legacy plaintext during reads. This avoids breaking existing local databases. A later migration can rewrite legacy plaintext rows once there is an explicit operational migration story.

Error handling: encryption/decryption failures should return wrapped internal errors. Provider config creation/update should not store a value if encryption fails. Completion-provider construction should fail before making any network request if decryption fails.

## JSON Request Body Limits

The JSON decoders in auth, project, agent, and skill endpoints currently decode from `r.Body` directly. This pass will introduce a shared JSON decoder with a `1 MiB` limit for ordinary JSON requests.

Archive and zip imports already have separate larger caps and should keep them. The shared decoder should preserve the current trailing-JSON rejection behavior.

Oversized JSON requests should return a bad request or payload-too-large response consistently through existing handler error paths. The implementation should prefer a small internal HTTP helper over repeated ad hoc handling.

## Login Rate Limiting

The login route currently allows unlimited attempts. This pass will add process-local rate limiting around `/api/auth/login`.

The limiter will be intentionally simple for self-hosted v1:

- Key attempts by remote IP address, using `X-Forwarded-For` only if the app already has a trusted-proxy abstraction. Otherwise use `r.RemoteAddr`.
- Use a small in-memory token bucket or sliding window.
- Return HTTP `429` when the limit is exceeded.
- Keep the state bounded with opportunistic cleanup of old entries.

This is not a distributed or persistent brute-force defense. It is a practical default until deployment topology and proxy trust are designed.

## Auth Secret Accessor

`auth.Service.Secret()` exports the raw JWT signing secret and is not needed by production callers. This pass will remove it. Tests that need tokens should use the known test secret directly.

## Search Candidate Limit

`SearchContentCandidates` currently returns all matching FTS rows. This pass will add a generous SQL-side limit, preferably parameterized if call sites can pass one cleanly, otherwise a fixed cap such as `1000`.

The goal is to prevent accidental large result sets before Go-side candidate filtering. Existing ranked ordering should remain unchanged.

## HTTP JSON Helpers

`writeJSON` is duplicated across packages and writes the status before encoding. This can commit a success status before a serialization failure. This pass will add a small shared helper that marshals before writing headers.

The helper should live in an internal package so app/auth/project/agent/skill can all use it without introducing an import cycle. Handler behavior should remain the same for normal serializable responses.

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

