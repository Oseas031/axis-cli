# Skills 系统设计

**Status**: Planned
**Implements**: `docs/specs/skills-system/requirements.md`
**Depends on**: `internal/tools/`（现有 ToolRegistry）、`internal/contextpack/`（可选集成）

## 概述

Skills System 是一个轻量级的按需知识加载机制。它在 `.axis/skills/` 目录下扫描 SKILL.md 文件，在系统启动时注入元数据到 system prompt，在 Agent 调用 `load_skill` 工具时返回完整内容。

```
.axis/skills/
  ├── pdf/
  │   └── SKILL.md
  └── code-review/
      └── SKILL.md
        │
        │  启动时：扫描并构建 SkillMeta 索引
        ▼
internal/skills/
  ├── loader.go      # Loader 接口实现
  ├── loader_test.go
  ├── discover.go    # Discover() - 扫描 .axis/skills/
  ├── discover_test.go
  ├── types.go       # SkillMeta, Skill 类型
  └── validate.go    # 验证 skill 目录结构
        │
        │  Layer 1: 注入元数据到 system prompt
        ▼
Agent System Prompt:
  "Skills available: pdf: ..., code-review: ..."
        │
        │  Layer 2: Agent 调用 load_skill("pdf")
        ▼
Tool Result:
  <skill name="pdf">
    完整的 PDF 处理指令...
  </skill>
```

无后台 goroutine。无远程获取。纯本地文件操作。

## 架构

```text
internal/skills/
  loader.go          # Loader 结构体，包含 Discover/Load/Validate 方法
  loader_test.go
  discover.go        # 扫描 .axis/skills/ 目录
  discover_test.go
  types.go           # SkillMeta, Skill, LoadSkillInput, LoadSkillOutput
  validate.go        # 验证 SKILL.md 格式和目录结构
  validate_test.go
  errors.go          # 类型化错误

cmd/axis/
  skills.go          # `axis skills ...` 子命令

internal/tools/
  skills_tool.go     # 注册 load_skill 工具（新文件）
```

## 核心数据模型

```go
package skills

import "time"

// SkillMeta 是用于发现的轻量级元数据。
// 这是注入到 system prompt 中的内容。
type SkillMeta struct {
    Name        string   `json:"name"`                  // kebab-case，必需
    Description string   `json:"description"`           // 单行描述，必需
    Tags        []string `json:"tags,omitempty"`        // 可选
    Version     string   `json:"version,omitempty"`     // 语义化版本，可选
    Author      string   `json:"author,omitempty"`      // 可选
}

// Skill 是 Load 返回的完整 skill 内容。
type Skill struct {
    Meta      SkillMeta `json:"meta"`
    Content   string    `json:"content"`    // 原始 markdown 主体（frontmatter 之后）
    Path      string    `json:"path"`       // SKILL.md 的绝对路径
    LoadedAt  time.Time `json:"loaded_at"`  // 加载时间
}

// LoadSkillInput 是工具输入 schema。
type LoadSkillInput struct {
    Name string `json:"name"` // kebab-case 格式的 skill 名称
}

// LoadSkillOutput 是工具输出。
type LoadSkillOutput struct {
    Name        string `json:"name"`
    Description string `json:"description"`
    Content     string `json:"content"`  // 完整 markdown 主体
}
```

## SKILL.md 格式

```yaml
---
name: skill-name
description: One-line description for discovery
tags: tag1, tag2
version: 1.0.0
author: author-name
---

# Skill Title

Markdown 内容...

## Section 1
...

## Section 2
...
```

Frontmatter 解析规则：
- 使用 `---` 分隔符
- 支持 YAML 格式
- `name` 和 `description` 必需
- `tags` 解析为逗号分隔的字符串数组
- 其余字段为可选

## 接口

```go
package skills

import "context"

// Loader 管理 skill 的发现和加载。
type Loader struct {
    skillsDir string                 // .axis/skills/ 的绝对路径
    index     map[string]SkillMeta   // name -> meta 缓存
    mu        sync.RWMutex
}

// NewLoader 创建新的 Loader。
func NewLoader(skillsDir string) *Loader {
    return &Loader{
        skillsDir: skillsDir,
        index:     make(map[string]SkillMeta),
    }
}

// Discover 返回所有可用的 skill 元数据。
// 首次调用时扫描 .axis/skills/ 目录，之后返回缓存索引。
func (l *Loader) Discover(ctx context.Context) ([]SkillMeta, error)

// Load 按名称返回完整的 skill 内容。
func (l *Loader) Load(ctx context.Context, name string) (*Skill, error)

// Validate 检查 skill 目录是否有效。
// 验证 SKILL.md 格式和必需的 frontmatter 字段。
func (l *Loader) Validate(ctx context.Context, name string) error

// Reload 重新扫描 skills 目录并重建索引。
func (l *Loader) Reload(ctx context.Context) error
```

