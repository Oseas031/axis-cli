# Occam's Razor Architecture Simplification Workflow

## Purpose

Prevent Axis from premature expansion under the temptation of grand designs, complex automation, or real Agent runtimes. Occam's Razor is not about weakening the vision, but about ensuring the vision unfolds in the correct order.

## Three Judgments

### 1. Is It Required Now?

Implement only the minimal capabilities confirmed for the current milestone.  
If an idea belongs to a subsequent bootstrap-loop / autogenesis-loop, write a spec or report first; do not mix it into the current implementation.

### 2. Is the Existing Lightweight Solution Sufficient?

Before adding the following complexities, you must explain why the existing lightweight solution is insufficient:

- UI: Why CLI / Shell is not enough
- Model Provider: Why MockProvider is not enough
- workflow: Why the existing routing in `workflow/entry.md` is not enough
- Automation: Why non-blocking reminders are not enough
- Dependencies: Why the standard library or existing dependencies are not enough
- Persistence / daemon: Why the current in-process semantics or file-backed state are not enough

When unable to explain, do not add by default.

### 3. Does It Break Scaffold-to-Self?

workflow, contract, permission rule, and spec are transitional structures.  
New rules must not disguise temporary scaffolding as permanent control.

## Design Philosophy Correction Rules

When an implementation is found to be inconsistent with **More Context, More Action, Zero Control, Controllable Evolution**, **bash is all you need, simple but robust, composable and extensible**, or **Competence earns autonomy, autonomy matches responsibility, evolution is controllable**:

1. First judge whether it can be corrected through error semantics, documentation, or tests.
2. If necessary, insert the minimal correction task and write it into the corresponding `tasks.md`.
3. Do not use correction tasks to introduce Web UI, complex TUI, external databases, daemons, or real LLM SDKs.
4. When new complexity is truly needed, create an independent spec first.

## Testability Design Constraints

When adding CLI commands or background functions:

1. Functions blocked on signals/global state should accept an injectable `context.Context` or shutdown channel.
2. Avoid calling `os.Exit()` directly inside functions; return errors to the caller for handling.
3. Avoid relying on non-portable APIs like `syscall.Kill`; Windows does not respond to programmatic signals.
4. If a path is not testable on the current platform, explicitly mark the reason in the test file; do not silently skip.

## Test Design Standards

Coverage improvement targets uncovered branch paths, not the number of tests:

```bash
go test -coverprofile=cov.out ./<package>/...
go tool cover -func=cov.out | grep -v "100.0%"
# Design test cases for functions with <100% coverage
# Priority: error handling > boundary conditions > concurrent paths > normal paths
```

## Notes

- No destructive edits.
- Maintain milestone boundaries.
- Move deprecated content to `docs/deprecated/`; do not erase history.
- Entry documents may update direction; implementation tasks must obey the current scope.
