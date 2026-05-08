# Axis CLI 项目交接文档

## 项目概述

**项目名称**：Axis CLI  
**项目定位**：面向自治 AI Agent 原生设计的命令行工具  
**核心创新**：将「人类操作主体」封装为 Agent 可无差别调用的自然语言函数  
**技术栈**：Go 1.21+（零外部依赖，仅 Go 标准库）  
**当前状态**：框架设计完成，未开始代码实现  

## ⚠️ 重要警告：禁止直接编码实现

**在开始任何编码工作之前，必须先完成以下步骤：**

1. **完整阅读所有架构文档**（docs/architecture/ 和 docs/protocols/）
2. **理解核心设计理念**（主 Agent 管控者、契约层、DAG 调度、Human-as-a-Function）
3. **确认当前里程碑目标**（里程碑 1：底座搭建）
4. **与原设计者确认理解无误后，再开始实施**

**绝对禁止的行为：**
- ❌ 不读文档直接写代码
- ❌ 跳过设计审查直接实现
- ❌ 提前实现后续里程碑的功能
- ❌ 修改核心架构设计

**正确的工作流程：**
1. 阅读 → 理解 → 提问确认 → 设计审查 → 编码实现 → 测试验证

## 项目目标

实现 Agent 原生 CLI 工具，打通 Agent 数字世界调度能力与物理世界执行能力，实现「Agent 主导全链路调度、人类作为按需履约的可调度子节点」的技术范式重构。

## 已完成的设计文档

### 核心架构文档
1. **核心模块架构** (`docs/architecture/core-modules.md`)
   - Agent 原生调度内核
   - 主 Agent 管控者架构
   - 全局共享层
   - Agent 契约层
   - Human-as-a-Function
   - 统一工具调用层
   - DAG 并行调度策略

2. **主 Agent 管控者架构** (`docs/architecture/orchestrator-architecture.md`)
   - 管控者三大核心职责
   - 绝对不介入原则
   - 全局共享层设计（状态存储、事件总线、数据管道、锁服务）
   - 子 Agent 间点对点通信
   - 异常兜底机制
   - 人类节点调度

3. **Agent 契约层设计** (`docs/architecture/agent-contract-design.md`)
   - 契约 5 要素（输入 Schema、输出 Schema、SLA 约定、准入规则、异常码体系）
   - 契约元数据结构
   - 准入规则引擎
   - 依赖解析机制
   - 与 Call Human 协议的同构映射

4. **DAG 并行调度策略** (`docs/architecture/dag-scheduling.md`)
   - DAG 数据结构与算法
   - 拓扑排序、循环依赖检测、关键路径计算
   - 失败处理策略（Fail Fast、Continue、Best Effort）
   - 资源感知调度
   - 性能优化（缓存、增量调度）

### 协议规范文档
5. **Call Human 协议规范** (`docs/protocols/call-human-spec.md`)
   - 调用请求/响应结构
   - 错误码体系
   - 超时管控机制
   - 重试策略与幂等性保证
   - 6 大核心技术落地场景

6. **LLM 提供商架构** (`docs/architecture/llm-provider.md`)
   - 提供商无关统一接口
   - OpenAI、Anthropic、本地模型支持
   - 模型能力适配
   - 配置管理与安全性

### 可选模块文档
7. **可选模块架构** (`docs/architecture/optional-modules.md`)
   - 人类兼容交互层（POSIX、HTTP、SDK）
   - AI 增强辅助模块（LLM 抽象、提示词管理）
   - 调试面板模块（日志、指标、链路追踪）

### 里程碑文档
8. **里程碑 1 检查清单** (`docs/milestones/milestone1-checklist.md`)
   - 内核可用性标准（99.9%）
   - Call Human 端到端闭环验证
   - POSIX CLI 兼容性（100%）
   - 核心与可选模块解耦验证

## 项目结构