## 存储与索引

- **Skills 目录**: `.axis/skills/`
- **运行时缓存**: 内存中的 `map[string]SkillMeta`，首次 `Discover()` 调用时构建
- **无持久化**: Skills 按需从文件系统加载

文件结构验证：
```
.axis/skills/<skill-name>/
  ├── SKILL.md           # 必需
  ├── scripts/           # 可选
  └── references/        # 可选
```

所有文件 I/O 使用 `path/filepath`。路径验证确保不会逃逸出 `.axis/skills/`。

## Frontmatter 解析策略

P0 使用手动 YAML frontmatter 解析，不引入外部依赖：

```go
// parseFrontmatter 从 markdown 内容中提取 YAML frontmatter。
// 返回 (meta map, markdown body, error)。
func parseFrontmatter(content string) (map[string]any, string, error) {
    // 1. 检查内容以 "---\n" 开头
    // 2. 找到结束的 "---\n"
    // 3. 解析分隔符之间的 YAML（P0 仅支持简单 key: value 解析）
    // 4. 返回剩余内容作为 markdown body
}
```

**P0 仅支持扁平 key-value 对**（不支持嵌套结构）：
```yaml
---
name: skill-name
description: One-line description
tags: tag1, tag2
version: 1.0.0
---
```

P1 如需嵌套 frontmatter，考虑引入 `gopkg.in/yaml.v3`。

## 发现流程

```
启动：
  │
  ├─► NewLoader(skillsDir)
  │
  ├─► 首次访问时调用 Discover()
  │      │
  │      ├─► 扫描 .axis/skills/ 子目录
  │      │
  │      ├─► 对每个子目录：
  │      │      ├─► 检查 SKILL.md 是否存在
  │      │      ├─► 解析 frontmatter
  │      │      ├─► 验证必需字段（name, description）
  │      │      └─► 添加到索引 map[name]SkillMeta
  │      │
  │      └─► 返回 []SkillMeta
  │
  └─► 将 SkillMeta 列表注入 system prompt
         "Skills available:\n  - pdf: Process PDF files...\n  - code-review: Review code..."
```

## 加载流程

```
Agent 调用 load_skill("pdf")：
  │
  ├─► 在索引中查找 "pdf"
  │      └─► 未找到则返回错误：ErrSkillNotFound
  │
  ├─► 读取 .axis/skills/pdf/SKILL.md
  │
  ├─► 解析 frontmatter 并提取 markdown body
  │
  └─► 返回 LoadSkillOutput{
         Name:        "pdf",
         Description: "Process PDF files...",
         Content:     "完整的 PDF 处理指令...",
       }
```

## 工具注册

```go
// 在 internal/tools/skills_tool.go 中

func RegisterSkillTools(registry *ToolRegistry, loader *skills.Loader) {
    registry.Register(&Tool{
        Name:        "load_skill",
        Description: "Load a skill by name to get detailed instructions. Use this when you need domain-specific knowledge.",
        InputSchema: map[string]any{
            "type": "object",
            "properties": map[string]any{
                "name": map[string]any{
                    "type":        "string",
                    "description": "Skill name in kebab-case (e.g., 'pdf', 'code-review')",
                },
            },
            "required": []string{"name"},
        },
        Handler: func(ctx context.Context, input map[string]any) (any, error) {
            name, ok := input["name"].(string)
            if !ok || name == "" {
                return nil, skills.ErrSkillNameRequired
            }
            skill, err := loader.Load(ctx, name)
            if err != nil {
                return nil, err
            }
            return &skills.LoadSkillOutput{
                Name:        skill.Meta.Name,
                Description: skill.Meta.Description,
                Content:     skill.Content,
            }, nil
        },
    })
}
```

## CLI 命令

