# Review Hardening Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Harden review-confirmed backend risks by encrypting provider API keys, bounding JSON requests, throttling login attempts, removing raw secret exposure, and capping FTS candidate search.

**Architecture:** Add focused internal helpers rather than broad refactors: `internal/crypto` owns API-key sealing/opening, and `internal/httputil` owns JSON decoding/encoding plus request-size errors. Keep existing service constructors stable by adding `agent.Service.SetEncryptionSecret`, wire it from `main.go`, and preserve legacy plaintext provider-key compatibility.

**Tech Stack:** Go standard library (`crypto/aes`, `crypto/cipher`, `crypto/rand`, `crypto/sha256`, `encoding/base64`, `net/http`, `sync`, `time`), chi, sqlc, SQLite.

---

## File Structure

- Create `internal/crypto/apikey.go`: AES-256-GCM key derivation, `EncryptAPIKey`, `DecryptAPIKey`, legacy plaintext fallback, dev/test fallback key.
- Create `internal/crypto/apikey_test.go`: focused crypto tests that do not require SQLite or FTS5.
- Create `internal/httputil/json.go`: `DecodeJSON`, `WriteJSON`, `IsRequestTooLarge`, and `RemoteIP`.
- Create `internal/httputil/json_test.go`: body-limit, trailing JSON, write success/failure, and remote IP tests.
- Modify `agent/service.go`: store encryption secret on `Service`, add `SetEncryptionSecret`, encrypt create/update provider keys, decrypt for provider construction.
- Modify `agent/service_test.go` and `agent/http_test.go`: update provider-key expectations and add legacy plaintext compatibility coverage.
- Modify `auth/service.go`: remove `Secret()`.
- Modify `auth/http.go`: add process-local login limiter, use `httputil.DecodeJSON` and `httputil.WriteJSON`.
- Modify `auth/http_test.go`: add login limiter and oversized body tests.
- Modify `app/app.go`, `project/http.go`, `agent/http.go`, `skill/http.go`: use shared JSON helpers and preserve package-specific domain errors.
- Modify HTTP tests in `auth/http_test.go` and `project/http_test.go` for representative 413 behavior. Existing `agent` and `skill` HTTP tests should keep passing through the shared helper changes.
- Modify `queries/project_core.sql`: add fixed `LIMIT 1000` to `SearchContentCandidates`.
- Regenerate `store/project_core.sql.go` and `store/querier.go` with `sqlc generate`.
- Modify `main.go`: call `agentService.SetEncryptionSecret(jwtSecret)`.

---

### Task 1: Add API-Key Encryption Helper

**Files:**
- Create: `internal/crypto/apikey.go`
- Create: `internal/crypto/apikey_test.go`

- [x] **Step 1: Write failing tests for encryption format, round trip, legacy fallback, and wrong-secret failure**

Create `internal/crypto/apikey_test.go`:

