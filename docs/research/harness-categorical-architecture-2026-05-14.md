# Harness Engineering as Categorical Architecture：对 Axis 的启示

**日期**: 2026-05-14
**论文**: arXiv:2605.12239 (Bogdan Banu, 2026-05-12)
**性质**: 竞品/参照分析

---

## 1. 论文核心

### 核心命题

Agent Harness（包围模型的系统层：prompts + tools + memory + orchestration）可以用范畴论的 Architecture triple `(G, Know, Φ)` 形式化：

| 组件 | 含义 | 对应 |
|------|------|------|
| **G** | 语法布线图（模块、端口、有向边） | Protocols / 信息流拓扑 |
| **Know** | 结构性知识/证书（不变量、保证） | 可验证的系统属性 |
| **Φ** | 部署映射（抽象能力槽 → 具体模型/工具） | 模型选择 |

### 四根支柱 → Architecture Triple 映射

| 外化支柱 | 范畴角色 | 具体组件 |
|----------|----------|----------|
| Memory | 余代数状态 (coalgebraic state) | 双时态记忆 |
| Skills | 操纵子组合对象 (operad-composed) | 原子技能模板 |
| Protocols | 语法布线 G | 类型化端口 + 光学 |
| Harness | 完整 Architecture `(G, Know, Φ)` | SkillOrganism |

### 关键贡献

1. **证书保持**：编译器函子将 Architecture 从一个框架映射到另一个，证书（结构保证）通过 identity + replay 验证存活
2. **模型参数化**：结构保证在 `Know` 层，不绑定特定模型。改 `Φ`（换模型）不影响 `Know`（保证）
3. **操纵子组合**：5 个原子编码技能（localize/edit/test/reproduce/review）通过 serial/parallel/trace 三种操作组合，无负干扰
4. **5 个编译器目标**：Swarms、DeerFlow、Ralph、Scion、LangGraph 全部保持 3 种证书类型

### 诚实局限（论文自述）

- 静态框架，不建模 harness 演化
- 单一参考实现（Operon）
- 证书只覆盖结构不变量，不覆盖行为属性（"不幻觉"）
- 升级实验只有 2 模型 1 任务
- SWE-bench-lite 在 8B 模型上 0 解决（格式纪律是瓶颈，不是架构）

---

## 2. 与 Axis 的对照

### 2.1 概念映射

| 论文概念 | Axis 对应 | 差异 |
|----------|-----------|------|
| Architecture triple `(G, Know, Φ)` | Kernel Abstraction Model（syscall + core abstractions + infra） | Axis 不用范畴论，用辩证法 |
| G（语法布线） | AgentContract + DAG scheduler | Axis 用 Contract 定义接口，不用类型化端口 |
| Know（证书） | SelfJudgement strategies + Contract admission | Axis 的"证书"是 admission validator + judgement |
| Φ（部署映射） | Provider profile + `axis provider use` | 几乎完全对应 |
| Memory as coalgebra | Long Horizon Memory（patterns/principles/narrative） | Axis 无双时态，但有遗忘策略 |
| Skills as operad | Skills system（`.axis/skills/`） | Axis 的 skills 是知识注入，不是可组合的执行单元 |
| Compiler functor | 无对应 | **Axis 缺失**：没有"编译到其他框架"的概念 |
| Certificate preservation | 无对应 | **Axis 缺失**：没有形式化的"属性在变换下保持" |
| Model parametricity | Provider 切换不影响 Contract | 已有，但未形式化 |

### 2.2 Axis 已有但论文缺失的

| Axis 概念 | 论文无对应 | 原因 |
|-----------|-----------|------|
| Staged Evolution | 无 | 论文是静态框架，不建模自我修改 |
| Autonomy Transition | 无 | 论文不讨论渐进信任 |
| 辩证方法论（SRS Loop） | 无 | 论文用数学形式化，不用哲学方法 |
| "绝不僭越"原则 | 无 | 论文的 harness 主动做决策（escalation） |
| Event log / 审计 | 无 | 论文关注结构保持，不关注可观测性 |

### 2.3 论文有但 Axis 缺失的

