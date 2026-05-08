# Agent 接手提示词

## 项目核心
Agent 原生调度系统 - 为 AI Agent 提供统一任务调度能力

## 当前状态（2026-05-08）
- ✅ 里程碑1核心功能已完成
- ✅ CI/CD流水线已建立
- ✅ 工作流改造完成
- ✅ staticcheck ST1003 已修复
- ⏳ 正在进行里程碑1验收

## 文档阅读顺序（必须按顺序）
1. HANDOVER.md - 项目交接文档（当前进度）
2. docs/QUICKSTART.md - 快速入门
3. WHITEPAPER.md - 项目定义
4. docs/milestones/milestone1-checklist.md - 里程碑1目标
5. docs/milestone1-acceptance-using-existing-workflows.md - 验收方案

## 设计原则
奥卡姆剃刀：最小可行，只实现验证核心概念所需的最小功能集

## 废稿警告
不要读取以下文件：
- WHITEPAPER-DRAFT.md - 已废弃
- 任何标记为 DRAFT 或 DEPRECATED 的文件

## 里程碑1范围
**包含**：FIFO 任务调度、简单依赖管理、输入输出验证、基础状态存储
**不包含**：DAG 并行调度、契约准入规则、SLA 约定、工具调用层

## 已完成的核心模块
- internal/kernel/sharedlayer/state_store - 状态存储
- internal/kernel/lifecycle - 生命周期管理
- internal/kernel/scheduler - 调度器（FIFO + 依赖管理）
- internal/contract/executor - 契约执行器（输入输出验证）
- internal/human/executor - 人类执行器
- internal/kernel/dispatcher - 分发器
- internal/kernel/orchestrator - 编排器
- cmd/axis - CLI 客户端

## 当前待处理任务
1. 观察CI workflow执行结果
2. 创建PR触发PR Quality Check和Security workflows
3. 生成里程碑1验收报告

## 已知问题
- ✅ staticcheck ST1003：shared_layer 包名包含下划线 - 已修复（2026-05-08）

## 开发优先级
1. 完成里程碑1验收（使用现有工作流）
2. 修复 staticcheck ST1003 错误
3. 生成里程碑1验收报告
4. 准备里程碑2设计（DAG并行调度、契约准入规则、SLA约定）
