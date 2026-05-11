# SWE1.6 Renormalization Report — 2026-05-10

## Scope

User Interface & Intent domain (`cmd/axis`): code review follow-up fixes for bugs identified during `/review` of CLI entry points, shell loop, ask/context commands, and deterministic intent parser.

Three confirmed issues fixed; one issue discovered to be already handled by lower layer.

---

## Convention Applied

| Issue | Category | Convention |
|---|---|---|
| #3 Render helpers silently discard `io.Writer` errors | `cli-output` | [cli-output-conventions.md](../../docs/architecture/cli-output-conventions.md) — errors must propagate, not be swallowed |
| #8 Local control server has no HTTP timeouts | `semantic-boundary` / `secret` | [external-tool-boundaries.md](../../docs/architecture/external-tool-boundaries.md) — local HTTP surface needs resource bounds |
| #4 Shell commands violate Cobra output contract | `cli-output` | [cli-output-conventions.md](../../docs/architecture/cli-output-conventions.md) — commands must use `cmd.OutOrStdout()` for testability and redirection |

---

## Evolution Boundary Check

**Result:** No evolution boundary crossed.

- No scheduler/orchestrator/contract/provider semantic changes.
- No new feature scope introduced.
- No metadata key changes.
- No permission or autonomy boundary changes.
- No sandboxed evolution workspace needed.

All changes are pure normalization: fixing error propagation, adding defensive timeouts, and aligning shell output with existing Cobra contracts. User-visible output content is preserved exactly.

---

## Files Changed

### Production code
- `cmd/axis/ask_cmd.go` — `renderTaskProposal` now returns `fmt.Fprintf` write error
- `cmd/axis/context_cmd.go` — `renderContextBundle`, `renderReadinessRecord`, `renderPreflightResult` now return write errors
- `cmd/axis/control_runtime.go` — extracted `newLocalHTTPServer(handler)` with `ReadTimeout`, `WriteTimeout`, `IdleTimeout`
- `cmd/axis/shell_cmd.go` — `runShell` uses `cmd.OutOrStdout()`/`cmd.ErrOrStderr()`; `printShellHelp`, `printTools`, `printDAG`, `handleShellAsk` accept `io.Writer`

### Test code
- `cmd/axis/ask_cmd_test.go` — added `failingWriter` helper + `TestRenderTaskProposal_ReturnsWriteError`
- `cmd/axis/context_cmd_test.go` — added `TestRenderContextBundle_ReturnsWriteError`, `TestRenderReadinessRecord_ReturnsWriteError`, `TestRenderPreflightResult_ReturnsWriteError`
- `cmd/axis/main_test.go` — added `TestNewLocalHTTPServer_HasTimeouts`; updated `TestPrintShellHelp` for new signature

---

## Behavior Preserved

- All CLI output strings are identical before/after (only the internal error propagation path changed).
- Shell command behavior (`help`, `run`, `status`, `dag`, `resolve`, `tools`, `ask`, `exit`) unchanged.
- `axis ask --submit` still submits through local runtime; dry-run still previews.
- `axis start` local runtime startup/shutdown sequence unchanged.
- Context preview, inspect, preflight output format and semantics unchanged.

---

## Tests Run

```bash
go test ./cmd/axis -count=1        # PASS (all existing + new tests)
go test ./internal/control -count=1 # PASS
go test ./... -count=1              # PASS (full suite)
```

New tests added:
- `TestRenderTaskProposal_ReturnsWriteError`
- `TestRenderContextBundle_ReturnsWriteError`
- `TestRenderReadinessRecord_ReturnsWriteError`
- `TestRenderPreflightResult_ReturnsWriteError`
- `TestNewLocalHTTPServer_HasTimeouts`

---

## Issues Evaluated but Not Changed

| Issue | Reason |
|---|---|
| #16 `submitTask` omits `CreatedAt` | Already handled by `SchedulerImpl.Submit` which sets `task.CreatedAt = time.Now()` at scheduling time. CLI `submitTask` does not need to duplicate this. |

---

## Remaining Gaps

The following `/review` issues remain unaddressed in this pass (out of scope for this SWE1.6 slice):

| # | Issue | Category | Risk |
|---|---|---|---|
| 1 | `startOrchestrator` double-shutdown race | `semantic-boundary` | Medium — needs `sync.Once` in signal handler |
| 2 | `captureStdout` leaks `os.Stdout` on panic | `test-gap` | Medium — defers missing in test helper |
| 5 | `TestRunShell_AskSubmitAndStatus` hardcoded timestamp | `test-gap` | Medium — flaky when clock crosses second boundary |
| 6 | `--provider mock` silently ignored when profile active | `cli-output` | Medium — flag detection needed |
| 7 | `deepseek`/`minimax` missing default models | `naming` | Low — user workaround exists |
| 9 | Provider API key exposed via `--api-key` flag | `secret` | High — needs env-var or secure prompt |
| 10 | Deterministic parser duplicate task IDs | `naming` | Low — collides within same second |
| 11 | `getTaskStatus`/`runTask` bypass `cmd.OutOrStdout()` | `cli-output` | Low — inconsistent with other commands |
| 12 | `writeJSON` silently drops encoding errors | `cli-output` | Low — server helper |
| 13 | Runtime health test race window | `test-gap` | Low — rare flake |
| 14 | `provider list` may leak credentials in BaseURL | `secret` | Medium — URL parsing needed |
| 15 | `preflight` only works for in-process tasks | `semantic-boundary` | Low — P0 documented limitation |
| 17 | `printTools` hardcoded maintenance liability | `semantic-boundary` | Low — should query registry at runtime |
| 18 | `providerconfig` temp file not cleaned up | `semantic-boundary` | Low — `.tmp` orphan on rename failure |
| 19 | `TestProviderCommand` doesn't test `provider list` secret leak | `test-gap` | Low — gap in existing test |

---

## Acceptance Checklist

- [x] Changes limited to declared scope (UI & Intent domain)
- [x] Behavior preserved or intentionally documented
- [x] Tests pass (`go test ./...`)
- [x] No new feature scope introduced
- [x] Relevant conventions referenced
- [x] Remaining gaps listed clearly
- [x] Evolution boundary check completed — no sandbox needed

---

*Report produced per [swe1-6-renormalization-guide.md](../../docs/architecture/swe1-6-renormalization-guide.md).*
