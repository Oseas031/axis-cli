# Skills 系统任务

**Status**: Planned
**Last Updated**: 2026-05-12
**Implements**: 本目录下的 `requirements.md` + `design.md`。

> 所有任务均为 P0，除非另有标注。P0 = 最小可行，无外部依赖，无后台工作。

---

## T1: Spec-RDT 定稿

- [x] requirements.md
- [x] design.md
- [x] tasks.md（本文件）
- [ ] Spec 三件套通过一致性检查（需求/设计无矛盾；已决策项已记录）
- [ ] 从 `docs/architecture/semantic-boundaries.md` 交叉链接
- [x] 状态从 Draft 提升为 Planned（按 `spec-lifecycle-conventions.md`）

**验收**: spec 三件套存在，内部一致，status = Planned。

---

## T2: 包骨架（`internal/skills/`）

创建包含类型和接口的包 —— 暂无实现逻辑。

### 2.1 文件

- `types.go` — `SkillMeta`、`Skill`、`LoadSkillInput`、`LoadSkillOutput`
- `errors.go` — 类型化错误（`ErrSkillNotFound`、`ErrSkillNameRequired`、`ErrInvalidSkillName`、`ErrInvalidPath`）
- `loader.go` — `Loader` 结构体及方法存根

### 2.2 验证

- `SkillMeta.Validate()` 强制：
  - `Name` 必需，匹配 `^[a-z][a-z0-9-]*[a-z0-9]$`
  - `Description` 必需，TrimSpace 后非空
- `ValidateSkillName(name string) error` — 验证 skill 名称格式

### 2.3 测试

- `types_test.go`: SkillMeta 验证、零值行为、JSON 往返
- `errors_test.go`: 使用 `errors.Is` 的错误包装

**验收**: `go test ./internal/skills/...` 通过。不引入 Go 标准库以外的依赖。

---

## T3: Skill 发现（`discover.go`）

### 3.1 实现

- `func (l *Loader) Discover(ctx context.Context) ([]SkillMeta, error)`
  - 首次调用时扫描 `.axis/skills/` 子目录
  - 对每个子目录，检查 `SKILL.md` 是否存在
  - 解析 frontmatter（`---` 分隔符之间的 YAML）
  - 验证必需字段
  - 缓存到 `l.index` map
  - 后续调用返回缓存

### 3.2 辅助函数

- `parseFrontmatter(content string) (map[string]any, string, error)` — 解析 YAML frontmatter，返回 meta map 和 markdown body
- `parseTags(raw string) []string` — 解析逗号分隔的标签

### 3.3 测试

- Discover 在 fixture 目录中找到所有 skills
- Discover 第二次调用返回缓存结果
- Discover 处理空 skills 目录
- Frontmatter 解析：有效 YAML、缺少分隔符、无效 YAML
- SkillMeta 验证：缺少 name、缺少 description、无效名称格式

**验收**: `go test ./internal/skills/...` 通过，包括边界情况。

---

## T4: Skill 加载（`loader.go`）

### 4.1 实现

- `func (l *Loader) Load(ctx context.Context, name string) (*Skill, error)`
  - 验证 skill 名称格式
  - 在索引中查找（索引为空时调用 Discover）
  - 读取 `SKILL.md` 文件
  - 解析 frontmatter 并提取 markdown body
  - 返回 `Skill` 结构体

### 4.2 路径安全

- `func safeSkillPath(baseDir, name string) (string, error)`
  - 拼接 `baseDir` 和 `name`
  - 解析为绝对路径
  - 验证结果在 `baseDir` 下（无路径逃逸）
  - 拒绝 `..`、`./`、名称中的绝对路径

### 4.3 测试

- Load 对有效 skill 返回内容
- Load 对不存在的 skill 返回 `ErrSkillNotFound`
- Load 拒绝路径逃逸尝试（`../escape`、`..\\escape`）
- Load 返回不含 frontmatter 的 markdown body
- LoadedAt 时间戳正确设置

**验收**: 路径安全测试通过，无路径逃逸可能。

---

## T5: Skill 验证（`validate.go`）

### 5.1 实现

- `func (l *Loader) Validate(ctx context.Context, name string) error`
  - 检查 skill 目录是否存在
  - 检查 `SKILL.md` 是否存在
  - 解析并验证 frontmatter
  - 检查 `scripts/` 和 `references/` 如果存在则必须是目录
  - 有效返回 nil，否则返回具体错误

### 5.2 验证规则

- 目录名必须匹配 skill 名称
- `SKILL.md` 必须是有效 UTF-8
- Frontmatter 必须包含 `name` 和 `description`
- Frontmatter 中的 `name` 必须匹配目录名

### 5.3 测试

- Validate 对有效 skill 返回 nil
- Validate 对缺少 SKILL.md 返回错误
- Validate 对无效 frontmatter 返回错误
- Validate 处理可选的 `scripts/` 和 `references/` 目录

**验收**: 所有验证规则均有测试覆盖。

---

## T6: CLI 命令（`cmd/axis/skills.go`）

### 6.1 子命令

- `axis skills list [--json]`
  - 列出所有可用 skills
  - 人类输出：表格格式
  - `--json`：JSON 数组

- `axis skills show <skill-name> [--json]`
  - 显示完整 skill 内容
  - 人类输出：格式化 markdown
  - `--json`：JSON 对象

