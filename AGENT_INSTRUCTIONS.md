# Agent 接手提示词

## 项目核心
Agent 原生调度系统 - 为 AI Agent 提供统一任务调度能力

## 文档阅读顺序（必须按顺序）
1. docs/QUICKSTART.md - 快速入门
2. WHITEPAPER.md - 项目定义
3. docs/milestones/milestone1-checklist.md - 里程碑1目标

## 设计原则
奥卡姆剃刀：最小可行，只实现验证核心概念所需的最小功能集

## 废稿警告
不要读取以下文件：
- WHITEPAPER-DRAFT.md - 已废弃
- 任何标记为 DRAFT 或 DEPRECATED 的文件

## 里程碑1范围
**包含**：FIFO 任务调度、简单依赖管理、输入输出验证、基础状态存储
**不包含**：DAG 并行调度、契约准入规则、SLA 约定、工具调用层

## 开发优先级
1. 理解里程碑1目标
2. 实现核心模块（见 docs/architecture/core-modules.md）
3. 验证功能（见 milestone1-checklist.md）
