# Wiki Schema — Agent 编辑规范

> 机器可解析。Agent 编辑 docs/ 前必须读取本文件。`axis docs lint` 验证合规性。

---

## 1. 约束边界（Constraint Boundaries）

### 1.1 目录权限矩阵

```yaml
directories:
  architecture/:
    create: requires_frontmatter
    modify: allowed
    delete: forbidden
    rename: forbidden
    max_files: 30
  specs/{name}/:
    create: requires_triplet  # requirements.md + design.md + tasks.md 同时创建
    modify: allowed
    delete: forbidden
    rename: forbidden
  research/:
    create: allowed
    modify: allowed
    delete: allowed_if_deprecated
    rename: forbidden
  status/:
    create: allowed
    modify: allowed
    delete: forbidden
    rename: forbidden
  lessons/:
    create: allowed
    modify: allowed
    delete: forbidden
    rename: forbidden
  guides/:
    create: requires_approval
    modify: allowed
    delete: forbidden
    rename: forbidden
```

### 1.2 命名边界

```yaml
naming:
  architecture: "^[a-z0-9]+(-[a-z0-9]+)*\\.md$"
  research:     "^[a-z0-9-]+-\\d{4}-\\d{2}-\\d{2}\\.md$"
  lessons:      "^[a-z0-9]+(-[a-z0-9]+)*\\.md$"
  specs_dir:    "^[a-z0-9]+(-[a-z0-9]+)*$"

forbidden_chars: "[A-Z_ ]"  # 无大写、无下划线、无空格
max_filename_length: 60
```

### 1.3 内容结构边界

```yaml
structure:
  h1_count: exactly_1
  h1_immutable: true  # 创建后 H1 不可修改（research/status 除外）
  max_heading_depth: 3
  max_file_bytes: 15360  # 15KB
  link_format: relative_only  # 禁止绝对路径
  link_separator: "/"  # 统一正斜杠
  external_links: references_section_only  # 外部链接只能在文末 ## References 下

frontmatter:
  required_in: [architecture/, specs/]
  schema:
    type: {enum: [architecture, spec, research, guide, status, lesson], required: true}
    status: {enum: [active, deprecated, draft], required: true}
    created: {format: "YYYY-MM-DD", required: true, immutable: true}
    last_verified: {format: "YYYY-MM-DD", required: true}
    related: {type: array, max_items: 5, items: relative_path}
```

### 1.4 关联完整性边界

```yaml
integrity:
  index_coverage: every_non_README_file_in_parent_README
  link_resolution: every_internal_link_resolves_to_existing_file
  bidirectional_soft: if_A_links_B_then_B_related_should_include_A  # lint warning, not error
  changelog_sync: every_create_delete_deprecate_appends_CHANGELOG
```

### 1.5 绝对禁令

```yaml
forbidden:
  - move_file_between_directories
  - create_file_in_docs_root  # 除 README/CHANGELOG/PURPOSE/WIKI-SCHEMA
  - modify_other_agents_frontmatter_created
  - delete_file_in_architecture_or_specs
  - rename_any_file
  - introduce_circular_related_chain  # A→B→C→A
  - embed_secrets_or_api_keys
  - modify_WIKI-SCHEMA.md  # 本文件只能由人类修改
```

---

## 2. 触发条件（Trigger Conditions）

### 2.1 何时必须执行 Ingest（更新索引）

| 触发事件 | 必须执行的动作 |
|----------|---------------|
| 新文件创建 | ① 追加到父目录 README.md ② 追加 CHANGELOG.md ③ 检查是否需要 related 双向链接 |
| Spec 状态变更 | ① 更新 specs/README.md 状态列 ② 追加 CHANGELOG.md |
| Phase III 结束 | ① 更新 status/current-progress.md ② 检查 lessons/ 是否需要新条目 |
| 研究完成 | ① 创建 research/{topic}-{date}.md ② 追加到 research/README.md ③ 追加 CHANGELOG.md |
| Bug fix 含教训 | ① 创建 lessons/{rule-name}.md ② 追加到 lessons/README.md |

### 2.2 何时必须执行 Lint（校验）

| 触发事件 | Lint 范围 |
|----------|-----------|
| 任何 docs/ 文件变更后 | `axis docs lint --check {changed_files}` |
| Phase III 退出前 | `axis docs lint` 全量 |
| 显式请求 | `axis docs lint` 全量 |
| CI push | `axis docs lint` 全量（exit 1 = 阻断） |

### 2.3 何时必须执行 Deprecate（而非删除）

| 条件 | 动作 |
|------|------|
| 文档内容完全过时 | set status=deprecated + 顶部加 deprecation notice |
| 被新文档取代 | set status=deprecated + notice 指向替代文档 |
| 30 天无人验证 + lint 标记 stale | 人类决定：更新 last_verified 或 deprecate |

### 2.4 何时禁止操作（硬阻断）

