# Production Hardening Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking. Commit steps are optional; only run them when the user or repository workflow explicitly wants commits.

**Spec:** `docs/superpowers/specs/2026-05-03-production-hardening.md`

**Goal:** Make PocketBase libSQL/Turso wrapper safer for production by adding explicit remote requirements, safer token handling, graceful final sync, tests, CI checks, release checksums, and production docs.

**Architecture:** Keep existing small codebase shape. Extract env/config and libSQL connection-string behavior into focused, testable helpers while leaving platform-specific DB opening in build-tagged files. Add production safety as opt-in strict mode so local development behavior remains unchanged unless `LIBSQL_REQUIRE_REMOTE=true`.

**Tech Stack:** Go 1.26.2, PocketBase v0.37.5, dbx, go-libsql, libsql-client-go, Cobra, GitHub Actions.

---

## File Structure

- Modify: `main.go`
  - Owns app startup, environment loading, production safety validation, command registration, and terminate hook.
  - Adds `appConfig`, `loadAppConfigFromEnv()`, `validateAppConfig()`, `isTruthyEnv()`, and uses config in `DBConnect`.

- Create: `main_test.go`
  - Unit tests for env parsing, production remote validation, and non-DB command detection.

- Create: `libsql_conn.go`
  - Owns safe construction of remote libSQL connection strings and URL masking helpers.
  - Used by remote fallback files.

- Create: `libsql_conn_test.go`
  - Unit tests for token URL encoding, preserving existing query params, empty token behavior, and masked URL output.

- Modify: `db_darwin_amd64.go`
  - Replaces manual `?authToken=` concatenation with `buildLibSQLConnectionString()`.
  - Logs masked remote URL only.

- Modify: `db_windows.go`
  - Replaces manual `?authToken=` concatenation with `buildLibSQLConnectionString()`.
  - Logs masked remote URL only.

- Modify: `db_embedded.go`
  - Logs masked remote URL only.
  - Runs final `Sync()` before connector close.

- Modify: `.env.example`
  - Documents `LIBSQL_REQUIRE_REMOTE` and production behavior.

- Modify: `.github/workflows/release.yml`
  - Adds test/vet job before build.
  - Adds smoke checks after build.
  - Adds `checksums.txt` to release assets.

- Modify: `README.md`
  - Adds production safety, env reference table, token handling note, graceful shutdown behavior, and release checksum usage.

---

### Task 1: Add Production Remote Validation

**Files:**
- Modify: `main.go`
- Create: `main_test.go`

- [ ] **Step 1: Write failing tests for env config and strict remote validation**

Create `main_test.go` with:

