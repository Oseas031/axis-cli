# Context Recommendation System — Design Document

> 实现 docs/architecture/agent-native-first-principles.md Context 提升行动质量原则

---

## 1. 概述

### 1.1 背景

当前 axis-cli 的上下文装配存在以下结构性问题：

| 问题 | 影响 | 根因 |
|------|------|------|
| ContextBundle 不注入 prompt | 上下文与执行路径分离 | 设计上 ContextBundle 仅做 preview |
| 四套独立预算系统 | 预算分配碎片化，无法全局优化 | 历史演进缺乏统一抽象 |
| Token 估算不一致 | 预算计算不可靠 | 无统一 tokenizer |
| 规则集极小（仅 6 条硬编码） | 无法适应多样化场景 | 规则与代码耦合 |
| TF-IDF 粒度粗（不分块） | 相关性评估精度低 | 无文档分块能力 |
| 中文支持差 | 中文文档召回率低 | tokenizer 未适配 CJK |
- 质量评估缺失 | 无法量化上下文装配效果 | 无反馈闭环

### 1.2 目标

引入上下文推荐系统，实现：

1. **多策略排序**：Relevance + Importance + Freshness 加权排序
2. **文档分块**：按语义边界切分长文档，提升召回精度
3. **配置化规则**：规则集外置为 YAML，支持运行时加载
4. **统一 Token 估算**：单一 tokenizer 覆盖所有预算计算
5. **质量反馈闭环**：量化评估上下文装配效果，驱动持续优化

### 1.3 非目标

- 不改变 ContextBundle 的 preview 定位（§1 绝对禁令：禁止 push-based context injection）
- 不引入外部依赖（§7 代码与架构风格）
- 不修改 scheduler/contract 语义（§6 语义边界）

---

## 2. 架构设计

### 2.1 整体架构

```
┌─────────────────────────────────────────────────────────┐
│                    Intent Parser                         │
│                  (用户意图 → 结构化查询)                   │
└──────────────────────┬──────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────┐
│              Candidate Generator                         │
│   (从 memory / docs / skills / vigil 收集候选上下文)       │
└──────────────────────┬──────────────────────────────────┘
                       │ candidates[]
                       ▼
┌─────────────────────────────────────────────────────────┐
│                    Ranker                                │
│          (多策略排序：Relevance + Importance + Freshness)  │
│                                                         │
│  ┌──────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ Chunker  │→ │ RuleEngine   │→ │ QualityAssessor│     │
│  │ (文档分块) │  │ (配置化规则)   │  │ (质量评估)     │     │
│  └──────────┘  └──────────────┘  └──────────────┘      │
└──────────────────────┬──────────────────────────────────┘
                       │ ranked candidates[]
                       ▼
┌─────────────────────────────────────────────────────────┐
│              Budget Allocator                            │
│         (统一 Token 估算 + 预算分配)                       │
└──────────────────────┬──────────────────────────────────┘
                       │ contextBundle
                       ▼
┌─────────────────────────────────────────────────────────┐
│                  ContextBundle                           │
│             (preview only, 不注入 prompt)                 │
└─────────────────────────────────────────────────────────┘
```

### 2.2 设计原则

1. **Pipeline 模式**：每个阶段独立、可测试、可替换
2. **配置驱动**：排序策略、分块参数、规则集均通过配置控制
3. **质量闭环**：每次装配结果可评估，评估结果驱动参数调整
4. **零控制**：系统推荐上下文，Agent 自主决定是否采纳

### 2.3 模块边界

| 模块 | 职责 | 不得做 |
|------|------|--------|
| Ranker | 对候选上下文排序 | 不修改上下文内容 |
| Chunker | 按语义边界切分文档 | 不决定分块是否被使用 |
| RuleEngine | 根据规则过滤/加权候选 | 不执行规则外的逻辑 |
| QualityAssessor | 评估装配质量 | 不修改装配结果 |
| BudgetAllocator | 分配 token 预算 | 不决定内容优先级 |

---

## 3. 核心组件

### 3.1 Ranker（多策略排序器）

#### 职责

对候选上下文进行多维度评分和排序。

#### 排序策略

| 策略 | 信号 | 权重（默认） | 说明 |
|------|------|-------------|------|
| Relevance | TF-IDF / BM25 余弦相似度 | 0.5 | 与查询的语义相关性 |
| Importance | 文档访问频率、引用次数 | 0.3 | 文档的历史重要性 |
| Freshness | 最后修改时间 | 0.2 | 时间衰减因子 |

