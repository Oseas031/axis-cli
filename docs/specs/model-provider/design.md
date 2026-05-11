# Model Provider Design

## Overview

The model provider layer gives Axis a replaceable execution abstraction between task dispatching and model-like output generation.

The first provider is `MockModelProvider`. It does not call a real model. It returns deterministic output that makes the execution chain visible and testable.

Implementation must follow [workflow-binding.md](workflow-binding.md): Meta-Workflow for documentation/code synchronization, Occam's Razor for scope control, and CI/PR/Document Audit workflows for validation.

## Architecture

```text
axis shell / axis run
  -> Orchestrator.SubmitTask
  -> Scheduler
  -> Dispatcher
  -> ContractExecutor
  -> ModelProvider
      -> MockModelProvider
  -> TaskResult
```

## Components

### ModelProvider interface

New package:

```text
internal/model/provider
```

Interface shape:

```go
type ModelProvider interface {
    Generate(ctx context.Context, request *ModelRequest) (*ModelResponse, error)
}
```

### ModelRequest

Carries enough context for model-like execution:

```go
type ModelRequest struct {
    TaskID     string
    ContractID string
    Input      map[string]any
}
```

### ModelResponse

Returns model-like output:

```go
type ModelResponse struct {
    Output map[string]any
}
```

### MockModelProvider

Behavior:

- reads `message` from input when present
- returns deterministic output
- includes provider name and task ID in output

Example output:

```go
map[string]any{
    "status": "completed",
    "message": "mock model executed task demo-task with input: test",
    "provider": "mock",
}
```

### ContractExecutor integration

Current `ContractExecutor.Execute` already represents the closest execution boundary. It should receive a provider and use it after input validation.

Minimal change:

```text
ContractExecutorImpl
  - contracts map
  - provider ModelProvider
```

Execution path:

1. Validate input
2. Call provider.Generate
3. Validate provider output
4. Return ExecutionResult

### Dispatcher integration

Dispatcher already calls contract validation and returns placeholder output. It should call `contractExecutor.Execute` instead of only `ValidateInput`.

This keeps model execution behind the contract executor and avoids pushing provider details into the dispatcher.

## File Structure

```text
internal/model/provider/provider.go       # interface and request/response types
internal/model/provider/mock.go           # MockModelProvider
internal/model/provider/mock_test.go      # provider tests
internal/contract/executor/executor.go    # provider-backed execution
internal/contract/executor/executor_test.go
internal/kernel/dispatcher/dispatcher.go  # call Execute instead of ValidateInput
cmd/axis/main.go                          # default executor uses mock provider through orchestrator construction
```

## Trade-offs

| Option | Decision | Rationale |
|---|---|---|
| Real provider first | Rejected | Requires keys/network and increases complexity |
| Mock provider first | Chosen | Runs locally and validates architecture |
| Provider inside Dispatcher | Rejected | Leaks model concerns into dispatch routing |
| Provider inside ContractExecutor | Chosen | Execution already belongs behind contract validation |
| Environment config now | Rejected | Not needed until real providers exist |

## Risks

| Risk | Mitigation |
|---|---|
| Mock output may look like real model output | Include `provider: mock` in output and docs |
| Contract output schema may reject provider fields | Keep required fields `status` and `message`; extra fields are allowed by current validation |
| Existing tests assume placeholder output | Update tests to assert provider-backed deterministic output |
| Over-expanding provider abstraction | Keep only `Generate` for now |

## Acceptance Mapping

- FR1: `internal/model/provider` interface
- FR2: `MockModelProvider`
- FR3: `ContractExecutor.Execute` uses mock provider by default construction path
- FR4: `axis shell` continues using `run <task-id>`
- FR5: no external SDK/config/API key
- FR6: provider and executor tests

## Provider Management Extension Design

### File layout

```text
.axis/
  providers.json
  backups/
    providers-<timestamp>.json
```

### Layers

```text
cmd/axis provider ...
  -> internal/model/providerconfig
  -> internal/model/provider
```

### Switching flow

1. Load `.axis/providers.json`.
2. Validate JSON and profiles.
3. Check selected profile exists and is not archived.
4. Backup the current config.
5. Set `active_profile`.
6. Save and reload through normal startup resolution.

The first implementation uses ordinary local file IO only. It does not add environment variable mutation, a daemon, a registry integration, or a global user configuration.
