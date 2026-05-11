# Engineering Guardrails — Root Cause Elimination Plan

**Principle**: Encode constraints into the compiler, linter, and CI, not into documents that require human memory.

**Source**: Cross-analysis of all fix logs and issue reports from 2026-05-08 to 2026-05-11.

---

## Problem → Root Cause → Structural Fix Mapping

### G1. Errors Silently Discarded (#3 High-Frequency Bug)

**Historical bug**: `safeMarshal` used `_` to ignore errors; render helpers discarded `io.Writer` errors; `2>/dev/null` swallowed gofmt errors.

**Root cause**: `.golangci.yml` excluded `G104` (Errors unhandled) from `gosec`, effectively disabling the check.

**Fix**:
- [x] Remove `gosec.excludes: G104` → `.golangci.yml`
- [x] Enable `nilerr` linter → `.golangci.yml`
- [x] Enable `errchkjson` (Go 1.21+) → `.golangci.yml`
- [x] CI uses `golangci-lint run` instead of standalone `staticcheck` → `.github/workflows/ci.yml`

---

### G2. Concurrent Resource Leaks (#2 High-Frequency Bug)

**Historical bug**: Goroutine leaks, channel double-close panics, busy-poll CPU waste, partial claims without rollback.

**Root cause**: No structured concurrency pattern; goroutine launches lack exit guarantees.

**Fix**:
- [x] Enable `copyloopvar` linter → `.golangci.yml`
- [x] Add `internal/safego` package → `internal/safego/safego.go` + `safego_test.go`
  - `Go(ctx, fn)` — auto-recover panic
  - `GoWithWaitGroup(ctx, wg, fn)` — recover panic + `wg.Done()` guarantee
- [ ] Establish goroutine discipline: all `go func()` must accept ctx and listen to `ctx.Done()`
  - *Progressive replacement: existing naked `go func()` in `dispatcher.go`/`orchestrator.go` are not forced to change all at once; new code should prefer safego*

---

### G3. Broken Context Propagation (#4 High-Frequency Bug)

**Historical bug**: Dispatcher passed `context.Background()` to executor; ContractExecutor interface had no ctx; human polling lacked `ctx.Done()`.

**Root cause**: The interface design phase did not enforce ctx as the first parameter.

**Fix**:
- [x] Enable `noctx` linter → `.golangci.yml`
- [x] Enable `revive` `context-as-argument` rule → `.golangci.yml`
- [x] Existing interfaces were fixed in core-engine-tdd-fixes; linter prevents regression

---

### G4. Windows / Cross-Platform Compatibility (#1 High-Frequency Bug)

**Historical bug**: CRLF stdin pollution, `signal.Notify` not working, hard-coded paths, port release delays.

**Root cause**: CI tests only ran on `ubuntu-latest`; Windows bugs were never caught before merge.

**Fix**:
- [x] CI test job adds `windows-latest` matrix → `.github/workflows/ci.yml`
- [x] CI build command corrected from `cmd/axis/main.go` to `./cmd/axis` → `.github/workflows/ci.yml`
- [x] Coverage/upload steps restricted to `ubuntu-latest` → `.github/workflows/ci.yml`

---

### G5. Missing Security and Boundary Validation (#5 High-Frequency Bug)

**Historical bug**: Unescaped URLs, HTTP client without timeout, Listen accepting non-loopback, file path sibling-prefix escape.

**Fix**:
- [x] Enable `bodyclose` linter → `.golangci.yml`
- [x] Enable `reassign` linter → `.golangci.yml`
- [x] Existing security fixes landed in context-preflight-fix and /review; linter prevents regression

---

### G6. golangci-lint Not Used in CI

**Root cause**: `.golangci.yml` existed but CI only ran standalone `staticcheck`, rendering all new linters ineffective.

**Fix**:
- [x] CI lint job: `staticcheck` → `golangci-lint run ./...` → `.github/workflows/ci.yml`

---

## Implementation Order

| Step | File | Change |
|---|---|---|
| 1 | `.golangci.yml` | Remove G104 exclusion, add 5 linters, add revive context-as-argument rule |
| 2 | `.github/workflows/ci.yml` | lint: golangci-lint, test: +windows-latest, build: fix command |
| 3 | `internal/safego/safego.go` | Structured goroutine launcher (ctx + panic recovery + exit guarantee) |
| 4 | `docs/architecture/engineering-guardrails.md` | This file; mark completion status |

**Expected outcome**: After steps 1+2 are complete, 4 of the historical top-5 bug categories will be automatically intercepted by CI without relying on human memory.

---

## What We Will Not Do (Avoid Over-Engineering)

- ❌ Do not write "coding standard documents" for people to read — use linters instead
- ❌ Do not add custom go vet analyzers — high maintenance cost; standard linters already cover it
- ❌ Do not enforce pre-commit hooks — CI is the single source of truth; keep local development flexible
- ❌ Do not add new document-sync CI checks — syncing 4 documents relies on Agent self-discipline, which is sufficient; automation would become over-control