#### 评分公式

```
score(c) = w_r × relevance(c, query) + w_i × importance(c) + w_f × freshness(c)
```

#### 接口

```go
type Ranker struct {
    chunker      *Chunker
    ruleEngine   *RuleEngine
    config       RankerConfig
}

type RankerConfig struct {
    Weights      StrategyWeights `yaml:"weights"`
    MaxCandidates int            `yaml:"max_candidates"`
}

type StrategyWeights struct {
    Relevance  float64 `yaml:"relevance"`
    Importance float64 `yaml:"importance"`
    Freshness  float64 `yaml:"freshness"`
}

func (r *Ranker) Rank(candidates []Candidate, query Query) ([]RankedCandidate, error)
```

#### 实现细节

1. **归一化**：各策略分数归一化到 [0, 1]
2. **权重校验**：`w_r + w_i + w_f = 1.0`，否则返回 error
3. **空值处理**：missing signal → 默认分数 0.5（中性）

### 3.2 Chunker（文档分块器）

#### 职责

将长文档按语义边界切分为可独立评估的块。

#### 分块策略

| 策略 | 适用场景 | 分块边界 |
|------|---------|---------|
| Heading | Markdown / 结构化文档 | `#`, `##`, `###` 标题 |
| Paragraph | 纯文本 | 双换行 |
| Fixed-size | 无结构文本 | 固定 token 数 |
| Semantic | 通用 | 句子边界 + 话题转换 |

#### 接口

```go
type Chunker struct {
    config ChunkerConfig
}

type ChunkerConfig struct {
    Strategy    string `yaml:"strategy"`    // heading | paragraph | fixed | semantic
    MaxTokens   int    `yaml:"max_tokens"`  // 分块最大 token 数
    Overlap     int    `yaml:"overlap"`     // 重叠 token 数（用于上下文衔接）
}

type Chunk struct {
    Content   string
    StartLine int
    EndLine   int
    Metadata  map[string]string  // namespace.chunk.*
}

func (c *Chunker) Chunk(document Document) ([]Chunk, error)
```

#### 分块边界规则

1. 永远不在句子中间切分
2. Heading 策略：标题行归入其后内容块
3. Overlap 确保上下文衔接：前一块末尾 `overlap` token 附加到后一块开头
4. 空块丢弃（不产生零内容候选）

### 3.3 RuleEngine（配置化规则引擎）

#### 职责

根据配置规则过滤和加权候选上下文。

#### 规则类型

| 类型 | 操作 | 示例 |
|------|------|------|
| Include | 包含匹配路径 | `paths: ["docs/architecture/**"]` |
| Exclude | 排除匹配路径 | `paths: ["**/*_test.go"]` |
| Boost | 提升权重倍数 | `paths: ["CLAUDE.md"]`, `factor: 2.0` |
| Demote | 降低权重倍数 | `paths: ["**/vendor/**"]`, `factor: 0.1` |
| Pin | 固定包含（不参与排序） | `paths: ["docs/specs/*/requirements.md"]` |

#### 规则文件格式

```yaml
# .axis/context-rules.yaml
version: 1
rules:
  - name: "core-docs"
    type: boost
    paths:
      - "docs/architecture/**"
      - "CLAUDE.md"
    factor: 1.5

  - name: "exclude-vendor"
    type: exclude
    paths:
      - "**/vendor/**"
      - "**/node_modules/**"

  - name: "pin-requirements"
    type: pin
    paths:
      - "docs/specs/*/requirements.md"
      - "docs/specs/*/design.md"

  - name: "deprioritize-old"
    type: demote
    paths:
      - "docs/archive/**"
    factor: 0.3
```

#### 接口

```go
type RuleEngine struct {
    rules []Rule
    mu    sync.RWMutex
}

type Rule struct {
    Name    string   `yaml:"name"`
    Type    string   `yaml:"type"`    // include | exclude | boost | demote | pin
    Paths   []string `yaml:"paths"`
    Factor  float64  `yaml:"factor"`  // 仅 boost/demote 使用
}

type RuleConfig struct {
    FilePath string `yaml:"file_path"` // .axis/context-rules.yaml
}

func (re *RuleEngine) Load(config RuleConfig) error
func (re *RuleEngine) Apply(candidates []Candidate) ([]Candidate, error)
func (re *RuleEngine) Reload() error  // 热重载
```

#### 规则求值顺序