```go
package crypto

import (
	"strings"
	"testing"
)

func TestEncryptAPIKeyRoundTrip(t *testing.T) {
	ciphertext, err := EncryptAPIKey("secret-one", "jwt-secret-32-bytes-minimum-value")
	if err != nil {
		t.Fatalf("EncryptAPIKey() error = %v", err)
	}
	if !strings.HasPrefix(ciphertext, encryptedAPIKeyPrefix) {
		t.Fatalf("ciphertext prefix = %q, want %q", ciphertext, encryptedAPIKeyPrefix)
	}
	if strings.Contains(ciphertext, "secret-one") {
		t.Fatalf("ciphertext includes plaintext: %q", ciphertext)
	}
	plaintext, err := DecryptAPIKey(ciphertext, "jwt-secret-32-bytes-minimum-value")
	if err != nil {
		t.Fatalf("DecryptAPIKey() error = %v", err)
	}
	if plaintext != "secret-one" {
		t.Fatalf("plaintext = %q, want secret-one", plaintext)
	}
}

func TestEncryptAPIKeyUsesRandomNonce(t *testing.T) {
	first, err := EncryptAPIKey("same-secret", "jwt-secret-32-bytes-minimum-value")
	if err != nil {
		t.Fatalf("EncryptAPIKey(first) error = %v", err)
	}
	second, err := EncryptAPIKey("same-secret", "jwt-secret-32-bytes-minimum-value")
	if err != nil {
		t.Fatalf("EncryptAPIKey(second) error = %v", err)
	}
	if first == second {
		t.Fatal("EncryptAPIKey produced identical ciphertexts for the same plaintext")
	}
}

func TestDecryptAPIKeyTreatsUnprefixedValuesAsLegacyPlaintext(t *testing.T) {
	plaintext, err := DecryptAPIKey("legacy-plaintext", "jwt-secret-32-bytes-minimum-value")
	if err != nil {
		t.Fatalf("DecryptAPIKey() error = %v", err)
	}
	if plaintext != "legacy-plaintext" {
		t.Fatalf("plaintext = %q, want legacy-plaintext", plaintext)
	}
}

func TestDecryptAPIKeyRejectsWrongSecret(t *testing.T) {
	ciphertext, err := EncryptAPIKey("secret-one", "jwt-secret-32-bytes-minimum-value")
	if err != nil {
		t.Fatalf("EncryptAPIKey() error = %v", err)
	}
	if _, err := DecryptAPIKey(ciphertext, "different-jwt-secret-32-bytes-value"); err == nil {
		t.Fatal("DecryptAPIKey() error = nil, want authentication failure")
	}
}

func TestEncryptAPIKeyFallsBackForEmptySecret(t *testing.T) {
	ciphertext, err := EncryptAPIKey("test-secret", "")
	if err != nil {
		t.Fatalf("EncryptAPIKey() error = %v", err)
	}
	plaintext, err := DecryptAPIKey(ciphertext, "")
	if err != nil {
		t.Fatalf("DecryptAPIKey() error = %v", err)
	}
	if plaintext != "test-secret" {
		t.Fatalf("plaintext = %q, want test-secret", plaintext)
	}
}
```

- [x] **Step 2: Run tests to verify they fail**

Run:

```bash
go test ./internal/crypto
```

Expected: FAIL because `EncryptAPIKey`, `DecryptAPIKey`, and `encryptedAPIKeyPrefix` do not exist.

- [x] **Step 3: Implement the helper**

Create `internal/crypto/apikey.go`:

```go
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	cryptorand "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	encryptedAPIKeyPrefix = "edda:v1:"
	encryptionLabel       = "open-edda-api-key-encryption"
	devTestSecret         = "test-encryption-key-for-edda-dev-only-32bytes"
)

var ErrInvalidCiphertext = errors.New("invalid encrypted API key")

func EncryptAPIKey(plaintext, jwtSecret string) (string, error) {
	gcm, err := apiKeyGCM(jwtSecret)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(cryptorand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate API key nonce: %w", err)
	}
	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return encryptedAPIKeyPrefix + base64.StdEncoding.EncodeToString(sealed), nil
}

func DecryptAPIKey(value, jwtSecret string) (string, error) {
	if !strings.HasPrefix(value, encryptedAPIKeyPrefix) {
		return value, nil
	}
	encoded := strings.TrimPrefix(value, encryptedAPIKeyPrefix)
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("%w: decode ciphertext: %w", ErrInvalidCiphertext, err)
	}
	gcm, err := apiKeyGCM(jwtSecret)
	if err != nil {
		return "", err
	}
	if len(raw) < gcm.NonceSize() {
		return "", ErrInvalidCiphertext
	}
	nonce := raw[:gcm.NonceSize()]
	ciphertext := raw[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("%w: authentication failed", ErrInvalidCiphertext)
	}
	return string(plaintext), nil
}

func apiKeyGCM(jwtSecret string) (cipher.AEAD, error) {
	key := deriveAPIKeyEncryptionKey(jwtSecret)
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("create API key cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create API key GCM: %w", err)
	}
	return gcm, nil
}

func deriveAPIKeyEncryptionKey(jwtSecret string) [32]byte {
	if jwtSecret == "" {
		jwtSecret = devTestSecret
	}
	return sha256.Sum256([]byte(encryptionLabel + jwtSecret))
}
```

- [x] **Step 4: Run helper tests**

Run:

```bash
go test ./internal/crypto
```

Expected: PASS.

- [x] **Step 5: Commit**