```go
package main

import (
	"strings"
	"testing"
	"time"
)

func TestIsTruthyEnv(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{name: "true", in: "true", want: true},
		{name: "uppercase true", in: "TRUE", want: true},
		{name: "trimmed true", in: "  true  ", want: true},
		{name: "one", in: "1", want: true},
		{name: "yes", in: "yes", want: true},
		{name: "on", in: "on", want: true},
		{name: "false", in: "false", want: false},
		{name: "empty", in: "", want: false},
		{name: "random", in: "maybe", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isTruthyEnv(tt.in)
			if got != tt.want {
				t.Fatalf("isTruthyEnv(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestLoadAppConfigFromEnv(t *testing.T) {
	t.Setenv("LIBSQL_DATABASE_URL", "libsql://example.turso.io")
	t.Setenv("LIBSQL_AUTH_TOKEN", "secret-token")
	t.Setenv("LIBSQL_SYNC_INTERVAL", "15")
	t.Setenv("LIBSQL_REQUIRE_REMOTE", "true")

	cfg := loadAppConfigFromEnv()

	if cfg.libsqlURL != "libsql://example.turso.io" {
		t.Fatalf("libsqlURL = %q", cfg.libsqlURL)
	}
	if cfg.libsqlToken != "secret-token" {
		t.Fatalf("libsqlToken = %q", cfg.libsqlToken)
	}
	if cfg.syncInterval != 15*time.Second {
		t.Fatalf("syncInterval = %v", cfg.syncInterval)
	}
	if !cfg.requireRemote {
		t.Fatalf("requireRemote = false, want true")
	}
}

func TestLoadAppConfigFromEnvUsesDefaultSyncIntervalForInvalidValue(t *testing.T) {
	t.Setenv("LIBSQL_SYNC_INTERVAL", "not-a-number")

	cfg := loadAppConfigFromEnv()

	if cfg.syncInterval != 60*time.Second {
		t.Fatalf("syncInterval = %v, want 60s", cfg.syncInterval)
	}
}

func TestValidateAppConfigAllowsLocalWhenRemoteNotRequired(t *testing.T) {
	cfg := appConfig{requireRemote: false}

	if err := validateAppConfig(cfg); err != nil {
		t.Fatalf("validateAppConfig() error = %v, want nil", err)
	}
}

func TestValidateAppConfigRequiresURLWhenStrict(t *testing.T) {
	cfg := appConfig{requireRemote: true, libsqlToken: "secret"}

	err := validateAppConfig(cfg)
	if err == nil {
		t.Fatalf("validateAppConfig() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "LIBSQL_DATABASE_URL is required") {
		t.Fatalf("error = %q, want LIBSQL_DATABASE_URL message", err.Error())
	}
}

func TestValidateAppConfigRequiresTokenWhenStrict(t *testing.T) {
	cfg := appConfig{requireRemote: true, libsqlURL: "libsql://example.turso.io"}

	err := validateAppConfig(cfg)
	if err == nil {
		t.Fatalf("validateAppConfig() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "LIBSQL_AUTH_TOKEN is required") {
		t.Fatalf("error = %q, want LIBSQL_AUTH_TOKEN message", err.Error())
	}
}

func TestShouldSkipDBInit(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{name: "no args", args: []string{"app"}, want: false},
		{name: "help flag", args: []string{"app", "--help"}, want: true},
		{name: "short help", args: []string{"app", "-h"}, want: true},
		{name: "version flag", args: []string{"app", "--version"}, want: true},
		{name: "short version", args: []string{"app", "-v"}, want: true},
		{name: "help command", args: []string{"app", "help"}, want: true},
		{name: "update command", args: []string{"app", "update"}, want: true},
		{name: "serve command", args: []string{"app", "serve"}, want: false},
		{name: "migrate command", args: []string{"app", "migrate"}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldSkipDBInitArgs(tt.args)
			if got != tt.want {
				t.Fatalf("shouldSkipDBInitArgs(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}
```

- [ ] **Step 2: Run tests to verify failure**

Run:

```bash
go test ./...
```

Expected: FAIL with undefined identifiers like:

```text
undefined: isTruthyEnv
undefined: loadAppConfigFromEnv
undefined: appConfig
undefined: validateAppConfig
undefined: shouldSkipDBInitArgs
```

- [ ] **Step 3: Implement config helpers and validation in `main.go`**

Modify `main.go` imports to include `errors` and `strings`:

```go
import (
	"errors"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/jsvm"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	"github.com/spf13/cobra"
)
```

Add this after `connectorsMu` var block:

```go
type appConfig struct {
	libsqlURL     string
	libsqlToken   string
	syncInterval  time.Duration
	requireRemote bool
}

func loadAppConfigFromEnv() appConfig {
	syncIntervalSec := getEnvInt("LIBSQL_SYNC_INTERVAL", 60)

	return appConfig{
		libsqlURL:     os.Getenv("LIBSQL_DATABASE_URL"),
		libsqlToken:   os.Getenv("LIBSQL_AUTH_TOKEN"),
		syncInterval:  time.Duration(syncIntervalSec) * time.Second,
		requireRemote: isTruthyEnv(os.Getenv("LIBSQL_REQUIRE_REMOTE")),
	}
}

func validateAppConfig(cfg appConfig) error {
	if !cfg.requireRemote {
		return nil
	}

	var missing []string
	if cfg.libsqlURL == "" {
		missing = append(missing, "LIBSQL_DATABASE_URL is required when LIBSQL_REQUIRE_REMOTE=true")
	}
	if cfg.libsqlToken == "" {
		missing = append(missing, "LIBSQL_AUTH_TOKEN is required when LIBSQL_REQUIRE_REMOTE=true")
	}

	if len(missing) > 0 {
		return errors.New(strings.Join(missing, "; "))
	}

	return nil
}

func isTruthyEnv(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
```

Replace env parsing block in `main()`:

```go
	libsqlURL := os.Getenv("LIBSQL_DATABASE_URL")
	libsqlToken := os.Getenv("LIBSQL_AUTH_TOKEN")
	syncIntervalSec := getEnvInt("LIBSQL_SYNC_INTERVAL", 60)
	syncInterval := time.Duration(syncIntervalSec) * time.Second
```

with:

```go
	cfg := loadAppConfigFromEnv()
	if err := validateAppConfig(cfg); err != nil {
		log.Fatalf("configuration error: %v", err)
	}
```

