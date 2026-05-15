# LLM Wiki 模式研究：Axis 文档知识库设计

> 研究日期：2026-05-15
> 来源：[Karpathy llm-wiki.md](https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f) + [nashsu/llm_wiki](https://github.com/nashsu/llm_wiki)
> 目标：将 Axis 文档系统搭建为 Agent-native 知识库

## 1. 核心洞察

Karpathy 的 LLM Wiki 模式解决了一个根本问题：**RAG 每次从零推导，知识不积累**。

替代方案：LLM 增量构建和维护一个持久化 wiki——结构化、互链的 Markdown 文件集合。知识编译一次，持续维护，而非每次查询重新推导。

关键隐喻：**Obsidian 是 IDE，LLM 是程序员，Wiki 是代码库。**

## 2. 三层架构

| 层 | Karpathy 原始 | nashsu 实现 | Axis 映射 |
|---|---|---|---|
| **Raw Sources** | 不可变源文档 | `raw/sources/` + 多格式解析 | 源码 `internal/` + `cmd/` |
| **Wiki** | LLM 生成的 Markdown | `wiki/` (entities/concepts/sources/synthesis) | `docs/` (architecture/specs/research/status) |
| **Schema** | 告诉 LLM 如何维护 wiki 的配置 | `schema.md` + `purpose.md` | `CLAUDE.md` + Skills |

**Axis 的独特之处**：三层已经存在，但缺少将它们作为"知识库"来运营的意识和工具。

## 3. 三个操作

### Ingest（摄入）
- **原始**：读源文档 → 写摘要 → 更新 index → 更新实体/概念页 → 追加 log
- **nashsu 增强**：两步 CoT（先分析再生成）、SHA256 增量缓存、持久化队列
- **Axis 适配**：新代码/新 spec 完成后，Agent 更新相关文档页面、交叉引用、进度状态

### Query（查询）
- **原始**：读 index → 找相关页 → 综合回答 → 好答案存回 wiki
- **nashsu 增强**：分词搜索 + 图扩展 + 向量搜索 + 预算控制
- **Axis 适配**：Agent 通过 `docs/README.md` 导航 → 读相关文档 → 回答。Skills 系统已支持按需加载。

### Lint（健康检查）
- **原始**：检测矛盾、过时声明、孤立页面、缺失交叉引用、知识空白
- **nashsu 增强**：Review 系统（异步人类审核）
- **Axis 适配**：**这是 Axis 最缺的能力**。文档间一致性无人维护。

## 4. 关键设计元素评估

### 4.1 采纳（精华）

**index.md 优化**
- 当前 `docs/README.md` 是人类导向的链接列表
- 需要：LLM-optimized 版本，每个条目带一行摘要 + 元数据标签
- 价值：Agent 读一个文件就能定位整个知识库

**purpose.md（方向意图）**
- llm_wiki 区分 schema（结构规则）和 purpose（为什么存在）
- Axis 有 CLAUDE.md（schema）但缺少 purpose 层
- 价值：声明文档系统的目标、关键问题、演化方向

**YAML frontmatter**
- 每个文档顶部添加结构化元数据：type、status、depends_on、last_verified
- 价值：Agent 可以程序化查询文档状态，支持 Lint 操作

**Lint 健康检查**
- 定期检测：过时声明、孤立文档、缺失引用、spec 与实现不一致
- 价值：文档系统自我维护，不依赖人类记忆

**Wikilink 交叉引用**
- 文档间使用 `[[target]]` 或显式相对路径互链
- 价值：知识网络化，Agent 可以遍历关联

### 4.2 扬弃（糟粕）

| 特性 | 扬弃理由 |
|------|----------|
| 桌面 GUI (Tauri/React) | Axis CLI-first 原则。axis-gui 已是 Observatory。 |
| 向量数据库 (LanceDB) | 引入外部依赖。Axis 文档规模（~100文件）不需要。index.md + grep 足够。 |
| Chrome Extension | 不符合 CLI-first。Axis 有 research-pipeline skill。 |
| 4信号知识图谱 | 过度工程化。简单 frontmatter `related: []` 字段足够。 |
| Louvain 社区检测 | 文档规模不需要自动聚类。目录结构已是人工聚类。 |
| 多格式文档解析 | Axis 全 Markdown，不需要 PDF/DOCX 解析。 |
| 持久化 Ingest 队列 | Axis Agent 单次处理，不需要桌面应用的后台队列。 |

### 4.3 Axis 独有优势（不需要从 llm_wiki 学的）

| Axis 已有 | 对应 llm_wiki 缺失 |
|---|---|
| CLAUDE.md 宪法 | llm_wiki 的 schema.md 远不如此强大 |
| Skills 可组合系统 | llm_wiki 是固定 pipeline |
| Vigil 跨会话追踪 | llm_wiki 只有 log.md |
| Spec-RDT 生命周期 | llm_wiki 文档无契约性质 |
| Staged Evolution Protocol | llm_wiki 无变更安全机制 |
| 辩证方法论 (SRS Loop) | llm_wiki 无方法论约束 |

## 5. 设计方案：Axis 文档知识库

### 5.1 架构映射

```
axis-cli/
├── CLAUDE.md                    # Schema 层（宪法 + 规则）
├── docs/
│   ├── PURPOSE.md               # [新增] 知识库方向意图
│   ├── README.md                # index.md 角色（优化为 LLM-friendly）
│   ├── CHANGELOG.md             # [新增] log.md 角色（文档变更时间线）
│   ├── architecture/            # Wiki 层：概念页
│   ├── specs/                   # Wiki 层：契约页
│   ├── research/                # Wiki 层：研究页
│   ├── status/                  # Wiki 层：状态页
│   └── guides/                  # Wiki 层：指南页
├── .axis/skills/
│   └── docs-knowledge-base/     # [新增] 知识库维护 Skill
│       └── SKILL.md
└── internal/ + cmd/             # Raw Sources 层
```

### 5.2 新增文件设计

**docs/PURPOSE.md** — 知识库的灵魂
```markdown
# Axis Documentation Purpose

## 目标
为 Axis 项目的所有参与者（人类 + Agent）提供可查询、可验证、自维护的知识基础设施。

## 关键问题
- Axis 的设计决策为什么这样做？（architecture/）
- 当前实现到什么程度？（status/）
- 下一步该做什么？（specs/ + vigil）
- 外部研究如何影响 Axis？（research/）

## 演化方向
从被动文档 → 主动知识库：Agent 不只读文档，还维护文档。
```

**docs/CHANGELOG.md** — 文档变更时间线
```markdown
# Documentation Changelog

## [2026-05-15] research | LLM Wiki Knowledge Base Design
- Added: docs/research/llm-wiki-knowledge-base-2026-05-15.md
- Added: docs/PURPOSE.md
- Added: .axis/skills/docs-knowledge-base/SKILL.md
```

**YAML Frontmatter 规范**
```yaml
---
type: architecture | spec | research | guide | status
status: active | deprecated | draft
created: 2026-05-15
last_verified: 2026-05-15
related:
  - architecture/agent-native-first-principles.md
  - specs/skills-system/requirements.md
tags: [knowledge-base, design-pattern]
---
```

### 5.3 Lint 操作设计

`axis docs lint` 检查项：
1. **孤立文档**：无入链的文档（不在 README.md 中，不被其他文档引用）
2. **过时声明**：`last_verified` 超过 30 天的文档
3. **Spec-实现不一致**：spec 标记 Completed 但代码不存在，或代码存在但 spec 未更新
4. **缺失交叉引用**：提到其他模块但未链接
5. **README.md 同步**：新文档未出现在 index 中

### 5.4 Ingest 操作设计

触发时机：
- 新 spec 创建 → 更新 README.md + CHANGELOG.md
- 代码完成 → 更新 status/current-progress.md + spec tasks.md
- 研究完成 → 更新 README.md + CHANGELOG.md + 相关 architecture 文档

### 5.5 实现为 Skill（不是代码）

关键决策：**知识库维护是 Skill，不是 Go 代码**。

理由：
- Axis 文档规模（~100文件）不需要程序化搜索引擎
- Agent 直接读 Markdown 比调用 API 更自然
- Skill 可以被任何 Agent 工具（Claude Code、Kiro、Codex）使用
- 符合 "bash is all you need" 原则

## 6. 与社区实现的对比

Gist 评论区的其他实现提供了有价值的视角：

| 项目 | 关键洞察 | Axis 适用性 |
|------|----------|-------------|
| Synthadoc | Routing layer + Candidates staging | 文档规模不需要路由，但 candidates 概念类似 spec Draft 状态 |
| Kompl | NLP before LLM（先 NER 再综合） | Axis 不需要，Agent 直接读 Markdown |
| jazzonenl 批判 | 事务性开销、链接完整性、时间退化 | 有效警告。Axis 用 git + frontmatter 缓解。 |
| nohmitaina | Identity/Level/Relationship 三问题 | Axis 用目录结构解决 Level，用 frontmatter 解决 Identity |
| Cogentia | Resumable judgment（可恢复判断点） | 类似 Axis 的 Review 系统 / Vigil |

## 7. 风险与缓解

| 风险 | 缓解 |
|------|------|
| frontmatter 维护负担 | Lint skill 自动检测缺失/过时 |
| CHANGELOG.md 被遗忘 | 纳入 Phase III 退出门禁 |
| 文档膨胀 | 定期 Lint 标记低价值文档 → deprecated/ |
| Agent 生成低质量文档 | Self-Judgement + 人类 Review |

## 8. 结论

Karpathy 的 LLM Wiki 模式的核心价值是**知识编译**思想——从"每次查询重新推导"转向"增量构建持久知识"。Axis 已经有了三层架构的骨架，缺的是：

1. **意识**：将文档系统视为知识库来运营
2. **工具**：Lint 健康检查、CHANGELOG 时间线
3. **规范**：frontmatter 元数据、PURPOSE 声明
4. **习惯**：每次工作结束更新知识库（已在 CLAUDE.md §0 rule #4 要求）

实现路径：添加 3 个文件（PURPOSE.md、CHANGELOG.md、docs-knowledge-base SKILL），优化 1 个文件（README.md），建立 frontmatter 规范。无需新代码。
