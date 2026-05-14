# 待办工作：2026-05-13 起

## 主要矛盾

**AI 生成的无限性 vs 交付的确定性**

所有工作都在解决这个矛盾的不同侧面：
- Determinateness 侧：Contract、Admission、Permission Ladder、token budget
- Sublation 侧：Judge、verify_bash、三层压缩、auto-dream

次要矛盾（主要矛盾未充分解决前不分散精力）：多 Agent 协作、Subagent、Background Tasks

---

## P0：全部完成 ✅

> 已滚入 `WORKFLOW-HUMAN/completed-5-13.md`

---

## 技术债（待基础设施就绪后解决）

| # | 问题 | 前置条件 |
|---|------|----------|
| 1 | defaultToolRegistry() 硬编码工具集 | 需要 Agent 身份 + per-task Contract tool_scopes |
| 2 | dispatcher 硬编码 AutonomyLevelLow | 需要 competence profile + autonomy rule engine |
| 3 | multi-turn loop cap=10 / 终止条件 | 需要 Agent 主动停止信号（`__done` tool） |
| 4 | Permission scopes 从未被检查 | 需要 Permission Ladder 基础设施 |
| 5 | syscall tools 是空壳 | 需要 scheduler 集成接口 |
| 6 | Set* 时序耦合 | 等 AgentConfig struct 重构（P1 #2） |
| 7 | circuit breaker=5 硬编码 | 可作为 Contract 字段，优先级低 |
| 8 | dispatcher 无审计日志 | 需要结构化事件发射器 |
| 9 | mock provider 静默默认 | 开发阶段合理，生产前需改 |
| 10 | FileWriteTool 路径验证弱于 FileReadTool | 安全修复，可独立做 |

---

## P1：中期

### Judge 系统：泛化性评估（MLS-Bench P1）

**核心**：Agent 不能只在熟悉任务上表现好就升级。

- [ ] 设计分布内/分布外任务集划分方案
- [ ] 实现泛化性得分：`OOD正确率 / ID正确率`
- [ ] 阈值：≥0.8 可升级，<0.5 降级

### Autonomy Transition 规则重构

替代"连续成功 N 次"，升级需同时满足：

- [ ] 资源约束下有效（固定预算内性能提升）
- [ ] 分布外有效（泛化性 ≥0.8）
- [ ] 架构驱动有效（不因 context 增大或 model 升级而自动升级）

### Permission Ladder Level 0 白名单

- [ ] 定义初始命令白名单：ls/cat/grep/wc/git/go build/go test
- [ ] 编写 FSM 流程 Contract 模板（可选，非全局强制）

### #1 拆分 orchestrator

抽出 `ToolRegistryBuilder` 和 `ExecutorFactory`，orchestrator 只负责调度循环。

### #2 引入 AgentConfig struct

统一持有 provider/tools/skills/compaction，一次性传入 executor，消灭 `Set*` 方法链。

### Subagent 上下文隔离

```
Parent Agent                    Subagent
+------------------+            +------------------+
| messages=[...]   |            | messages=[]      |  ← fresh context
|                  │  dispatch  |                  │
| tool: task       │ ─────────> │ while tool_use:  │
|   prompt="..."   │            │   call tools     │
|                  │  summary   │   append results │
|   result = "..." │ <───────── │ return last text │
+------------------+            +------------------+
```

### Guarantee Registry（扬弃自 arXiv:2605.12239 Know 层）

系统当前持有哪些结构保证？每个保证的验证方式？在什么条件下失效？

- [ ] 定义 `Guarantee` 结构（ID / Description / VerifiedBy / BreaksWhen）
- [ ] 从已有代码提取保证清单（admission / judgement / contract / circuit breaker）
- [ ] Sandboxed Evolution promote 前检查：本次修改是否破坏已声明保证

**价值**：让 Axis 自我修改时有可审计的安全网。

### Quality-Gated Model Escalation（扬弃自 arXiv:2605.12239 Φ 层）

Judge 评分低于阈值 → 自动切换到更强模型重新执行。

- [ ] Dispatcher 执行后调用 SelfJudgement 评分
- [ ] 分数 < 阈值 → 切换 provider tier → 重新 dispatch 同一任务
- [ ] 升级次数上限（防无限循环）
- [ ] 与 SLA failure_class 集成（retryable 才升级，fatal 不升级）

**价值**：用便宜模型处理简单任务，贵模型只在需要时介入。Token 成本自动优化。

### Background Tasks 异步执行

慢操作丢后台，Agent 继续想下一步。

---

## P2：长期演进

### Feature 渐进开放

| Feature | Level 0 | Level 1 | Level 2 |
|---|---|---|---|
| BashTool | 白名单命令 | 白名单+受限管道 | 完全 bash |
| Sandboxed Evolution | 修改指定文件 | 沙箱内任意文件 | 提议架构变更 |
| Tool 权限 | 只读+执行 | 读写（Contract 授权） | 完全访问 |
| 任务生成 | 子任务 | 同级任务 | follow-up 链 |
| Contract 自定义 | 预定义 | 提议修改 | 创建新 Contract |
| 自我评判 | 外部 Judge | 自判+抽检 | 自判为主 |

升级触发：30 任务 + 正确性≥90% + 泛化性≥0.8 + 可靠性≥95% + 零安全违规

### #3 统一能力注册

`Capability` 接口（tool + prompt injection + config），新能力实现接口 + 注册一行。

### #5 provider 语义分层

Execute（对话主循环）vs Utility（摘要/评判/嵌入），独立配额和模型选择。

### 多 Agent 协作

- JSONL 邮箱协议
- 自动认领机制
- Worktree 深度集成

### Skills 组合建议元数据（扬弃自 arXiv:2605.12239 operad 层）

Skills 声明输入/输出类型和组合建议，但不强制 pipeline（保持 Zero Control）。

- [ ] SKILL.md frontmatter 扩展：`input_type` / `output_type` / `composable_with`
- [ ] Agent 可查询"哪些 skills 可以接在当前 skill 后面"
- [ ] 不强制执行顺序——建议权，非修改权

**前置条件**：CodingAgent 实践验证 Phase 1→2→3→4 流程后再决定是否需要

---

## 待更新文档

> 已全部完成（2026-05-14），见 `completed-5-13.md`

---

## 研究任务

### Gemma 4 MTP（Multi-Token Prediction）推测解码

**来源**: Google DeepMind, 2026-05-05
**核心**: 74M 草稿模型 + 主模型并行验证，无损加速 1.8x-3.0x

**对 Axis 的潜在价值**：
- **provider 语义分层**：MTP 的 drafter/verifier 架构映射到 Axis 的 Execute/Utility 分层——轻量模型做草稿，重量模型做验证
- **token budget 优化**：MTP 降低实际推理成本，影响 TokenBudget 的阶段分配策略
- **自举加速**：autogenesis loop 中大量简单任务（文件操作、格式化）可用 MTP 加速
- **judge 系统**：drafter/verifier 模式可借鉴到 self-judgement——轻量判断 + 重量抽检

**TODO**：
- [ ] 评估 Axis provider 层集成 MTP 的可行性（需要框架支持：vLLM/TensorRT-LLM）
- [ ] 评估对 token accounting 和 cost estimation 的影响
- [ ] 如有价值，写入 `docs/research/`

---

## 参考

- `docs/research/mls-bench-constraint-integration.md` — 完整设计方案
- `WORKFLOW-HUMAN/mls-bench-constraint-integration.md` — 讨论记录
- Commits: `36b9e54` (feat), `d9c5c77` (refactor align philosophy)