Replace `DBConnect` body:

```go
		DBConnect: func(dbPath string) (*dbx.DB, error) {
			return dbConnect(dbPath, libsqlURL, libsqlToken, syncInterval)
		},
```

with:

```go
		DBConnect: func(dbPath string) (*dbx.DB, error) {
			return dbConnect(dbPath, cfg.libsqlURL, cfg.libsqlToken, cfg.syncInterval)
		},
```

Replace `shouldSkipDBInit()` with:

```go
func shouldSkipDBInit() bool {
	return shouldSkipDBInitArgs(os.Args)
}

func shouldSkipDBInitArgs(args []string) bool {
	if len(args) < 2 {
		return false
	}

	cmd := args[1]

	if cmd == "update" || cmd == "--help" || cmd == "-h" || cmd == "--version" || cmd == "-v" || cmd == "help" {
		return true
	}

	return false
}
```

- [ ] **Step 4: Run tests to verify pass**

Run:

```bash
go test ./...
```

Expected:

```text
ok  	myapp	...
```

- [ ] **Step 5: Commit production validation (optional)**

Run:

```bash
git add main.go main_test.go
git commit -m "feat: add production remote validation"
```

Expected: commit succeeds.

---

### Task 2: Add Safe libSQL Connection String Helpers

**Files:**
- Create: `libsql_conn.go`
- Create: `libsql_conn_test.go`
- Modify: `db_darwin_amd64.go`
- Modify: `db_windows.go`

- [ ] **Step 1: Write failing tests for connection string and masking**

Create `libsql_conn_test.go` with:

```go
package main

import "testing"

func TestBuildLibSQLConnectionStringEmptyToken(t *testing.T) {
	got, err := buildLibSQLConnectionString("libsql://example.turso.io", "")
	if err != nil {
		t.Fatalf("buildLibSQLConnectionString() error = %v", err)
	}
	want := "libsql://example.turso.io"
	if got != want {
		t.Fatalf("connection string = %q, want %q", got, want)
	}
}

func TestBuildLibSQLConnectionStringAddsEncodedToken(t *testing.T) {
	got, err := buildLibSQLConnectionString("libsql://example.turso.io", "secret token+/=")
	if err != nil {
		t.Fatalf("buildLibSQLConnectionString() error = %v", err)
	}
	want := "libsql://example.turso.io?authToken=secret+token%2B%2F%3D"
	if got != want {
		t.Fatalf("connection string = %q, want %q", got, want)
	}
}

func TestBuildLibSQLConnectionStringPreservesExistingQuery(t *testing.T) {
	got, err := buildLibSQLConnectionString("libsql://example.turso.io?tls=1", "secret")
	if err != nil {
		t.Fatalf("buildLibSQLConnectionString() error = %v", err)
	}
	want := "libsql://example.turso.io?authToken=secret&tls=1"
	if got != want {
		t.Fatalf("connection string = %q, want %q", got, want)
	}
}

func TestBuildLibSQLConnectionStringInvalidURL(t *testing.T) {
	_, err := buildLibSQLConnectionString("://bad-url", "secret")
	if err == nil {
		t.Fatalf("buildLibSQLConnectionString() error = nil, want error")
	}
}

func TestMaskLibSQLURL(t *testing.T) {
	got := maskLibSQLURL("libsql://example.turso.io?authToken=secret&tls=1")
	want := "libsql://example.turso.io?authToken=***&tls=1"
	if got != want {
		t.Fatalf("maskLibSQLURL() = %q, want %q", got, want)
	}
}

func TestMaskLibSQLURLInvalidInput(t *testing.T) {
	got := maskLibSQLURL("://bad-url")
	want := "<invalid-url>"
	if got != want {
		t.Fatalf("maskLibSQLURL() = %q, want %q", got, want)
	}
}
```

- [ ] **Step 2: Run tests to verify failure**

Run:

```bash
go test ./...
```

Expected: FAIL with undefined identifiers:

```text
undefined: buildLibSQLConnectionString
undefined: maskLibSQLURL
```

- [ ] **Step 3: Implement helper file**

Create `libsql_conn.go` with:

```go
package main

import (
	"net/url"
	"strings"
)

func buildLibSQLConnectionString(rawURL string, token string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	if u.Scheme == "" || u.Host == "" {
		return "", errMissingSchemeOrHost{}
	}
	if token == "" {
		return rawURL, nil
	}

	q := u.Query()
	q.Set("authToken", token)
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func maskLibSQLURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "<invalid-url>"
	}

	if u.RawQuery == "" {
		return u.String()
	}

	parts := strings.Split(u.RawQuery, "&")
	for i, part := range parts {
		key, _, _ := strings.Cut(part, "=")
		decodedKey, err := url.QueryUnescape(key)
		if err != nil {
			return "<invalid-url>"
		}
		if decodedKey == "authToken" {
			parts[i] = key + "=***"
		}
	}
	u.RawQuery = strings.Join(parts, "&")

	return u.String()
}

type errMissingSchemeOrHost struct{}

func (errMissingSchemeOrHost) Error() string {
	return "missing URL scheme or host"
}
```

- [ ] **Step 4: Update remote fallback files to use helper**

In `db_darwin_amd64.go`, replace log block:

```go
	if !seen[dbPath] {
		log.Printf("[DB] Connecting to remote libSQL (macOS Intel fallback):")
		log.Printf("     Remote: %s", url)
		log.Printf("     Note: Embedded replica is not supported on macOS Intel")
		seen[dbPath] = true
	}

	connStr := url
	if token != "" {
		if strings.Contains(connStr, "?") {
			connStr += "&authToken=" + token
		} else {
			connStr += "?authToken=" + token
		}
	}

	return dbx.Open("libsql", connStr)
```

with:

```go
	connStr, err := buildLibSQLConnectionString(url, token)
	if err != nil {
		return nil, err
	}

	if !seen[dbPath] {
		log.Printf("[DB] Connecting to remote libSQL (macOS Intel fallback):")
		log.Printf("     Remote: %s", maskLibSQLURL(connStr))
		log.Printf("     Note: Embedded replica is not supported on macOS Intel")
		seen[dbPath] = true
	}

	return dbx.Open("libsql", connStr)
```

Remove unused `strings` import from `db_darwin_amd64.go` if compiler reports it.

In `db_windows.go`, replace log block:

```go
	if !seen[dbPath] {
		log.Printf("[DB] Connecting to remote libSQL (Windows fallback):")
		log.Printf("     Remote: %s", url)
		log.Printf("     Note: Embedded replica is not supported on Windows")
		seen[dbPath] = true
	}

	// Windows fallback uses HTTP driver via connection string
	connStr := url
	if token != "" {
		if strings.Contains(connStr, "?") {
			connStr += "&authToken=" + token
		} else {
			connStr += "?authToken=" + token
		}
	}

	return dbx.Open("libsql", connStr)
```

with:

```go
	connStr, err := buildLibSQLConnectionString(url, token)
	if err != nil {
		return nil, err
	}

	if !seen[dbPath] {
		log.Printf("[DB] Connecting to remote libSQL (Windows fallback):")
		log.Printf("     Remote: %s", maskLibSQLURL(connStr))
		log.Printf("     Note: Embedded replica is not supported on Windows")
		seen[dbPath] = true
	}

	return dbx.Open("libsql", connStr)
```

Remove unused `strings` import from `db_windows.go` if compiler reports it.

- [ ] **Step 5: Run tests to verify pass on current platform**

Run:

```bash
go test ./...
```

Expected:

```text
ok  	myapp	...
```

- [ ] **Step 6: Compile remote fallback targets**

Run:

```bash
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go test ./...
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go test ./...
```

Expected for each:

```text
ok  	myapp	...
```

- [ ] **Step 7: Commit safe connection string helpers (optional)**

Run:

```bash
git add libsql_conn.go libsql_conn_test.go db_darwin_amd64.go db_windows.go
git commit -m "fix: encode libsql auth token in remote fallbacks"
```

Expected: commit succeeds.

---

### Task 3: Add Final Sync Before Embedded Connector Close

**Files:**
- Modify: `db_embedded.go`

- [ ] **Step 1: Update embedded logs to mask remote URL**

In `db_embedded.go`, replace:

```go
	log.Printf("[DB] Creating embedded replica for main db:")
	log.Printf("     Local:  %s", dbPath)
	log.Printf("     Remote: %s", url)
	log.Printf("     Sync:   every %v", syncInterval)
```

with:

```go
	log.Printf("[DB] Creating embedded replica for main db:")
	log.Printf("     Local:  %s", dbPath)
	log.Printf("     Remote: %s", maskLibSQLURL(url))
	log.Printf("     Sync:   every %v", syncInterval)
```

- [ ] **Step 2: Add final sync before close**