```go
// 在 cmd/axis/skills.go 中

// axis skills list
func runSkillsList(cmd *cobra.Command, args []string) error {
    loader := skills.NewLoader(skillsDir)
    metas, err := loader.Discover(context.Background())
    if err != nil {
        return err
    }
    // 按 cli-output-conventions.md 格式化输出
    // 默认：人类可读表格
    // --json：JSON 数组
}

// axis skills show <skill-name>
func runSkillsShow(cmd *cobra.Command, args []string) error {
    name := args[0]
    loader := skills.NewLoader(skillsDir)
    skill, err := loader.Load(context.Background(), name)
    if err != nil {
        return err
    }
    // 格式化输出
}

// axis skills validate [<skill-name>]
func runSkillsValidate(cmd *cobra.Command, args []string) error {
    loader := skills.NewLoader(skillsDir)
    if len(args) > 0 {
        // 验证指定 skill
        return loader.Validate(context.Background(), args[0])
    }
    // 验证所有 skills
    metas, _ := loader.Discover(context.Background())
    for _, meta := range metas {
        if err := loader.Validate(context.Background(), meta.Name); err != nil {
            return err
        }
    }
    return nil
}

// axis skills create <skill-name>
func runSkillsCreate(cmd *cobra.Command, args []string) error {
    name := args[0]
    // 验证名称格式（kebab-case）
    // 创建目录 .axis/skills/<name>/
    // 创建 SKILL.md 模板内容
}
```

## Layer 1 注入

```go
// 在 internal/agent/prompt.go（或等效位置）中

func buildSystemPrompt(skillsLoader *skills.Loader) string {
    var sb strings.Builder
    
    sb.WriteString("You are an Axis agent...\n\n")
    
    // Layer 1: 注入可用 skills
    metas, err := skillsLoader.Discover(context.Background())
    if err == nil && len(metas) > 0 {
        sb.WriteString("Skills available:\n")
        for _, meta := range metas {
            sb.WriteString(fmt.Sprintf("  - %s: %s\n", meta.Name, meta.Description))
        }
        sb.WriteString("\nUse load_skill(name) to load detailed instructions.\n\n")
    }
    
    // ... 系统提示的其余部分
    
    return sb.String()
}
```

预估 token 开销：基础约 20 tokens + 每个 skill 约 80 tokens = 约 100 tokens/skill。

## 边界强制测试

以下测试是强制性的：

1. `TestKernelDoesNotImportSkills` — `go list -deps ./internal/kernel/...` 不得包含 `internal/skills`
2. `TestSchedulerDoesNotImportSkills` — 同上，针对 `internal/scheduler/`
3. `TestLoadSkillIsOptIn` — 未调用 `load_skill` 时，skill 内容不得出现在 agent 上下文中
4. `TestSkillPathSafety` — skill 名称 `../escape` 必须被拒绝
5. `TestSkillNameFormat` — skill 名称必须匹配 `^[a-z][a-z0-9-]*[a-z0-9]$`

## 并发

- 单个 `map[string]SkillMeta`，由 `sync.RWMutex` 保护
- 无 goroutine
- Discover、Load、Validate 均为同步操作

## 跨平台安全

- 所有路径通过 `path/filepath`
- 斜杠规范化：`filepath.ToSlash()`
- 仅使用 LF 换行符
- 要求 UTF-8 编码

## 非目标（从需求文档强化）

- 不支持嵌套 skills（skill 依赖另一个 skill）
- 不支持版本冲突解决
- 不支持远程 skill 仓库
- 不支持脚本运行时隔离
- 不自动注入到 provider prompts
- 不修改 scheduler/contract

## 已决策项

### D1: Skills 存放在 `.axis/skills/`

- **决策**: Skills 存储在 `.axis/skills/<skill-name>/SKILL.md`
- **原因**: 遵循 Axis 将配置存储在 `.axis/` 目录的惯例
- **逆转条件**: 需要支持用户全局 skills 时，添加 `~/.axis/skills/` 作为次要位置

### D2: P0 不做 skill 版本管理

- **决策**: 每个 skill 只有一个版本（最新）。不做版本冲突解决。
- **原因**: Karpathy §2 — 最少代码。版本管理增加复杂度，目前无明确用例。
- **逆转条件**: 实际使用表明需要同一 skill 的多个版本

### D3: P0 不支持嵌套 skills

- **决策**: 一个 skill 不能依赖另一个 skill。
- **原因**: 保持加载逻辑简单。无需依赖解析。
- **逆转条件**: 实际使用表明 skill 组合有价值

### D4: 脚本放在 `scripts/` 子目录

- **决策**: Skill 可以有 `scripts/` 子目录存放辅助脚本。
- **原因**: 保持 skill 自包含，同时将知识与可执行代码分离。
- **逆转条件**: 安全审查表明需要脚本隔离

这些决策已记录，供未来 Spec-RDT 审查。
