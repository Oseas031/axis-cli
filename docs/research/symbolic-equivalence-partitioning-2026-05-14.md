# Symbolic Equivalence Partitioning：对 Axis CodingAgent 的启示

> 基于 arXiv:2604.06485 — David Cho, Yifan Wang, Fanping Sui, Ananth Grama (2026-04)

## 1. 核心机制（≤10行）

SEP 解决的问题：从 LLM 生成的 N 个候选程序中选出正确解。
核心观察：正确解尽管语法/算法不同，在合法输入上收敛到相同函数行为；错误解也可能形成共识（correlated wrong）。
流程：
1. 用题目公开样例做轻量有效性过滤（淘汰明显错误）
2. 对剩余候选用**符号执行**计算有界函数等价类（bounded functional equivalence classes）
3. 选**主导等价类**（dominant class）中的解作为最终输出

无需辅助测试生成、无需 learned verifier、无需额外 LLM 推理。

## 2. 关键数据

| 基准 | Baseline (N=10) | SEP (N=10) | 提升 |
|------|-----------------|------------|------|
| HumanEval+ | 0.754 | 0.826 | +7.2pp |
| LiveCodeBench | 0.565 | 0.647 | +8.2pp |

- 仅依赖符号执行分区，不引入额外模型调用
- 对比 majority voting / test-based filtering 均有显著优势

## 3. 对 Axis 的启示

- **P2 (Refutation) 原则对齐**：SEP 的等价类分区本质是"用结构化方法暴露行为差异"——不是证明对了，而是把"行为不同"的候选分开，让错误解自我暴露
- **比 verify_bash 更强**：Axis 当前 verify_bash 只能判"通过/不通过"，SEP 提供了"N 个候选之间的结构化比较"维度
- **可借鉴**：多候选生成 → 行为等价类分区 → 选主导类。Axis 可用差分测试（differential testing）近似此效果，无需符号执行引擎
- **不能借鉴**：符号执行引擎（如 KLEE/angr）是重依赖，与 Axis "bash is all you need" 原则冲突，不引入

## 4. 可行动建议

| 优先级 | 模块 | 建议 |
|--------|------|------|
| P0 | `internal/agent` | CodingAgent 生成 N≥3 候选解（temperature sampling），而非单次生成 |
| P1 | `internal/agent/judge` | 新增 DifferentialJudge 策略：对候选解跑相同随机输入集，按输出分组为等价类，选主导类 |
| P2 | `internal/contract` | 扩展 AgentContract 支持 `candidates_min: N` 声明，触发多候选流程 |
| P3 | `internal/model` | 复用现有 provider 的 temperature/n 参数，单次 API 调用获取多候选（降低延迟） |

核心收益：用 O(N) 次 bash 执行替代符号执行，获得等价类分区的大部分选择增益。