In `closeConnectors()` in `db_embedded.go`, replace:

```go
	for path, c := range connectors {
		if c != nil {
			log.Printf("[DB] Closing embedded replica connector for %s...", path)
			if err := c.Close(); err != nil {
				log.Printf("[DB] Warning closing connector for %s: %v", path, err)
			}
		}
	}
```

with:

```go
	for path, c := range connectors {
		if c != nil {
			log.Printf("[DB] Performing final sync for embedded replica %s...", path)
			if _, err := c.Sync(); err != nil {
				log.Printf("[DB] Warning during final sync for %s: %v", path, err)
			} else {
				log.Printf("[DB] Final sync complete for %s", path)
			}

			log.Printf("[DB] Closing embedded replica connector for %s...", path)
			if err := c.Close(); err != nil {
				log.Printf("[DB] Warning closing connector for %s: %v", path, err)
			}
		}
	}
```

- [ ] **Step 3: Run tests**

Run:

```bash
go test ./...
```

Expected:

```text
ok  	myapp	...
```

- [ ] **Step 4: Commit final sync behavior (optional)**

Run:

```bash
git add db_embedded.go
git commit -m "fix: sync embedded replica before shutdown"
```

Expected: commit succeeds.

---

### Task 4: Document Production Environment Behavior

**Files:**
- Modify: `.env.example`
- Modify: `README.md`

- [ ] **Step 1: Update `.env.example`**

Replace `.env.example` content with:

```env
# Main libSQL Database (Turso remote URL)
LIBSQL_DATABASE_URL=libsql://your-db-name.turso.io
LIBSQL_AUTH_TOKEN=your-auth-token

# Production safety guard.
# When true, the app fails fast if LIBSQL_DATABASE_URL or LIBSQL_AUTH_TOKEN is missing.
# Recommended for production deployments.
LIBSQL_REQUIRE_REMOTE=false

# Background sync interval in seconds (optional, defaults to 60)
# LIBSQL_SYNC_INTERVAL=60
```

- [ ] **Step 2: Add README production safety section**

In `README.md`, after Setup section env block, add:

```markdown
### Environment variables

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `LIBSQL_DATABASE_URL` | Production: yes | empty | Turso/libSQL remote database URL, for example `libsql://your-db-name.turso.io`. If empty and `LIBSQL_REQUIRE_REMOTE=false`, main DB falls back to local SQLite. |
| `LIBSQL_AUTH_TOKEN` | Production: yes | empty | Turso auth token. Required when `LIBSQL_REQUIRE_REMOTE=true`. |
| `LIBSQL_REQUIRE_REMOTE` | no | `false` | Set to `true` in production to fail fast when remote URL or token is missing. Prevents accidental local SQLite usage. |
| `LIBSQL_SYNC_INTERVAL` | no | `60` | Embedded replica background sync interval in seconds. |

### Production safety

Set this in production:

```env
LIBSQL_REQUIRE_REMOTE=true
```

With this enabled, startup fails if `LIBSQL_DATABASE_URL` or `LIBSQL_AUTH_TOKEN` is missing. This prevents accidental production boot using local SQLite.

Remote URLs printed in logs are masked when an `authToken` query parameter is present.
```

- [ ] **Step 3: Add README graceful shutdown section**

In `README.md`, after `How it Works` section platform bullet list, add:

```markdown
## Shutdown behavior

On Linux and macOS arm64 embedded replica builds, the app attempts a final `Sync()` before closing the libSQL connector during PocketBase termination. If final sync fails, shutdown continues and logs a warning because network failures can happen during process termination.
```

- [ ] **Step 4: Run tests**

Run:

```bash
go test ./...
```

Expected:

```text
ok  	myapp	...
```

- [ ] **Step 5: Commit documentation (optional)**

Run:

```bash
git add .env.example README.md
git commit -m "docs: document production remote safety"
```

Expected: commit succeeds.

---

### Task 5: Harden Release Workflow

**Files:**
- Modify: `.github/workflows/release.yml`

- [ ] **Step 1: Add test job before build**

In `.github/workflows/release.yml`, replace the existing top-level `jobs:` line and beginning of the `build:` job with this single top-level `jobs:` block. Do not create a duplicate `jobs:` key:

```yaml
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache: true

      - name: Run tests
        run: go test ./...

      - name: Run vet
        run: go vet ./...

  build:
```

Keep the rest of the existing `build:` job contents after this header:

```yaml
  build:
    needs: test
    runs-on: ${{ matrix.os }}