```bash
git add internal/crypto/apikey.go internal/crypto/apikey_test.go
git commit -m "feat: add provider key encryption helper"
```

---

### Task 2: Encrypt Provider Config Keys In Agent Service

**Files:**
- Modify: `agent/service.go`
- Modify: `main.go`
- Modify: `agent/service_test.go`
- Modify: `agent/http_test.go`

- [x] **Step 1: Write failing provider-key storage and compatibility tests**

In `agent/service_test.go`, replace the existing plaintext storage assertion in the provider-config test near the current `SELECT api_key_encrypted` check with this shape:

```go
	var storedKey string
	if err := db.QueryRowContext(ctx, `SELECT api_key_encrypted FROM provider_configs WHERE id = ?`, provider.ID).Scan(&storedKey); err != nil {
		t.Fatalf("get stored API key: %v", err)
	}
	if storedKey == "secret-key" {
		t.Fatal("stored API key is plaintext")
	}
	if !strings.HasPrefix(storedKey, "edda:v1:") {
		t.Fatalf("stored API key = %q, want encrypted edda:v1 prefix", storedKey)
	}
```

Add this test to `agent/service_test.go` near the provider config tests:

```go
func TestProviderConfigLegacyPlaintextKeyStillWorks(t *testing.T) {
	db := openMigratedTestDB(t)
	ctx := context.Background()
	projectService := project.NewService(db)
	service := NewService(db, projectService, nil)

	if _, err := db.ExecContext(ctx, `
		INSERT INTO provider_configs (id, author_id, name, base_url, api_key_encrypted, created_at, updated_at)
		VALUES ('provider-legacy', 'author-1', 'Legacy', 'https://api.example.invalid/v1', 'legacy-plaintext-key', '2026-06-20T00:00:00Z', '2026-06-20T00:00:00Z')
	`); err != nil {
		t.Fatalf("seed legacy provider: %v", err)
	}
	model, err := service.CreateModelVariant(ctx, CreateModelVariantInput{
		AuthorID:         "author-1",
		ProviderConfigID: "provider-legacy",
		Name:             "Legacy model",
		Model:            "legacy-model",
		Temperature:      0.7,
		MaxOutputTokens:  1000,
	})
	if err != nil {
		t.Fatalf("CreateModelVariant() error = %v", err)
	}
	providerConfig, err := service.queries.GetProviderConfig(ctx, store.GetProviderConfigParams{
		ID:       "provider-legacy",
		AuthorID: "author-1",
	})
	if err != nil {
		t.Fatalf("GetProviderConfig() error = %v", err)
	}
	_, err = service.completionProvider(providerConfig, model)
	if err != nil {
		t.Fatalf("completionProvider() error = %v", err)
	}
}
```

Ensure `agent/service_test.go` imports `git.inkyquill.net/inky/writer/store`. The file already imports `strings`; keep that import for the encrypted-prefix assertion.

- [x] **Step 2: Run tests to verify they fail**

Run:

```bash
go test ./agent -run 'Test.*Provider.*Key|TestProviderConfigLegacyPlaintextKeyStillWorks' -count=1
```

Expected: FAIL because provider keys are still stored as plaintext.

- [x] **Step 3: Add encryption secret state and setter**

Modify the `Service` struct in `agent/service.go`:

```go
type Service struct {
	db               *sql.DB
	queries          *store.Queries
	projectService   *project.Service
	provider         Provider
	skillService     SkillProvider
	encryptionSecret string
}
```

Add this method near `SetSkillService`:

```go
func (s *Service) SetEncryptionSecret(jwtSecret string) {
	s.encryptionSecret = jwtSecret
}
```

- [x] **Step 4: Encrypt create/update provider keys**

Import the helper with this alias:

```go
edcrypto "git.inkyquill.net/inky/writer/internal/crypto"
```

In `CreateProviderConfig`, before the `CreateProviderConfig` query:

```go
	encryptedAPIKey, err := edcrypto.EncryptAPIKey(input.APIKey, s.encryptionSecret)
	if err != nil {
		return ProviderConfig{}, fmt.Errorf("encrypt provider API key: %w", err)
	}
```

Use `encryptedAPIKey` in the query:

