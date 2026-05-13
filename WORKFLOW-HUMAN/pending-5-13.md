# 待办工作：2026-05-13 起

## P0：已完成 ✅

### BashTool 执行验证强化

- [x] `verify_bash` 独立工具（Agent 自主调用，Zero Control）
- [x] 退出码 + 输出内容 + 副作用三维验证
- [x] 注册到 defaultToolRegistry

### SLA 策略引擎补充 token 预算

- [x] `internal/kernel/budget/` 独立包
- [x] 阶段分配（prototype 10% / small 30% / large 60%）
- [x] `ErrTokenBudgetExhausted` 错误码
- [x] `SLAKeyTokenBudget` 元数据键

### failure_class 收紧

- [x] `MaxRetryLimit=3` 常量
- [x] admission validator 显式拒绝超限任务（Contract is Structure）
- [x] parseSLA 不再静默截断

### 三层上下文压缩

- [x] `Compactor` 统一接口
- [x] `ThreeLayerCompaction`：micro（每轮）+ auto（阈值+transcript 持久化）+ manual
- [x] `TranscriptStore` 压缩前保存完整历史到磁盘
- [x] 替换 orchestrator 默认 pipeline

### 设计哲学对齐

- [x] verify_bash 作为可选工具（Zero Control）
- [x] 重试限制通过 admission 而非执行时静默截断（Contract is Structure）
- [x] Compactor 接口统一（无并行机制）
- [x] TokenBudget 移至 kernel/budget（types 仅存纯数据类型）

---

## P0：剩余

### TodoWrite 增强

- [ ] 添加 nag 提醒机制：Agent 遗忘任务跟踪时主动提醒

### 结构性冲突审计与重构

调查 Axis 已有实现，找出与当前设计哲学有结构性冲突的地方。

**已修复 ✅**：
- [x] fatal+retries 矛盾 → admission 显式拒绝
- [x] degradable success override → 移除，系统不替 Agent 判断成功语义
- [x] fabricated ErrTaskTimeout → 新增 ErrDispatchFailed 准确报告

**待基础设施就绪后解决（记录为技术债）**：

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

---

## 待更新文档

| 文档 | 更新内容 |
|------|----------|
| `CLAUDE.md` | 添加 Skills 系统语义边界约束 |
| `docs/architecture/semantic-boundaries.md` | 添加 SkillLoader "must NOT do" 列表 |
| `docs/architecture/module-and-naming-conventions.md` | 添加 `.axis/skills/` 目录规范 + `internal/kernel/budget/` |
| `internal/skills/BOUNDARY.md` | Never push skill content into provider prompts; never change scheduler semantics |

---

## 参考

- `docs/research/mls-bench-constraint-integration.md` — 完整设计方案
- `WORKFLOW-HUMAN/mls-bench-constraint-integration.md` — 讨论记录
- Commits: `36b9e54` (feat), `d9c5c77` (refactor align philosophy)
