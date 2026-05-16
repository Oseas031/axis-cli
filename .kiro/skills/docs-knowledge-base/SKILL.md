# Skill: docs-knowledge-base

> 文档知识库维护技能。将 Axis 文档系统作为持久化知识库运营。

## 触发条件

- 新文档创建后
- Spec 状态变更后
- Phase III 结束时（A8 写回）
- 显式请求 `axis docs lint`

## 操作

### Ingest（摄入新知识）

新文档/变更完成后执行：

1. 更新 `docs/README.md`：添加条目（链接 + 一行摘要）
2. 追加 `docs/CHANGELOG.md`：`## [日期] 操作 | 主题` 格式
3. 检查交叉引用：新文档是否应链接到已有文档？已有文档是否应链接到新文档？
4. 如果影响 status：更新 `docs/status/current-progress.md`

### Query（查询知识）

Agent 查找信息的路径：

1. 读 `docs/README.md`（index）定位相关文档
2. 读具体文档获取详情
3. 通过 frontmatter `related:` 字段发现关联文档
4. 好的综合回答应存回 `docs/research/` 或相关 architecture 文档

### Lint（健康检查）

定期检查项：

1. **孤立文档**：存在于目录中但未出现在 `docs/README.md` 的文档
2. **过时标记**：frontmatter `last_verified` 超过 30 天
3. **Spec-代码不一致**：spec tasks.md 标记完成但代码不存在，或反之
4. **死链接**：引用了不存在的文档路径
5. **缺失 frontmatter**：architecture/ 和 specs/ 下的文档应有 frontmatter

### 输出格式

Lint 结果示例：
```
docs lint: 3 issues found
  [orphan] docs/architecture/secret-handling.md — not in README.md
  [stale]  docs/specs/m4/tasks.md — last_verified: 2026-04-01 (44 days ago)
  [dead-link] docs/architecture/kernel-abstraction-model.md:15 → docs/architecture/axis-system-conventions.md (not found)
```

## Frontmatter 规范

architecture/ 和 specs/ 下的文档应包含：

```yaml
---
type: architecture | spec | research | guide | status
status: active | deprecated | draft
created: YYYY-MM-DD
last_verified: YYYY-MM-DD
related:
  - path/to/related-doc.md
---
```

- `type`：文档类型
- `status`：当前状态（active = 有效，deprecated = 已废弃，draft = 草稿）
- `created`：创建日期
- `last_verified`：最后一次确认内容仍然准确的日期
- `related`：相关文档路径列表

## 约束

- 不自动修改文档内容（只报告问题）
- 不引入外部依赖（纯 Markdown + 文件系统操作）
- 不替代 Spec-RDT 流程（Lint 是辅助，不是门禁）
- frontmatter 是渐进采纳的（新文档必须有，旧文档逐步补充）
