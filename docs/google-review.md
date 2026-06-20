# Code Quality, Security, and Readability Review for Open Edda

This document outlines the findings of a comprehensive code quality, security, and readability review conducted on the core components of the Open Edda codebase.

---

## 1. Security Findings

### Critical Risk: Lack of OS-Level Sandboxing in Skill Script Runtime
* **Files:** 
  * [runner.go](file:///home/inky/Development/writer/skill/runtime/runner.go#L54-L67)
  * [service.go](file:///home/inky/Development/writer/skill/service.go#L97-L116)
  * [script_runtime.go](file:///home/inky/Development/writer/skill/script_runtime.go#L55)
* **Description:** 
  The script runtime package executes approved skill scripts using standard command line execution (`exec.CommandContext`) via `sh -c` (on Unix/Linux) or `cmd /C` (on Windows). There is no OS-level isolation (e.g., gVisor, Docker containers, systemd sandboxing, or Linux namespaces/cgroups). Any script runs with the exact operating system privileges of the Go service user.
  Additionally, the script import process in [service.go](file:///home/inky/Development/writer/skill/service.go#L97-L116) automatically writes hardcoded audit records with `destructive_operations = 0` and `network_access = 0` rather than performing actual static code analysis.
* **Impact:** 
  If a local administrator approves a skill script (which they might do under a false sense of security provided by the auto-generated "0 risk" audits), a malicious or bug-prone script can execute arbitrary commands on the host, read/write/delete system files, or open connections to the outside network.
* **Recommendations:**
  * Implement true OS-level isolation (e.g., using lightweight containers, gVisor, or restricted child process namespaces) to execute scripts.
  * Re-evaluate the auto-generation of audit records; if a script cannot be statically analyzed reliably, mark its risk as "high/manual inspection required" instead of default zero values.

### Medium Risk: Predictable/Sequential ID Generation
* **Files:**
  * [newID](file:///home/inky/Development/writer/project/service.go#L966-L968) in `project`
  * [newID](file:///home/inky/Development/writer/agent/service.go#L1992-L1994) in `agent`
  * [newID](file:///home/inky/Development/writer/skill/service.go#L531-L533) in `skill`
* **Description:**
  Unique identifiers for entities such as projects, chapters, sessions, messages, and skills are generated using `fmt.Sprintf("%s-%d-%d", prefix, time.Now().UTC().UnixNano(), atomic.AddUint64(&idCounter, 1))`.
* **Impact:**
  While completely acceptable for a single-author self-hosted setup, sequential timestamp-based IDs are predictable. If Open Edda eventually supports multi-tenancy or collaboration, this predictability exposes the system to ID enumeration (Insecure Direct Object Reference) attacks.
* **Recommendations:**
  * Consolidate ID generation using cryptographically secure random identifiers (e.g., UUIDv4 or random hex strings), similar to the ID helper in [auth/helpers.go](file:///home/inky/Development/writer/auth/helpers.go#L9-L15) which uses `crypto/rand`.

### Low Risk: Missing Rate Limiting on Login Route
* **File:** [http.go](file:///home/inky/Development/writer/auth/http.go#L16)
* **Description:**
  The login API route `/api/auth/login` is exposed directly without any rate limiting middleware.
* **Impact:**
  Allows unlimited rapid password-guessing and brute-force attempts against the single-author instance.
* **Recommendations:**
  * Apply a basic rate-limiting middleware (e.g., token bucket or sliding window) to `/api/auth/login` to deter brute-force attacks.

---

## 2. Code Quality & Readability Findings

### Web Helper Redundancy and Unhandled Serialization Errors
* **Files:**
  * [auth/http.go](file:///home/inky/Development/writer/auth/http.go#L47-L51)
  * [project/http.go](file:///home/inky/Development/writer/project/http.go#L432-L436)
  * [skill/http.go](file:///home/inky/Development/writer/skill/http.go#L302-L306)
  * [app/app.go](file:///home/inky/Development/writer/app/app.go#L104-L108)
* **Description:**
  The HTTP helper `writeJSON` is copy-pasted identically across four different files. Moreover, it calls `w.WriteHeader(status)` *before* starting the JSON serialization. If serialization fails, the response status has already been committed as `200 OK` (or the requested code) and a truncated, invalid payload is sent to the client while the error is discarded (`_ =`).
* **Recommendations:**
  * Move `writeJSON` to a centralized utilities file.
  * Marshal the JSON to a temporary buffer first, check for errors, and write the header/status only upon successful serialization.

### Fragile Custom YAML Frontmatter Parser
* **File:** [elysium.go](file:///home/inky/Development/writer/markdownio/elysium.go#L461-L505)
* **Description:**
  Elysium layouts use custom string splitting and line-by-line matching to parse frontmatter blocks.
* **Impact:**
  It is fragile and does not support nested YAML, maps within lists, or multiline strings cleanly. If value fields contain unquoted colons, it can easily misparse.
* **Recommendations:**
  * Replace this custom logic with a well-maintained, standard YAML parser library, such as `gopkg.in/yaml.v3`.

### Middleware Access Check Robustness
* **File:** [app.go](file:///home/inky/Development/writer/app/app.go#L61-L86)
* **Description:**
  The project authorization middleware `requireProjectAccess` extracts the `projectID` from the raw URL path using custom prefix checks (`projectIDFromPath`) rather than utilizing Chi's path parameters.
* **Impact:**
  This approach makes assumptions about the path structure and is prone to break if endpoints are refactored or mounted differently.
* **Recommendations:**
  * Read the path parameter directly from the route context using Chi's context features (`chi.URLParam(r, "projectID")`) once the route is matched.

### Duplicated SQLite Error Checks
* **Files:**
  * [auth/service.go](file:///home/inky/Development/writer/auth/service.go#L133-L136)
  * [project/service.go](file:///home/inky/Development/writer/project/service.go#L974-L977)
* **Description:**
  Functions to determine if database errors correspond to SQLite constraints (`isSQLiteUniqueConstraint` and `isSQLiteConstraint`) are defined separately.
* **Recommendations:**
  * Extract database-specific error matching helpers to a shared `store` helper package.

---

## 3. Highly Commendable Implementations

* **Anti-Zip Slip and Anti-Zip Bomb Protection:** The zip extraction utility inside [http.go](file:///home/inky/Development/writer/project/http.go#L315-L363) utilizes Go's modern `filepath.IsLocal` to prevent Zip Slip path traversal attacks and imposes strict limits on file count (`maxElysiumFiles = 512`) and uncompressed size (`maxElysiumUncompressedBytes = 50 << 20`) to safely block Zip Bomb denial-of-service attempts.
* **Parametrized Queries with SQLC:** The database layer is generated using SQLC. By avoiding raw string concatenation in SQL queries and enforcing parametrized queries, the codebase is secure against SQL Injection vulnerabilities.
* **Robust SQLite Session Settings:** SQLite databases are initialized with foreign key checks (`_foreign_keys=on`), Write-Ahead Logging (`_journal_mode=WAL`), and a busy timeout in [open.go](file:///home/inky/Development/writer/store/open.go#L20-L27), providing excellent write-concurrency protection.
* **CSRF Immunity:** Stateless authentication relying on bearer tokens passed via authorization headers avoids cross-origin request forgery concerns without needing heavy CSRF tokens.