```

- [ ] **Step 2: Add smoke check after build**

In `Build binaries` step, after both `go build` commands, append:

```bash
          if [ "${{ matrix.goos }}" = "linux" ] && [ "${{ matrix.goarch }}" = "amd64" ]; then
            "dist/${OUT_NORMAL}${EXT}" --help >/dev/null
            "dist/${OUT_NORMAL}${EXT}" --version >/dev/null
            "dist/${OUT_SLIM}${EXT}" --help >/dev/null
            "dist/${OUT_SLIM}${EXT}" --version >/dev/null
          fi
```

- [ ] **Step 3: Add checksums to release assets**

In `Collect assets` step, replace run block:

```yaml
        run: |
          mkdir -p release_assets
          # Flatten nested artifact directories into a single folder for release
          find dist_all -type f -maxdepth 3 -print -exec cp {} release_assets/ \;
```

with:

```yaml
        run: |
          mkdir -p release_assets
          # Flatten nested artifact directories into a single folder for release
          find dist_all -maxdepth 3 -type f -print -exec cp {} release_assets/ \;
          cd release_assets
          find . -maxdepth 1 -type f ! -name checksums.txt -printf '%f\0' | sort -z | xargs -0 sha256sum > checksums.txt
```

- [ ] **Step 4: Validate workflow YAML syntax by printing relevant file**

Run:

```bash
sed -n '1,220p' .github/workflows/release.yml
```

Expected: one top-level `jobs:` block, `test` job before `build`, `build` has `needs: test`, release still has `needs: build`, and checksum generation excludes `checksums.txt` itself.

- [ ] **Step 5: Run local tests**

Run:

```bash
go test ./...
go vet ./...
```

Expected:

```text
ok  	myapp	...
```

and `go vet` exits 0.

- [ ] **Step 6: Commit release workflow hardening (optional)**

Run:

```bash
git add .github/workflows/release.yml
git commit -m "ci: test and checksum release builds"
```

Expected: commit succeeds.

---

### Task 6: Rename Go Module

**Files:**
- Modify: `go.mod`
- Modify: `go.sum` if Go updates it

- [ ] **Step 1: Change module path**

In `go.mod`, replace:

```go
module myapp
```

with:

```go
module github.com/fadlee/pocketbase-libsql
```

- [ ] **Step 2: Run module tidy and tests**

Run:

```bash
go mod tidy
go test ./...
```

Expected:

```text
ok  	github.com/fadlee/pocketbase-libsql	...
```

- [ ] **Step 3: Commit module rename (optional)**

Run:

```bash
git add go.mod go.sum
git commit -m "chore: rename go module"
```

Expected: commit succeeds.

---

### Task 7: Final Verification

**Files:**
- No code changes expected

- [ ] **Step 1: Run full local verification**

Run:

```bash
go test ./...
go vet ./...
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go test ./...
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go test ./...
```

Expected: all commands exit 0.

- [ ] **Step 2: Build Linux amd64 binary**

Run:

```bash
CGO_ENABLED=1 go build -trimpath -ldflags="-s -w" -o /tmp/pocketbase-libsql .
/tmp/pocketbase-libsql --help >/dev/null
/tmp/pocketbase-libsql --version >/dev/null
```

Expected: all commands exit 0.

- [ ] **Step 3: Check git diff**

Run:

```bash
git status --short
git log --oneline -7
```

Expected if optional commits were created: working tree clean and recent commits include:

```text
feat: add production remote validation
fix: encode libsql auth token in remote fallbacks
fix: sync embedded replica before shutdown
docs: document production remote safety
ci: test and checksum release builds
chore: rename go module
```

---

## Self-Review

**Spec coverage:**
- Production fail-fast covered by Task 1 and Task 4.
- Safer token handling covered by Task 2.
- Final embedded replica sync covered by Task 3.
- Tests covered by Tasks 1, 2, and 7.
- CI and release checksums covered by Task 5.
- Module rename covered by Task 6.
- Production docs covered by Task 4.

**Placeholder scan:**
- No `TBD`, `TODO`, `implement later`, or vague edge-case instructions.
- Each code-changing step includes concrete code.
- Each test step includes exact command and expected result.

**Type consistency:**
- `appConfig`, `loadAppConfigFromEnv()`, `validateAppConfig()`, `isTruthyEnv()`, and `shouldSkipDBInitArgs()` names match across tests and implementation.
- `buildLibSQLConnectionString()` and `maskLibSQLURL()` names match across tests and platform files.