1. **Pin** 规则最先执行：固定包含的候选直接进入结果集
2. **Exclude** 规则其次：排除匹配候选
3. **Include** 规则：过滤仅保留匹配候选（如果存在 include 规则）
4. **Boost/Demote** 规则：调整权重倍数

### 3.4 QualityAssessor（质量评估器）

#### 职责

评估上下文装配效果，提供反馈闭环。

#### 评估维度

| 维度 | 指标 | 权重 | 说明 |
|------|------|------|------|
| Coverage | 查询覆盖率 | 0.3 | 候选是否覆盖查询的所有关键概念 |
| Precision | 精度 | 0.3 | 候选中相关部分占比 |
| Efficiency | 效率 | 0.2 | token 使用效率（有效信息 / 总 token） |
| Diversity | 多样性 | 0.2 | 候选来源的多样性（避免单一来源） |

#### 评估时机

1. **装配后**：ContextBundle 生成后立即评估
2. **执行后**：Agent 使用上下文执行任务后评估（需 Agent 反馈）
3. **批量评估**：定期对历史装配结果进行批量评估

#### 接口

```go
type QualityAssessor struct {
    config QualityConfig
}

type QualityConfig struct {
    Weights         QualityWeights `yaml:"weights"`
    EnableFeedback  bool           `yaml:"enable_feedback"`
    HistoryLimit    int            `yaml:"history_limit"`
}

type QualityWeights struct {
    Coverage  float64 `yaml:"coverage"`
    Precision float64 `yaml:"precision"`
    Efficiency float64 `yaml:"efficiency"`
    Diversity float64 `yaml:"diversity"`
}

type QualityScore struct {
    Overall   float64            `yaml:"overall"`
    Breakdown map[string]float64 `yaml:"breakdown"`
    Timestamp time.Time          `yaml:"timestamp"`
}

type Feedback struct {
    AssemblyID string
    Useful     bool
    Comments   string
}

func (qa *QualityAssessor) Assess(assembly ContextAssembly, result ExecutionResult) (*QualityScore, error)
func (qa *QualityAssessor) RecordFeedback(feedback Feedback) error
func (qa *QualityAssessor) History(limit int) ([]QualityScore, error)
```

---

## 4. 数据流

### 4.1 主流程

```
用户意图
    │
    ▼
Intent Parser ──→ 结构化查询 (Query)
    │
    ▼
Candidate Generator
    │  ├── memory.Recall(query)
    │  ├── docs.Search(query)
    │  ├── skills.Match(intent)
    │  └── vigil.ActiveItems()
    │
    ▼
candidates[] ──→ RuleEngine.Apply() ──→ filtered candidates[]
    │
    ▼
Chunker.Chunk() ──→ chunks[]（长文档分块）
    │
    ▼
Ranker.Rank()
    │  ├── relevanceScore = TF-IDF(query, chunk)
    │  ├── importanceScore =访问频率(chunk)
    │  ├── freshnessScore = 时间衰减(chunk)
    │  └── finalScore = Σ(w_i × score_i)
    │
    ▼
ranked candidates[]
    │
    ▼
BudgetAllocator
    │  ├── totalBudget = config.max_tokens
    │  ├── usedBudget = 0
    │  └── for c in ranked:
    │       if usedBudget + c.tokens <= totalBudget:
    │           add to bundle
    │           usedBudget += c.tokens
    │
    ▼
ContextBundle
    │
    ▼
QualityAssessor.Assess() ──→ QualityScore
    │
    ▼
反馈写入 history ──→ 驱动参数调整
```

### 4.2 Token 估算流程

```
文本输入
    │
    ▼
统一 Tokenizer（基于分词规则，非外部依赖）
    │  ├── 英文：空格分词 + 标点
    │  ├── 中文：字符级 bigram（无外部依赖）
    │  └── 通用：UTF-8 字节计数（fallback）
    │
    ▼
token count
```

### 4.3 规则热重载流程

```
文件监控（.axis/context-rules.yaml 变更）
    │
    ▼
RuleEngine.Reload()
    │  ├── 解析 YAML
    │  ├── 校验规则格式
    │  ├── 原子替换规则集（RWMutex）
    │  └── 日志记录重载事件
    │
    ▼
下次 Rank 调用使用新规则
```

---

## 5. API 设计

### 5.1 公共接口

