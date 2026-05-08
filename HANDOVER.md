# Axis 项目交接文档

## 项目概述

**项目名称**：Axis（Agent 原生调度系统）
**项目定位**：面向 AI Agent 设计的统一调度平台
**核心能力**：任务调度、依赖管理、输入输出验证
**技术栈**：Go 1.26+
**当前状态**：框架设计完成，未开始代码实现

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

## 已完成的设计文档

1. **核心模块架构** (`docs/architecture/core-modules.md`) - 简化版
2. **契约层设计** (`docs/architecture/agent-contract-design.md`) - 简化版
3. **DAG 调度策略** (`docs/architecture/dag-scheduling.md`) - 简化版
4. **里程碑1检查清单** (`docs/milestones/milestone1-checklist.md`) - 简化版

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
│   │   └── shared_layer/ # 共享状态存储
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

## 文档索引

### 快速导航
- Agent 接手提示词：`AGENT_INSTRUCTIONS.md`
- 快速入门：`docs/QUICKSTART.md`
- 系统架构可视化：`docs/DIAGRAMS.md`
- 项目定义：`docs/WHITEPAPER.md`
- 里程碑1检查清单：`docs/milestones/milestone1-checklist.md`
- 项目演化路线图：`docs/ROADMAP.md`

### 配置文件
- 生产环境配置：`configs/config.yaml`
- 开发环境配置：`configs/config.dev.yaml`

---

**交接时间**：2026-05-08
**交接状态**：框架设计完成，等待实施
**下一步行动**：开始里程碑1实施