```go
		ApiKeyEncrypted: encryptedAPIKey,
```

In `UpdateProviderConfig`, before `UpdateProviderConfig`:

```go
	encryptedAPIKey, err := edcrypto.EncryptAPIKey(input.APIKey, s.encryptionSecret)
	if err != nil {
		return ProviderConfig{}, fmt.Errorf("encrypt provider API key: %w", err)
	}
```

Use `encryptedAPIKey` in the update params.

- [x] **Step 5: Decrypt before provider construction**

Change `completionProvider` from returning `Provider` to returning `(Provider, error)`:

```go
func (s *Service) completionProvider(providerConfig store.ProviderConfig, model ModelVariant) (Provider, error) {
	if s.provider != nil {
		return s.provider, nil
	}
	apiKey, err := edcrypto.DecryptAPIKey(providerConfig.ApiKeyEncrypted, s.encryptionSecret)
	if err != nil {
		return nil, fmt.Errorf("decrypt provider API key: %w", err)
	}
	return NewOpenAICompatibleClient(providerConfig.BaseUrl, apiKey, model), nil
}
```

Update the call site in `RunChatTurn` from:

```go
	provider := s.completionProvider(providerConfig, modelVariantFromStore(model))
```

to:

```go
	provider, err := s.completionProvider(providerConfig, modelVariantFromStore(model))
	if err != nil {
		return ChatTurnResult{}, err
	}
```

Update the call site in `runQuickActionCompletion` from:

```go
	provider := s.completionProvider(providerConfig, modelVariant)
```

to:

```go
	provider, err := s.completionProvider(providerConfig, modelVariant)
	if err != nil {
		return "", Session{}, err
	}
```

- [x] **Step 6: Wire production secret from main**

Modify `main.go` after `agentService := agent.NewService(...)`:

```go
	agentService.SetEncryptionSecret(jwtSecret)
```

- [x] **Step 7: Run focused tests**

Run:

```bash
go test ./internal/crypto ./agent -run 'Test.*Provider|TestProviderConfigLegacyPlaintextKeyStillWorks' -count=1
```

Expected: PASS unless the package-level migration tests hit the local FTS5 build issue. If FTS5 blocks this command, record the exact `no such module: fts5` failure and continue after running `go test ./internal/crypto`.

- [x] **Step 8: Commit**

```bash
git add agent/service.go agent/service_test.go agent/http_test.go main.go
git commit -m "feat: encrypt provider API keys at rest"
```

---

 Add Shared HTTP JSON Helpers

**Files:**
- Create: `internal/httputil/json.go`
- Create: `internal/httputil/json_test.go`

- [x] **Step 1: Write failing helper tests**

Create `internal/httputil/json_test.go`:

```go
package httputil

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeJSONRejectsOversizedBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"value":"`+strings.Repeat("x", 20)+`"}`))
	rec := httptest.NewRecorder()
	var payload struct {
		Value string `json:"value"`
	}
	err := DecodeJSON(rec, req, &payload, 8)
	if err == nil {
		t.Fatal("DecodeJSON() error = nil, want size error")
	}
	if !IsRequestTooLarge(err) {
		t.Fatalf("DecodeJSON() error = %v, want request too large", err)
	}
}

func TestDecodeJSONRejectsTrailingData(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"value":"ok"} {}`))
	rec := httptest.NewRecorder()
	var payload struct {
		Value string `json:"value"`
	}
	err := DecodeJSON(rec, req, &payload, 1024)
	if err == nil {
		t.Fatal("DecodeJSON() error = nil, want trailing data error")
	}
	if IsRequestTooLarge(err) {
		t.Fatalf("DecodeJSON() error = %v, did not expect request too large", err)
	}
}

func TestWriteJSONWritesSuccess(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteJSON(rec, http.StatusCreated, map[string]string{"status": "ok"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("content type = %q, want application/json", got)
	}
	if strings.TrimSpace(rec.Body.String()) != `{"status":"ok"}` {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestWriteJSONHandlesMarshalFailureBeforeRequestedStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteJSON(rec, http.StatusCreated, map[string]any{"bad": func() {}})
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	if !strings.Contains(rec.Body.String(), "internal server error") {
		t.Fatalf("body = %q, want internal server error", rec.Body.String())
	}
}

func TestRemoteIPStripsPort(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.RemoteAddr = "192.0.2.10:54321"
	if got := RemoteIP(req); got != "192.0.2.10" {
		t.Fatalf("RemoteIP() = %q, want 192.0.2.10", got)
	}
}

func TestIsRequestTooLargeUnwrapsMaxBytesError(t *testing.T) {
	err := &http.MaxBytesError{Limit: 1}
	if !IsRequestTooLarge(err) {
		t.Fatal("IsRequestTooLarge(MaxBytesError) = false")
	}
	if !IsRequestTooLarge(errors.Join(errors.New("decode"), err)) {
		t.Fatal("IsRequestTooLarge(joined MaxBytesError) = false")
	}
}
```

