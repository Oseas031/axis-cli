# Agent 接手提示词

## 项目核心
Agent 原生调度系统 - 为 AI Agent 提供统一任务调度能力

## 当前状态（2026-05-08 12:54）
- ✅ 里程碑1核心功能已完成
- ✅ CI/CD流水线已建立
- ✅ 工作流改造完成
- ✅ 文档系统完善
- ✅ 工作流索引创建
- ✅ staticcheck ST1003 已修复
- ✅ 文档审查工作流已创建
- ✅ 工作流注册表验证器已创建
- ⏳ 正在进行里程碑1验收

## 文档阅读顺序（必须按顺序）
1. docs/current-progress.md - 当前工作进度（必须首先阅读）
2. HANDOVER.md - 项目交接文档（当前进度）
3. reports/daily/daily-retrospective-2026-05-08.md - 最新复盘
4. docs/QUICKSTART.md - 快速入门
5. WHITEPAPER.md - 项目定义
6. docs/milestones/milestone1-checklist.md - 里程碑1目标
7. docs/milestone1-acceptance-using-existing-workflows.md - 验收方案

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
1. 观察文档审查工作流执行结果
2. 观察工作流注册表验证器执行结果
3. 处理 release.yml 与 cd-workflow 重复问题（本周）
4. 创建PR触发PR Quality Check和Security workflows
5. 生成里程碑1验收报告
6. 完成里程碑1验收

## 已知问题
- ✅ staticcheck ST1003：shared_layer 包名包含下划线 - 已修复（2026-05-08）
- ⚠️ release.yml 与 cd-workflow 重复 - 待处理（本周）
- ⚠️ sign-artifacts job 未使用 - 待处理（里程碑1后）

## Claude Code 工作流衔接

### 接手工作流
1. 读取 `docs/current-progress.md` 了解当前工作进度
2. 读取 `docs/claude-code-workflow-continuity-guide.md` 了解工作流衔接指南
3. 读取 `reports/daily/daily-retrospective-YYYY-MM-DD.md` 了解最新复盘
4. 读取 `workflows/README.md` 了解工作流索引
5. 检查 `.github/workflows/registry.yml` 了解工作流状态
6. 检查 GitHub Actions 了解 CI/CD 状态
7. 完成交接检查清单
8. 更新记忆系统加载项目上下文

### 交接工作流
1. 更新 `docs/current-progress.md` 记录工作进度
2. 创建每日复盘文档
3. 更新 `HANDOVER.md` 和 `AGENT_INSTRUCTIONS.md`
4. 提交并推送所有变更
5. 确保 CI/CD 通过

### 重要文档
- `docs/current-progress.md` - 当前工作进度（必须首先阅读）
- `docs/claude-code-workflow-continuity-guide.md` - 工作流衔接指南
- `HANDOVER.md` - 项目交接文档
- `reports/daily/daily-retrospective-YYYY-MM-DD.md` - 最新复盘
- `workflows/README.md` - 工作流索引

## 开发优先级
1. 处理 release.yml 重复问题（本周）
2. 完成里程碑1验收（使用现有工作流）
3. 生成里程碑1验收报告
4. 准备里程碑2设计（DAG并行调度、契约准入规则、SLA约定）