```
axis-cli/
├── cmd/                    # 命令行入口
│   ├── axis/              # 主 CLI 命令
│   └── agentd/            # Agent 守护进程
├── internal/              # 核心内部模块
│   ├── kernel/           # Agent 原生调度内核
│   │   ├── orchestrator/ # 管控者架构
│   │   │   ├── contract_manager/
│   │   │   ├── dag_orchestrator/
│   │   │   ├── exception_handler/
│   │   │   ├── human_scheduler/
│   │   │   └── observer/
│   │   ├── shared_layer/ # 全局共享层
│   │   │   ├── state_store/
│   │   │   ├── event_bus/
│   │   │   ├── data_pipeline/
│   │   │   └── lock_service/
│   │   ├── scheduler/    # 调度器
│   │   ├── dispatcher/   # 分发器
│   │   └── lifecycle/    # 生命周期管理
│   ├── human/           # Human-as-a-Function
│   │   ├── protocol/    # 协议层
│   │   ├── executor/    # 执行器
│   │   └── parser/      # 解析器
│   ├── contract/        # Agent 契约层
│   │   ├── registry/    # 契约注册表
│   │   ├── executor/    # 契约执行器
│   │   ├── admission/   # 准入引擎
│   │   └── dependency/  # 依赖解析器
│   ├── tools/           # 统一工具调用层
│   │   ├── registry/    # 工具注册表
│   │   └── invoker/     # 调用器
│   └── adapter/        # 适配层
│       ├── posix/       # POSIX 适配器
│       ├── http/        # HTTP 适配器
│       └── sdk/         # SDK 适配器
├── pkg/                 # 可选公共包
│   ├── llm/            # LLM 提供商抽象
│   ├── permissions/    # 权限管控
│   └── observability/  # 可观测性
├── configs/            # 配置文件
│   ├── config.yaml    # 生产环境配置
│   └── config.dev.yaml # 开发环境配置
├── docs/               # 文档
│   ├── architecture/  # 架构文档
│   ├── protocols/     # 协议规范
│   └── milestones/    # 里程碑文档
├── scripts/            # 脚本
├── go.mod             # Go 模块定义
└── README.md          # 项目说明
```

## 四大核心创新

### 1. 主 Agent 管控者架构
**理念**：主 Agent 从「执行者」升级为「管控者」  
**职责**：
- 契约制定与 DAG 编排
- 异常兜底（仅处理 P0 级异常）
- 人类节点调度（P0 级审批）

**约束**：绝对不介入子 Agent 执行细节，不做指令中转

### 2. Agent 契约层
**理念**：主 Agent 从「发指令」变为「定契约」  
**契约 5 要素**：
- 输入 Schema（强类型、最小必要信息）
- 输出 Schema（验收标准、格式要求、合规规则）
- SLA 约定（超时、重试、熔断）
- 准入规则（前置校验条件）
- 异常码体系（标准化错误分类）

**特性**：与 Call Human 协议完全同构

### 3. DAG 并行调度
**理念**：基于依赖关系的并行调度  
**性能提升**：总耗时从「各环节时长之和」压缩为「最长关键路径的单环节时长」  
**典型加速比**：2-5 倍

### 4. Human-as-a-Function
**理念**：将人类操作者封装为可调用的函数  
**6 大场景**：代码评审、生产环境高权限操作审批、线下环境配置执行、需求边界与变更确认、合规性校验、物理世界信息采集与执行

## 开发路线

### 里程碑 1：自举起点 - 底座搭建
**目标**：完成解耦式 Agent 原生内核最小实现

**交付物**：
- Agent 原生调度内核最小实现（FIFO 策略）
- 全局共享层基础框架（内存状态存储、事件总线）
- Agent 契约层基础框架（契约注册、基础准入规则）
- Human-as-a-Function 端到端最小闭环
- POSIX CLI 兼容性验证
- 基础权限管控框架

**技术准入标准**：
- 内核可用性达到 99.9%
- call human 端到端调用闭环验证通过
- POSIX CLI 兼容性 100% 符合规范
- 核心与可选模块解耦验证通过

### 里程碑 2：自举核心 - 双向兼容
**目标**：实现 Agent 与人类双主体的调度语义兼容

**交付物**：
- Agent 原生调用层（HTTP/SDK 接口）
- 主 Agent 管控者架构实现
- DAG 并行调度策略实现
- 异常兜底机制实现
- Human-as-a-Function 全量能力实现
- 多 Agent 并发调度与全生命周期管控

**技术准入标准**：
- Agent 端调用占比 ≥ 80%
- call human 核心任务闭环成功率 ≥ 99.9%
- 多 Agent 并发调度无状态冲突与语义错乱

### 里程碑 3：自举完成 - 终态成熟
**目标**：完成 Agent 原生 CLI 终态技术演化

**交付物**：
- 内核重构为 Agent 原生调度内核（守护进程模式）
- 人类兼容交互层降级为最小调试入口
- 企业级多租户管控、多人类节点并行调度、全链路合规审计

**技术准入标准**：
- Agent 端调用占比 ≥ 70%
- 多租户与多节点调度能力无语义冲突
- 全链路合规审计可完整追溯

## 关键技术决策

### 技术栈选择
- **语言**：Go 1.21+
- **理由**：
  - Go 1 兼容性官方承诺，避免架构破坏性变更
  - Goroutine + Channel 并发模型，支持数万级并发
  - 单静态二进制文件，符合 POSIX CLI 分发范式
  - 零外部依赖，仅 Go 标准库