- [x] **Step 2: Run tests to verify they fail**

Run:

```bash
go test ./internal/httputil
```

Expected: FAIL because package/functions do not exist.

- [x] **Step 3: Implement helper package**

Create `internal/httputil/json.go`:

```go
package httputil

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
)

const DefaultJSONBodyLimit int64 = 1 << 20

func DecodeJSON(w http.ResponseWriter, r *http.Request, value any, limit int64) error {
	if limit <= 0 {
		limit = DefaultJSONBodyLimit
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, limit))
	if err := decoder.Decode(value); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("trailing JSON data")
	}
	return nil
}

func WriteJSON(w http.ResponseWriter, status int, value any) {
	body, err := json.Marshal(value)
	if err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(append(body, '\n'))
}

func IsRequestTooLarge(err error) bool {
	var maxBytesErr *http.MaxBytesError
	return errors.As(err, &maxBytesErr)
}

func RemoteIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}
```

- [x] **Step 4: Run helper tests**

Run:

```bash
go test ./internal/httputil
```

Expected: PASS.

- [x] **Step 5: Commit**

```bash
git add internal/httputil/json.go internal/httputil/json_test.go
git commit -m "feat: add shared HTTP JSON helpers"
```

---

### Task 4: Apply JSON Helpers And Body Limits To Handlers

**Files:**
- Modify: `app/app.go`
- Modify: `auth/http.go`
- Modify: `auth/middleware.go`
- Modify: `project/http.go`
- Modify: `agent/http.go`
- Modify: `skill/http.go`
- Modify: `auth/http_test.go`
- Modify: `project/http_test.go`

- [x] **Step 1: Add failing oversized JSON tests**

Add this test to `auth/http_test.go` or create the file if absent:

```go
func TestLoginRejectsOversizedJSON(t *testing.T) {
	db := openMigratedTestDB(t)
	service := NewService(db, "test-secret-32-bytes-minimum-value")
	handler := chi.NewRouter()
	RegisterRoutes(handler, service)

	body := strings.NewReader(`{"email":"test@example.invalid","password":"` + strings.Repeat("x", int(httputil.DefaultJSONBodyLimit)+1) + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", body)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusRequestEntityTooLarge, rec.Body.String())
	}
}
```

Imports needed:

```go
import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"git.inkyquill.net/inky/writer/internal/httputil"
	"github.com/go-chi/chi/v5"
)
```

Add one representative protected handler test in `project/http_test.go`:

```go
func TestHTTPCreateProjectRejectsOversizedJSON(t *testing.T) {
	db := openMigratedTestDB(t)
	handler := newTestProjectHTTP(NewService(db))
	req := httptest.NewRequest(http.MethodPost, "/projects", strings.NewReader(`{"title":"`+strings.Repeat("x", int(httputil.DefaultJSONBodyLimit)+1)+`"}`))
	req = req.WithContext(auth.ContextWithAuthorID(req.Context(), "author-1"))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusRequestEntityTooLarge, rec.Body.String())
	}
}
```

If `newTestProjectHTTP` does not expose `POST /projects`, wrap `project.RegisterRoutes` in a `chi.NewRouter()` in the test and serve the request through that router.

- [x] **Step 2: Run tests to verify they fail**

Run:

```bash
go test ./auth ./project -run 'TestLoginRejectsOversizedJSON|TestHTTPCreateProjectRejectsOversizedJSON' -count=1
```

Expected: FAIL because handlers return `400` or accept unbounded JSON.

- [x] **Step 3: Replace local writeJSON wrappers**

In each package, import:

```go
httpjson "git.inkyquill.net/inky/writer/internal/httputil"
```

Change each local `writeJSON` function body to:

```go
func writeJSON(w http.ResponseWriter, status int, value any) {
	httpjson.WriteJSON(w, status, value)
}
```

This preserves local call sites and keeps the diff small. Remove `encoding/json` imports from files that no longer need them.

- [x] **Step 4: Replace local decodeJSON wrappers**

In `project/http.go`, `agent/http.go`, and `skill/http.go`, change `decodeJSON` signatures to:

```go
func decodeJSON(w http.ResponseWriter, r *http.Request, value any) error {
	return httpjson.DecodeJSON(w, r, value, httpjson.DefaultJSONBodyLimit)
}
```

Update call sites from:

```go
if err := decodeJSON(r, &input); err != nil {
	writeJSON(w, http.StatusBadRequest, errorResponse{Error: "malformed JSON"})
	return
}
```

to:

```go
if err := decodeJSON(w, r, &input); err != nil {
	writeMalformedJSON(w, err)
	return
}
```

Add this package-local helper where needed:

```go
func writeMalformedJSON(w http.ResponseWriter, err error) {
	if httpjson.IsRequestTooLarge(err) {
		writeJSON(w, http.StatusRequestEntityTooLarge, errorResponse{Error: "request body too large"})
		return
	}
	writeJSON(w, http.StatusBadRequest, errorResponse{Error: "malformed JSON"})
}
```

In `auth/http.go`, decode login with:

```go
	if err := httpjson.DecodeJSON(w, r, &req, httpjson.DefaultJSONBodyLimit); err != nil {
		if httpjson.IsRequestTooLarge(err) {
			writeJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"error": "request body too large"})
			return
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
```

- [x] **Step 5: Run HTTP body-limit tests**

Run:

```bash
go test ./internal/httputil ./auth ./project ./agent ./skill -run 'Test.*OversizedJSON|TestDecodeJSON|TestWriteJSON' -count=1
```

Expected: PASS unless migration-backed package tests hit FTS5.

- [x] **Step 6: Commit**

```bash
git add internal/httputil app auth project agent skill
git commit -m "fix: cap JSON request bodies"
```

---

 Add Login Rate Limiting

**Files:**
- Modify: `auth/http.go`
- Modify: `auth/http_test.go`

- [x] **Step 1: Write failing login limiter test**

Add to `auth/http_test.go`:

```go
func TestLoginRateLimiterReturnsTooManyRequests(t *testing.T) {
	db := openMigratedTestDB(t)
	service := NewService(db, "test-secret-32-bytes-minimum-value")
	if _, err := service.Register(context.Background(), "limit@example.invalid", "correct-password"); err != nil {
		t.Fatalf("register author: %v", err)
	}
	handler := chi.NewRouter()
	RegisterRoutes(handler, service)

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(`{"email":"limit@example.invalid","password":"wrong-password"}`))
		req.RemoteAddr = "192.0.2.44:12345"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d status = %d, want %d; body = %s", i+1, rec.Code, http.StatusUnauthorized, rec.Body.String())
		}
	}

	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(`{"email":"limit@example.invalid","password":"wrong-password"}`))
	req.RemoteAddr = "192.0.2.44:54321"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("limited status = %d, want %d; body = %s", rec.Code, http.StatusTooManyRequests, rec.Body.String())
	}
}
```

- [x] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./auth -run TestLoginRateLimiterReturnsTooManyRequests -count=1
```

Expected: FAIL because the sixth login attempt is not throttled.

- [x] **Step 3: Implement process-local limiter**

In `auth/http.go`, add imports:

```go
	"sync"
	"time"

	httpjson "git.inkyquill.net/inky/writer/internal/httputil"
```

Add package-level limiter:

```go
var defaultLoginLimiter = newLoginLimiter(5, 2*time.Second, 1000)

type loginLimiter struct {
	mu          sync.Mutex
	entries     map[string]*loginLimitEntry
	burst       float64
	refillEvery time.Duration
	maxEntries  int
}

type loginLimitEntry struct {
	tokens float64
	seenAt time.Time
}

func newLoginLimiter(burst int, refillEvery time.Duration, maxEntries int) *loginLimiter {
	return &loginLimiter{
		entries:     make(map[string]*loginLimitEntry),
		burst:       float64(burst),
		refillEvery: refillEvery,
		maxEntries:  maxEntries,
	}
}

func (l *loginLimiter) allow(key string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	entry := l.entries[key]
	if entry == nil {
		entry = &loginLimitEntry{tokens: l.burst, seenAt: now}
		l.entries[key] = entry
	}
	elapsed := now.Sub(entry.seenAt)
	if elapsed > 0 {
		entry.tokens += float64(elapsed) / float64(l.refillEvery)
		if entry.tokens > l.burst {
			entry.tokens = l.burst
		}
		entry.seenAt = now
	}
	if len(l.entries) > l.maxEntries {
		l.prune(now)
	}
	if entry.tokens < 1 {
		return false
	}
	entry.tokens--
	return true
}

func (l *loginLimiter) prune(now time.Time) {
	staleAfter := time.Duration(l.burst) * l.refillEvery
	for key, entry := range l.entries {
		if entry.tokens >= l.burst && now.Sub(entry.seenAt) >= staleAfter {
			delete(l.entries, key)
		}
	}
}
```

At the start of `login`, before decoding:

```go
	if !defaultLoginLimiter.allow(httpjson.RemoteIP(r), time.Now()) {
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "too many login attempts"})
		return
	}
```

- [x] **Step 4: Run auth tests**

Run:

```bash
go test ./auth -run 'TestLogin|Test.*RateLimiter' -count=1
```

Expected: PASS.

- [x] **Step 5: Commit**

```bash
git add auth/http.go auth/http_test.go
git commit -m "fix: rate limit login attempts"
```

---

### Task 6: Remove Exported Auth Secret Accessor

**Files:**
- Modify: `auth/service.go`
- Modify tests if any use `Service.Secret()`

- [x] **Step 1: Find usages**

Run:

```bash
rg -n "Secret\\(\\)" auth app agent project skill main.go main_test.go
```

Expected before change: `auth/service.go` defines `Secret()`. If other usages appear, update those tests to use the known test secret value directly.

- [x] **Step 2: Remove method**

Delete this method from `auth/service.go`:

```go
func (s *Service) Secret() string {
	return s.secret
}
```

- [x] **Step 3: Run auth tests**

Run:

```bash
go test ./auth
```

Expected: PASS.

- [x] **Step 4: Commit**

```bash
git add auth/service.go
git commit -m "chore: remove exported auth secret accessor"
```

---

### Task 7: Cap SearchContentCandidates And Regenerate sqlc

**Files:**
- Modify: `queries/project_core.sql`
- Modify: `store/project_core.sql.go`
- Modify: `store/querier.go`
- Modify: `project/service_test.go` or `store/db_test.go`

- [x] **Step 1: Write failing cap test**

Add a store-level test in `store/db_test.go` because it can assert the generated query result directly:

```go
func TestSearchContentCandidatesHasSafetyLimit(t *testing.T) {
	db := openMigratedProjectCoreDB(t)
	ctx := context.Background()
	queries := New(db)
	now := "2026-06-20T00:00:00Z"

	if err := queries.CreateAuthor(ctx, CreateAuthorParams{
		ID:           "author-limit",
		Email:        "limit@example.invalid",
		PasswordHash: "hash",
		CreatedAt:    now,
	}); err != nil {
		t.Fatalf("create author: %v", err)
	}
	if err := queries.CreateStoryProject(ctx, CreateStoryProjectParams{
		ID:        "project-limit",
		AuthorID:  "author-limit",
		Title:     "Limit Test",
		Language:  "en",
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		t.Fatalf("create project: %v", err)
	}
	for i := 0; i < 1005; i++ {
		if err := queries.CreateContentItem(ctx, CreateContentItemParams{
			ID:              fmt.Sprintf("content-limit-%04d", i),
			ProjectID:       "project-limit",
			Kind:            "chapter",
			Title:           fmt.Sprintf("Needle %04d", i),
			Slug:            fmt.Sprintf("needle-%04d", i),
			BodyMarkdown:    "needle search body",
			MetadataJson:    "{}",
			SortOrder:       int64(i),
			CurrentRevision: 1,
			CreatedAt:       now,
			UpdatedAt:       now,
		}); err != nil {
			t.Fatalf("create content %d: %v", i, err)
		}
	}
	items, err := queries.SearchContentCandidates(ctx, SearchContentCandidatesParams{
		Query:     "needle",
		ProjectID: "project-limit",
	})
	if err != nil {
		t.Fatalf("SearchContentCandidates() error = %v", err)
	}
	if len(items) != 1000 {
		t.Fatalf("len(items) = %d, want 1000", len(items))
	}
}
```

