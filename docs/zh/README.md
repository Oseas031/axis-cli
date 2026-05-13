# Axis 文档入口

文档按用途分类。优先读入口文档，再按任务进入对应目录。

## 1. 必读入口

- [项目 README](../README.md)
- [交接文档](../HANDOVER.md)
- [Agent 接手提示词](../AGENT_INSTRUCTIONS.md)
- [Claude Code 入口](../CLAUDE.md)

## 2. 产品与方向

- [原生场景白皮书](product/axis-native-scenarios-whitepaper.md)
- [Roadmap（M1-M6 ✅）](product/ROADMAP.md)
- [Agent 原生第一性原理](architecture/agent-native-first-principles.md) **← 编码前必读**
- [Autogenesis 设计报告](../reports/strategy/axis-autogenesis-design-2026-05-08.md)
- [自因化自举差距分析](../reports/strategy/bootstrap-gap-analysis-2026-05-08.md)
- [Agent 原生场景与原理分析](../reports/strategy/agent-native-scenario-principles-analysis-2026-05-11.md)

## 3. 使用指南

- [快速入门](guides/QUICKSTART.md)
- [小白快速上手指南（axis-up）](guides/BEGINNER_GUIDE.md)
- [axis-up 外部上手工具](tools/axis-up.md)
- [Bash is All You Need](architecture/bash-is-all-you-need.md)

## 4. 架构参考

- [Agent 原生第一性原理](architecture/agent-native-first-principles.md) **← 编码前必读**
- [Axis 系统规范总纲](architecture/axis-system-conventions.md)
- [模块与命名规范](architecture/module-and-naming-conventions.md)
- [语义边界规范](architecture/semantic-boundaries.md)
- [Metadata Key 规范](architecture/metadata-key-conventions.md)
- [CLI 输出规范](architecture/cli-output-conventions.md)
- [Spec 生命周期规范](architecture/spec-lifecycle-conventions.md)
- [数据模型演进规范](architecture/data-model-evolution.md)
- [错误码规范](architecture/error-code-conventions.md)
- [外部工具边界规范](architecture/external-tool-boundaries.md)
- [Secret Handling](architecture/secret-handling.md)
- [重构与迁移规范](architecture/refactor-migration-conventions.md)
- [SWE1.6 再规范执行指南](architecture/swe1-6-renormalization-guide.md)
- [Bash is All You Need](architecture/bash-is-all-you-need.md)

## 5. 规格文档

- [Milestone 2](specs/milestone2/)
- [M3 Phase 3](specs/m3-phase3/)
- [M4](specs/m4/)
- [M5](specs/m5/)
- [M6](specs/m6/)
- [Model Provider](specs/model-provider/)
- [Interactive Shell](specs/interactive-shell/)
- [Natural Language Scheduling](specs/natural-language-scheduling/)
- [Adaptive Context Assembly](specs/adaptive-context-assembly/)
- [Execution-time Context Consumption](specs/execution-context-consumption/)
- [Sandboxed Evolution Protocol](specs/sandboxed-evolution/)
- [Local Control Plane](specs/local-control-plane/)

## 6. 状态与验收

- [当前进度](status/current-progress.md)
- [Milestone 1 验收报告](status/acceptance/milestone1-acceptance-report.md)
- [Milestone 1 工作流验收](status/acceptance/milestone1-acceptance-using-existing-workflows.md)
- [Milestone 1 技术准入检查清单](status/acceptance/milestone1-checklist.md)
- [Adaptive Context Assembly 验收报告](status/acceptance/adaptive-context-assembly-acceptance-report.md)

## 7. Workflow 文档

- [工作流入口](workflow/entry.md)
- [Meta-Workflow 管理](workflow/meta-workflow-management.md)
- [奥卡姆剃刀工作流](workflow/occams-razor-architecture-simplification.md)
- [Claude Code 工作流衔接](workflow/claude-code-workflow-continuity-guide.md)
- [工作流最佳实践](workflow/workflow-best-practices.md)
- [工作流架构图](workflow/workflow-architecture.drawio)
- [工作流索引](../workflows/README.md)

## 8. 历史与归档

- [报告索引](../reports/README.md)
- [废弃文档](deprecated/README.md)

## 9. 外部参考

- [learn-claude-code](https://github.com/shareAI-lab/learn-claude-code) — 教学仓库：从零开始构建 Claude Code 风格 Agent（12 渐进式课程：工具调用 → 任务规划 → 子 Agent → 上下文压缩 → 多 Agent 协作 → 自主执行）

**当前权威设计文档**：`docs/architecture/agent-native-first-principles.md`（原理 + 设计哲学合并版）。编码前必读。