```yaml
hard_block:
  - condition: "file.path matches architecture/ AND operation == delete"
    message: "Architecture docs cannot be deleted. Use deprecate."
  - condition: "file.path matches specs/ AND operation == delete"
    message: "Spec docs cannot be deleted. Use status=deprecated."
  - condition: "frontmatter.created != original_value AND operation == modify"
    message: "frontmatter.created is immutable after creation."
  - condition: "h1_text != original_h1 AND file.path matches architecture/"
    message: "H1 title is immutable in architecture docs."
  - condition: "file_size > 15360 AND operation == create"
    message: "File exceeds 15KB. Split into multiple files."
```

---

## 3. 验收机制（Acceptance Mechanisms）

### 3.1 自动验收（`axis docs lint` 输出）

```yaml
lint_checks:
  - id: orphan
    severity: warning
    rule: "file exists in docs/ but not referenced in any README.md"
    auto_fix: false

  - id: dead-link
    severity: error
    rule: "internal markdown link points to non-existent file"
    auto_fix: false
    blocks_ci: true

  - id: no-frontmatter
    severity: error
    rule: "file in architecture/ or specs/ lacks YAML frontmatter"
    auto_fix: false
    blocks_ci: true

  - id: bad-name
    severity: error
    rule: "filename does not match naming pattern for its directory"
    auto_fix: false
    blocks_ci: true

  - id: too-large
    severity: warning
    rule: "file exceeds 15KB"
    auto_fix: false

  - id: stale
    severity: warning
    rule: "last_verified older than 30 days"
    auto_fix: false

  - id: deep-heading
    severity: warning
    rule: "heading depth > 3 (H4 or deeper)"
    auto_fix: false

  - id: multi-h1
    severity: error
    rule: "file has more than one H1 heading"
    auto_fix: false
    blocks_ci: true

  - id: abs-path
    severity: error
    rule: "internal link uses absolute path"
    auto_fix: false
    blocks_ci: true

  - id: missing-index
    severity: error
    rule: "new file not added to parent README.md"
    auto_fix: false
    blocks_ci: true
```

### 3.2 验收流程（每次文档变更）

```
Agent 编辑 docs/ 文件
       │
       ▼
┌─────────────────┐
│ Pre-write check │  ← 检查 naming + directory permission + frontmatter
└────────┬────────┘
         │ pass
         ▼
┌─────────────────┐
│   Write file    │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Post-write lint │  ← axis docs lint --check {file}
└────────┬────────┘
         │
    ┌────┴────┐
    │         │
  error     pass
    │         │
    ▼         ▼
 REVERT    ACCEPT
 + fix     + append CHANGELOG
```

### 3.3 验收标准量化

| 指标 | 阈值 | 测量方式 |
|------|------|----------|
| lint error 数 | 0 | `axis docs lint` exit code == 0 |
| lint warning 数 | < 20 | `axis docs lint 2>&1 | grep -c warning` |
| 孤立文件率 | < 10% | orphan_count / total_files |
| frontmatter 覆盖率（architecture/specs） | 100% | files_with_fm / total_files_in_scope |
| dead link 数 | 0 | 硬阻断 |
| 平均 last_verified 年龄 | < 14 天 | avg(today - last_verified) |

### 3.4 违规处理

```yaml
violation_response:
  error_severity:
    - action: block_commit  # CI 阻断
    - action: agent_must_fix_before_proceeding
    - escalation: none  # Agent 自行修复

  warning_severity:
    - action: log_to_lint_report
    - action: add_to_vigil_if_count_exceeds_20
    - escalation: human_review_at_phase_end

  repeated_violation:  # 同一 Agent 连续 3 次触发同一 error
    - action: downgrade_agent_docs_permission
    - action: require_human_approval_for_next_edit
```

### 3.5 Schema 自身的验收

```yaml
meta_validation:
  - this_file_is_immutable_by_agent  # 只有人类可修改 WIKI-SCHEMA.md
  - every_rule_has_lint_check_id     # 每条规则对应一个 lint 检查
  - every_lint_check_has_severity    # 每个检查有明确严重级别
  - no_rule_contradicts_CLAUDE_md    # 不得与宪法冲突
```

---

## 4. 快速参考（Agent 操作清单）

### RDM 判定（先检查）

```
操作是否满足 CLAUDE.md §2.1 rdm_predicate？
  ✅ 全部满足 → 使用 RDM 快速路径（下方简化流程）
  ❌ 任一不满足 → 使用完整流程（Phase 声明 + 检查清单）
```

### 创建文件（RDM 路径）

```
1. 确认目录允许创建
2. 确认文件名符合 naming pattern
3. 如在 architecture/specs/：添加 frontmatter
4. 确认文件 < 15KB，H1 唯一
5. 写入文件
6. 追加到父目录 README.md
7. 追加到 docs/CHANGELOG.md
8. 运行 axis docs lint --check {file}
```

### 修改文件

```
1. 读取当前内容
2. 确认 H1 未变（architecture/lessons）
3. 确认 frontmatter.created 未变
4. 确认无新增死链接
5. 写入修改
6. 更新 frontmatter.last_verified = today
7. 运行 axis docs lint --check {file}
```

### 废弃文件

```
1. 设置 frontmatter.status = deprecated
2. 文件顶部添加: > ⚠️ DEPRECATED: {reason}. See [{replacement}]({path}).
3. 不删除文件
4. README.md 中标记 ~~strikethrough~~
5. 追加 CHANGELOG.md
```