- `axis skills validate [<skill-name>]`
  - 验证 skill 格式
  - 不指定名称：验证所有 skills
  - 有效退出 0，否则非零

- `axis skills create <skill-name>`
  - 创建 skill 目录和 `SKILL.md` 模板
  - 先验证名称格式

### 6.2 输出格式

按 `cli-output-conventions.md`：

```
$ axis skills list
NAME         DESCRIPTION
pdf          Process PDF files - extract text, create PDFs
code-review  Review code for quality and security

$ axis skills show pdf
Name: pdf
Description: Process PDF files - extract text, create PDFs
---
# PDF Processing Skill

You now have expertise in PDF manipulation...

$ axis skills create new-skill
Created skill directory: .axis/skills/new-skill/
Created: .axis/skills/new-skill/SKILL.md
Edit the file to add your instructions.
```

### 6.3 测试

- 人类输出的 golden-file 测试
- `--json` 输出的 JSON schema 测试
- 帮助文本测试

**验收**: `axis skills list` 对真实 `.axis/skills/` 目录有效。

---

## T7: 工具注册（`internal/tools/skills_tool.go`）

### 7.1 实现

- 创建 `load_skill` 工具
- 注册到 ToolRegistry
- Handler 调用 `loader.Load()`

### 7.2 工具 Schema

```json
{
  "name": "load_skill",
  "description": "Load a skill by name to get detailed instructions. Use this when you need domain-specific knowledge.",
  "input_schema": {
    "type": "object",
    "properties": {
      "name": {
        "type": "string",
        "description": "Skill name in kebab-case (e.g., 'pdf', 'code-review')"
      }
    },
    "required": ["name"]
  }
}
```

### 7.3 测试

- 工具已注册到 ToolRegistry
- 工具 handler 对有效名称返回 skill 内容
- 工具 handler 对无效名称返回错误
- 工具输出是有效 JSON

**验收**: `load_skill` 工具可从 agent 循环中调用。

---

## T8: Layer 1 系统提示注入

### 8.1 实现

- 扩展系统提示构建器以包含 skills 元数据
- 启动时调用 `loader.Discover()`
- 格式化为：`Skills available:\n  - <name>: <description>\n`

### 8.2 Token 预算

- 预估：基础约 20 tokens + 每个 skill 约 80 tokens
- 如果总量超过阈值，截断列表（未来：按上下文优先级排序）

### 8.3 测试

- 系统提示包含 skills 元数据
- 系统提示不包含 skill 内容（仅元数据）
- 空 skills 目录：提示中无 skills 部分

**验收**: Skills 列在系统提示中，内容未包含。

---

## T9: 边界强制测试

在 `boundary_test.go` 或包级测试中：

- T9.1 `TestKernelDoesNotImportSkills` — `go list -deps ./internal/kernel/...` 排除 `internal/skills`
- T9.2 `TestSchedulerDoesNotImportSkills` — 同上，针对 `internal/scheduler/`
- T9.3 `TestLoadSkillIsOptIn` — 验证未调用 `load_skill` 时 skill 内容不在 agent 上下文中
- T9.4 `TestSkillPathSafety` — 路径逃逸尝试被拒绝
- T9.5 `TestSkillNameFormat` — 无效名称被拒绝

**验收**: 所有边界测试在 `go test ./...` 下通过。

---

## T10: 文档

- [ ] 在 `docs/architecture/semantic-boundaries.md` 中为 Skills System 边界添加行
- [ ] 在 `docs/architecture/module-and-naming-conventions.md` 中为 `.axis/skills/` 目录添加条目
- [ ] 创建 `internal/skills/BOUNDARY.md` 包含约束规则
- [ ] 更新 `WORKFLOW-HUMAN/today-5-12-learnclaude.md` 的完成状态

**验收**: `grep -r "skills" docs/architecture/` 返回上述添加内容。

---

## T11: 示例 Skills

创建用于测试和文档的示例 skills：

- `axis skills create pdf`
- `axis skills create code-review`
- `axis skills create database`

每个包含真实的 SKILL.md 内容。

**验收**: `.axis/skills/` 中至少存在一个示例 skill。

---

## T12: P1 后续（首次实现范围外）

- 远程 skill 仓库（Git clone）
- Skill 版本管理和冲突解决
- 嵌套 skills（skill 依赖）
- 脚本运行时隔离
- Skill 市场 / 发现服务
- 按上下文相关性自动排序 skills

**关于 `scripts/` 和 `references/` 目录的说明**：
P0 仅验证这些目录存在（如果有的话）。实际使用推迟到 P1：
- `scripts/` — Agent 可能调用的辅助脚本（需要安全审查）
- `references/` — 补充文件（模板、schema），可能在 skill 内容中被引用

**验收**: 每个 P1 项目已记录供未来规划。

---

## 完成定义（整个 Spec）

- 上述所有 P0 任务已完成
- `go test -race ./...` 绿色
- `go vet`、`staticcheck`、`gosec` 干净
- `go.mod` 无新条目（纯标准库）
- 边界测试（T9）全部通过
- `axis skills list` 和 `axis skills show` 端到端工作
- `load_skill` 工具可被 Agent 调用
- Status = Planned（按 `spec-lifecycle-conventions.md`）