```go
package contextrec

// Ranker 对候选上下文进行多策略排序
type Ranker interface {
    Rank(candidates []Candidate, query Query) ([]RankedCandidate, error)
}

// Chunker 将文档切分为可独立评估的块
type Chunker interface {
    Chunk(document Document) ([]Chunk, error)
}

// RuleEngine 根据配置规则过滤和加权候选
type RuleEngine interface {
    Load(config RuleConfig) error
    Apply(candidates []Candidate) ([]Candidate, error)
    Reload() error
}

// QualityAssessor 评估上下文装配效果
type QualityAssessor interface {
    Assess(assembly ContextAssembly, result ExecutionResult) (*QualityScore, error)
    RecordFeedback(feedback Feedback) error
    History(limit int) ([]QualityScore, error)
}

// BudgetAllocator 根据 token 预算分配上下文
type BudgetAllocator interface {
    Allocate(ranked []RankedCandidate, budget int) ([]Candidate, error)
}
```

### 5.2 数据类型

```go
// Query 结构化查询
type Query struct {
    Intent    string            `json:"intent"`
    Keywords  []string          `json:"keywords"`
    Metadata  map[string]string `json:"metadata"`  // intent.*
}

// Candidate 候选上下文
type Candidate struct {
    ID        string            `json:"id"`
    Source    string            `json:"source"`    // memory | docs | skills | vigil
    Path      string            `json:"path"`
    Content   string            `json:"content"`
    Tokens    int               `json:"tokens"`
    Metadata  map[string]string `json:"metadata"`  // context.*
}

// RankedCandidate 排序后的候选
type RankedCandidate struct {
    Candidate
    Score     float64           `json:"score"`
    Breakdown map[string]float64 `json:"breakdown"` // 各策略得分
}

// ContextAssembly 上下文装配结果
type ContextAssembly struct {
    Query     Query              `json:"query"`
    Bundles   []Candidate       `json:"bundles"`
    TotalTokens int             `json:"total_tokens"`
    Metadata  map[string]string `json:"metadata"`
}

// ExecutionResult 执行结果（用于质量评估）
type ExecutionResult struct {
    Success   bool              `json:"success"`
    Duration  time.Duration     `json:"duration"`
    Feedback  *Feedback         `json:"feedback,omitempty"`
}
```

### 5.3 CLI 集成

```bash
# 预览上下文推荐结果（不执行）
axis context recommend --intent "重构调度器" --json

# 评估历史装配质量
axis context quality --history 10

# 管理规则
axis context rules list
axis context rules validate .axis/context-rules.yaml
axis context rules reload
```

---

## 6. 配置选项

### 6.1 全局配置

```yaml
# .axis/context-recommendation.yaml
version: 1

ranker:
  weights:
    relevance: 0.5
    importance: 0.3
    freshness: 0.2
  max_candidates: 50

chunker:
  strategy: heading       # heading | paragraph | fixed | semantic
  max_tokens: 512
  overlap: 50

rules:
  file_path: .axis/context-rules.yaml

budget:
  max_tokens: 4096
  reserve_tokens: 512     # 为响应预留的空间

quality:
  weights:
    coverage: 0.3
    precision: 0.3
    efficiency: 0.2
    diversity: 0.2
  enable_feedback: true
  history_limit: 100
```

### 6.2 配置优先级

1. CLI flag（`--max-tokens` 等）
2. 环境变量（`AXIS_CONTEXT_MAX_TOKENS`）
3. 项目配置（`.axis/context-recommendation.yaml`）
4. 用户配置（`~/.config/axis/context-recommendation.yaml`）
5. 默认值

### 6.3 配置校验

- 权重总和必须为 1.0（±0.01 容差）
- `max_tokens` 必须 > 0
- `overlap` 必须 < `max_tokens`
- 规则文件必须存在且可解析

---

## 7. 实现计划

### Phase 1：统一 Token 估算（1天）

**目标**：消除四套独立预算系统，建立统一 tokenizer。

**任务**：
1. 创建 `internal/contextrec/tokenizer.go`
2. 实现英文分词（空格 + 标点）
3. 实现中文字符级 bigram
4. 实现 UTF-8 字节计数 fallback
5. 替换现有四处 token 估算逻辑
6. 添加单元测试

**验收标准**：
- 所有现有测试通过
- 新增 tokenizer 覆盖英文、中文、混合文本
- 无外部依赖引入

**依赖**：无

---

### Phase 2：RecoveryContext（2天）

**目标**：为中断任务提供恢复上下文。

