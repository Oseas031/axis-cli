# Skills 系统需求规格

**Status**: Planned
**Inspired by**: learn-claude-code s05: Skills — 按需知识加载
**Related**: `docs/architecture/agent-native-first-principles.md`, `reports/analysis/learn-claude-code-agent-execution-layer-analysis-2026-05-12.md`

## 概述

Skills System 是一个按需知识加载机制，让 Agent 能够在需要时动态获取领域知识，而不是在系统启动时预加载所有知识。这实现了 First Principles 中的 "Query is Context" 原则 —— Agent 可主动查询领域知识，而非被动接受系统推送。

Skills System 采用两层加载架构：
- **Layer 1 (cheap)**: 在 system prompt 中列出可用的 skill 名称和简短描述（约 100 tokens/skill）
- **Layer 2 (on demand)**: Agent 调用 `load_skill` 工具时，返回完整的 skill 内容

## 设计哲学

### 按需知识加载

Agent 不应在启动时背负所有可能用到的知识。Skills 系统让 Agent "用到什么知识，临时加载什么知识"。这减少了初始上下文占用，同时保持了知识获取的灵活性。

### 两层注入架构

```
┌─────────────────────────────────────────────────────────────┐
│ Layer 1: System Prompt (cheap, ~100 tokens/skill)          │
│   Skills available:                                          │
│     - pdf: Process PDF files...                             │
│     - code-review: Review code...                           │
└─────────────────────────────────────────────────────────────┘
                              │
                              │ Agent calls load_skill("pdf")
                              ▼
┌─────────────────────────────────────────────────────────────┐
│ Layer 2: Tool Result (on demand, full content)             │
│   <skill name="pdf">                                        │
│     Full PDF processing instructions...                     │
│   </skill>                                                  │
└─────────────────────────────────────────────────────────────┘
```

### Contract is Structure

SKILL.md 文件是结构化契约，包含 frontmatter 元数据和 markdown 主体内容。所有 Skill 遵循统一格式，可被程序解析和验证。

### Zero Control

Skills 系统不控制 Agent 的行为路径。它只提供知识，Agent 自主决定是否使用、如何使用。Skills 不改变调度器语义，不修改执行路径。

## 用户

- Agents 需要领域知识来完成任务（如处理 PDF、代码审查、数据库操作）
- 开发者创建和维护 Skills
- CI/CD 管道验证 Skills 格式和一致性

## 功能需求

### FR1: SKILL.md 格式

每个 Skill 必须是一个目录，包含 `SKILL.md` 文件：

```yaml
---
name: skill-name           # 必需，kebab-case 格式
description: One-line description for discovery  # 必需，单行描述
tags: tag1, tag2          # 可选，逗号分隔
version: 1.0.0            # 可选，语义化版本
author: author-name       # 可选
---

# Skill Title

Detailed instructions in markdown format...

## Section 1
...

## Section 2
...
```

目录结构示例：
```
.axis/
├── skills/
│   ├── pdf/
│   │   └── SKILL.md
│   ├── code-review/
│   │   └── SKILL.md
│   └── agent-builder/
│       ├── SKILL.md
│       ├── scripts/           # 可选辅助脚本
│       │   └── validate.py
│       └── references/        # 可选参考文件
│           └── template.yaml
```

### FR2: 两层加载

**Layer 1: Discovery**

系统启动时，扫描 `.axis/skills/` 目录，将所有 Skill 的元数据（name, description）注入到 system prompt：

```
Skills available:
  - pdf: Process PDF files - extract text, create PDFs, merge documents.
  - code-review: Review code for quality, security, and best practices.
  - database: Query and manage databases with SQL and ORM support.
```

Layer 1 开销：每个 Skill 约 100 tokens（仅名称和描述）。

**Layer 2: Loading**

Agent 调用 `load_skill` 工具时：
1. 验证 skill 名称存在
2. 读取 SKILL.md 文件
3. 返回完整内容作为 tool_result

```go
// Tool definition
type LoadSkillInput struct {
    Name string `json:"name"` // skill name in kebab-case
}

// Tool result
type LoadSkillOutput struct {
    Name        string `json:"name"`
    Description string `json:"description"`
    Content     string `json:"content"`  // full markdown body
}
```

