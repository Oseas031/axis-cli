# 已完成工作：2026-05-12

## 1. Skills 系统完整实现

**状态**: Done
**参考**: `reports/analysis/learn-claude-code-agent-execution-layer-analysis-2026-05-12.md`

### 核心设计

将 learn-claude-code 的 Harness 机制迁移到 Axis，实现按需知识加载：

- **SKILL.md 格式**：YAML frontmatter（name/description/tags）+ Markdown body
- **两层加载架构**：
  - Layer 1: System Prompt 注入 metadata（~100 tokens/skill），Agent 可发现
  - Layer 2: Agent 调用 `load_skill("name")` 按需加载完整内容
- **目录结构**：`.axis/skills/<skill-name>/SKILL.md`

### 完成的任务

| Task | 内容 | 产出 |
|------|------|------|
| T2 | 包骨架 | `internal/skills/` — types.go, errors.go, loader.go |
| T3 | Skill 发现 | discover.go + parseFrontmatter + testdata |
| T4 | Skill 加载 | Load() + safeSkillPath() 路径安全 |
| T5 | Skill 验证 | validate.go（目录/frontmatter/名称一致性） |
| T6 | CLI 命令 | `axis skills list/show/validate/create` |
| T7 | 工具注册 | load_skill tool + orchestrator 注册 |
| T8 | Layer 1 注入 | BuildSkillsPromptSection + ModelRequest.SystemPrompt |
| T9 | 边界强制测试 | scheduler 隔离、opt-in、path safety、name format |

### Spec 文档

- `docs/specs/skills-system/requirements.md` ✅
- `docs/specs/skills-system/design.md` ✅
- `docs/specs/skills-system/tasks.md` ✅
- `docs/zh/specs/skills-system/` 中文同步 ✅

### Git 提交

| Commit | 内容 |
|--------|------|
| `b43acbb` | feat(skills): add internal/skills package skeleton with Discover and Load (T2-T4) |
| `7a99085` | feat(skills): complete T5-T9 — validate, CLI, tool registration, prompt injection, boundary tests |
| `f8cba11` | docs(zh): add Chinese design.md and tasks.md for skills-system |

### 验证

```bash
go test -race ./internal/skills/... ./cmd/axis/... ./internal/model/tool/... \
  ./internal/model/provider/... ./internal/contract/executor/... \
  ./internal/kernel/orchestrator/...
# 全部通过
```

---

## 2. 记忆系统设计方案

**状态**: 设计完成（未实现）

### 核心思路

短视野 + 长视野 → 被动注入 + 主动检索

| 维度 | 被动注入（系统主动） | 主动检索（Agent 主动） |
|------|----------------------|------------------------|
| 核心职责 | 覆盖 90% 高频通用信息 | 覆盖 10% 低频高价值信息 |
| 触发时机 | Agent 思考之前预加载 | Agent 思考过程中按需调用 |
| 信息范围 | 最近 N 条历史、核心文档、全局规则 | 任意历史、特定文件、精确关键词 |
| 优先级 | 低优先级，上下文末尾 | 高优先级，上下文开头 |
| 可审计性 | 每次注入写入 history | 每次检索写入 history |

---

## 3. learn-claude-code 机制分析

### 核心哲学

```
Agent = Model + Harness
Harness = Tools + Knowledge + Observation + Action Interfaces + Permissions
```

### 12 个机制评估

| 机制 | Axis 状态 |
|------|-----------|
| Agent Loop | ✅ Orchestrator + Dispatcher |
| Tool Dispatch | ✅ Tool interface + ToolRegistry |
| Skills | ✅ 本次实现 |
| Task System | ✅ AgentTask + Scheduler |
| EventBus | ✅ .axis/events/tasks.jsonl |
| TodoWrite | 🟡 部分实现 |
| Context Compact | 🟡 部分实现 |
| Worktree Isolation | 🟡 与 Sandboxed Evolution 结合 |
| Subagent | 🔴 待设计 |
| Background Tasks | 🔴 待设计 |
| Agent Teams | 🔴 P2 |
| Team Protocols | 🔴 P2 |

---

## 复盘

### 保留

- Harness = Model + 供给 哲学
- 两层 Skills 加载模式
- dispatch map 工具分发
- safe_path 路径安全检查

### 修正

- Python → Go 实现映射
- 与 Contract + Permission ladder 集成
- 与 Sandboxed Evolution Protocol 对齐

### 剔除

- 教学用简化实现（无并发安全的 TaskManager）
- 硬编码黑名单（改用 Contract）
- 单进程多线程模型（改用 Axis 进程模型）
