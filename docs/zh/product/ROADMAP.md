# 项目演化路线图

**最后更新**: 2026-05-11

## 演化阶段

```
M1 基础调度 → M2 并行调度 → M3 执行生态 → M4 真实 LLM → M5 自举循环 → M6 自我评判
  已完成        已完成         已完成         已完成        已完成        已完成
```

## 里程碑1：基础调度 ✅ 已完成

**核心能力**
- FIFO 任务调度
- 简单依赖管理
- 输入输出验证
- 基础状态存储
- 基础 CLI

**验证标准**
- 任务提交 → 调度 → 执行 → 结果返回 闭环
- 依赖任务完成后才执行后续任务
- 输入输出 Schema 验证通过

## 里程碑2：并行调度 ✅ 已完成

**新增能力**
- DAG 并行调度（`GetReadyTasks` ready-set API）
- 契约准入规则（`internal/contract/admission`）
- SLA 约定（超时、重试、`failure_class`、退避策略、优先级排序）
- 异常码体系（`AgentError` / `ErrorCode`）
- 并行执行循环（worker pool + bounded goroutines）

**性能提升**
- 无依赖任务并行执行
- 总耗时从「各环节时长之和」压缩为「关键路径时长」

## 里程碑3：执行生态 ✅ 已完成

**Phase 1 — 执行路径打通**
- ModelProvider 接口 + Mock/Echo 实现
- Dispatcher → ContractExecutor → ModelProvider 执行链
- `ErrDependencyNotReady` + `sla.failure_class`

**Phase 2 — 生态成熟**
- Provider 可配置化（functional options）
- HumanExecutor 路由
- DAG 增强（`GetAllTasks` / `GetDependencyGraph`）

**Phase 3 — SLA 策略引擎 + 工具调用层**
- `failure_class` 路由（retryable / fatal / degradable）
- 退避策略（fixed / linear / exponential）
- 优先级排序（`sla.priority` 0-255）
- Tool 接口 + ToolRegistry + BashTool
- ModelRequest/ModelResponse 多轮扩展
- ContractExecutor 多轮 tool-use 循环（max 10 turns）

## 里程碑4：真实 LLM 集成 ✅ 已完成

**新增能力**
- Anthropic Provider（Claude 系列，直接 HTTP 调用）
- OpenAI Provider（GPT 系列，兼容 DeepSeek / MiniMax）
- Provider 配置（functional options: `WithModel` / `WithAPIKey` / `WithBaseURL`）
- Token 计费（`InputTokens` / `OutputTokens`）
- 安全 JSON 序列化

**扩展工具**
- `file_read` — 带路径验证的文件读取
- `file_write` — 带路径验证的文件写入
- `http_request` — 带主机白名单的 HTTP 客户端
- 工具权限作用域
- 熔断机制（连续错误阈值）

## 里程碑5：自举循环（Bootstrap Loop）✅ 已完成

**核心组件**
- AgentExecutor 接口 + MockAgentExecutor
- AgentRuntimeAdapter
- SelfContext 数据结构
- ContextBuilder / ContextCompressor
- 自迭代契约（analyze / implement / validate / update / review / spawn）
- BootstrapOrchestrator（自循环任务调度）
- FollowUpTaskGenerator
- AutonomyTransition（5 级自主层级）
- RuleEngine（基于 competence evidence 的规则引擎）

**Sandboxed Evolution Protocol**
- 隔离工作空间（`.axis/evolution/<run-id>/`）
- 原子进化步骤 + append-only 追踪账本
- 显式验证 / 晋升 / 丢弃门控
- 完整审计痕迹

## 里程碑6：自我评判（Self-Judgement）✅ 已完成

**核心组件**
- JudgementCriteria / JudgementResult / JudgementItem
- SelfJudgementEngine
- 5 种内置验证策略：Syntax / Semantic / Contract / Test / Coverage
- `self/judge-execution` 契约
- BootstrapOrchestrator 评判集成
- CLI `axis judge` 诊断命令

## 演化原则

- 每个里程碑独立可验证
- 后续里程碑不修改核心调度语义
- 保持向后兼容
- 奥卡姆剃刀：最小可行，渐进增强
- **More Context, More Action, Zero Control, Controllable Evolution**
