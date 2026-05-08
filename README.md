# Axis CLI

面向自治 AI Agent 原生设计的命令行工具，将「人类操作主体」封装为 Agent 可无差别调用的自然语言函数。

## 核心创新

- **主 Agent 管控者架构**：主 Agent 从「执行者」升级为「管控者」，仅负责契约制定、异常兜底、人类节点调度，子 Agent 间通过全局共享层点对点通信
- **Human-as-a-Function**：将人类操作者封装为可调用的函数
- **Agent 契约层**：主 Agent 从「发指令」变为「定契约」，子 Agent 封装为标准化无状态函数，与 call human 协议完全同构
- **DAG 并行调度**：基于依赖关系的并行调度，总耗时从「各环节时长之和」压缩为「最长关键路径的单环节时长」
- **Agent 原生调度**：Agent 主导全链路调度，人类作为按需履约的可调度子节点
- **POSIX CLI 兼容**：原生兼容 POSIX CLI 规范，可作为 AI 增强终端工具使用
- **渐进式自举**：从传统 CLI 形态向 Agent 原生终态的渐进式演化

## 技术架构

### 核心模块（必选）
- **Agent 原生调度内核**：管控者架构、任务调度、分发、生命周期管理
- **全局共享层**：子 Agent 间点对点通信基础设施
- **Agent 契约层**：契约注册、契约执行、准入规则、依赖解析
- **Human-as-a-Function**：协议层、执行器、解析器
- **统一工具调用层**：工具注册表、调用器

### 可选模块（可插拔）
- **人类兼容交互层**：POSIX 适配器、HTTP 适配器、SDK
- **AI 增强辅助模块**：LLM 提供商抽象、提示词管理、上下文管理
- **调试面板模块**：日志、指标、链路追踪、调试接口

### 模块铁律
核心模块与可选模块 100% 解耦，可选模块的迭代不得修改核心内核的 Agent 原生设计。

## 开发路线

### 里程碑 1：自举起点 - 底座搭建
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
- 内核重构为 Agent 原生调度内核（守护进程模式）
- 人类兼容交互层降级为最小调试入口
- 企业级多租户管控、多人类节点并行调度、全链路合规审计

**技术准入标准**：
- Agent 端调用占比 ≥ 70%
- 多租户与多节点调度能力无语义冲突
- 全链路合规审计可完整追溯

## 快速开始

### 安装
```bash
# 从源码构建
go build -o axis cmd/axis/main.go

# 或下载预编译二进制文件（未来提供）
```

### 基础使用
```bash
# 传统 CLI 模式
axis ls
axis run my-task

# Agent 调用模式（里程碑 2）
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{"type": "human_call", "params": {...}}'
```

## 文档

- [核心模块架构](docs/architecture/core-modules.md)
- [主 Agent 管控者架构](docs/architecture/orchestrator-architecture.md)
- [Agent 契约层设计](docs/architecture/agent-contract-design.md)
- [DAG 并行调度策略](docs/architecture/dag-scheduling.md)
- [可选模块架构](docs/architecture/optional-modules.md)
- [Call Human 协议规范](docs/protocols/call-human-spec.md)
- [LLM 提供商架构](docs/architecture/llm-provider.md)
- [里程碑 1 检查清单](docs/milestones/milestone1-checklist.md)

## 技术栈

- **语言**：Go 1.21+
- **核心依赖**：零外部依赖（仅 Go 标准库）
- **并发模型**：Goroutine + Channel
- **部署形态**：单静态二进制文件

## 贡献指南

详见 [CONTRIBUTING.md](CONTRIBUTING.md)

## 许可证

MIT License
