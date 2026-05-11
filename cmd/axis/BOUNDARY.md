# CLI BOUNDARY — Edit This Directory Only If You Accept These Constraints

## What CLI Must NEVER Do

1. **Never introduce Web/TUI frameworks** — Axis is CLI-native, scriptable, composable. No React, Vue, gin, echo, fiber
2. **Never hide control plane** — `axis start` is explicit; no auto-daemon, no hidden background processes
3. **Never break output contract** — all CLI output must follow `docs/architecture/cli-output-conventions.md`
4. **Never leak secrets in output** — API keys, tokens must never appear in stdout/stderr or `--help`

## Before Modifying This Directory

- [ ] Read `docs/architecture/cli-output-conventions.md`
- [ ] Read `docs/architecture/metadata-key-conventions.md` if adding new flags
- [ ] Confirm: new command is scriptable (parsable output, stdin-friendly)
- [ ] Confirm: no secrets in output (grep for "api_key", "token", "secret" in proposed output)

## Common Traps

| Trap | Why It Is Wrong |
|---|---|
| Adding interactive prompt as default | Breaks pipe/script usage; use `--confirm` flag for interactive mode |
| Changing output format without updating tests | Breaks downstream scripts and CI parsers |
| Embedding API key in error message | Security leak; violates `docs/architecture/secret-handling.md` |