| 论文概念 | Axis 缺失 | 价值评估 |
|----------|-----------|----------|
| **形式化证书系统** | Axis 的 admission/judgement 是代码，不是可验证的形式证书 | 🟡 中等——形式化有理论价值但工程收益不明确 |
| **编译器函子** | Axis 不需要编译到其他框架（单一实现） | ⚪ 低——Axis 是自包含系统 |
| **操纵子组合** | Axis 的 skills 不可组合（只是知识注入） | 🟠 高——如果 skills 能组合为 workflow，价值大 |
| **双时态记忆** | Axis 的 memory 无 valid_time/record_time 区分 | 🟡 中等——"Agent 在决策点 t 知道什么"是有用的问题 |
| **质量驱动升级** | Axis 有 Judge 但没有自动 model escalation | 🟠 高——直接映射到 P2 动态模型路由 |

---

## 3. 可行动的借鉴

### 3.1 质量驱动模型升级（P2 优先级最高）

论文的 escalation 实验：Phi-3 Mini 产出质量 < 阈值 → 自动升级到 Gemma 4。

**映射到 Axis**：
```
CodingAgent 用 deepseek-v4-flash（快/便宜）执行
→ SelfJudgement 评分 < 阈值
→ 自动切换到 claude-3-5-sonnet（慢/贵）重新执行
→ 结构保证：升级逻辑在 Contract 层，不在 Agent 层
```

这正是 Axis P2 #5 "provider 语义分层"的具体实现路径。论文验证了这条路可行。

### 3.2 Skills 可组合化（P2 方向验证）

论文的 5 个原子技能（localize/edit/test/reproduce/review）通过 operad 组合。

**对 Axis 的启示**：当前 `.axis/skills/` 只是知识注入（Layer 1 metadata + Layer 2 load）。如果 skills 能声明输入/输出类型并组合为 pipeline，就能实现：

```
skill: localize → skill: edit → skill: test
```

这与 CodingAgent 的 Phase 1→2→3→4 流程完全对应。但当前不需要行动——CodingAgent spec 已经定义了这个流程，只是没用 operad 形式化。

### 3.3 双时态记忆（记录但不行动）

"Agent 在决策点 t 知道什么"是一个有用的调试问题。当前 Axis 的 memory 只有 `fetched_at`（record time），没有 `valid_time`。

**评估**：当前阶段不需要。等 Axis 有多 Agent 协作 + 跨会话状态追踪时再考虑。记录为 aspirational。

---

## 4. 哲学层面的分歧

### 论文的方法论：数学形式化

- 用范畴论证明属性保持
- 证书是可机械重放的形式证明
- 组合通过类型系统保证安全

### Axis 的方法论：辩证否定

- 用实践反馈证明规则有效
- "证书"是 admission validator + go test
- 组合通过 Contract 边界 + 语义边界保证安全

**判断**：两种方法论解决不同问题。

- 论文解决的是**可移植性**（同一 harness 编译到不同框架）——Axis 不需要这个
- Axis 解决的是**可演化性**（harness 自我修改后仍然安全）——论文明确说不做这个

两者互补，不冲突。Axis 不需要引入范畴论，但可以借鉴"证书"概念让 SelfJudgement 更结构化。

---

## 5. 结论

| 维度 | 评估 |
|------|------|
| 对 Axis 的威胁 | ⚪ 无——解决不同问题（可移植性 vs 可演化性） |
| 可借鉴的工程模式 | 🟠 质量驱动模型升级（P2 直接可用） |
| 可借鉴的设计方向 | 🟡 Skills 可组合化（验证方向正确，不需要立即行动） |
| 需要引入的理论 | ⚪ 无——范畴论对 Axis 当前阶段是 over-engineering |
| 值得持续关注 | ✅ 是——Operon 框架的后续发展可能产出可复用的编译器模式 |

**一句话**：这篇论文用范畴论做了 Axis 用辩证法做的同一件事——形式化 Agent Harness。方法不同，结论互相验证。Axis 的"Contract is Structure"≈ 论文的"Know-level certificates"。Axis 的"Provider profile switching"≈ 论文的"Φ deployment map"。最大借鉴点是质量驱动模型升级的具体实现路径。
