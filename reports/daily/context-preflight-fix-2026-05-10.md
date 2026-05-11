# Context & Preflight Domain Fix Report — 2026-05-10

## Scope

`internal/contextpack` + `internal/control` domain: code review follow-up fixes for bugs identified during `/review` of the context assembly, readiness registry, preflight state machine, execution consumer, and local control plane client/server/event log.

Five confirmed issues fixed; all changes are defensive and non-invasive.

---

## Convention Applied

| Issue | Category | Convention |
|---|---|---|
| #1 `taskID` not URL-escaped in client path | `semantic-boundary` | [external-tool-boundaries.md](../../docs/architecture/external-tool-boundaries.md) — HTTP surface must handle special characters safely |
| #8 `http.DefaultClient` fallback has no timeout | `semantic-boundary` | [external-tool-boundaries.md](../../docs/architecture/external-tool-boundaries.md) — local HTTP client needs resource bounds |
| #2 Negative `PacketCount` from corrupted metadata passes as `ready` | `semantic-boundary` | [metadata-key-conventions.md](../../docs/architecture/metadata-key-conventions.md) — metadata consumers must validate inbound values |
| #7 `Listen` accepts non-loopback addresses | `secret` / `semantic-boundary` | [external-tool-boundaries.md](../../docs/architecture/external-tool-boundaries.md) — local-only surface must not expose to network |
| #9 GET requests carry empty-but-non-nil body | `cli-output` | [cli-output-conventions.md](../../docs/architecture/cli-output-conventions.md) — behavior should be idiomatic and minimal |

---

## Evolution Boundary Check

**Result:** No evolution boundary crossed.

- No scheduler/orchestrator/contract/provider semantic changes.
- No new feature scope introduced.
- No metadata key additions or schema changes.
- No permission or autonomy boundary changes.
- No sandboxed evolution workspace needed.

All changes are pure defensive hardening: URL escaping, bounds validation, timeout defaults, address enforcement, and request body cleanup. No user-visible behavior changes.

---

## Files Changed

### Production code
- `internal/control/client.go` — `Status` uses `url.PathEscape(taskID)`; `NewClient` defaults to `&http.Client{Timeout: 30s}`; `do` passes `nil` body reader for GET instead of `bytes.NewReader(nil)`
- `internal/contextpack/preflight.go` — `PacketCount <= 0` treated as `missing` (was `== 0`)
- `internal/control/server.go` — `Listen` enforces loopback via `ResolveTCPAddr` + `IsLoopback()` check

### Test code
- `internal/control/client_test.go` — added `TestClientStatusURLEscapesTaskID`, `TestClientGETDoesNotSendBody`, `TestClientUsesTimeout`
- `internal/control/server_test.go` — added `TestControlServerListenRejectsNonLoopback`
- `internal/contextpack/preflight_test.go` — added `TestPreflightNegativePacketCount`

---

## Behavior Preserved

- All CLI command output strings identical before/after.
- `axis ask --submit` cross-process submission unchanged.
- `axis status <task-id>` unchanged.
- Context preview, inspect, preflight semantics unchanged.
- Local control plane startup/shutdown sequence unchanged.

---

## Tests Run

```bash
go test ./internal/control -count=1              # PASS (all existing + 4 new tests)
go test ./internal/contextpack -count=1          # PASS (all existing + 1 new test)
go test ./cmd/axis ./internal/kernel/... ./internal/types -count=1 -timeout=120s  # PASS
```

Full suite `go test ./...` shows one pre-existing flake in `internal/model/provider` (`TestOpenAIProvider_Execute_HistoryIncludesToolCallID`) when run in batch, but passes in isolation. This is an existing test interdependency issue, unrelated to this change.

---

## Remaining Gaps (from original review, out of scope)

| # | Issue | Category | Risk |
|---|---|---|---|
| 4 | Bundle ID field-boundary collision with `\n` in goal/contract | `semantic-boundary` | Low — P0 rules don't produce `\n` in goal |
| 5 | `ExecutionContextConsumer.Summarize` double registry lookup (TOCTOU) | `semantic-boundary` | Low — P0 in-process only |
| 6 | `TaskEventLog` `EventID` relies on clock resolution | `test-gap` | Low — serialized by mutex |
| 10 | `preflight` nil registry path not tested | `test-gap` | Low — covered by consumer tests |
| 11 | `execution_consumer` nil consumer not tested | `test-gap` | Low — behavior is simple fallback |
| 12 | `registry` concurrent access not explicitly tested | `test-gap` | Low — `sync.RWMutex` present |
| 13 | `assembler` duplicate packet IDs not deduplicated | `semantic-boundary` | Low — rule packets have unique IDs |
| 14 | `packet.Validate` doesn't check relevance/type bounds | `semantic-boundary` | Low — P0 rule packets are valid |
| 15 | `locator.Load` doesn't validate empty address/protocol | `semantic-boundary` | Low — runtime writes valid records |

---

## Acceptance Checklist

- [x] Changes limited to declared scope (Context & Preflight domain)
- [x] Behavior preserved or intentionally documented
- [x] Tests pass (`go test ./internal/control ./internal/contextpack`)
- [x] No new feature scope introduced
- [x] Relevant conventions referenced
- [x] Remaining gaps listed clearly
- [x] Evolution boundary check completed — no sandbox needed

---

*Report produced per [swe1-6-renormalization-guide.md](../../docs/architecture/swe1-6-renormalization-guide.md).*
