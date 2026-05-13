# AXIS AGENT CONSTITUTION (v1.1 — 2026-05-12)

Project: **Axis** — Agent-native scheduling system; early execution substrate for Agent autogenesis.
Core proposition: **More Context, More Action, Zero Control, Controllable Evolution**.

> This file is the single source of truth for all coding constraints.
> Other files (`HANDOVER.md`, `AGENT_INSTRUCTIONS.md`, `workflow/entry.md`) reference this file; they do NOT redefine constraints.

## 0. 作者工作方法论（AI 必读）

作者使用 **辩证开发方法论（Dialectical Development Methodology）** 工作。

### 本体论层：Construct-Constraint-Judge

| 环节 | 含义 | AI 的角色 |
|------|------|-----------|
| **Construct**（对象化） | 意图转化为客观存在 | AI 执行生成 |
| **Constraint**（规定性） | 划定质的界限 | AI 在边界内工作 |
| **Judge**（扬弃） | 保留内核，否定偏差 | 人类判断，AI 辅助 |

### 操作层：SRS Loop（三阶段）

每个工作单元按三个辩证阶段推进：

**Phase I — Objectification**：外化意图（A0 Posture → A1 Externalize）
- 退出条件：想法已成为可被规定的客观文本

**Phase II — Determinateness**：确立边界（A2 Inventory → A3 Diagnose → A4 Realign → A5 Minimize）
- 退出条件：下一步行动的边界清晰到不可误解

**Phase III — Sublation**：执行+扬弃（A6 Execute → A7 Distill → A8 Sublate）
- 退出条件：新规则已写回系统

### AI 行为约束

1. **不跳过 Phase II 直接执行**。收到任务后先确认边界，再动手。
2. **Phase 转换时主动声明**。说明当前在哪个 Phase，退出条件是否满足。
3. **执行失败回退到 Phase II**（重选最小单元），方向错误回退到 Phase I（重新外化）。
4. **工作终点是规则更新（A8）**，不是代码合并。每次工作结束问：什么规则需要写回？
5. **每次工作开始时输出 Phase 声明**，格式固定为三行：
   ```
   Phase: <I/II/III> (<名称>)
   主要矛盾侧面: <Construct/Determinateness/Sublation> — <本次工作的核心张力>
   退出条件: <可验证的完成标准>
   ```
   未声明不动手。
6. **编码实现委派 subagent**。主上下文负责 Phase I/II 决策和 A8 写回，Phase III 的 A6 Execute 交给 subagent，避免实现细节污染决策上下文。
7. **Subagent 产出必须验收**。主上下文跑 `go test` + 抽查关键逻辑路径，不盲信。
8. **v1 简化显式标记**。简化处加 `// v1: <说明>. TODO: <改进方向>`，区分"故意简化"和"遗漏"。
9. **push 后自动监控 CI**。git push 后派 subagent 轮询 `gh run watch`，CI 失败则自动修复并重新推送，直到通过。主上下文不需要介入。

### 参考文档

- **本体论层**：`docs/architecture/dialectical-development-methodology.md`
- **操作层**：`docs/guides/YOUR_IMPLICIT_METHODOLOGY.md`
- **AI 精简版**：`docs/guides/SRS-LOOP-AI-REFERENCE.md`

---

## 1. Absolute Prohibitions

1. **No Web/TUI frameworks** in core or CLI (no React, Vue, gin, echo, fiber)
2. **No hidden daemons or auto-spawn** (explicit `axis start` only)
3. **No scheduler/dispatcher/contract semantic changes** without updated Spec-RDT
4. **No new Agent autonomy** without sandboxed-evolution protocol (isolated workspace -> verification -> promote/discard). Promotion is gated by machine-checkable verification criteria, NOT by promoter identity: an Agent that has passed all defined criteria MAY promote its own change. Humans are auditors of last resort, not a required gate.
5. **No push-based context injection** into provider prompts (contextpack is preview-only, opt-in, non-invasive)
6. **No unnamespaced metadata keys** (prefix with `context.*`, `tool.*`, `sla.*`, `evolution.*`, `intent.*`, `provider.*`, or `axis.*`)
7. **No secrets in output** — never log/output API keys, bearer tokens, private keys, credentials

## 2. Pre-Coding Checklist

Before proposing any change, confirm ALL of the following:

- [ ] Read `AGENT_INSTRUCTIONS.md` (handover entry) if this is your first task
- [ ] New behavior is observable via CLI / filesystem / event log (not hidden)
- [ ] CLI output follows `docs/architecture/cli-output-conventions.md`
- [ ] Metadata keys use namespaced prefix
- [ ] Tests include boundary/safety assertions (not just functional correctness)
- [ ] If touching `internal/kernel/`, `cmd/axis/`, `internal/contextpack/`, `internal/agent/`, or `internal/memory/`, read the adjacent `BOUNDARY.md`
- [ ] Cross-platform safety: Windows behavior considered (paths, signals, stdin/stdout, process mgmt)
- [ ] Defensive validation: nil checks, empty values, boundary conditions at public function entry points
- [ ] Context propagation: I/O and cancellable operations accept `ctx context.Context` as first parameter
- [ ] Concurrency hygiene: every goroutine has a defined exit path; no busy-poll; no double-close channels
- [ ] Bug fix includes a regression test that fails before the fix
- [ ] Docs synchronized if change affects milestones, specs, or conventions

> If a change feels quick but violates any item above, it is the wrong change. Write the Spec-RDT first.

## 3. Mandatory Reading

| Priority | File | Why |
|----------|------|-----|
| P0 | `docs/architecture/agent-native-first-principles.md` | Six first principles. Violating any = architectural breach. |
| P1 | `docs/architecture/semantic-boundaries.md` | What each kernel module must NOT do. |
| P2 | `docs/architecture/spec-lifecycle-conventions.md` | When to write/update a Spec-RDT. |

## 4. Directory Boundaries

Read the neighboring `BOUNDARY.md` before editing files in these directories:

| Directory | Boundary |
|-----------|----------|
| `internal/kernel/` | Scheduler must NOT call provider directly; must NOT inject context assembly |
| `cmd/axis/` | No Web/TUI; no hidden daemons; scriptable output; no secret leaks |
| `internal/contextpack/` | Never push context into provider prompts; never change scheduler semantics |
| `internal/agent/` | Never bypass contract layer; never inject context metadata into provider input |
| `internal/memory/` | Never push into provider prompts; never physical-delete; no external deps; no background tasks; LF-only line terminators |

## 5. Spec-First Protocol

Axis is spec-first. Specs are implementation contracts, not decorative notes.

- Non-trivial features or structural changes: check if a Spec-RDT is needed BEFORE coding
- Required shape: `docs/specs/<feature>/requirements.md`, `design.md`, `tasks.md`
- Structure-modifying changes = evolution work, not ordinary edits
- Use Sandboxed Evolution Protocol for: permission semantics, contract changes, workflow changes, context rules, autonomy surfaces
- **Promotion gate is verification quality, not promoter identity.** A change is promoted when (a) all declared verification criteria pass, (b) a `spec.promoted` (or analogous) event is appended to the event log referencing the verification artifacts and source digest, and (c) the spec status is updated atomically. The promoter may be human or Agent. Verification criteria MUST be machine-checkable and reproducible from the recorded workspace digest; criteria that depend on subjective human judgement are not valid criteria and must be reframed.

## 6. Semantic Boundaries

Each module has a strict "must NOT do" list:

- **AgentTask**: must NOT own execution logic, provider selection, hidden permissions
- **AgentContract**: must NOT own scheduler policy, provider credentials, context retrieval
- **Scheduler**: must NOT make model calls, shell execution, provider config, NL parsing
- **Orchestrator**: must NOT store provider profiles, render CLI, make hidden policy decisions
- **Dispatcher**: must NOT set admission policy, manage provider credentials
- **Provider**: must NOT manage task lifecycle, scheduler state, credential persistence
- **Tool**: must NOT manage global permissions, task scheduling
- **Intent Parser**: must NOT execute directly, assemble context, escalate permissions
- **ContextBundle**: must NOT escalate authority, change scheduler, submit tasks
- **EvolutionRun**: must NOT implicitly mutate main tree, hide execution policy, auto-escalate authority

## 7. Code & Architectural Style

- Language: Go 1.26; CLI framework: spf13/cobra only
- No external dependencies unless absolutely justified
- Metadata format: `namespace.key` (e.g., `context.bundle_id`, `tool.allowed_paths`)
- All state mutations leave observable traces (event logs, metadata)
- Natural language produces structure; it does NOT execute by itself
- Context improves action quality; it does NOT control or authorize action

## 8. CLI Output Contracts

- Human-readable by default; `--json` for machine mode (stable snake_case fields)
- Success: what happened + primary ID + suggested next command
- Error: failed action + object ID + concise cause + next step
- Preview commands clearly state they did NOT mutate state
- No color-only meaning; keep output line-oriented; keep existing human output stable