### FR3: Skill 目录结构

```
.axis/skills/
├── <skill-name>/
│   ├── SKILL.md           # 必需：知识主体
│   ├── scripts/           # 可选：辅助脚本
│   │   └── *.py
│   └── references/        # 可选：参考文件
│       └── *.yaml
```

约束：
- Skill 名称必须是 kebab-case（`^[a-z][a-z0-9-]*[a-z0-9]$`）
- `SKILL.md` 是必需的
- `scripts/` 和 `references/` 目录是可选的
- 不允许嵌套 Skill 目录

### FR4: Skill Loader 接口

```go
package skills

// Loader manages skill discovery and loading
type Loader interface {
    // Discover returns all available skill metadata
    Discover(ctx context.Context) ([]SkillMeta, error)
    
    // Load returns full skill content by name
    Load(ctx context.Context, name string) (*Skill, error)
    
    // Validate checks if a skill directory is valid
    Validate(ctx context.Context, name string) error
}

type SkillMeta struct {
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Tags        []string `json:"tags,omitempty"`
    Version     string   `json:"version,omitempty"`
    Author      string   `json:"author,omitempty"`
}

type Skill struct {
    Meta    SkillMeta `json:"meta"`
    Content string    `json:"content"`  // raw markdown body
    Path    string    `json:"path"`     // absolute path to SKILL.md
}
```

### FR5: CLI 命令

P0 命令：

```
axis skills list [--json]
  列出所有可用的 Skills

axis skills show <skill-name> [--json]
  显示完整的 Skill 内容

axis skills validate [<skill-name>]
  验证 Skill 格式（不指定名称则验证所有）

axis skills create <skill-name>
  创建新 Skill 目录和 SKILL.md 模板
```

输出规则遵循 `docs/architecture/cli-output-conventions.md`。

### FR6: Tool 注册

`load_skill` 工具必须注册到 ToolRegistry：

```go
// In internal/tools/skills.go
func NewLoadSkillTool(loader skills.Loader) *Tool {
    return &Tool{
        Name:        "load_skill",
        Description: "Load a skill by name to get detailed instructions",
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
            name, _ := input["name"].(string)
            return loader.Load(ctx, name)
        },
    }
}
```

### FR7: 非侵入式边界

Skills System MUST NOT:
- 自动注入 Skill 内容到 provider prompts（必须通过 Agent 调用 `load_skill`）
- 修改调度器行为或执行路径
- 阻止或延迟任何任务的执行
- 改变权限模型或 Contract 语义
- 拥有任何后台 goroutine 或 watcher

### FR8: 跨平台安全

- 所有文件操作使用 `path/filepath`
- 路径检查：Skill 目录必须在 `.axis/skills/` 下，禁止路径逃逸
- 换行符统一使用 LF
- 文件编码必须是 UTF-8

### FR9: 与 Contextpack 集成

Contextpack 可以选择性地包含已加载 Skills 的引用：

```go
type ContextPack struct {
    // ... existing fields
    LoadedSkills []string `json:"loaded_skills,omitempty"`  // names of loaded skills
}
```

但这只是记录，不自动注入内容。

## 非目标

- 不支持嵌套 Skill（一个 Skill 依赖另一个 Skill）
- 不支持 Skill 版本冲突解决（P0 只加载最新版本）
- 不支持远程 Skill 仓库（只加载本地 `.axis/skills/`）
- 不支持 Skill 运行时隔离（Scripts 在主 Agent 上下文中执行）
- 不自动推送 Skill 内容到 provider prompts
- 不修改 scheduler 或 contract 语义
- 不创建 GUI/TUI/Web UI

## 验收标准

- `docs/specs/skills-system/{requirements,design,tasks}.md` 存在
- `axis skills list` 列出 `.axis/skills/` 下所有 Skill
- `axis skills show <name>` 返回完整 Skill 内容
- `axis skills validate` 验证所有 Skill 格式正确
- `load_skill` 工具注册到 ToolRegistry，Agent 可调用
- 没有 code path 在 `internal/kernel/` 或 `internal/scheduler/` 中读取 Skills
- Layer 1 注入的 token 开销 ≤ 150 tokens/skill
- `go test -race ./internal/skills/...` 通过
