# learn-claude-code 机制迁移方案

**日期**: 2026-05-12
**状态**: Done（Skills 系统已完整实现）
**参考**: `reports/analysis/learn-claude-code-agent-execution-layer-analysis-2026-05-12.md`
**来源**: [shareAI-lab/learn-claude-code](https://github.com/shareAI-lab/learn-claude-code)

---

## A0 Posture：本次工作的"姿态"

**姿态**: 长出新能力

**目标**: 将 learn-claude-code 的成熟 Harness 机制迁移到 Axis，增强 Agent 的执行能力层

**与 First Principles 的对齐**:
- ✅ More Action：更多工具、更多执行能力
- ✅ Interface is Existence：Skills 作为可发现、可加载的能力接口
- ✅ Contract is Structure：SKILL.md 作为结构化契约
- ✅ Capability is Decision Right：Agent 按需加载知识，自主决策

---

## A1 Externalize：核心想法

### 核心洞察

learn-claude-code 的核心设计哲学：

```
Agent = Model + Harness

Harness = Tools + Knowledge + Observation + Action Interfaces + Permissions
```

Agency（感知、推理、行动能力）来自模型训练，不是外部代码编排。Harness 是载具，Model 是驾驶者。

### 可迁移的 12 个机制

| 课程 | 机制 | 格言 | Axis 适用性 |
|------|------|------|-------------|
| s01 | Agent Loop | One loop & Bash is all you need | ✅ 已实现 |
| s02 | Tool Dispatch | 加一个工具, 只加一个 handler | ✅ 已实现 |
| s03 | TodoWrite | 没有计划的 agent 走哪算哪 | 🟡 可增强 |
| s04 | Subagent | 大任务拆小, 每个小任务干净的上下文 | 🔴 待设计 |
| s05 | Skills | 用到什么知识, 临时加载什么知识 | ✅ 已实现 |
| s06 | Context Compact | 上下文总会满, 要有办法腾地方 | 🟡 可增强 |
| s07 | Task System | 大目标要拆成小任务, 排好序, 记在磁盘上 | ✅ 已实现 |
| s08 | Background Tasks | 慢操作丢后台, agent 继续想下一步 | 🔴 待设计 |
| s09 | Agent Teams | 任务太大一个人干不完, 要能分给队友 | 🔴 P2 规划 |
| s10 | Team Protocols | 队友之间要有统一的沟通规矩 | 🔴 P2 规划 |
| s11 | Autonomous Agents | 队友自己看看板, 有活就认领 | 🔴 P2 规划 |
| s12 | Worktree Isolation | 各干各的目录, 互不干扰 | 🟡 与 Sandboxed Evolution 结合 |

### 迁移优先级

```
P0 (短期):
├── Skills 系统（SKILL.md 格式 + 两层加载）
├── 三层上下文压缩（micro/auto/manual）
└── TodoWrite 增强（nag 提醒）

P1 (中期):
├── Subagent 上下文隔离
├── Background Tasks 异步执行
└── EventBus 统一事件格式

P2 (长期):
├── 多 Agent 协作（JSONL 邮箱协议）
├── 自动认领机制
└── Worktree 深度集成
```

---

## A2 Inventory：已有机制

### Axis 已有的对应机制

| learn-claude-code | Axis 对应 | 状态 |
|-------------------|-----------|------|
| Agent Loop | Orchestrator + Dispatcher | ✅ 完成 |
| Tool Dispatch | Tool interface + ToolRegistry | ✅ 完成 |
| TodoWrite | AgentTask.status | 🟡 部分实现 |
| Task System | AgentTask + Scheduler | ✅ 完成 |
| Context Compact | contextpack | 🟡 部分实现 |
| Worktree | Sandboxed Evolution | 🟡 部分实现 |
| EventBus | .axis/events/tasks.jsonl | ✅ 完成 |

### Axis 的独特优势

| 维度 | learn-claude-code | Axis |
|------|-------------------|------|
| 语言 | Python | Go |
| 架构 | 单文件教学实现 | 模块化生产架构 |
| 权限模型 | 黑名单 | Contract + Permission ladder |
| 演化机制 | 无 | Sandboxed Evolution Protocol |
| 持久化 | JSON 文件 | 结构化 internal/memory |

---

## A3 Diagnose：差距诊断

### 当前缺陷

| # | 缺陷 | 严重性 | 影响 |
|---|------|--------|------|
| A | ~~无按需知识加载机制~~ | ~~高~~ | ~~Agent 无法动态获取领域知识~~ ✅ 已解决 |
| B | 上下文压缩策略单一 | 高 | 长会话上下文溢出 |
| C | 无 Subagent 隔离 | 高 | 复杂任务污染主上下文 |
| D | 无后台任务执行 | 中 | 长时间操作阻塞 Agent |
| E | 无多 Agent 协作原语 | 中 | 无法并行处理独立子任务 |
| F | TodoWrite 缺少 nag 机制 | 低 | Agent 可能遗忘任务跟踪 |

### 与 First Principles 的违背

| 违背 | 说明 |
|------|------|
| ~~Query is Context~~ | ~~无 Skills 系统，Agent 无法主动查询领域知识~~ ✅ 已解决 |
| Ladder is Boundary | 无 Subagent 隔离，无法为子任务分配更低的权限阶梯 |
| Layered Isolation is Collaboration | 无后台任务机制，无法实现异步协作 |

---

## A4 Realign：映射回 Four Principles + Six First Principles

### More Context

| 机制 | 贡献 |
|------|------|
| Skills | Agent 可主动查询领域知识，实现 "Query is Context" |
| Context Compact | 延长会话生命周期，保持更多历史上下文 |

### More Action

| 机制 | 贡献 |
|------|------|
| Subagent | 大任务可拆分为独立子任务并行执行 |
| Background Tasks | 长时间操作不阻塞 Agent 主循环 |
| TodoWrite + nag | 强制 Agent 跟踪任务进度 |

### Zero Control

| 机制 | 贡献 |
|------|------|
| Skills 两层加载 | 知识按需注入，不预设 Agent 行为路径 |
| Subagent 独立上下文 | 子任务自主决策，主 Agent 不干预 |

### Controllable Evolution

| 机制 | 贡献 |
|------|------|
| Worktree Isolation | 与 Sandboxed Evolution 深度集成 |
| EventBus | 所有状态变更可审计 |

---

## A5 Minimize：奥卡姆三问

### 三问

1. **最小可用形态是什么？** → Skills 系统
2. **今天能完成什么？** → Skills 完整实现（T2-T9）
3. **什么可以推迟？** → 多 Agent 协作 (P2)

### 今日最小一步

**目标**: 完成 Skills 系统从 Spec 到实现的全流程

**产出**:
- `docs/specs/skills-system/requirements.md` ✅
- `docs/specs/skills-system/design.md` ✅
- `docs/specs/skills-system/tasks.md` ✅
- `internal/skills/` 完整包实现 ✅
- `cmd/axis/skills_cmd.go` CLI 命令 ✅
- `internal/model/tool/skills_tool.go` 工具注册 ✅
- Layer 1 系统提示注入 ✅
- 边界强制测试 ✅

---

## A6 Execute & Verify：实施与验证

### 今日任务清单

- [x] 创建 Skills 系统 Spec-RDT 目录结构
- [x] 编写 requirements.md（需求规格）
- [x] 编写 design.md（设计规格）
- [x] 编写 tasks.md（任务清单）
- [x] T2: 包骨架 — types.go, errors.go, loader.go
- [x] T3: Skill 发现 — discover.go + parseFrontmatter + testdata fixtures
- [x] T4: Skill 加载 — Load() + safeSkillPath() 路径安全
- [x] T5: Skill 验证 — validate.go（目录/frontmatter/名称一致性）
- [x] T6: CLI 命令 — axis skills list/show/validate/create
- [x] T7: 工具注册 — load_skill tool + orchestrator 注册
- [x] T8: Layer 1 注入 — BuildSkillsPromptSection + ModelRequest.SystemPrompt
- [x] T9: 边界强制测试 — scheduler 隔离、opt-in、path safety、name format
- [x] 中英文文档同步（docs/zh/specs/skills-system/）

### 验证结果

```bash
# 所有 skills 相关包测试通过（含 -race）
go test -race ./internal/skills/... ./cmd/axis/... ./internal/model/tool/... \
  ./internal/model/provider/... ./internal/contract/executor/... \
  ./internal/kernel/orchestrator/...
# ok  github.com/axis-cli/axis/internal/skills
# ok  github.com/axis-cli/axis/cmd/axis
# ok  github.com/axis-cli/axis/internal/model/tool
# ok  github.com/axis-cli/axis/internal/model/provider
# ok  github.com/axis-cli/axis/internal/contract/executor
# ok  github.com/axis-cli/axis/internal/kernel/orchestrator
```

### Git 提交记录

| Commit | 内容 |
|--------|------|
| `b43acbb` | feat(skills): add internal/skills package skeleton with Discover and Load (T2-T4) |
| `7a99085` | feat(skills): complete T5-T9 — validate, CLI, tool registration, prompt injection, boundary tests |
| `f8cba11` | docs(zh): add Chinese design.md and tasks.md for skills-system |

---

## A7 Distill：复盘四栏

### 保留

- learn-claude-code 的 Harness = Model + 供给 哲学
- 两层 Skills 加载模式（metadata in system prompt, body on demand）
- dispatch map 工具分发模式
- safe_path 路径安全检查
- 三层上下文压缩策略

### 修正

- 将 Python 实现模式映射到 Go
- 与 Axis 现有 Contract + Permission ladder 集成
- 与 Sandboxed Evolution Protocol 对齐

### 剔除

- 教学用途的简化实现（如无并发安全的 TaskManager）
- 硬编码的黑名单安全检查（改用 Contract）
- 单进程多线程模型（改用 Axis 的进程模型）

### 沉淀

- Skills 系统将写入 `docs/specs/skills-system/`
- 三层压缩策略将集成到 `internal/contextpack/`
- Subagent 隔离将影响 `internal/agent/` 架构

---

## A8 Sublate：写回规范

### 待更新的文档

| 文档 | 更新内容 |
|------|----------|
| `CLAUDE.md` | 添加 Skills 系统的语义边界约束 |
| `docs/architecture/semantic-boundaries.md` | 添加 SkillLoader 的 "must NOT do" 列表 |
| `docs/architecture/module-and-naming-conventions.md` | 添加 `.axis/skills/` 目录规范 |

### 待创建的 BOUNDARY.md

| 目录 | 边界约束 |
|------|----------|
| `internal/skills/` | Never push skill content into provider prompts; never change scheduler semantics |

---

## 附录 A：Skills 系统设计摘要

### SKILL.md 格式

```yaml
---
name: skill-name
description: One-line description for discovery
tags: tag1, tag2
---

# Skill Title

Detailed instructions...

## Section 1
...

## Section 2
...
```

### 两层加载架构

```
┌─────────────────────────────────────────────────────────────┐
│ Layer 1: System Prompt (cheap, ~100 tokens/skill)          │
│   Skills available:                                          │
│     - pdf: Process PDF files...                             │
│     - code-review: Review code...                           │
└─────────────────────────────────────────────────────────────┘
                              │
                              │ Agent calls load_skill("pdf")
                              ▼
┌─────────────────────────────────────────────────────────────┐
│ Layer 2: Tool Result (on demand, full content)             │
│   <skill name="pdf">                                        │
│     Full PDF processing instructions...                     │
│   </skill>                                                  │
└─────────────────────────────────────────────────────────────┘
```

### 目录结构

```
.axis/
├── skills/
│   ├── pdf/
│   │   └── SKILL.md
│   ├── code-review/
│   │   └── SKILL.md
│   └── agent-builder/
│       ├── SKILL.md
│       ├── scripts/
│       │   └── validate.py
│       └── references/
│           └── template.yaml
```

---

## 附录 B：三层上下文压缩策略

### Layer 1: micro_compact（每轮自动）

```python
# 保留最近 3 个 tool_result，旧的使用占位符替换
# 保留 read_file 结果（避免重复读取）
if len(tool_results) > KEEP_RECENT:
    for result in tool_results[:-KEEP_RECENT]:
        if result.tool_name not in PRESERVE_RESULT_TOOLS:
            result.content = f"[Previous: used {result.tool_name}]"
```

### Layer 2: auto_compact（阈值触发）

```python
# 当 token 估计 > 阈值时
if estimate_tokens(messages) > THRESHOLD:
    # 1. 保存完整 transcript 到磁盘
    save_transcript(messages)
    # 2. 请求 LLM 摘要
    summary = llm.summarize(messages)
    # 3. 替换所有 messages 为摘要
    messages = [{"role": "user", "content": f"[Compressed] {summary}"}]
```

### Layer 3: compact tool（手动触发）

```python
# Agent 主动调用 compact 工具
def handle_compact():
    return auto_compact(messages)
```

---

## 附录 C：Subagent 隔离模式

### 架构

```
Parent Agent                    Subagent
+------------------+            +------------------+
| messages=[...]   |            | messages=[]      |  ← fresh context
|                  │  dispatch  |                  │
| tool: task       │ ─────────> │ while tool_use:  │
|   prompt="..."   │            │   call tools     │
|                  │  summary   │   append results │
|   result = "..." │ <───────── │ return last text │
+------------------+            +------------------+
          │
Parent context stays clean.
Subagent context is discarded.
```

### Axis 集成点

| 组件 | 集成方式 |
|------|----------|
| Orchestrator | 作为 Subagent 的父调度器 |
| Dispatcher | 为 Subagent 创建独立的执行上下文 |
| ToolRegistry | 为 Subagent 提供工具子集（无递归 spawning） |
| ContractExecutor | 为 Subagent 应用更严格的 Contract |

---

## 参考文档

- `reports/analysis/learn-claude-code-agent-execution-layer-analysis-2026-05-12.md`
- `EXTERNAL/learn-claude-code-main/README-zh.md`
- `EXTERNAL/learn-claude-code-main/agents/s_full.py`
- `CLAUDE.md`
- `docs/architecture/agent-native-first-principles.md`
- `docs/architecture/spec-lifecycle-conventions.md`
