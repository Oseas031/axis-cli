# Bash is All You Need

**[Chinese version / 中文版](../zh/architecture/bash-is-all-you-need.md)**

## Summary

Axis's interaction design follows **"bash is all you need, simple but robust, composable and extensible"**: prioritizing shell-native, composable, scriptable, reliable, and extensible capabilities rather than building heavy Web UIs or complex TUIs.

This philosophy does not reject UI, but emphasizes that Axis's default interaction surface should be the command line and standard I/O, because they are best suited for Agent-native scheduling systems.

## Relationship to Core Design Philosophy

`bash is all you need, simple but robust, composable and extensible` is the concrete realization of **More Context, More Action, Zero Control, Controllable Evolution** at the interaction layer.

### More Context

CLI output should provide sufficient context:

- Current operation target
- Execution result
- Error reason
- Actionable next-step suggestions

### More Action

Shell interfaces should let Agents and users take direct action:

- Submit tasks
- Query status
- Compose commands
- Extend capabilities through pipes and scripts

### Zero Control

Shell interfaces should not force users into a fixed workflow:

- Commands should be independently executable
- Errors should provide context rather than terminating the entire session
- Interactive shell should provide guidance without making decisions for the user

### Controllable Evolution

Shell interfaces should make high-risk actions, permission changes, and self-modification processes observable, confirmable, and rollback-safe:

- High-risk actions should trigger confirmation
- Command results should be kept in auditable records
- Extension interfaces should maintain backward compatibility

## Design Principles

### 1. CLI First

Axis's primary interaction surface is CLI:

```bash
axis run task-1
axis status task-1
axis shell
```

### 2. Shell Native

Commands should be suitable for use in bash, PowerShell, CI, scripts, and Agent tool invocations.

### 3. Composable and Extensible

Output and command design should facilitate composition:

```bash
axis status task-1
axis run task-2
```

JSON output can be added later, but is not mandatory for Milestone 1.

### 4. Simple but Robust

Keep interactions minimal, but add necessary fault tolerance, confirmation, rollback, and observability capabilities to reduce operational errors.

### 5. Minimal UI

No default Web UI or complex TUI. Only consider heavier interfaces when CLI cannot express necessary context.

### 6. Interactive When Useful

Interactive shell is an enhancement layer on top of CLI, not a replacement:

```text
axis> help
axis> run task-1
axis> status task-1
axis> exit
```

## Non-Goals

- Do not make Axis a Web-first product
- Do not turn the interaction layer into core architecture
- Do not sacrifice scriptability for UI effects
- Do not introduce complex UI frameworks in Milestone 1

## Implementation Priority

When adding new interaction capabilities, the priority order is:

1. Standard CLI commands
2. Interactive Shell
3. TUI
4. Web UI

Only move to the next layer when the current one cannot meet requirements.

## Conclusion

Axis's default interaction form should be:

> CLI as the primitive, shell as the interface, workflows as composition.

That is:

> **bash is all you need, simple but robust, composable and extensible**.
