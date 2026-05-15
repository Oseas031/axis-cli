# mattpocock/skills：Agent Skill 设计哲学研究

> 基于 https://github.com/mattpocock/skills — 82.5k stars, MIT License

## 1. 核心机制

mattpocock/skills 是一套面向 Claude Code 的工程 skill 集合，解决四个 Agent 编码失败模式：

1. **对齐失败** → Grilling（拷问式对齐）
2. **过度冗长** → CONTEXT.md（共享语言/术语表）
3. **代码不工作** → TDD + Diagnose（反馈循环优先）
4. **架构腐化** → Deep Modules + Architecture Review

安装机制：`npx skills@latest add mattpocock/skills`（skills.sh CLI），写入 `.claude/` 目录。

## 2. 关键设计决策

### 2.1 SKILL.md 格式

```yaml
---
name: skill-name
description: 一句话描述 + "Use when [触发条件]"
disable-model-invocation: true  # 可选：只注入 prompt 不触发
---
```

- description 是 Agent 选择 skill 的唯一依据（≤1024 chars）
- 正文是 markdown 指令，Agent 按指令执行
- 可附带子文件（REFERENCE.md, EXAMPLES.md, scripts/）

### 2.2 CONTEXT.md 共享语言

核心创新：项目级术语表，不是 spec，不是 scratch pad，**纯粹是 glossary**。

格式：
- Language 段：定义术语（名称 + 定义 + 避免用词）
- Relationships 段：术语间关系
- Flagged ambiguities 段：已解决的歧义

效果：
- 变量/函数/文件命名一致
- Agent 思考 token 减少（用精确术语替代冗长描述）
- 代码库更易导航

### 2.3 ADR（Architecture Decision Records）

只在三个条件同时满足时创建：
1. 难以逆转
2. 没有上下文会让未来读者困惑
3. 是真正的权衡（有替代方案）

### 2.4 Vertical Slices（纵向切片）

to-issues 将计划拆分为端到端薄切片，每个切片：
- 穿透所有集成层（schema → API → UI → tests）
- 独立可验证/可演示
- 标记为 HITL（需人类）或 AFK（Agent 可独立完成）

### 2.5 Deep Modules 哲学

来自 John Ousterhout《A Philosophy of Software Design》：
- **深模块** = 小接口 + 大实现 = 高杠杆
- **浅模块** = 接口几乎和实现一样复杂 = 低价值
- **Deletion Test**：删除模块后复杂度消失 → 它是 pass-through；复杂度散布到 N 个调用者 → 它在创造价值
- **接口即测试面**

### 2.6 Feedback Loop First（diagnose skill）

Debug 方法论的核心洞察：**先建反馈循环，再修 bug**。

10 种构建反馈循环的方式（按优先级）：
1. 失败测试
2. curl/HTTP 脚本
3. CLI + fixture + diff
4. 无头浏览器
5. 重放捕获的 trace
6. Throwaway harness
7. Property/fuzz loop
8. Bisection harness
9. Differential loop
10. HITL bash script

"Build the right feedback loop, and the bug is 90% fixed."

## 3. 对 Axis 的启示

### 当前状态（Axis 有什么）

- CLAUDE.md 宪法（比 mattpocock 的 CLAUDE.md 重得多）
- `.axis/skills/` 目录 + SKILL.md 格式（兼容）
- Phase II 对齐机制（对应 Grilling）
- semantic-boundaries.md（对应 Deep Modules 边界思想）
- vigil 跨会话追踪（对应 handoff 但更重）

### mattpocock 做了什么不同的

| 维度 | mattpocock | Axis |
|------|-----------|------|
| 复杂度 | 极简（每个 skill <100 行） | 重型（宪法 + 方法论 + 治理） |
| 共享语言 | CONTEXT.md 独立文件 | 散布在多个 docs 中 |
| ADR | 轻量、三条件门控 | 无独立 ADR 机制 |
| Debug | 结构化 6 阶段 + 反馈循环优先 | 无专门 debug skill |
| 跨会话 | handoff（一次性文档） | vigil（持久化追踪） |
| 架构审查 | 定期 skill 触发 | 无定期审查机制 |
| Token 节省 | caveman mode | 无 |

### 可借鉴

1. **CONTEXT.md 理念**：Axis 应有独立的项目术语表（不是 CLAUDE.md 的一部分）
2. **diagnose skill**：直接安装，补充 Axis 工程实践
3. **Deep Modules 语言**：improve-codebase-architecture 的术语体系（Module/Interface/Implementation/Depth/Seam/Adapter）可直接用于 Axis 架构审查
4. **caveman mode**：token 节省在长会话中价值巨大
5. **prototype skill**：throwaway code 理念对 Axis 探索性工作有价值
6. **Vertical Slices**：to-issues 的切片方法可用于 Axis 任务分解

### 不能借鉴

1. **setup-matt-pocock-skills**：Axis 有自己的 skill 管理机制
2. **triage**：Axis 不用 GitHub Issues 做工作追踪（用 vigil）
3. **to-prd / to-issues**：Axis 用 Spec-RDT 而非 PRD/Issue 驱动
4. **grill-with-docs 的 CONTEXT.md 写入**：Axis 的 CLAUDE.md 有严格修改协议（A8 写回）

## 4. 可行动建议

| 优先级 | 行动 | 模块 |
|--------|------|------|
| P1 | 安装 diagnose/tdd/improve-codebase-architecture/caveman/prototype/handoff 到 `.axis/skills/` | skills |
| P1 | 适配 skill 内容为 Axis 语境（替换 GitHub Issues → vigil，CONTEXT.md → CLAUDE.md §语义边界） | skills |
| P2 | 创建 Axis 项目级 CONTEXT.md（纯术语表，从 CLAUDE.md 中提取） | docs |
| P2 | 引入 ADR 机制到 `docs/adr/`（三条件门控） | docs |
| P3 | 定期架构审查 workflow（每周触发 improve-codebase-architecture） | workflow |
