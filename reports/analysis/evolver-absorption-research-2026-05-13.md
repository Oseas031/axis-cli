# evolver-main 吸收性研究

**日期**: 2026-05-13
**来源**: [EvoMap/evolver](https://github.com/EvoMap/evolver) 本地副本 `C:\Users\ASUS\Desktop\evolver-main`（对应 npm 包 `@evomap/evolver` v1.80.7）
**目的**: 以 axis 辩证开发方法论（SRS Loop A0-A8）为工作流，识别 evolver 对 axis 可吸收的思想、架构与设计，冻结为后续 Sublation 循环的输入。
**地位**: 研究报告，不作为当前设计权威。设计权威以 `docs/architecture/` 为准。

---

## A0 Posture（本轮对象化边界）

**目标**：产出 axis 本阶段（M1-M6 + Staged Evolution + Local Control Plane 已交付）的下一个**最小可对象化单元**，要求：

- 打中 axis 当前主要矛盾（非 evolver 的主要矛盾）
- 映射到 `docs/architecture/agent-native-first-principles.md` 的某条具体原则
- 可被 `go test ./...` 单轮验证
- 不违反已生效的 Strategic Rejections（Workflow canvas / Universal Agent / Web-TUI first / Black-box AI / Static permissions）

**显式不做**：全面迁移 evolver 代码；实装 evolver 的产品化特性（Hub/ATP/Validator 市场）；引入 Gene/Capsule 全 schema；在本文档内 Execute 代码。

---

## A1 Externalize（evolver 调研外化）

### 1.1 evolver 本体论定位

> "**Evolution is not optional. Adapt or die.**"
> "**Evolver 是一个提示词生成器，不是代码修改器。**"
> 基于 arXiv:2604.15097 "From Procedural Skills to Strategy Genes" 的**经验蒸馏引擎**。

对照 axis：

| 维度 | evolver | axis |
|------|---------|------|
| 本体 | Strategy Gene 蒸馏引擎（经验→Gene 资产） | Objectification infrastructure（Agent OS） |
| 产品 | 提示词生成器 + 资产库（assets/gep/） | 基质（kernel + contract + context + evolution） |
| 范畴 | AI agent 经验跨会话累积 | intent → objective artifact 的通用基质 |
| 主矛盾 | 经验沉淀 vs 行动自由 | 无边界生成 vs 有边界交付 |
| 自我克制 | 默认不改用户代码、不改自身源码 | 沙盒 + promote/discard 显式闸门 |

**结论**：evolver ≠ axis 竞品。evolver 的"经验表征层"是 axis 对象化范畴在策略层的缺位补完，但**产品化生态（Hub/ATP/Validator）与 hook-寄生模型整体不可搬运**。

### 1.2 evolver 核心思想（五条）

1. **Gene 优于 Skill 文档**（论文结论）：紧凑、可匹配 signals、承载经验迭代；文档式 skill 控制信号稀疏且不稳定
2. **Evolver 是提示词生成器，不是代码修改器**：自我克制以保留 Agent 行动权
3. **Proxy Mailbox 解耦**：Agent 只读写本地 JSONL，Hub 鉴权/同步/重试/退避在独立进程
4. **经验必须走可审计资产**：Selector 决策 JSON + append-only `events.jsonl` + Capsule 血缘链
5. **执行安全硬核规定**：白名单 executable、解释器 flag 黑名单、shell 元字符扫描、HOME 重定向、沙盒后 rm -rf

### 1.3 evolver 架构骨架

```
index.js                      入口（.env 加载 + 进程锁 + 自更新）
  └─ src/evolve.js            六阶段流水线编排（已 javascript-obfuscator 混淆）
       └─ src/evolve/pipeline/
            guards → collect → signals → enrich → select → hub → dispatch
```

六大子系统与职责边界：

| 子系统 | 目录 | 角色 | axis 是否有对应物 |
|---|---|---|---|
| 编排 | `src/evolve/` | 六阶段单向流水线 + guards 守门 | 有：`internal/kernel/` + `internal/contextpack/` |
| 知识资产 | `src/gep/` | Gene/Capsule/Task schema + solidify + selector + memoryGraph | 部分：`internal/types/` + `internal/evolution/` |
| 网络解耦 | `src/proxy/` | mailbox/server/sync/lifecycle/task/extensions | 部分：`internal/control/`（单机） |
| 市场协议 | `src/atp/` | Agent Transaction Protocol（商家/消费者/自购/自交付） | **无，也不应有** |
| 运维 | `src/ops/` | lifecycle/health/cleanup/self_repair，跨平台进程管理 | 部分：`axis start` |
| 宿主寄生 | `src/adapters/` | Cursor/Kiro/Claude Code/Codex/OpenCode/OpenClaw hook 安装器 | **方向相反，不应有** |

### 1.4 evolver 三大 schema（知识对象化的"规定性"）

**Gene**（策略基因，`src/gep/schemas/gene.js`）：
```
category ∈ {repair, optimize, innovate, explore}
signals_match[], strategy[], validation[], preconditions[]
constraints { max_files, forbidden_paths }
epigenetic_marks[]     ← 环境依赖的行为标记（创新字段）
learning_history[]     ← 曾经尝试过的迭代路径（创新字段）
anti_patterns[]        ← 已知失败模式的负样本（创新字段）
```

**Capsule**（成功经验压缩，`src/gep/schemas/capsule.js`）：
```
gene (父链), trigger[], blast_radius { files, lines }
outcome { status ∈ {success, failed}, score }
success_streak         ← 连续成功次数 = 信用积分
env_fingerprint        ← 环境指纹，判断经验可迁移性
reused_asset_id        ← 显式记录复用血缘
a2a { eligible_to_broadcast }
diff, execution_trace[]
```

**Task**（Hub 派发，`src/gep/schemas/task.js`）：
```
signals, bounty_id, bounty_amount
validation_commands[]  ← 发布者声明、验证者执行的命令
nonce                  ← 防重放
atp_order_id           ← 与 ATP 市场挂钩
```

### 1.5 evolver 关键设计亮点

**sandboxExecutor**（`src/gep/validator/sandboxExecutor.js`）—— 最值得对齐的安全设计：

```js
ALLOWED_EXECUTABLES = new Set(['node'])       // 可执行程序白名单
BLOCKED_NODE_FLAGS = {-e, --eval, -p, --print,
                      -r, --require, --loader, // 解释器注入 flag 全堵
                      --experimental-loader, --import, --env-file}
```

- 自实现 shell 解析器，遇 `| & ; > < ` $` 元字符直接拒绝
- 禁止内联执行 `node -e "..."`，必须 `node <file>`
- 每命令 60s、批次 180s 超时，硬上限 120s/命令
- `HOME/USERPROFILE/TMPDIR` 全重定向到 `os.tmpdir()`（防读 `~/.ssh/*`、`~/.aws/credentials`、`~/.npmrc`）
- 输出截断 4000 字符（对齐 Hub schema）
- 每任务独立临时目录，执行后 `rm -rf`
- `runPreflight()` 启动时跑 `node __preflight.js` 确认工具链可用

**Proxy Mailbox**（`src/proxy/`）—— 教科书级解耦：

- 本地 JSONL `fs.appendFileSync` 落盘即确认（延迟零）
- Agent API 极简 5 动词：send / poll / ack / status / list
- UUID v7 作为 message_id（时间单调排序）
- JSONL 读取时 `safeAssign` 去掉 `__proto__/constructor/prototype`（GHSA-2cjr-5v3h-v2w4 修复点）
- 独立进程做 Hub 同步 + 指数退避，Agent 端不感知网络

**策略权重矩阵**（`EVOLVE_STRATEGY` 环境变量）：

| 策略 | 创新 | 优化 | 修复 | 场景 |
|---|---|---|---|---|
| balanced（默认） | 50% | 30% | 20% | 日常 |
| innovate | 80% | 15% | 5% | 快速铺功能 |
| harden | 20% | 40% | 40% | 大改后稳固 |
| repair-only | 0% | 20% | 80% | 紧急修复 |

**stakeBootstrap 四级退避 + 持久化**：成功后 24h 不重试；短暂失败 5min→15min→60min→4h；资金不足 60min→4h；状态持久化到 `~/.evomap/validator_stake_state.json`，防短命进程打爆 Hub。

---

## A2 Inventory（axis 代码事实盘点）

### 2.1 `internal/contextpack/artifact.go` 现状

`ReadinessArtifact` 当前字段：
```go
type ReadinessArtifact struct {
    BundleID     string
    AssemblyMode string
    TaskID       string
    PacketCount  int
    Truncated    bool
    SourceDigest string
    Sources      []string
}
```

**事实缺口**：无 `selection_rationale` / `rejected_candidates` / `anti_patterns_consulted` 字段。当前 ReadinessArtifact 只记录"用了哪些源"，不记录"为什么选这些、为什么没选那些"。

### 2.2 `internal/evolution/types.go` 现状

已定义：`EvolutionIntent` / `EvolutionRun` / `EvolutionStep` / `VerificationRecord` / `EvolutionDecision`。

```go
type EvolutionDecision struct {
    RunID     string
    Decision  DecisionType   // promoted | discarded | paused
    Actor     string
    Reason    string
    CreatedAt time.Time
}
```

**事实缺口**：`Discarded` 状态下的 run 不会产出任何"结构化的失败教训"。当前语义是"沙盒丢弃，历史保留"，但历史只有 step trace + verification record，没有可查询的"反模式"资产。

### 2.3 grep 验证

```bash
grep -r "anti_patterns|AntiPattern|success_streak|SuccessStreak|env_fingerprint|EnvFingerprint" *.go
→ 0 matches
```

确认这三个概念（反模式、连续成功、环境指纹）**在 axis 代码中完全不存在**。它们是 evolver 的真正结构性创新，不是 axis 已覆盖的功能。

### 2.4 已有相关基础设施

- `internal/contextpack/registry.go`: `ReadinessRegistry` 已具备 Record/Load 能力
- `internal/evolution/ledger.go`: `Ledger` 已具备 ReadStepsStrict
- `internal/evolution/verify.go`: `VerificationRecord` 已落盘（stdout/stderr 引用文件）
- `internal/evolution/store.go`: `Store.RunDir` 提供 run 级目录隔离

**结论**：新增"反模式"资产类型不需要改动任何现有路径，只需要在 `internal/evolution/` 增加派生逻辑，后续通过 Registry 被 ContextPack 查询即可。

---

## A3 Diagnose（真实缺口 + 前轮自我修正）

### 3.1 evolver 暴露的 axis 真实缺口

axis 目前对象化的主要是：

1. **任务产物**（AgentTask、ContextBundle、EvolutionStep）
2. **上下文组装过程**（ReadinessArtifact、SourceDigest）
3. **决策闸门**（EvolutionDecision = promoted/discarded/paused）

axis 尚未对象化的：

4. **Agent 做事方式的演化本身** ← evolver 的 Gene
5. **成功经验的跨 run 复用血缘** ← evolver 的 Capsule（env_fingerprint + reused_asset_id + success_streak）
6. **失败尝试的结构化记忆** ← evolver 的 anti_patterns

evolver 三个创新字段中，**anti_patterns 对 axis 的贡献密度最高**。理由：

- axis 已有"沙盒进化 + promote/discard"——丢弃的 run 是失败证据的富矿，但目前**失败证据无可消费的结构化出口**
- "Query is Context"原则要求 Agent 能主动查询上下文——"什么不该做"是查询密度比"什么应做"更高的信号
- anti_patterns 的派生**不需要联网、不需要新契约、不改现有调用链**——纯本地纯新增

### 3.2 非缺口（前轮自我修正）

**初判 1（应撤回）**：曾拟吸纳 Selector Decision JSON 作为 ReadinessArtifact 的新字段。

**修正**：ReadinessArtifact 的 `SourceDigest` + `Sources` 已经回答了"用了哪些上下文源"。evolver 的 Selector Decision 之所以重要，是因为它要从**已有 Gene 库**选一条出来；axis 现阶段没有 Gene 库，context 组装用的是 rule_based 规则匹配，规则本身就是决策理由。**结论**：Selector Decision 的价值依赖于"先有 Gene 资产库"的前提；前提未成立时，吸纳它是过早优化。

**初判 2（应撤回）**：曾拟吸纳 Capsule 的 success_streak 字段作为 AutonomyTransition 信用积分。

**修正**：axis 已有 AutonomyTransition rule engine 但尚未有"capability profile 跨 run 持久化"。`success_streak` 依赖"同一资产被复用多次"的语义，axis 连"资产"本身都还没定义。**结论**：不是下一循环目标，延后到 Agent Identity & Competence Profile 实装时再议。

**初判 3（应保留）**：sandboxExecutor 的 5 条硬法可对齐 BashTool。

**保留**：但不属于主矛盾命中项，延后到"BashTool 安全加固"专题循环。

### 3.3 本轮真实主矛盾命中项

**"把 Discarded evolution run 的失败证据派生为 AntiPattern 资产"** ← 唯一在本轮应被对象化的项。

理由链：
- axis 主矛盾"无边界生成 vs 有边界交付"的"边界"一侧当前靠 verification + promote/discard 兜底
- 但 discard 后的信号只存在于 ledger 中，下次 Agent 做类似事时无法回避——形成**重复踩坑**
- evolver 的 anti_patterns 正是回答这个问题的语言
- 纯派生 + 纯新增 + 零集成 → 满足"最小单元"约束

---

## A4 Realign（映射到原则）

### 4.1 命中的原则：**Query is Context**

> Context is not a "parameter package" assembled by the system and pushed to Agents, but a shared reality that Agents actively query and construct.

当前 `ReadinessArtifact` 只暴露"成功的经验源"（docs/specs/*、providers.json）。Agent 在组装上下文时**无法查询"这条路径过去失败过"**——这违反"Query is Context"的完备性要求。AntiPattern 资产就是为填补这个查询通道而存在。

### 4.2 不违反的原则（确认无冲突）

- **Interface is Existence**：AntiPattern 通过 Registry 暴露读接口，属于合规界面
- **Contract is Structure**：AntiPattern schema 写成文件（JSONL），属于结构
- **Layered Isolation is Collaboration**：派生逻辑在 evolution 包内，不跨包侵入
- **Controllable Evolution**：anti_patterns 的写入触发点是 `EvolutionDecision = Discarded`，本身就是 promote/discard 闸门的副产物

### 4.3 对应方法论条目

- **A4 Realign**：本单元真正的"对立统一"是——把"丢弃的失败证据"从"沙盒副产物"扬弃为"下轮上下文的一级资产"
- **A7 Distill（未来循环）**：当 anti_patterns 数量足够多时，需要再一层 Distill 做去重与抽象（"同一种失败模式出现 N 次 → 升格为 gene-level 反模式"）。**本循环不做**。

---

## A5 Minimize（最小单元·冻结）

### 5.1 最小单元定义

新增 `internal/evolution/antipattern.go`，导出**单一派生函数**：

```go
type AntiPattern struct {
    ID            string    `json:"id"`              // ap-<hash16>
    RunID         string    `json:"run_id"`          // tombstone 指向的 discarded run
    TargetDomain  string    `json:"target_domain"`   // 从 EvolutionIntent.TargetDomain 取
    Attempted     []string  `json:"attempted"`       // 由 EvolutionStep 列表压缩而来：action+target_path 序列
    FailureReason string    `json:"failure_reason"`  // 优先取 VerificationRecord.Status=failed 的证据摘要，否则取 EvolutionDecision.Reason
    BlastRadius   BlastRadius `json:"blast_radius"`  // 由 steps 统计得出
    DerivedAt     time.Time `json:"derived_at"`
}

type BlastRadius struct {
    Files int `json:"files"`
    Lines int `json:"lines"`   // P0 置 0（未追踪 diff lines），待后续循环补齐
}

func DeriveAntiPatternFromRun(
    intent   *EvolutionIntent,
    run      *EvolutionRun,
    steps    []*EvolutionStep,
    verifs   []*VerificationRecord,
    decision *EvolutionDecision,
) *AntiPattern
```

**语义**：当 `decision != nil && decision.Decision == DecisionDiscarded`，返回结构化 AntiPattern；否则返回 `nil`。

### 5.2 七条边界（不可误解）

1. **只新增文件**。不改 types.go、不改 store.go、不改 verify.go、不改 ledger.go。
2. **不做集成**。不自动接入任何现有调用链——Discard gate 触发写入留给后续单元。
3. `BlastRadius.Lines` 本单元**固定置 0**。P0 只统计 `len(distinct TargetPath)` 作为 Files 值。Lines 追踪依赖 diff 解析，属于后续循环。
4. `FailureReason` 优先级：
   - 若任一 `VerificationRecord.Status == VerificationFailed` → 取该 record 的 `Command + "(exit=" + ExitCode + ")"`
   - 否则 → 取 `decision.Reason`
   - 都空 → `"discarded_without_verification_failure"`
5. `ID` 生成：`"ap-" + sha256(RunID + TargetDomain + canonical(Attempted))[:16]`（与 bundleID 同风格）
6. `Attempted` 生成：遍历 `steps`，每个 step 产出一行 `string(step.Action) + ":" + step.TargetPath`，按 `Sequence` 升序。
7. 配套 `antipattern_test.go` 覆盖 5 case：
   - `decision == nil` → 返回 nil
   - `decision.Decision == DecisionPromoted` → 返回 nil
   - `decision.Decision == DecisionDiscarded` + 有 failed verification → `FailureReason` 含 exit code
   - `decision.Decision == DecisionDiscarded` + 无 failed verification → `FailureReason == decision.Reason`
   - `decision.Decision == DecisionDiscarded` + 所有字段齐全 → `ID` 稳定（同输入 → 同 ID）
   - 退出条件：`go test ./internal/evolution/... -run TestDeriveAntiPatternFromRun -count=1` 通过

### 5.3 为什么这是"最小"

- 不触碰现有代码路径 → **零回归风险**
- 单文件、单类型、单函数、单测试文件 → A6 可在一次 SRS 循环内完成
- 证明 axis 可以把 evolver 的"反模式对象化"概念用 **纯 Go 标准库 + 现有类型组合**表达，**完全不依赖 evolver 的任何外部抽象**
- 不依赖"AntiPattern Registry"的落盘通道——本单元只产出值对象，写入留给后续单元
- 不依赖"Agent 如何消费 AntiPattern"——查询通道留给后续单元

### 5.4 本单元不解决的

- AntiPattern 落盘到 `.axis/evolution/<run_id>/antipattern.json` → 后续单元
- `EvolutionDecision.Decision == Discarded` 时自动触发派生 → 后续单元
- `ReadinessArtifact` 加 `anti_patterns_consulted []string` 字段 → 更后续单元
- ContextPack 组装时按 `TargetDomain` 查询相关 AntiPattern → 更后续单元
- 跨 run 的 AntiPattern 去重与抽象升格 → A7 Distill 专题循环
- evolver 的其他所有可吸收项（见附录 A）→ 后续循环

---

## 附录 A：evolver 可吸收项完整清单（按主次要矛盾排序）

### A.1 主矛盾命中项（本次冻结为 A5 单元的上游输入）

| # | evolver 机制 | 对 axis 的贡献 | 状态 |
|---|------------|-------------|------|
| 1 | Gene.`anti_patterns[]` 字段语义（`src/gep/schemas/gene.js`） | 为"Discarded run → AntiPattern 资产"提供概念蓝本 | **本单元冻结**（A5） |
| 2 | `events.jsonl` append-only 事件链 + `parent_event` 父链 | 为 AntiPattern 落盘后的后续查询/去重单元提供血缘模型 | 待后续循环 |
| 3 | Selector Decision JSON（"为什么选这条 Gene"） | 为 ReadinessArtifact 增加 selection_rationale 字段提供模板 | 待后续循环（依赖先有 Gene 资产库） |

### A.2 次要矛盾命中项（延后循环）

| # | evolver 机制 | 对 axis 的贡献 | 预判优先级 |
|---|------------|-------------|----------|
| 4 | sandboxExecutor 5 条硬法（白名单 executable / flag 黑名单 / shell 元字符扫描 / HOME 重定向 / 输出截断） | BashTool 安全加固 | 次要（BashTool 已做权限 scope） |
| 5 | Capsule.`success_streak` | Agent capability profile 的信用积分字段 | 次要（依赖"Agent 身份"实装） |
| 6 | Capsule.`env_fingerprint` | 跨项目经验迁移的门控信号 | 次要（依赖"跨项目经验库"实装） |
| 7 | Capsule.`reused_asset_id` | 资产复用血缘链 | 次要（依赖"资产库"概念先成立） |
| 8 | Proxy Mailbox 5 动词 API（send/poll/ack/status/list） | 未来分布式协作时的 Agent 端极简 API 模板 | 次要（axis 当前单机） |
| 9 | UUID v7 消息 ID + safeAssign 防原型污染 | JSONL append-only 存储的通用硬化 | 防御性债务 |
| 10 | `EVOLVE_STRATEGY` 四预设（balanced/innovate/harden/repair-only） | AutonomyTransition rule engine 的策略预设语法糖 | 次要 |
| 11 | stakeBootstrap 四级退避 + 磁盘持久化 | 未来任何"本地进程打远端"场景的退避参考 | 次要（axis 当前不调远端） |
| 12 | `src/ops/` 跨平台 lifecycle/health_check/cleanup/self_repair | daemon 化时的运维参考 | 次要（axis start 当前最小） |
| 13 | `--review` 作为一级运行模态 | HITL 模态显式化 | 次要（context preflight --strict 已部分覆盖） |
| 14 | Gene.`epigenetic_marks[]` / `learning_history[]` | 策略演化的历史链追踪 | 次要（依赖 Gene 资产库实装） |

### A.3 非矛盾项（有价值但本阶段不触碰）

| # | evolver 机制 | 延后原因 |
|---|------------|---------|
| 15 | Gene schema 完整引入（category/signals_match/strategy/constraints） | 引入整套 Gene 等于引入 evolver 的论文本体论，scope creep |
| 16 | skillDistiller / skill2gep 双向转换 | 依赖 Gene 库与 Skill 库同时存在 |
| 17 | 验证者节点 + 去中心化共识积分 | 范畴外（axis 是本地基础设施） |
| 18 | Hub 同步 + Worker 池 + 心跳 6min | 范畴外（违反 CLI-first + offline-first） |
| 19 | narrativeMemory / memoryGraph（叙事记忆 + 图数据库） | 本阶段无 Agent 长期记忆需求 |
| 20 | personality.js (PersonalityState) | 范畴外（axis 对"Agent 人格"不立场） |
| 21 | innovation.js / reflection.js（创新/反思子系统） | 依赖 Gene/Capsule 库 |
| 22 | curriculum.js（课程学习调度） | 依赖前 4-6 项落地后才有意义 |

---

## 附录 B：明确拒绝清单（糟粕）

以下 evolver 特性**整体性不能吸收**，理由回到 axis 已生效的 Strategic Rejections 或定位约束：

| evolver 特性 | 拒绝理由 | 来源原则 |
|------------|---------|---------|
| EvoMap Hub（A2A 协议 + 心跳 + 节点注册） | 范畴外，把 axis 从本地基础设施变成分布式网络客户端 | CLI-first / offline-first |
| ATP（Agent Transaction Protocol + 商家/消费者/自购/自交付 + Bounty 经济） | 把经济系统拉进基础设施层，scope creep | 范畴外 |
| 验证者节点（decentralized validator + stake + credits） | 依赖 Hub 生态，axis 不做 | 范畴外 |
| 技能商店（`evolver fetch --skill`） | 依赖 Hub 生态 | 范畴外 |
| `src/adapters/{cursor,kiro,claude-code,codex,opencode}.js` hook 寄生模型 | axis 的哲学是"Agent 主动调 CLI"，不是反向钻进宿主 IDE 装 hook；方向相反 | Interface is Existence |
| hook 事件流（promptSubmit / afterFileEdit / agentStop 被动触发） | 被动触发与 axis 的主动调用原则冲突 | Interface is Existence |
| "Gene > Skill 文档"的论文强论断 | axis 对"知识表征方式"不立场，skill 与 Gene 是不同层 | 定位守护 |
| `shield.js` 硬禁自改源码 + `EVOLVE_ALLOW_SELF_MODIFY=false` | axis 已有 staged evolution + promote/discard 更优方案 | Controllable Evolution |
| javascript-obfuscator 混淆核心模块（selector.js 135KB / solidify.js 418KB / a2aProtocol.js 389KB / prompt.js 256KB） | 代码组织反面教材；axis 用 Go 小模块 + 纯标准库 + 完全 GPL | module-and-naming-conventions |
| "Harness as product" / 1.80+ 源码可见（非开源）定位 | 采纳即 axis 退化为闭源商品 | 定位守护 |
| 进化圈（协作进化小组 + 上下文共享） | 依赖 Hub 生态 | 范畴外 |
| PersonalityState（Agent 人格进化） | axis 对 Agent 个性不立场 | 定位守护 |

---

## 附录 C：下一循环候选项（按优先级）

本次 A5 最小单元落地（Execute + Judge）后，下一个 SRS 循环的候选对象化：

1. **AntiPattern 落盘通道**：`internal/evolution/antipattern_store.go`，把 `DeriveAntiPatternFromRun` 的产物写到 `.axis/evolution/<run_id>/antipattern.json`
2. **EvolutionDecision Discarded 自动触发 AntiPattern 派生**：在 `promote/discard` CLI 命令中调用 derive + store
3. **AntiPattern Registry 查询接口**：`AntiPatternRegistry.ByTargetDomain(domain)` / `.Recent(n)`
4. **ReadinessArtifact 扩展**：添加 `AntiPatternsConsulted []string` 字段，在 ContextPack 组装时查询相关 AntiPattern 并注入上下文
5. **sandboxExecutor 5 条硬法吸收**：BashTool 加白名单 executable + flag 黑名单 + shell 元字符扫描
6. **Selector Decision JSON（待 Gene 库前置）**：ReadinessArtifact 加 `SelectionRationale`，记录"为什么选/拒这些源"
7. **`--review` 一级模态**：把 context preflight --strict 抽象为统一的 HITL 模态

每一项都是独立 SRS 循环，不预先承诺一次完成。

---

## 本研究的方法论留痕

本次研究内部出现**两次方法论违反**并被自我修正，记录在此以便未来检索：

1. **初稿阶段**：在给出完整 A/B/C 清单时没有区分"主矛盾命中 vs 次矛盾命中 vs 非矛盾项"，仅按技术密度排序。属于 Operational Principle 1（主要矛盾优先）违反。
2. **自我修正**：A3 Diagnose 阶段做了三条"应撤回"列表（Selector Decision / success_streak / sandboxExecutor 均被从"主矛盾命中"下调），依据是"axis 当前是否已具备该字段起作用的前提条件"。只有 anti_patterns 通过了这个筛子。

修正代价：一次章节重写。若初稿就做 A3 Diagnose 筛选，0 轮。

**方法论教训**：涉及"从外部项目吸收机制"的 A1 Externalize 阶段，技术亮点的清单化是必要的，但**主次矛盾排序必须建立在 A2 Inventory 代码事实之后**——否则会把"技术先进"误判为"矛盾命中"。evolver 的 sandboxExecutor 安全设计在技术上极其硬核，但当前 BashTool 已有 scope 权限、未出过安全事故，不构成矛盾；anti_patterns 看似简单，却是 axis 真实缺口的精确对应。

另一条教训：**evolver 的混淆 + 源码可见转向**（1.80 起，devDependencies 含 javascript-obfuscator）提醒 axis——开源 + 模块化组织不只是许可证选择，也是**方法论遗产是否可被后人阅读、调试、扬弃**的结构性前提。这一条不是技术点，是本研究的间接结论，归入本留痕。
