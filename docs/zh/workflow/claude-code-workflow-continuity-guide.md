# Claude Code 工作流衔接指南

本指南用于 Claude Code / 外部 Agent 接手 Axis 项目时快速恢复上下文。当前原则是轻量、可追溯、不过度自动化。

## 接手顺序

按 `AGENT_INSTRUCTIONS.md`（唯一入口）执行：

1. 读 `CLAUDE.md` — 宪法（所有约束的唯一来源）。
2. 读 `docs/status/current-progress.md` — 里程碑状态的 ground truth。
3. 读 `docs/architecture/agent-native-first-principles.md` — 设计原则。
4. 读 `workflow/entry.md` — 将任务路由到最小 workflow 组合。
5. 读 `HANDOVER.md` — 参考：项目结构、已完成工作、已知问题。
6. 如任务涉及 CI/CD，读 `.github/config/registry.yml` 和对应 `.github/workflows/*.yml`。

## 交接同步

任务完成或阶段切换时，按影响范围同步：

| 文件 | 何时更新 |
|---|---|
| `docs/status/current-progress.md` | 当前进度、阶段状态、最近验证结果变化 |
| `HANDOVER.md` | 已知问题、下一步行动、项目结构变化 |
| `docs/README.md` | 文档入口或目录结构变化 |
| `workflows/README.md` | workflow 状态、路径、路由变化 |

## 工作流维护规则

1. 不为单个任务新增 workflow。
2. 不把建议性经验检查升级为硬门禁。
3. GitHub Actions 的 active 状态必须对应实际存在的 `.yml` 文件。
4. 已合并、删除、乱码、路径失效的流程必须移出活跃索引。
5. 自动化脚本只在收益明确且可验证时引入。
6. 长期设想放入报告或 deprecated，不进入当前执行路径。

## 验证清单

工作流相关改动完成后至少检查：

```text
1. workflows/README.md 与 .github/config/registry.yml 状态一致
2. workflow/entry.md 不引用 deprecated workflow
3. Markdown 相对链接无断链
4. active GitHub Actions 文件真实存在
5. 文档路径符合 docs/README.md 的分类结构
```

## 当前废弃项

- `wf-dev`：本地开发检查不再作为独立 GitHub Actions。
- `wf-release`：发布链路合并到 `wf-cd`。
- `wf-docs`：文档生成合并到 `wf-ci`。
- `wf-occams`：奥卡姆剃刀作为约束原则内置，不再作为独立 workflow。
- 旧 Entry Point Workflow：由 `workflow/entry.md` 替代。

## 设计边界

工作流只提供上下文、路由和验证边界，不替 Agent 决策。新增流程必须同时满足：

1. 当前任务确实复用不了现有 workflow。
2. 有明确触发条件和退出条件。
3. 有可运行的验证方式。
4. 不违反 `CLAUDE.md` 第 1 节（绝对禁止项）。
