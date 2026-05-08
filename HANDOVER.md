# Axis 项目交接文档

## 项目概述

**项目名称**：Axis（Agent 原生调度系统）
**项目定位**：面向 AI Agent 设计的统一调度平台
**核心能力**：任务调度、依赖管理、输入输出验证
**技术栈**：Go 1.26+
**当前状态**：里程碑1核心功能已完成，CI/CD已建立，工作流改造完成

**重要说明**：
- **终态**：Agent 原生调度系统
- **CLI 定位**：CLI 只是调度系统的一个客户端，不是核心
- **设计原则**：奥卡姆剃刀 - 最小可行，只实现验证核心概念所需的最小功能集

## ⚠️ 重要警告：禁止直接编码实现

**在开始任何编码工作之前，必须先完成以下步骤：**

1. **阅读 AGENT_INSTRUCTIONS.md** - Agent 接手提示词
2. **阅读 docs/QUICKSTART.md** - 快速入门
3. **阅读 docs/DIAGRAMS.md** - 系统架构可视化
4. **阅读 docs/WHITEPAPER.md** - 项目定义
5. **阅读 docs/milestones/milestone1-checklist.md** - 里程碑1目标

**绝对禁止的行为：**
- ❌ 不读文档直接写代码
- ❌ 提前实现里程碑2+的功能
- ❌ 修改核心架构设计
- ❌ 将 CLI 当成核心

## 项目目标

为 AI Agent 提供统一的任务调度能力。

**里程碑1目标**：
- FIFO 任务调度
- 简单依赖管理
- 输入输出验证
- 基础状态存储
- 基础 CLI

## 已完成的工作

### 里程碑1核心功能（已完成）
- ✅ 核心数据结构实现（AgentTask, TaskStatus, TaskResult, FieldDef等）
- ✅ 状态存储模块（internal/kernel/sharedlayer/state_store）
- ✅ 生命周期管理器（internal/kernel/lifecycle）
- ✅ 调度器（FIFO + 依赖管理）（internal/kernel/scheduler）
- ✅ 契约执行器（输入输出验证）（internal/contract/executor）
- ✅ 人类执行器（internal/human/executor）
- ✅ 分发器（internal/kernel/dispatcher）
- ✅ 编排器（internal/kernel/orchestrator）
- ✅ CLI 客户端（cmd/axis）
- ✅ 单元测试（覆盖率≥60%）

### CI/CD流水线（已完成）
- ✅ CI Workflow（format, vet, staticcheck, test, build）
- ✅ CD Workflow（多平台构建、Docker镜像、Release、签名）
- ✅ Security Workflow（SAST, SCA, Secret Scan, License Compliance）
- ✅ PR Quality Check Workflow（质量门禁、代码审查）
- ✅ Monitoring Workflow（性能基准、覆盖率趋势、CI指标）
- ✅ Dev Workflow（本地开发自动化）
- ✅ Document Audit Workflow（文档格式、链接、内容一致性、里程碑对齐）
- ✅ Registry Validator Workflow（注册表验证、索引生成）

### 工作流改造（已完成）
- ✅ 基于软件工程范式（TDD, Quality Gates, DevSecOps, CD, Observability）
- ✅ 工作流注册表（.github/workflows/registry.yml）
- ✅ Meta-Workflow架构决策（Git存储、双轨绑定、三层版本控制、显式依赖）
- ✅ 工作流架构图（docs/workflow-architecture.drawio）
- ✅ 进度报告（docs/workflow-progress-report.md）
- ✅ Claude Code 工作流衔接系统（docs/claude-code-workflow-continuity-guide.md）
- ✅ 工作流索引（workflows/README.md）
- ✅ 文档审查工作流（.github/workflows/document-audit.yml）
- ✅ 工作流注册表验证器（.github/workflows/registry-validator.yml）
- ✅ 文件夹重组（reports/ 和 docs/deprecated/workflows/）
- ✅ 工作流经验总结与完善

### Bug修复（已完成）
- ✅ 调度器 GetNextTask 标记任务为已调度
- ✅ 分发器 goroutine 泄漏修复
- ✅ 编排器 busy-wait 模式修复
- ✅ 契约执行器线程安全修复
- ✅ CLI nil指针风险修复
- ✅ 编排器 Start 逻辑错误修复
- ✅ 生命周期检查的 mutex 保护
- ✅ 契约执行器枚举验证修复
- ✅ 分发器 context shadowing 修复

### 文档（已完成）
- ✅ 里程碑1检查清单（docs/milestones/milestone1-checklist.md）
- ✅ 工作流验收文档（docs/milestone1-acceptance-using-existing-workflows.md）
- ✅ 进度报告（docs/workflow-progress-report.md）
- ✅ Claude Code 工作流衔接指南（docs/claude-code-workflow-continuity-guide.md）
- ✅ 工作流索引（workflows/README.md）
- ✅ 工作流组织报告（reports/workflow/workflow-organization-report.md）
- ✅ 文件夹组织评估（reports/folder-organization-evaluation.md）
- ✅ 工作流废弃内容检查（reports/workflow-deprecated-content-check.md）
- ✅ 工作流经验总结（reports/workflow-experience-summary.md）
- ✅ 每日复盘（reports/daily/daily-retrospective-2026-05-08.md）

## 当前待处理任务