## 9. Build & Test

```bash
go build -o axis-dev.exe ./cmd/axis   # Windows dev binary
go test -race ./...                    # must pass before any commit
gofmt -w . && go vet ./...            # formatting + vet
staticcheck ./... && gosec ./...       # static analysis + security
```

### CI Rules
- 集成测试（启动真实进程）加 `testing.Short()` skip，CI 用 `-short`。
- `ci.yml` 必须在自身的 `paths` 触发列表中。
- **Traceability**: every commit message MUST reference a Spec-RDT ID (e.g. `M6 T13`, `M5 Phase 5.4`) or an explicit milestone/scope tag. Pure "fix typo" / "wip" commits are not allowed on `main`.
- **No build artifacts**: never stage `axis-dev.exe`, `*.exe`, `*.test`, `coverage.out`, `dist/`, `.cache/`, editor scratch files, or any generated binary. `.gitignore` is the first line of defense; the author is the second. If `git status` shows such a file before commit, fix `.gitignore` rather than `git add`-ing selectively.
- **Bisect-safe**: every commit MUST independently compile (`go build ./...`) and pass `go vet ./...`. Tests should pass too; if a commit is intentionally test-red (e.g. failing regression test before fix), the message MUST start with `wip(red):` and the next commit MUST turn it green. Never split a build-breaking change across two commits on `main`.

## 10. Engineering Practices

### Cross-Platform Safety (Windows is first-class)
- Use `path/filepath` for paths; never hardcode `/` or `\\`
- Process termination: use `os.Process.Kill()` or platform abstraction, not `syscall.Kill`
- Signal handling: graceful shutdown that degrades on Windows
- Shell scripts/hooks: portable across WSL / Git Bash / PowerShell / macOS / Linux
- File sharing: when a file may be held by another process for writing (e.g. event logs, JSONL), use `os.ReadFile` (snapshot read) not `os.Open` + streaming scanner — Windows default share mode blocks concurrent readers
- Batch scripts: use `start "" /MIN` for independent background processes, not `start /B` (which dies with the parent console)

### Defensive Programming
- Every public function validates inputs at entry: nil, empty, negative, special chars, boundary
- Never return `(nil, nil)` — use sentinel errors or typed zero values
- Never use `strings.HasPrefix` for security-critical path checks — use `filepath.Clean` + segment matching
- Never silently swallow errors

### Concurrency Safety
- Every goroutine has a defined exit path: `context.Cancel`, done channel, `sync.WaitGroup`
- Never close a channel from multiple goroutines
- Scheduler and orchestrator must include crash-recovery for orphaned `Running` tasks

### Context Propagation
- `context.Context` as first parameter for all I/O, long-running, or cancellable operations
- Never use `context.Background()` as convenience default in business logic
- Graceful shutdown = context-cancellation cascade, not `os.Exit`

### Testing
- Every bug fix includes a regression test
- Design tests around risk paths, not coverage targets
- Destructive tests mandatory: bad JSON, network timeout, 404/500, malformed schema, empty body
- Never make real external network calls in tests — use `httptest` or mock fixtures
- Provider contract tests: URL construction, request schema, response deserialization for every provider

### Config Externalization
- Never hardcode timeouts, ports, buffer sizes, retry counts in business logic
- Extract tunables to constants, config structs, or CLI flags with sensible defaults
- Scripts must not modify user-global config; validate environment before destructive actions

## 11. Evolution Principles

- Stable surfaces, replaceable internals
- Safety defaults: dry-run, preview, redaction, validation, explicit submit
- Auditable by design: every important decision leaves a trace
- Small contracts over large control planes
- Progressive evolution: deterministic/local first, adaptive later
- **Audit, not approval, is the trust mechanism.** Agent-promoted changes leave the same audit trail as human-promoted ones (verification artifacts, source digest, event log entry). Humans intervene by reading the trail and exercising revert/quarantine after the fact, not by gatekeeping promotion before it.

## 12. Naming & Structure

- Module layout follows `docs/architecture/module-and-naming-conventions.md`
- Spec statuses: Draft -> Planned -> In Progress -> Completed | Paused | Deprecated | Cancelled
- A task is Completed only when: code done, tests pass, docs synchronized, user-visible behavior described
- Metadata promotion rule: move to typed field when multiple core modules require it, validation depends on it, tests need stable access, or CLI/API consumers rely on it