### 核心约束
- **模块铁律**：核心模块与可选模块 100% 解耦
- **零外部依赖**：核心模块只依赖 Go 标准库
- **管控者不干预**：主 Agent 绝对不介入子 Agent 执行细节
- **向后兼容**：核心接口变更必须保证向后兼容

### 设计原则
- **理解比执行更重要**：优先理解用户真实意图
- **分离生产与质检**：生成后必须质检
- **保持交接清晰**：关键状态写回文档，不靠对话记忆

## 分支策略

### main 分支（简化方案 - 当前）
采用**最小化里程碑 1**策略，只实现最核心的功能：

**里程碑 1 简化范围**：
- ✅ FIFO 调度器
- ✅ 基础任务分发
- ✅ 简单的 Human 调用（协议 + 执行）
- ✅ 内存状态存储（仅 Layer 1）
- ✅ 基础契约注册（仅元数据）

**暂不实现**：
- ❌ DAG 并行调度（里程碑 2 再评估）
- ❌ 分级左移（里程碑 2 再评估）
- ❌ 准入规则引擎（里程碑 2 再评估）
- ❌ 三层存储（仅 Layer 1）
- ❌ 主 Agent 管控者（里程碑 2 再评估）

### feature/full-architecture 分支（完整架构设计）
包含所有复杂方案的完整设计：
- 分级左移（Level 1/2/3）
- 三层状态存储架构（Layer 1/2/3）
- DAG 并行调度
- 主 Agent 管控者完整架构
- 准入规则引擎

**用途**：
- 作为未来扩展的参考
- 需要时可以 cherry-pick 特性
- 评估复杂方案的实际价值

### 切换策略
```bash
# 查看简化方案（当前）
git checkout main

# 查看完整架构设计
git checkout feature/full-architecture

# 如果需要某个复杂特性，可以 cherry-pick
git checkout feature/full-architecture -- docs/architecture/xxx.md
```

## 下一步实施建议

### 立即开始（里程碑 1 - 简化版）
1. 实现调度器基础框架（FIFO 策略）
2. 实现全局共享层（仅内存状态存储、事件总线）
3. 实现基础契约注册表（仅元数据，无准入规则）
4. 实现 Human-as-a-Function 最小闭环（协议 + 执行）
5. 验证 POSIX CLI 兼容性

### 实施顺序
按照依赖关系顺序实施：
1. 先实现基础设施（全局共享层）
2. 再实现核心模块（调度器、契约层）
3. 最后实现集成（Human-as-a-Function、CLI）

### 测试策略
- 每个模块必须有单元测试
- 核心模块测试覆盖率 ≥ 80%
- 集成测试覆盖 6 大场景
- 压力测试验证并发性能

## 重要注意事项

### 必须遵守的铁律
1. **核心模块零外部依赖**：只使用 Go 标准库
2. **可选模块可插拔**：移除后核心仍可独立运行
3. **管控者不干预执行**：主 Agent 绝对不介入子 Agent 执行细节
4. **接口向后兼容**：核心接口变更必须保证向后兼容

### 架构设计已完成
- 所有核心架构文档已完成
- 所有协议规范已完成
- 项目结构已搭建
- 配置文件已创建
- 无需重新设计，直接实现即可

### 实施时优先级
1. 里程碑 1 是当前唯一目标
2. 不要提前实现里程碑 2/3 的功能
3. 严格按里程碑 1 检查清单验收
4. 达到技术准入标准后方可进入里程碑 2

## 文档索引

### 快速导航
- 项目概述：本文档
- 核心架构：`docs/architecture/core-modules.md`
- 管控者架构：`docs/architecture/orchestrator-architecture.md`
- 契约层设计：`docs/architecture/agent-contract-design.md`
- DAG 调度：`docs/architecture/dag-scheduling.md`
- Call Human 协议：`docs/architecture/call-human-spec.md`
- LLM 架构：`docs/architecture/llm-provider.md`
- 里程碑 1 检查清单：`docs/milestones/milestone1-checklist.md`

### 配置文件
- 生产环境配置：`configs/config.yaml`
- 开发环境配置：`configs/config.dev.yaml`

### 项目入口
- 主 CLI 命令：`cmd/axis/`
- Agent 守护进程：`cmd/agentd/`

## 联系方式

如有疑问，请参考架构文档或查看代码注释。所有设计决策都在文档中有详细说明。

---

**交接时间**：2026-05-08  
**交接状态**：框架设计完成，等待实施  
**下一步行动**：开始里程碑 1 实施
