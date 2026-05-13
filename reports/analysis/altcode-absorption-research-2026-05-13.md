# altcode-main 吸收性研究

**日期**: 2026-05-13
**来源**: [jiayaoqijia/altcode](https://github.com/jiayaoqijia/altcode) 本地副本 `C:\Users\ASUS\Desktop\altcode-main`
**目的**: 以 axis 辩证开发方法论（SRS Loop A0-A8）为工作流，识别 altcode 对 axis 可吸收的思想、架构与设计，冻结为后续 Sublation 循环的输入。
**地位**: 研究报告，不作为当前设计权威。设计权威以 `docs/architecture/` 为准。

---

## A0 Posture（本轮对象化边界）

**目标**：产出 axis 本阶段（M1-M6 后）的下一个**最小可对象化单元**，要求：
- 打中 axis 当前主要矛盾（非 altcode 的主要矛盾）
- 映射到 `docs/architecture/agent-native-first-principles.md` 的某条具体原则
- 可被 `go test ./...` 单轮验证
- 不违反已生效的 Strategic Rejections（Workflow canvas / Universal Agent / Web-TUI first / 等）

**显式不做**：全面迁移 altcode 代码；实装 altcode 的产品化特性；在本文档内 Execute 代码。

---

## A1 Externalize（altcode 调研外化）

### 1.1 altcode 本体论定位

> "**The harness is the product. The model is the engine.**"
> 跨 13 个 LLM 提供商的**编码垂类脚手架**。

对照 axis：
| 维度 | altcode | axis |
|------|---------|------|
| 本体 | Coding harness（跨模型脚手架） | Objectification infrastructure（Agent OS） |
| 产品 | Harness 是产品 | Harness 是构件；系统是基质 |
| 范畴 | coding 垂类 | intent → objective artifact 的通用基质 |
| 主矛盾 | 模型能力 vs 脚手架假设的时效性 | 无边界生成 vs 有边界交付 |
| 扬弃态度 | "Every eval failure is a harness signal" | 主体见之于客体的辩证否定 |

**结论**：altcode ≠ axis 竞品。altcode 的工程硬核可被 axis 吸收，但**定位不可整体搬运**。

### 1.2 altcode 核心思想（五条）

1. **Harness as Product**：agent 行为差异主要来自 harness，而非 model tuning
2. **Action beats narration**：先拿证据，少描述 intent
3. **Verification is part of generation**：验证不是后置环节，是生成循环成员
4. **Generator/Evaluator 物理隔离**：独立 worktree + 独立 session 消除 self-evaluation bias
5. **Every eval failure is a harness signal. Fix the environment, not the symptom.**（meta-rule）

### 1.3 altcode 架构亮点（按模块）

| 模块 | 核心机制 | 代码位点 |
|------|---------|---------|
| `internal/engine/intent.go` | TaskClass（8 类）+ RiskLevel（3 档）关键词分类 | `ClassifyIntent(prompt)` |
| `internal/engine/loopguard.go` | sha256(tool\|input) 去抖，soft=3 注入合成 tool_result / hard=8 熔断 | `LoopGuard` |
| `internal/compact/summarize.go` | cutoff walk-backward 不切断 tool_use/tool_result 对 | `Summarizer.Compact()` |
| `internal/compact/budget.go` | OpenAI 扁平 `Content` + Anthropic `Parts` 双模式截断 | `BudgetCompactor.Apply()` |
| `internal/permission/permission.go` | Allow/Deny/Ask + doom-loop 检测（连问 3 次转 deny） | `Evaluator.Check()` |
| `internal/sandbox/sandbox.go` | Shell-syntax 拦截（`>` `<` `` ` `` `;` `\|` `&` `$(...)`），引号/转义感知 | `unsafeShellSyntax()` |
| `internal/hooks/hooks.go` | 13 类事件（PreToolUse/PostToolUse/PreCompact/SubagentStop/…） | `KnownEvents()` |
| `internal/workspace/worktree.go` | git worktree + symlink deps + checkpoint commit | `WorktreeWorkspace` |
| `internal/lifecycle/lifecycle.go` | 12 态 PR 流程状态机（spawning→pr_open→ci_checking→…→merged） | `Manager.Advance()` |
| `internal/memory/memory.go` | MEMORY.md 索引 + path-traversal 防御 | `validateMemoryID()` |
| `internal/wfdef/wfdef.go` | YAML 工作流 DSL（phases + depends_on + parallel + on_failure） | `WorkflowDef` |
| `internal/workflow/ralph.go` | 持续执行模式，终止信号 JSON `{"done":true,"reason":"..."}` | `RalphPrompt()` |
| `internal/mcp/` | MCP stdio + SSE，`mcp__server__tool` 命名空间 | `Manager.RegisterAll()` |
| `internal/lsp/tool.go` | LSP diagnostics/definition/hover 暴露为 tool | `NewLSPTool()` |
| `internal/agent/team.go` | Team + Mailbox + depth-limited registry for sub-agents | `NewChildTeam()` |
| `docs/world-class-coding-agent-design.md` §12 | Verification Ladder：parse → unit → package → build → lint → reviewer | 设计文档 |
| `docs/world-class-coding-agent-design.md` §13 | 10 类失败 × 明确 recovery policy 矩阵 | 设计文档 |
| `docs/world-class-coding-agent-design.md` §15 | 4 层 state：raw history / structured task / decision ledger / long-term memory | 设计文档 |

---

## A2 Inventory（axis 代码事实盘点）

**基于 2026-05-13 对仓库的实际读取**（不靠记忆推断）。

### 2.1 `internal/agent/judgement/engine.go`

```go
func (e *Engine) Judge(input any, criteria []strategies.JudgementCriteria) (*JudgementResult, error) {
    // ...
    for _, c := range criteria {
        // 平铺遍历，无顺序保证，无短路语义
    }
}
```

**事实**：
- Engine 结构上**已无状态**（只接受 `input any` + `criteria`，不接收 session history）
- criteria **平铺遍历**，无显式阶梯顺序，无失败短路
- 默认策略：Syntax / Test / Coverage；可选：Semantic / Contract（需要 provider/executor 注入）

### 2.2 `internal/agent/followup.go`

```go
func GenerateFollowUpTasks(result *AgentExecutionResult, parentTask *types.AgentTask) []*types.AgentTask {
    if result == nil || len(result.FollowUpTasks) == 0 { return nil }
    // ... 只是遍历 result.FollowUpTasks 构建 AgentTask 壳子
}
```

**事实**：
- `GenerateFollowUpTasks` **不是决策者**，是 task-shape builder
- "下一步做什么"的决策由上游填入 `result.FollowUpTasks`，即**由 LLM 在生成阶段自由决定**
- 函数与 `JudgementResult` **完全无关**（签名里没有 judgement 入参）

### 2.3 `internal/agent/bootstrap.go`

- `WithJudgementEngine` option 存在
- 但 BootstrapOrchestrator 实际未在 follow-up 分支里**调用** Judgement Engine
- `metadata["parent_validation_passed"]` 仅作为 tag 写入，**不是因果变量**

### 2.4 `docs/status/current-progress.md` Gap H

> **Gap H — 状态 Open**：Execution feedback loop is fully broken; no quality assessment or system improvement from results.

**Gap H 精确定位**（本文补强）：不是"没有 quality assessment"（已有 JudgementEngine），而是 **assessment 的结果没有在结构上驱动 next iteration 的 task shape**。

---

## A3 Diagnose（真实缺口 + 前轮自我修正）

### 3.1 前轮（对话 round 1-2）错误主张撤回

| 前轮主张 | A2 代码事实 | 处置 |
|---------|----------|------|
| "Judge 在同 session 自评，需引入 Evaluator Fresh Context 隔离 self-evaluation bias" | `Engine.Judge` 不接收 session；结构上已无状态 | ❌ 撤回。该问题不存在于当前实现 |
| "LoopGuard 是中优先级，打中轻度问题" | 真实主矛盾不在 loop 层，在 Sublation→FollowUp 断链 | ⬇️ 降为次要矛盾延后 |
| "应按 Ideas > Architecture > Designs 批量吸收" | 违反方法论 Operational Principle 1（主要矛盾优先） | ❌ 撤回 |

**方法论违反**：前两轮违反 Operational Principle 5（调查研究——读文件、跑测试、grep，不靠记忆推断），直接给出 Tier A/B/C 清单式建议，属于 A6 风格的过早输出。

### 3.2 真实缺口（代码确证）

> **`JudgementResult` 与 `GenerateFollowUpTasks` 之间没有因果连接。Sublation 的"扬弃发现"（Judge）与"下一次 Objectification 的边界"（FollowUp）在代码里是两条不交的轨道。决策谁做了下一步？—— LLM，通过 `AgentExecutionResult.FollowUpTasks` 的自由填写。**

这正是 Gap H 的**机制级根因**。

---

## A4 Realign（映射到原则）

### 4.1 被违反的原则：**Contract is Structure**

结构性决策"失败 X → 下一步做 Y"当前寄生在 LLM 生成里，而不是 Contract 声明里。这让"规定性的动态化"仍然是**承诺**而非**结构**。

### 4.2 被误解的原则：**Zero Control**

Zero Control 约束的是 Agent 的**行动路径**，**不是**在规定性内的路径选择。失败→恢复的映射属于 Determinateness（方法论 A2/A3 阶段的分类产物），不属于 LLM 的 free will。

前轮把"LLM 决定下一步"当成 Zero Control 合规实现——概念偷换。

### 4.3 对应方法论条目

- **A7 Distill**："保留/修正/剔除/沉淀" 必须由**失败证据的分类**机械驱动
- 当前实现把 A7 让给了 LLM，LLM 在没有系统性调查失败证据的前提下自创 follow-up，同时违反 Operational Principle 5

---

## A5 Minimize（最小单元·冻结）

### 5.1 最小单元定义

新增 `internal/agent/recovery.go`，导出**单一函数**：

```go
func DeriveRecoveryTasks(
    judgement *judgement.JudgementResult,
    parentTask *types.AgentTask,
) []*types.AgentTask
```

**语义**：当 `judgement != nil && !judgement.Passed`，按**失败 criterion → recovery contract ID** 的静态映射表，为每个失败条目合成一个 recovery `AgentTask`，`dependencies = [parentTask.TaskID]`。否则返回 `nil`。

### 5.2 七条边界（不可误解）

1. **只新增文件**。不改 Engine、不改 BootstrapOrchestrator、不改 GenerateFollowUpTasks。
2. **不做集成**。不自动接入任何现有调用链——让后续单元处理接入。
3. 映射表**只预置 1 条**：`syntax → "axis.recover.syntax"`。表在文件顶层常量，后续扩展。
4. 未知 criterion 类型 → **跳过并在 metadata 里记** `recovery.skipped_reason=unmapped`，不报错。
5. 输入 `judgement == nil` 或 `judgement.Passed == true` → 返回 `nil`。
6. 每个 recovery task 的 metadata 必含：
   - `recovery.source_criterion` = 失败 criterion 名
   - `recovery.parent_task_id` = parentTask.TaskID
   - `recovery.derivation = automatic`
7. 配套 `recovery_test.go` 覆盖 4 case：
   - all-pass → nil
   - mixed-fail-mapped → 仅生成 mapped criterion 的 recovery
   - fail-unmapped → 返回空列表且记录 skipped_reason
   - nil-input → nil
   - 退出条件：`go test ./internal/agent/... -run TestDeriveRecoveryTasks -count=1` 通过

### 5.3 为什么这是"最小"

- 不触碰现有代码路径 → **零回归风险**
- 单文件、单函数、单测试文件 → A6 可在一次 SRS 循环内完成
- 证明"Sublation 结果 → Objectification 边界"**可以脱离 LLM**用代码表达
- 不依赖还未存在的 `axis.recover.syntax` 契约——函数只产出 task shape，契约缺失由后续单元解决

### 5.4 本单元不解决的

- `axis.recover.syntax` 契约的实际内容 → 后续单元
- 接入 BootstrapOrchestrator / FollowUpTaskGenerator → 后续单元
- 映射表从静态走向 contract-declared → 更后续单元
- altcode 的其他所有可吸收项（见附录 A）→ 后续循环

---

## 附录 A：altcode 可吸收项完整清单（按主次要矛盾排序）

### A.1 主矛盾命中项（本次冻结为 A5 单元的上游输入）

| # | altcode 机制 | 对 axis 的贡献 | 状态 |
|---|------------|-------------|------|
| 1 | Verification Ladder（`docs/world-class-coding-agent-design.md` §12） | 为 JudgementEngine 的 criteria 提供**顺序化与短路语义** | 本单元冻结，后续循环落地 |
| 2 | Failure class → recovery policy 矩阵（§13） | 为映射表提供初始词典 | 本单元冻结，后续循环扩展映射表 |
| 3 | Ralph 终止信号 JSON（`internal/workflow/ralph.go`） | Bootstrap Loop 的 termination 从 LLM 自由解释升级为契约级信号 | 待后续循环 |

### A.2 次要矛盾命中项（延后循环）

| # | altcode 机制 | 对 axis 的贡献 | 预判优先级 |
|---|------------|-------------|----------|
| 4 | LoopGuard（`internal/engine/loopguard.go`） | tool-call 粒度重复/熔断检测 | 次要矛盾（SLA + failure_class 已覆盖 80%） |
| 5 | Compact 边界 walk-backward（`internal/compact/summarize.go`） | 防未来多轮压缩踩 tool_use/tool_result 切断坑 | 防御性债务 |
| 6 | Shell-syntax 拦截（`internal/sandbox/sandbox.go`） | BashTool 安全硬化 | 防御性债务 |
| 7 | Hook 13 事件分类 | Event log schema 完备性 | 待 `axis audit` 实装时 |
| 8 | Permission Allow/Deny/Ask + doom-loop | 交互增强 | 低 |
| 9 | Memory ID path-traversal 防御 | 未来 `.axis/memory/` 实装时采纳 | 低 |
| 10 | Activity JSONL tail-read | `.axis/events/tasks.jsonl` O(n²) poll 优化 | 低 |
| 11 | Auto-extending backtick fence | FileReadTool 输出格式 | 低 |

### A.3 非矛盾项（有价值但本阶段不触碰）

| # | altcode 机制 | 延后原因 |
|---|------------|---------|
| 12 | MCP 集成 | Scope creep，本阶段主矛盾未解 |
| 13 | LSP as tool | 依赖外部 LSP server，本阶段非必需 |
| 14 | git worktree as evolution backend | Sandboxed Evolution P0 已交付，Plan B（手工 worktree）已验证 |
| 15 | Task Classifier（TaskClass/RiskLevel） | `intent/` 模块增强候选，但非本阶段关键 |
| 16 | Cost Budget USD 微分累计 | Partial：T27 已做 `CostEstimateUSD`，预算约束延后 |
| 17 | 4 层 state 模型（含 decision ledger） | M8+ 讨论，本阶段不触碰 |
| 18 | Evidence Packet 子代理输出约束 | 依赖前面 1-3 落地后才有意义 |
| 19 | Harness Entropy Check（规则过期检测） | 需要先有足够规则历史，延后 |
| 20 | Generator/Evaluator 物理隔离 | A2 事实表明 Engine 已无状态，本项无效 |

---

## 附录 B：明确拒绝清单（糟粕）

以下 altcode 特性**整体性不能吸收**，理由回到 axis 已生效的 Strategic Rejections 或定位约束：

| altcode 特性 | 拒绝理由 | 来源原则 |
|------------|---------|---------|
| WFDef YAML 工作流 DSL（phases/depends_on/parallel/on_failure） | 将 Agent 行动路径锁入独立 DSL 层 | Reject Workflow canvas |
| 重型 Bubbletea TUI（22 文件，70K+ 行） | 违反 CLI-first | Reject Web/TUI first |
| Gateway 子模块（WhatsApp/Telegram/Slack/WeCom/Discord/…12 种 IM 集成） | 与 CLI-first OS 本体论正交 | 范畴外 |
| OAuth PKCE + device-code 登录流程 | 项目本地 provider profile 已满足 | 过度工程 |
| 多订阅凭证跨 subagent 复用 | SaaS 多租户语义，axis 是项目本地基础设施 | 范畴外 |
| 12 态 workspace PR lifecycle（spawning→pr_open→ci_checking→merged） | 绑死 coding 垂类；axis 的 lifecycle 必须通用 | 范畴约束 |
| "Harness is the product" 定位 | 采纳即 axis 退化为 framework | 定位守护 |

---

## 附录 C：下一循环候选项（按优先级）

本次 A5 最小单元落地（Execute + Judge）后，下一个 SRS 循环的候选对象化：

1. **接入 `DeriveRecoveryTasks` 到 BootstrapOrchestrator/FollowUpTaskGenerator 调用链**
2. **起草 `axis.recover.syntax` 契约内容**（对应映射表第 1 条）
3. **把 JudgementEngine 的 criteria 从平铺改为阶梯**（吸收 Verification Ladder 顺序 + 短路）
4. **扩展映射表到覆盖 Test/Contract/Semantic/Coverage 五策略的 recovery 契约**

每一项都是独立 SRS 循环，不预先承诺一次完成。

---

## 本研究的方法论留痕

本次研究内部出现**两次方法论违反**并被自我修正，记录在此以便未来检索：

1. **Round 1-2**：违反 Operational Principle 1（主要矛盾优先）和 5（调查研究），直接给出 Tier A/B/C 清单式吸收建议，属于 A6 风格过早输出。A3 阶段缺失。
2. **Round 4（A2 代码实读后）**：撤回 Round 3 "Evaluator Fresh Context" 主张，基于 `internal/agent/judgement/engine.go` 事实（Engine 本就无状态）。

修正时间：2026-05-13。修正代价：一轮对话。若早读代码，0 轮。

**方法论教训**：涉及 axis 代码的 A2 Inventory，必须先 `glob` + `read` 实际文件，再进入 A3 Diagnose。任何"印象中 axis 的 X 是怎么实现的"都必须被代码事实替换，否则整个 A3/A4/A5 链条建立在沙基之上。