### 立即待处理
- ⏳ 观察文档审查工作流执行结果
- ⏳ 观察工作流注册表验证器执行结果
- ⏳ 处理 release.yml 与 cd-workflow 重复问题（本周）
- ⏳ 创建PR触发PR Quality Check和Security workflows
- ⏳ 生成里程碑1验收报告

### 已知问题
- ✅ staticcheck ST1003：shared_layer 包名包含下划线 - 已修复（2026-05-08）
- ⚠️ release.yml 与 cd-workflow 重复 - 待处理（本周）
- ⚠️ sign-artifacts job 未使用 - 待处理（里程碑1后）

### 下一步行动
1. 使用现有工作流完成里程碑1验收（不创建新工作流）
2. 处理 release.yml 重复问题
3. 生成里程碑1验收报告
4. 准备里程碑2设计（DAG并行调度、契约准入规则、SLA约定）

## 项目结构

```
axis-cli/
├── cmd/
│   ├── axis/              # 主 CLI 命令
│   └── agentd/            # Agent 守护进程（里程碑2+）
├── internal/
│   ├── kernel/           # 调度内核
│   │   ├── scheduler/    # 调度器
│   │   ├── dispatcher/   # 分发器
│   │   ├── lifecycle/    # 生命周期管理
│   │   └── sharedlayer/ # 共享状态存储
│   ├── contract/        # 契约层
│   │   └── executor/    # 契约执行器
│   └── human/           # Human-as-a-Function
│       └── executor/    # Human 执行器
├── docs/                 # 文档
├── configs/              # 配置文件
├── go.mod               # Go 模块定义
└── README.md            # 项目说明
```

## 开发路线

### 里程碑1：基础调度
**目标**：验证 Agent 调度核心能力

**包含**：
- FIFO 任务调度
- 简单依赖管理
- 输入输出验证
- 基础状态存储
- 基础 CLI

**不包含**（里程碑2+）：
- DAG 并行调度
- 契约准入规则
- SLA 约定
- 工具调用层

### 里程碑2：并行调度
- DAG 并行调度
- 契约准入规则
- SLA 约定

### 里程碑3：生态成熟
- 工具调用层
- 完整异常处理
- 多客户端支持

## 关键技术决策

### 技术栈选择
- **语言**：Go 1.26+
- **理由**：
  - Goroutine + Channel 并发模型
  - 单静态二进制文件
  - 零外部依赖，仅 Go 标准库

### 核心约束
- **零外部依赖**：核心模块只依赖 Go 标准库
- **向后兼容**：核心接口变更必须保证向后兼容
- **奥卡姆剃刀**：最小可行，渐进增强

## 下一步实施建议

### 立即开始（里程碑1）
**实施顺序**：
1. 实现基础任务调度（FIFO）
   - 任务队列（内存实现）
   - 任务提交/消费
   - 任务状态跟踪
2. 实现简单任务编排
   - 任务依赖管理
   - 循环依赖检测
3. 实现输入输出验证
   - 输入 Schema 验证
   - 输出 Schema 验证
4. 实现基础状态存储
   - 内存状态存储
   - 状态查询
5. 实现 CLI 客户端（使用 cobra 框架）
   - 基础命令解析
   - 信号处理

### 测试策略
- 核心调度能力必须有单元测试
- 任务调度单元测试覆盖率 ≥ 60%
- 基础端到端集成测试验证通过

## 重要注意事项

### 必须遵守的铁律
1. **里程碑1是当前唯一目标**
2. **不要提前实现里程碑2+的功能**
3. **严格按里程碑1检查清单验收**
4. **CLI 只是客户端，不是核心**

### 废稿警告
不要读取以下文件（已废弃）：
- docs/deprecated/whitepapers/WHITEPAPER-DRAFT.md
- docs/deprecated/architecture/orchestrator-architecture-DRAFT.md
- docs/deprecated/architecture/llm-provider-DRAFT.md
- docs/deprecated/architecture/optional-modules-DRAFT.md
- docs/deprecated/protocols/call-human-spec-DRAFT.md
- docs/deprecated/workflows/ci-cd-quality-improvement-workflow.md
- docs/deprecated/workflows/comprehensive-automation-workflows.md
- docs/deprecated/workflows/entry-workflow.md
- docs/deprecated/workflows/software-engineering-paradigm-workflow-improvement.md
- docs/deprecated/workflows/workflow-improvement-plan.md

## 文档索引

### 快速导航
- Agent 接手提示词：`AGENT_INSTRUCTIONS.md`
- 当前工作进度：`docs/current-progress.md`
- Claude Code 工作流衔接指南：`docs/claude-code-workflow-continuity-guide.md`
- 快速入门：`docs/QUICKSTART.md`
- 系统架构可视化：`docs/DIAGRAMS.md`
- 项目定义：`docs/WHITEPAPER.md`
- 里程碑1检查清单：`docs/milestones/milestone1-checklist.md`
- 项目演化路线图：`docs/ROADMAP.md`
- 工作流索引：`workflows/README.md`
- 报告索引：`reports/`

### 配置文件
- 生产环境配置：`configs/config.yaml`
- 开发环境配置：`configs/config.dev.yaml`

---

**交接时间**：2026-05-08 12:54
**交接状态**：里程碑1核心功能已完成，CI/CD已建立，工作流改造完成，文档系统完善
**下一步行动**：处理 release.yml 重复问题，完成里程碑1验收
