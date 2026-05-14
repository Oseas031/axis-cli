# 当前工作进度

**更新时间**: 2026-05-14
**当前里程碑**: Milestone 1 ✅ | Milestone 2 ✅ | Milestone 3 Phase 1-3 ✅ | Milestone 4 ✅ | Milestone 5 ✅ | Milestone 6 ✅ | Coding Agent P0 ✅ | 架构批判修复 ✅

## 当前设计定位

Axis 当前不再仅定义为普通 Agent 调度平台，而是 Agent 自因化的早期执行底座。

核心判断：

- 自举起点已经发生：外部 Agent 正在向 Axis 注入可观固化、执行、反思和演化的思想
- M2 不是普通并行调度里程碑，而是未来 Autogenesis Loop 的执行底座
- workflow 是临时脚手架，contract 是成长边界，permission rule 是逐进自主权机制，spec 是种子

## 架构批判修复（2026-05-14 下午）

### P0 — 架构硬伤修复
- [x] **Docker SandboxedBashTool**：真正的进程/网络/文件系统隔离。只读挂载项目目录、`--network none`、内存/CPU 限制、per-turn 超时。Docker 不可用时优雅降级。
- [x] **重命名 "Sandboxed Evolution" → "Staged Evolution"**：诚实命名——该协议使用隔离工作区 + 审查门，不是 OS 级沙箱。

### 技术债务批量消化（8 项）
- [x] circuit breaker 硬编码 → `multiturn.LoopConfig.MaxErrors` 外部化
- [x] Set* 时序耦合 → `ExecutorConfig` struct + `factory.go`
- [x] multi-turn cap=10 → `LoopConfig.MaxIterations` 外部化
- [x] mock provider 静默 → stderr 打印警告
- [x] dispatcher 无审计 → `SetAuditFunc` + dispatch_start/end 事件
- [x] 硬编码 tool set → `BuildToolRegistry(root, filter)` 可选过滤
- [x] 硬编码 AutonomyLevelLow → 从 task metadata 读取
- [x] Permission scopes 从未检查 → 显式 log 标记（v1 gap）

### 结构改进
- [x] Orchestrator 拆分：`factory.go` 提取 `BuildToolRegistry` + `BuildContractExecutor`
- [x] Multi-turn 循环统一：`internal/model/multiturn/loop.go`
- [x] Orphan cleanup 加固：共享实例、错误传播、256KB buffer
- [x] terminationProvider 重试上限：防止无限循环
- [x] Tool 执行超时：工具现在也受 TurnTimeout 约束
- [x] 前端源码入 git：16 个文件 tracked

### Vigil 状态
- 总计 44 | 完成 32 | 待做 12（0 P0 / 0 P1 / 11 P2 / 1 P3）

## Agent 自迭代验证（2026-05-14 晚）

### 真实 Coding Task 验证
- [x] 18/18 测试全部 PASS（fix bug / implement function / fix race / fix panic）
- [x] Agent 展示完整 read→write→verify 循环
- [x] Runaway 检测在真实场景中触发

### 自迭代机制
- [x] ExecutionJudge 5 条件（error rate / trailing failures / intent alignment / output substance / go test oracle）
- [x] Judgement 失败 → 5s 退避 → 自动 retry + 纠正 feedback
- [x] Memory 集成：失败教训写入 `.axis/memory/patterns/` → 下次 recall 注入 prompt
- [x] FallbackProvider + rate limiter：429 失败率 60% → 0%

### 评估方法论
- [x] `docs/architecture/agent-evaluation-principles.md`：tests decide, not judges
- [x] go test oracle 作为权威判定

## Coding Agent P0（2026-05-14）

### 战略与设计
- [x] 前端战略定位：Observatory（CLAUDE.md §14）
- [x] Agent 设计第一性原理文档（`docs/architecture/agent-design-first-principles.md`）
- [x] LLM 缺陷分类：2 个架构性缺陷 + 3 个当前形态约束
- [x] Harness 借鉴分析：结论为不借鉴具体机制，只内化方法论思维

### 实现
- [x] `LLMAgentExecutor` — 多轮 LLM ↔ Tool 循环（258 行）
- [x] 7 个单元测试覆盖核心路径
- [x] 通过 `WithAgentExecutor` + `SetToolRegistry` 接入 orchestrator
- [x] Coding Agent Spec-RDT（`docs/specs/coding-agent/`）

### 安全与技术债务
- [x] FileWriteTool 路径验证加固
- [x] BashTool 权限阶梯 L0/L1/Unrestricted
- [x] `agent_id` + `quality_signal` 字段预留（止血不可逆信息损失）

### 重构
- [x] `ExecutorConfig` struct 替代 Set* 方法链
- [x] Orchestrator 使用 `NewContractExecutorWithConfig`

### 新模块
- [x] Guarantee Registry（`internal/guarantee/`）— 可验证的系统保证

### 前端
- [x] axis-gui 前端重写：Dashboard/Tasks/Providers/Events/Chat 五个完整页面

### 基础设施
- [x] 项目复制到 WSL
- [x] WSL 安装 Kiro CLI v2.3.0
- [x] MindMagnifier skill 注册

## 已完成里程碑摘要

| 里程碑 | 核心内容 | 状态 |
|--------|----------|------|
| M1 | 基础调度、CLI 框架、CI | ✅ |
| M2 | DAG 并行调度、Contract Admission、SLA | ✅ |
| M3 | LLM Provider 集成、Tool 系统、Multi-turn | ✅ |
| M4 | 自然语言调度、Context Assembly、Control Plane | ✅ |
| M5 | Staged Evolution、Self-Judgement | ✅ |
| M6 | Bootstrap Loop、Follow-up Generation、Autonomy Rules | ✅ |
| Coding Agent P0 | LLMAgentExecutor、Permission Ladder、Guarantee Registry | ✅ |

## 核心架构差距（Top Gaps）

| # | 差距 | 严重度 | 状态 |
|---|------|--------|------|
| A | 跨进程 context 断裂 | Critical | 部分解决 |
| B | Orchestrator 伪并行，无 inter-Agent 协作 | Critical | Open |
| C | Event log 无结构化查询 | High | Open |
| E | Tool 边界是静态围栏 | High | 部分（L0 白名单已实现） |
| F | Model routing 手动挡 | High | 部分（cost tracking 有） |
| G | 无 Agent 身份 | High | 部分（agent_id 字段已预留） |
| H | 执行反馈环断裂 | High | 部分（quality_signal 字段已预留） |

## Vigil 工作追踪

活跃 P1 项：5/11 已完成，6 项 pending。详见 `.axis/vigil/items.json`。

## 推荐下一步优先级

1. Split orchestrator（依赖 AgentConfig 重构已完成）
2. Judge generalization scoring
3. Autonomy Transition rule refactor
4. Quality-Gated Model Escalation
5. Cross-process state persistence
