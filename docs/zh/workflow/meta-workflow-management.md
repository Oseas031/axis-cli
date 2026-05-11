---
description: 管理工作流的工作流（Meta-Workflow）
---

# Meta-Workflow

本项目采用轻量级 workflow 机制。实际执行以本文件和 `workflow/entry.md` 为准。

## 权威入口

```text
workflow/entry.md
```

Agent 接到任务后，先读入口，再选择最小上游 workflow 组合。

## 当前有效规则

1. **文档先行**：新功能先有 `requirements.md`、`design.md`、`tasks.md`、`workflow-binding.md`。
2. **最小 workflow**：优先复用现有 workflow，不新增重型流程。
3. **门禁克制**：构建、测试、安全可阻塞；经验类检查只提醒。
4. **状态同步**：进行中状态写 `docs/status/current-progress.md`，交接状态写 `HANDOVER.md`。
5. **唯一归类**：复盘时每个工作项只归入一个唯一上游 workflow。
6. **设计主权**：用户已交接设计主权时，Agent 主动组织设计路线和文档落盘；只在破坏性或高风险操作前请求确认。

## 本体定位变化规则

当项目核心定位变化时：

1. 先更新核心设计哲学或架构报告。
2. 再成组检查入口文档：`README.md`、`docs/README.md`、`docs/guides/QUICKSTART.md`、`docs/product/WHITEPAPER.md`、`docs/status/current-progress.md`、`HANDOVER.md`。
3. 明确当前做什么、不做什么、后续 spec 是什么。
4. 不因宏大设计自动扩大当前 milestone scope。

## 当前不启用的长期设想

- Prometheus / Grafana 工作流监控
- 自动创建/部署 workflow
- 工作流性能测试平台
- 复杂回滚系统
- 多层版本自动发布机制
- 独立 workflow 调度器

## 并行开发隔离规则

Agent `isolation: "worktree"` 存在已知缺陷：创建的 worktree 基于默认分支 `main` HEAD，而非当前分支 HEAD。并行开发时使用手动 worktree：

```bash
# 创建（从当前 commit）
git worktree add -b <branch-name> .claude/worktrees/<name> <commit>

# 使用（主会话通过 EnterWorktree --path 进入）

# 回收
git worktree remove --force .claude/worktrees/<name>
git branch -D <branch-name>
```

同一分支不能同时在两个 worktree checkout，必须基于 commit SHA 创建新分支。

## 文档同步清单

进度更新必须同步以下 4 个文件：

| 文件 | 更新内容 |
|---|---|
| `CLAUDE.md` | Current Status + Architecture + Runtime Flow + 约束/缺陷 |
| `HANDOVER.md` | 交接状态 + 覆盖率 + 已知问题 + 下一步行动 |
| `AGENT_INSTRUCTIONS.md` | 当前状态摘要 |
| `docs/status/current-progress.md` | 已完成/进行中/待处理 + 最近提交 |

## 与设计哲学的关系

- **More Context**：workflow 提供上下文和路由。
- **More Action**：workflow 给出下一步可执行动作。
- **Zero Control**：workflow 不替 Agent 决策，不制造过度阻塞。
- **Bash is All You Need**：优先用 CLI、脚本、简单文档实现流程。
