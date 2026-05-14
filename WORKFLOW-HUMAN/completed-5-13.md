# 已完成工作：2026-05-13

## P0：已完成

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

### TodoWrite 增强

- [x] Nag 提醒机制：多轮执行中超过 5 轮无 checkpoint/store_memory 时，System Prompt 追加软提醒

### 记忆系统：Long Horizon 实现

- [x] 目录结构 + markdown frontmatter 格式
- [x] `recall_memory` / `store_memory` tools（Agent 驱动，Zero Control）
- [x] Dream 回放引擎（聚类失败 → 蒸馏 pattern → 去重写入）
- [x] 遗忘策略（narrative >7d 归档，>30d 删除；patterns/principles 永不删除）
- [x] Principles 注入 System Prompt（Layer 1，零检索成本）
- [x] 按需检索（Agent 主动调用 recall_memory）
- [x] 垃圾记忆过滤（title 前缀去重 + dream 写入前检查）
- [x] 免疫分层（L1 immunity → L2 patterns recall → L3 dream+人类）

### 结构性冲突审计（已修复部分）

- [x] fatal+retries 矛盾 → admission 显式拒绝
- [x] degradable success override → 移除，系统不替 Agent 判断成功语义
- [x] fabricated ErrTaskTimeout → 新增 ErrDispatchFailed 准确报告

### CodingAgent 第一性原则

- [x] `docs/specs/coding-agent/first-principles.md` 立稿
- [x] 5 原则：Conjecture / Refutation / Explore-Think-Act / Decompose / Failure=Info
- [x] 三层哲学硬编码（prompt / process / format）
- [x] CLAUDE.md rules #7 #8（subagent 验收 + v1 简化标记）

### 宪法层 / 理论基础层重构

- [x] §13 矛盾治理框架（稳定/发展分层 + 横向仲裁 + 理论实践预判）
- [x] 方法论理论基础重写（从 Axis 实践提炼，非毛选引用）
- [x] 哲学四层（ontology + axiology + 历史观 + 逻辑）
- [x] L1/L2/L3 纵向贯通审计 + 死内容删除 + aspirational 标记
- [x] 认识论：fiction as premature objectification
- [x] Phase 声明触发条件明确化 + bypass 记录协议
- [x] 3 条补充原则 + L1 权威层级 + L3 死内容修复

---

## Git 提交记录（5/13）

| Commit | 内容 |
|--------|------|
| `cad3f73` | docs(epistemology): fiction as premature objectification + aspirational directions |
| `0d1e3d2` | refactor(L2→L3): add 3 supplementary principles + L1 authority hierarchy + fix L3 dead content |
| `5f5c4e6` | refactor(L2): deep audit-driven reform — delete dead, strip duplicates, mark aspirational |
| `19694b7` | refactor(L2→L3): top-down philosophical alignment — no level-skipping |
| `9c53346` | docs(epistemology): correct over-humility — theory IS practice crystallization |
| `61bed18` | docs(negation²): self-audit theory foundation — 5 contradictions resolved |
| `c7e89c7` | docs(philosophy): ontology + axiology + historical view + logic — from Axis practice |
| `cd940bd` | docs(governance): clarify Phase trigger condition + bypass recording protocol |
| `b623c5d` | docs(negation²): harness self-audit via its own rules — 4 contradictions resolved |
| `11409db` | docs(methodology): rewrite theory foundation from Axis practice, not Mao quotes |
| `44ae349` | docs(methodology): 理论基础层 — 毛选五大原理 + 溯源声明 + 三不原则 + 术语映射 |
| `49c5ca3` | feat(governance): §13 矛盾治理框架 — 稳定/发展分层 + 横向仲裁 + 理论实践预判 |
| `feb2151` | refactor(harness): CLAUDE.md as single authority with bootstrap |
| `ddd77c2` | docs: CodingAgent first principles + CLAUDE.md rules #7 #8 |
| `6c84dfe` | feat(executor): soft nag reminder after 5 turns without progress update |
| `685965b` | docs: cascade dialectical methodology + mark Long Horizon complete |
| `b80c6d3` | feat(memory): garbage dedup — pattern deduplication + dream skip-on-existing |
