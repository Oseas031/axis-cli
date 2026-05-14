# CLI Execution Semantics Design

> 展开自 CLAUDE.md §1（No hidden daemons）+ §8（CLI Output Contracts）

**Status**: In Progress

## Command Classification

| Command | Mode | Requires Runtime | Behavior |
|---------|------|-----------------|----------|
| `axis run <id>` | In-process | No | Synchronous execute → print → exit |
| `axis shell` | In-process | No | REPL with shared orchestrator |
| `axis ask <prompt>` | In-process | No | Parse + preview (dry-run) |
| `axis ask --submit` | Runtime | Yes | Submit to persistent runtime |
| `axis status <id>` | Runtime | Yes | Query persistent runtime |
| `axis start` | Runtime | N/A (creates it) | Start persistent runtime |
| `axis provider *` | Local config | No | Read/write `.axis/providers.json` |
| `axis skills *` | Local files | No | Read `.axis/skills/` |
| `axis memory *` | Local files | No | Read/write `.axis/memory/` |
| `axis context *` | In-process | No | Preview context assembly |
| `axis judge` | In-process | No | Diagnostic (synthetic or real) |
| `axis evolve *` | Local files | No | Read `.axis/evolution/` |

## Architecture Change

### Current (broken)

```
axis run → initOrchestrator() → submitTask() → print "submitted" → exit
                                                  ↑ task never executes (no Start)
```

### Target

```
axis run → initOrchestrator() → orch.Start(ctx) → submitTask() → waitForCompletion() → print result → orch.Shutdown() → exit
```

## Key Design Decisions

### D1: `runTask()` becomes synchronous

```go
func runTask(cmd *cobra.Command, args []string) error {
    root := project.MustResolveRoot()
    initOrchestrator()

    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()

    if err := orch.Start(ctx); err != nil {
        return err
    }
    defer orch.Shutdown(context.Background())

    task := buildTask(args[0], inputFlag, promptFlag)
    if err := orch.SubmitTask(task); err != nil {
        return err
    }

    result, err := waitForCompletion(ctx, args[0])
    // print result, return appropriate exit code
}
```

### D2: `waitForCompletion()` polls orchestrator

Simple polling loop (not event-driven — v1 simplification):

```go
func waitForCompletion(ctx context.Context, taskID string) (types.TaskStatus, error) {
    ticker := time.NewTicker(100 * time.Millisecond)
    defer ticker.Stop()
    for {
        select {
        case <-ctx.Done():
            return "", ctx.Err()
        case <-ticker.C:
            status, err := orch.GetTaskStatus(taskID)
            if err != nil {
                return "", err
            }
            if status.IsTerminal() {
                return status, nil
            }
        }
    }
}
```

// v1: polling at 100ms. TODO: event-driven notification from orchestrator.

### D3: Project root unified via `App.root`

```go
func (app *App) initOrchestrator() {
    app.orchOnce.Do(func() {
        app.root = project.MustResolveRoot()  // single source of truth
        // ... rest of init
    })
}
```

All commands that need root access `defaultApp.root` after init.

### D4: `getTaskStatus()` checks runtime availability

```go
func getTaskStatus(cmd *cobra.Command, args []string) error {
    root := project.MustResolveRoot()
    locator := control.NewRuntimeLocator(root)
    if !locator.Exists() {
        return fmt.Errorf("no local runtime running\n  Start one with: axis start")
    }
    // ... existing HTTP client logic
}
```

### D5: `submitTask()` requires explicit input

Remove hardcoded `Input: map[string]any{"message": "test"}`. Replace with:

```go
func buildTask(taskID, inputJSON, prompt string) (*types.AgentTask, error) {
    if inputJSON == "" && prompt == "" {
        return nil, fmt.Errorf("task input required\n  Use: axis run %s --prompt \"...\" or --input '{...}'", taskID)
    }
    // parse inputJSON or use intent parser for prompt
}
```

## Output Contract

### Success
```
Task <id> completed (2.3s)
Output: <JSON or summary>
```

### Failure
```
Task <id> failed: <reason>
  Retries exhausted (3/3)
  Last error: <message>
```

### No runtime (for status/submit)
```
Error: no local runtime running
  Start one with: axis start
  Or use: axis run <id> --prompt "..." for one-shot execution
```

## Migration Safety

- `axis shell` unchanged (already calls `orch.Start`)
- `axis start` unchanged (already calls `runLocalRuntime` which starts orchestrator)
- Only `axis run` and `axis status` behavior changes
- Existing tests for shell remain valid
