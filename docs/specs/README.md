# Specs Index

> **收敛约束（2026-11-15）**：冻结新增独立 Spec 的横向扩张，优先跑通并验证现有核心闭环（详见 `CLAUDE.md` §0 & §5）。

每个 spec 含 requirements.md + design.md + tasks.md。

| Spec | 状态 | 一句话 |
|------|------|--------|
| [coding-agent](coding-agent/) | In Progress | 多候选生成 + 差分测试 |
| [actor-comm](actor-comm/) | In Progress | 多 Agent JSONL mailbox 通信 |
| [vigil](vigil/) | Completed | 跨会话工作追踪 |
| [cli-execution-semantics](cli-execution-semantics/) | In Progress | run/ask/shell 语义统一 |
| [context-compaction](context-compaction/) | Planned | 上下文压缩策略 |
| [skills-system](skills-system/) | Completed | 按需知识技能加载 |
| [immunity-memory](immunity-memory/) | Completed | 免疫记忆（错误模式学习） |
| [layered-memory-model](layered-memory-model/) | Completed | 分层记忆架构 |
| [model-provider](model-provider/) | Completed | LLM provider + 语义分层 |
| [sandboxed-evolution](sandboxed-evolution/) | Completed | 隔离工作区 + 验证门控 |
| [local-control-plane](local-control-plane/) | Completed | 本地运行时 + 跨进程通信 |
| [natural-language-scheduling](natural-language-scheduling/) | Completed | NL → AgentTask |
| [interactive-shell](interactive-shell/) | Completed | axis shell 交互会话 |
| [adaptive-context-assembly](adaptive-context-assembly/) | Completed | 自适应上下文组装 |
| [execution-context-consumption](execution-context-consumption/) | Completed | 执行时上下文消费 |
| [human-agent-scheduling](human-agent-scheduling/) | In Progress | 人机互调度：双向任务流对称执行者 |
| [cost-budget](cost-budget/) | Completed | 任务级 USD 预算 + 自动降级 |
| [swarm-topology](swarm-topology/) | Completed | Swarm 多候选并行投票 |
| [context-recommendation](context-recommendation/) | Planned | 上下文推荐系统（design 先行） |
