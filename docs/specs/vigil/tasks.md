# Vigil Tasks

**Status**: In Progress
**Implements**: requirements.md + design.md

---

## T1: Core data layer

- `internal/vigil/types.go` — Item, Status, Origin, StatusChange, ID generation
- `internal/vigil/store.go` — Store interface + JSONStore (Load/Save/Add/Update/Get/List/Archive)
- `internal/vigil/store_test.go` — CRUD + persistence across restarts + concurrent safety
- `internal/vigil/triage.go` — TriageItems(items) → mutations (stale/upgrade/archive)
- `internal/vigil/triage_test.go` — 7d stale, dependency upgrade, 48h archive

## T2: CLI commands

- `cmd/axis/vigil_cmd.go`:
  - `axis vigil resume` — in_progress + recent completed + top pending + suggestion
  - `axis vigil list [--priority] [--tag] [--status] [--json]`
  - `axis vigil add "title" [--priority] [--tag] [--origin] [--depends-on] [--notes]`
  - `axis vigil start <id>`
  - `axis vigil done <id> [--commit]`
  - `axis vigil show <id>`
  - `axis vigil triage`
- `cmd/axis/vigil_cmd_test.go` — 每个子命令至少一个 happy path + 一个 error path

## T3: git hook

- `scripts/post-commit-vigil` — bash script parsing `vigil:<id>` from commit message
- 文档：README 或 CLAUDE.md 说明如何安装 hook

## T4: CLAUDE.md integration

- §0 新增规则：工作开始 `axis vigil resume`，commit 用 `vigil:<id>` 标记
- §4 Directory Boundaries 添加 `internal/vigil/` 条目

## T5: Migration

- `axis vigil import` — 从 `WORKFLOW-HUMAN/pending-*.md` 解析并导入现有待办
- 导入后 pending markdown 保留为只读参考（不删除）

---

## Definition of Done

- `axis vigil resume` 在空状态输出引导信息
- `axis vigil add` + `start` + `done` 完整生命周期
- `axis vigil triage` 正确标记 stale 和归档
- git hook 解析 `vigil:<id>` 自动完成
- `go test -race ./internal/vigil/... ./cmd/axis/...` 通过
- `go build ./cmd/axis` 通过
- CLAUDE.md 已更新
