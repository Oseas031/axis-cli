# Vigil Design

> 展开自 CLAUDE.md §7（Code & Architectural Style）+ §10（Engineering Practices）

**Status**: In Progress

## Data Model

```go
type Item struct {
    ID          string            `json:"id"`          // vigil-<短hash>
    Title       string            `json:"title"`
    Priority    string            `json:"priority"`    // P0/P1/P2
    Status      Status            `json:"status"`      // pending/in_progress/completed/stale
    Tags        []string          `json:"tags"`
    Origin      Origin            `json:"origin"`      // 溯源
    DependsOn   []string          `json:"depends_on"`  // 其他 item ID
    Notes       string            `json:"notes"`
    CreatedAt   time.Time         `json:"created_at"`
    StartedAt   *time.Time        `json:"started_at,omitempty"`
    CompletedAt *time.Time        `json:"completed_at,omitempty"`
    CommitHash  string            `json:"commit_hash,omitempty"`
    History     []StatusChange    `json:"history"`
}

type Status string
const (
    StatusPending    Status = "pending"
    StatusInProgress Status = "in_progress"
    StatusCompleted  Status = "completed"
    StatusStale      Status = "stale"
)

type Origin struct {
    Type string `json:"type"` // manual/judge/test_failure/followup/code_scan
    Ref  string `json:"ref"`  // 关联引用（commit hash、task id、test name）
}

type StatusChange struct {
    From      Status    `json:"from"`
    To        Status    `json:"to"`
    At        time.Time `json:"at"`
    Reason    string    `json:"reason,omitempty"`
}
```

## Storage

```
.axis/vigil/
  items.json          # active items (pending + in_progress + stale)
  archive/
    2026-05.json      # 按月归档的 completed items
```

- `items.json`：读写用 `os.ReadFile` + `os.WriteFile`（原子快照，Windows 安全）
- 归档：completed 超过 48h 的项在 triage 时移入 archive

## Package Layout

```
internal/vigil/
  types.go       # Item, Status, Origin, StatusChange
  store.go       # Store interface + JSONStore implementation
  store_test.go  # 持久化测试
  triage.go      # triage 逻辑（stale 标记、依赖升级、归档）
  triage_test.go
```

## CLI Commands

```
cmd/axis/vigil_cmd.go    # 所有 vigil 子命令
cmd/axis/vigil_cmd_test.go
```

## resume 输出格式

```
=== Vigil Resume ===

In Progress (1):
  [vigil-a3f] P0 fix axis run synchronous execution
    Started: 2h ago

Pending (3):
  [vigil-b7c] P1 Guarantee Registry
  [vigil-d2e] P1 Quality-Gated Model Escalation
  [vigil-f4a] P2 Skills composable metadata

Recently Completed (2, last 24h):
  [vigil-c1d] ✓ CLI execution semantics spec (3h ago)
  [vigil-e5b] ✓ Unify project root resolution (2h ago)

Suggested next: vigil-b7c (P1, no blockers)
```

## ID Generation

`vigil-` + 前 3 字节的 SHA256(title + created_at)，hex 编码 = 6 字符。
碰撞时追加一位。

## git hook 集成

`scripts/post-commit-vigil`（bash，可被 `.git/hooks/post-commit` 调用）：

```bash
#!/bin/bash
msg=$(git log -1 --format=%B)
if echo "$msg" | grep -qoP 'vigil:\K[a-z0-9-]+'; then
    ids=$(echo "$msg" | grep -oP 'vigil:\K[a-z0-9-]+')
    for id in $ids; do
        axis vigil done "$id" --commit "$(git rev-parse HEAD)" 2>/dev/null || true
    done
fi
```

约束：hook 失败不阻塞 commit（`|| true`）。

## triage 规则

| 条件 | 动作 |
|------|------|
| 被 ≥3 项 depends_on 引用 | priority 升为 P0 |
| status=pending 且 >7 天未变更 | status → stale |
| status=completed 且 >48h | 移入 archive |

## 与 CLAUDE.md 集成

§0 新增规则：
```
工作开始时执行 `axis vigil resume`，以此为工作起点。
commit message 中用 `vigil:<id>` 标记关联的工作项。
```

## 不做的事

- 不做 event log 双向同步（P2，等 event log 有结构化查询后再做）
- 不做 memory 集成（P2，等 vigil 稳定后再考虑 context signal）
- 不做 ingest 自动触发（P2，只预留接口）
- 不做 AgentTask 批量调度（P2）