**任务**：
1. 创建 `internal/contextrec/recovery.go`
2. 定义 RecoveryContext 数据结构
3. 实现从 event log 恢复上下文
4. 实现 RecoveryContext 与 Candidate 的转换
5. 集成到 Candidate Generator
6. 添加集成测试

**验收标准**：
- RecoveryContext 可从 event log 正确恢复
- 恢复的上下文可作为候选参与排序
- 恢复过程不阻塞主流程

**依赖**：Phase 1（token 估算）

---

### Phase 3：配置化规则（3天）

**目标**：将硬编码规则外置为 YAML 配置。

**任务**：
1. 创建 `internal/contextrec/ruleengine.go`
2. 定义规则 YAML schema
3. 实现 Include/Exclude/Boost/Demote/Pin 规则
4. 实现规则求值顺序
5. 实现热重载（文件监控）
6. 创建 `axis context rules` CLI 命令
7. 添加单元测试和集成测试

**验收标准**：
- 规则文件可正确加载和求值
- 热重载不阻塞 Rank 调用
- CLI 命令可列出、校验、重载规则
- 默认规则文件包含合理的初始规则

**依赖**：无

---

### Phase 4：文档分块（3天）

**目标**：将长文档按语义边界切分为可独立评估的块。

**任务**：
1. 创建 `internal/contextrec/chunker.go`
2. 实现 Heading 分块策略
3. 实现 Paragraph 分块策略
4. 实现 Fixed-size 分块策略
5. 实现重叠（overlap）逻辑
6. 集成到 Ranker 流程
7. 添加单元测试

**验收标准**：
- 长文档可正确分块
- 分块边界不破坏句子完整性
- Overlap 确保上下文衔接
- 空块被丢弃

**依赖**：Phase 1（token 估算）

---

### Phase 5：多策略排序（5天）

**目标**：实现 Relevance + Importance + Freshness 加权排序。

**任务**：
1. 创建 `internal/contextrec/ranker.go`
2. 实现 TF-IDF 相关性评分
3. 实现重要性评分（访问频率、引用次数）
4. 实现新鲜度评分（时间衰减）
5. 实现归一化和加权合并
6. 集成 RuleEngine 和 Chunker
7. 创建 `axis context recommend` CLI 命令
8. 添加单元测试和基准测试

**验收标准**：
- 排序结果符合预期（相关性高的排在前面）
- 权重配置可调整
- 性能：100 个候选排序 < 100ms
- CLI 命令可预览排序结果

**依赖**：Phase 1, 3, 4

---

### Phase 6：质量评估（5天）

**目标**：建立质量反馈闭环，量化上下文装配效果。

**任务**：
1. 创建 `internal/contextrec/quality.go`
2. 实现 Coverage 评估
3. 实现 Precision 评估
4. 实现 Efficiency 评估
5. 实现 Diversity 评估
6. 实现反馈记录和历史查询
7. 创建 `axis context quality` CLI 命令
8. 添加集成测试

**验收标准**：
- 质量评估结果可量化
- 反馈可记录和查询
- 历史趋势可可视化（CLI 输出）
- 评估过程不阻塞主流程

**依赖**：Phase 5

---

## 8. 风险与缓解

| 风险 | 概率 | 影响 | 缓解措施 |
|------|------|------|---------|
| TF-IDF 性能不足 | 中 | 排序延迟 | 限制候选数量；缓存 IDF |
| 中文分块不准 | 高 | 分块质量低 | 字符级 bigram；人工校验 |
| 规则热重载竞态 | 低 | 排序错误 | RWMutex 保护 |
| Token 估算偏差 | 中 | 预算溢出 | 预留 12% buffer |
| 质量评估偏差 | 中 | 反馈不可靠 | 多维度加权；人工校验 |

---

## 9. 后续演进

1. **BM25 替代 TF-IDF**：提升相关性评估精度
2. **语义分块**：基于 embedding 的话题边界检测
3. **学习排序（LTR）**：从反馈数据训练排序模型
4. **向量检索**：引入 embedding 相似度作为额外排序信号
5. **预算动态调整**：根据任务复杂度自适应调整 token 预算

---

## 10. 参考文档

- `docs/architecture/agent-native-first-principles.md` — Context 提升行动质量原则
- `docs/architecture/semantic-boundaries.md` — 模块语义边界
- `docs/architecture/dialectical-development-methodology.md` — 辩证开发方法论
- `docs/guides/SRS-LOOP-AI-REFERENCE.md` — SRS Loop 操作手册
