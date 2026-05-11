---
name: safe-edit
description: Playbook for safely editing structured files (Markdown / YAML / JSON / code) on Windows + CRLF projects. Invoke before precise text replacements, especially when an `edit` call has just failed with "String not found", when working on files with CRLF line endings or UTF-8 BOM, or when an edit touches fenced code blocks / structural boundaries.
triggers:
  - user
  - model
---

# Safe Edit Playbook (Windows / CRLF / structured files)

Use this before doing precise text replacements in files like `CLAUDE.md`, `AGENTS.md`, `docs/**/*.md`, `*.yaml`, `*.json`, or source files — especially on Windows projects where CRLF is the norm.

## Why this exists

Three failure modes show up repeatedly in this project:

1. **Encoding mismatch**: `read` returns content with LF, but the file on disk is CRLF (+ optional BOM). `edit` does byte-literal matching, so a copy-pasted `old_string` from `read` output silently fails to match.
2. **Local match, global breakage**: a `old_string` is unique, but the insertion point sits inside a structural boundary (Markdown fence, JSON bracket, indented block). The edit succeeds but corrupts the structure.
3. **Poisoned shell**: one bash command with mis-escaped backticks / `$` leaves the persistent shell session in a state where subsequent commands return weird exit codes with no output.

## Five rules (ordered by importance)

### R1. Probe before edit
Before any precise edit on an unfamiliar file, run once:
```bash
file <path>
head -c 200 <path> | od -c | tail
```
Confirms encoding (UTF-8 / UTF-8 BOM / UTF-16) and line endings (LF / CRLF). Costs <1s, eliminates the most common failure.

### R2. Read the neighborhood, not just the line
Before precise replacement, `read` ±5–10 lines around the target. Identify the **enclosing structural boundary**: fenced code block, JSON object, indented YAML block, function body. "Unique match" only solves *where to change*, not *what structure encloses the change*.

### R3. Prefer atomic writes over incremental fixes
If the edit spans multiple lines, crosses a code-fence, or modifies a structural boundary, write the full corrected block in one `edit` call (including closing delimiters). Incremental "insert now, fix later" turns one bug into nested bugs.

### R4. Anchor on single lines for line-ending-fragile tools
When the `edit` tool does byte-literal matching:
- Single-line `old_string` (no `\n`) → safe; CRLF/LF differences live only at line boundaries.
- Multi-line `old_string` → downgrade to a byte-level script:
  ```bash
  py -c "b=open('F','rb').read(); old=b'...\r\n...'; new=b'...\r\n...'; assert old in b; open('F','wb').write(b.replace(old,new))"
  ```
  Always print `assert` / `old in b` as a self-check before writing.

### R5. Burn the poisoned shell
If a shell starts returning `Exit code: N` with empty output after a mis-escaped command (especially anything with literal `` ` `` or `$` in `bash -c "..."`):
- **Do not** try to diagnose it. Open a fresh shell with the next `exec` call.
- Reopening costs one tool call; debugging a corrupted session costs many.

## Pre-edit checklist

```
[ ] 1. file <path>  → encoding + line ending known
[ ] 2. read ±10 lines around target → structural boundary identified
[ ] 3. choose anchor:
        - single-line unique → edit tool, single-line old_string
        - multi-line / crosses fence → py byte-level replace
[ ] 4. self-check: "found: True" / "assert old in b" before write
[ ] 5. read the edited region + neighbors → verify no broken structure
[ ] 6. if shell behaves weirdly → open a new shell, don't debug
```

## Anti-patterns (stop on sight)

- **"Let me just try pasting it in"** — there is no "try" in structured files; it either matches structurally or breaks it.
- **`bash -c "...containing literal backticks or unquoted $..."`** — use a heredoc, a temp script file, or `\x60` / `\$` escaping. In `py -c`, build backtick bytes as `b'\x60\x60\x60'`.
- **Reusing a shell that just gave a weird exit code** — that session is dead; move on.
- **`replace_all` to fix a multi-anchor mistake** — prefer multiple precise edits; never amplify a wrong pattern across the file.

## One-line summary

> Verify environment, read the neighborhood, pick a line-safe anchor; if the edit breaks, open a new shell and restart — never patch a half-broken edit in place.
