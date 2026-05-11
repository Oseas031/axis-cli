# Natural Language Scheduling Requirements

## Summary

Add a non-destructive natural language scheduling layer to Axis. The feature converts user intent into structured Axis tasks and routes the result through the existing CLI, contract, orchestrator, and scheduler system.

This is an intent-to-task compiler, not a chatbot control plane. Natural language is an input interface, while `types.AgentTask`, contracts, metadata, and scheduler state remain the authoritative execution structure.

## Philosophy Alignment

- **Interface is existence**: natural language becomes one more task submission interface, equal to CLI commands and future Agent interfaces.
- **Query is context**: intent parsing should query only the context it needs; it must not push large implicit context bundles by default.
- **Contract is structure**: parsed intent must become structured task data before scheduling.
- **More Context, More Action, Zero Control**: the feature should make task creation easier while preserving observable task metadata and avoiding central control logic.
- **Bash is all you need**: the first interface must be CLI-native, scriptable, stdin/stdout friendly, and usable by humans or Agents.

## Users

- Developers who want to submit Axis tasks without remembering exact command syntax.
- Agents that need a lightweight text-to-task entrypoint.
- Operators who want dry-run visibility before submitting generated tasks.

## Functional Requirements

### FR1: CLI natural language entrypoint

Axis should provide a CLI entrypoint such as:

```bash
axis ask "check the current provider configuration"
```

The command must not replace existing `run`, `status`, or `shell` commands.

### FR2: Dry-run by default for generated structure

The first safe behavior should show the parsed task structure before execution unless the user explicitly submits it.

Example:

```bash
axis ask "check provider config" --dry-run
```

### FR3: Explicit submission

Users can submit a parsed task explicitly:

```bash
axis ask "check provider config" --submit
```

### FR4: Structured AgentTask output

The parser must produce a `types.AgentTask`-compatible structure with at least:

- task ID
- contract ID
- input map containing the original user goal
- metadata recording natural language provenance

### FR5: Provenance metadata

Submitted tasks must preserve:

- source: `natural_language`
- original prompt
- parser mode
- confidence or deterministic marker when available

### FR6: Shell integration as extension

`axis shell` may later support:

```text
axis> ask check provider config
```

This must reuse the same intent-to-task path as `axis ask`.

### FR7: No core scheduler changes for P0

The P0 implementation must submit ordinary tasks into the existing orchestrator/scheduler path. It must not change scheduler semantics, dispatcher semantics, or contract execution semantics.

## Non-Goals for P0

- No autonomous multi-step execution.
- No direct shell command execution from raw natural language.
- No hidden execution without task materialization.
- No Web UI or full TUI.
- No new permission/control center.
- No mandatory LLM dependency.
- No persistent database beyond existing Axis state paths.

## Acceptance Criteria

- `axis ask "..." --dry-run` prints the proposed task without submitting it.
- `axis ask "..." --submit` submits an ordinary `AgentTask`.
- The original prompt is preserved in task metadata.
- Existing `run`, `status`, `shell`, provider management, and scheduler behavior remain compatible.
- Tests cover dry-run, submit, metadata, and non-destructive defaults.