Add `fmt` to the `store/db_test.go` imports for the `fmt.Sprintf` calls in this test.

- [x] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./store -run TestSearchContentCandidatesHasSafetyLimit -count=1
```

Expected: FAIL with `len(items) = 1005, want 1000`, unless local SQLite FTS5 is unavailable.

- [x] **Step 3: Add SQL limit**

Modify `queries/project_core.sql`:

```sql
-- name: SearchContentCandidates :many
SELECT content_items.*
FROM content_search(CAST(sqlc.arg(query) AS TEXT))
JOIN content_items ON content_items.rowid = content_search.rowid
WHERE content_items.project_id = sqlc.arg(project_id)
ORDER BY rank
LIMIT 1000;
```

- [x] **Step 4: Regenerate sqlc**

Run:

```bash
pnpm exec sqlc generate
```

If `pnpm exec sqlc generate` fails because `sqlc` is not installed in the frontend workspace, run:

```bash
sqlc generate
```

Expected: `store/project_core.sql.go` updates the query SQL constant; `store/querier.go` should remain compatible or receive only mechanical generated changes.

- [x] **Step 5: Run store test**

Run:

```bash
go test ./store -run TestSearchContentCandidatesHasSafetyLimit -count=1
```

Expected: PASS unless FTS5 is unavailable.

- [x] **Step 6: Commit**

```bash
git add queries/project_core.sql store/project_core.sql.go store/querier.go store/db_test.go
git commit -m "fix: cap content search candidates"
```

---

### Task 8: Final Verification And Review

**Files:**
- Modify: Go files changed in Tasks 1-7 when `gofmt` reports formatting differences.

- [x] **Step 1: Format Go files**

Run:

```bash
gofmt -w internal/crypto internal/httputil app auth agent project skill main.go
```

Expected: no output.

- [x] **Step 2: Run narrow tests that avoid broad frontend work**

Run:

```bash
go test ./internal/crypto ./internal/httputil ./auth
```

Expected: PASS.

- [x] **Step 3: Run backend test suite**

Run:

```bash
go test ./...
```

Expected: PASS in an environment with SQLite FTS5. If it fails with `no such module: fts5`, record that exact limitation and include the successful narrow package tests in the final report.

- [x] **Step 4: Inspect remaining diff**

Run:

```bash
git status --short
git diff --stat
```

Expected: no unstaged formatting churn. Any remaining files should belong to this plan.

- [x] **Step 5: Commit final verification fixes when files changed**

If Task 8 changed formatting or small test fixes, commit them:

```bash
git add .
git commit -m "test: verify review hardening"
```

If `git status --short` is empty after Step 4, do not create an empty commit.

---

## Self-Review

- Spec coverage: provider-key AES-GCM encryption is covered by Tasks 1-2; JSON body limits and safe JSON writing by Tasks 3-4; login rate limiting by Task 5; `Secret()` removal by Task 6; fixed `LIMIT 1000` by Task 7; verification by Task 8.
- Exclusions: script runtime boundaries and audit policy remain untouched and are already recorded in `data/backlog.md`.
- Type consistency: helper names are stable across tasks: `EncryptAPIKey`, `DecryptAPIKey`, `SetEncryptionSecret`, `DecodeJSON`, `WriteJSON`, `IsRequestTooLarge`, and `RemoteIP`.
- Environment risk: migration-backed tests may fail if SQLite lacks FTS5; the plan includes narrow tests and explicit reporting for that case.
