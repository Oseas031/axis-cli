# 废弃文档

此文件夹包含已废弃的设计文档。

## 废弃原因

这些文档基于旧的设计思想，已被奥卡姆剃刀原理简化后的新文档替代，或属于早期里程碑设计稿，已被实际代码和后续规格文档取代。

## 文件列表

### whitepapers/
- **WHITEPAPER-DRAFT.md** — 旧版白皮书，已被 WHITEPAPER.md v1.1 替代
- **WHITEPAPER-v1.md** — 白皮书 v1，内容已被 `docs/product/ROADMAP.md` 和根目录 `README.md` 覆盖

### architecture/
- **orchestrator-architecture-DRAFT.md** — 管控者架构详细设计，已在简化版中移除
- **llm-provider-DRAFT.md** — LLM 提供商架构，M4 已完成真实 LLM 集成
- **optional-modules-DRAFT.md** — 可选模块架构，M3 已完成执行生态
- **core-modules-M1-DRAFT.md** — M1 核心模块设计稿，已被实际代码取代
- **agent-contract-design-M1-DRAFT.md** — M1 契约设计稿，已被实际代码取代
- **dag-scheduling-M1-DRAFT.md** — M1 DAG 调度设计稿，已被实际代码取代
- **DIAGRAMS-M1.md** — M1 架构图，仅反映 M1 阶段
- **agent-native-design-philosophy-v1.md** — 设计哲学 v1，内容已合并至 `agent-native-first-principles.md`

### protocols/
- **call-human-spec-DRAFT.md** — Call Human 协议规范，M1 不需要

### workflows/
- **ci-cd-quality-improvement-workflow.md** — 早期 CI/CD 设置过程，当前 CI/CD 已完成
- **comprehensive-automation-workflows.md** — 7 个核心工作流架构设计，已被简化
- **entry-workflow.md** — 工作流调度器设计，过度设计不符合奥卡姆剃刀
- **software-engineering-paradigm-workflow-improvement.md** — 理论化的 CI/CD 改进，与实际需求脱节
- **workflow-improvement-plan.md** — 工作流改进计划，高优先级问题已修复

## 当前文档

请阅读以下当前文档：
- `README.md` — 项目总览
- `docs/product/ROADMAP.md` — 里程碑路线图（M1-M6 ✅）
- `docs/architecture/agent-native-first-principles.md` — **编码前必读**
- `docs/guides/QUICKSTART.md` — 快速入门
- `docs/status/current-progress.md` — 当前进度
