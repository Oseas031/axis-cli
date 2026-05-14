# CLI Execution Semantics Tasks

**Status**: In Progress
**Implements**: requirements.md + design.md

---

## T1: Fix `axis run` — synchronous execution

- Modify `runTask()`: call `orch.Start(ctx)` + `waitForCompletion()` + `orch.Shutdown()`
- Add `waitForCompletion()` polling function (100ms tick, respect ctx timeout)
- Default timeout: 60s, overridable via `--timeout` flag
- Print result on completion, exit 1 on failure
- Regression test: `TestRunTask_ExecutesAndCompletes`

## T2: Fix `axis run` — require explicit input

- Add `--prompt` and `--input` flags to run command
- Remove hardcoded `Input: map[string]any{"message": "test"}` from `submitTask()`
- `--prompt`: parse via `intent.DeterministicParser` → extract input
- `--input`: parse as JSON map
- Neither provided → error with usage hint
- Regression test: `TestRunTask_RequiresInput`

## T3: Unify project root

- Replace all hardcoded `"."` in cmd/axis/ with `project.MustResolveRoot()`
- Affected: `getTaskStatus()`, `startOrchestrator()`, `App.resolveProvider()`
- Store resolved root in `App.root` field (set once in `initOrchestrator`)
- Regression test: verify commands work from subdirectory

## T4: Fix `axis status` — clear error when no runtime

- Check `locator.Exists()` before attempting HTTP request
- If no runtime.json: print "No local runtime running. Start one with: axis start"
- Exit code 1
- Regression test: `TestGetTaskStatus_NoRuntime`

## T5: Documentation update

- README CLI Commands table: add "Requires Runtime" column
- Each command's `--help` Long description states mode (in-process / runtime)

---

## Definition of Done

- `axis run test-task --prompt "hello"` executes synchronously, prints result, exits
- `axis run test-task` (no input) prints usage error
- `axis status X` without runtime prints clear error
- `axis shell` unchanged behavior
- `go test -race ./cmd/axis/...` passes
- `go build ./cmd/axis` passes
- No hardcoded `"."` in cmd/axis/*.go for project root
