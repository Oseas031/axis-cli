---
description: 用户任务到项目工作流的最小入口路由
---

# Workflow Entry

本文件是 Agent 接到任务后的第一入口。它只做一件事：把用户任务路由到最小必要 workflow 组合。

## 一、总原则

1. **先读入口，再行动**：不要绕过 `workflow/entry.md` 直接大改。
2. **最小组合**：优先复用现有 workflow，不为单个任务新增 workflow。
3. **文档先行**：新功能必须有 `requirements.md`、`design.md`、`tasks.md`、`workflow-binding.md`。
4. **验证诚实**：未运行或工具缺失必须写“未验证”，不能写“通过”。
5. **范围克制**：宏大设计可更新方向，但不能自动扩大当前 milestone scope。
6. **设计主权**：用户已交接设计主权时，Agent 主动组织设计路线和文档落盘；只在破坏性或高风险操作前请求确认。

---

## 二、四类路由

### 1. 功能实现 / Bug 修复

```text
wf-pr-check + wf-ci + wf-doc-006
```

执行要求：定位根因 -> 最小实现 -> 测试验证 -> 同步文档。  
CLI / Shell 行为变化必须补行为测试或可执行验证命令。

### 2. 新功能 / 新规格

```text
wf-doc-004 + wf-occams + wf-pr-check + wf-ci + wf-doc-006
```

执行要求：先写或确认 specs 与 `workflow-binding.md`，再实现。  
缺少 binding 时，先补 binding，不直接编码。

### 3. 文档 / 设计 / 工作流调整

```text
wf-doc-004 + wf-doc-006 + wf-occams
```

执行要求：先更新核心设计或报告，再更新入口文档和交接状态。  
项目本体定位变化时，入口类文档必须成组检查：

```text
README.md
docs/README.md
docs/QUICKSTART.md
docs/WHITEPAPER.md
docs/current-progress.md
HANDOVER.md
```

### 4. CI/CD / 发布 / 安全 / 监控

```text
wf-ci 或 wf-pr-check 或 wf-security 或 wf-cd 或 wf-monitoring + wf-doc-004
```

执行要求：先查 `.github/config/registry.yml`，再改对应 workflow。  
构建、测试、安全可作为硬门禁；经验类检查默认只提醒。

---

## 三、复盘规则

工作复盘使用：

```text
wf-doc-004 + wf-occams + wf-doc-006
```

要求：

1. 每个工作项只归入一个唯一上游 workflow。
2. 每类提炼成功做法、问题根因、临时方案、阻塞待办。
3. 按保留 / 修正 / 剔除 / 沉淀评审。
4. 只把可执行、不过度控制的规则沉淀回 workflow。

---

## 四、默认执行顺序

```text
1. 读 workflow/entry.md
2. 选择最小上游 workflow 组合
3. 读相关 workflow 文档或 registry
4. 写/更新 spec、计划或报告
5. 实施最小变更
6. 运行必要验证
7. 更新 current-progress / HANDOVER / 报告
```

---

## 五、禁止事项

- 不读工作流直接大改。
- 不为单个 feature 新增独立 workflow。
- 不把建议性规范升级成硬门禁。
- 不因 Autogenesis、自举、真实 Agent runtime 等宏大方向扩大当前 M2 scope。
- 不用删除重建方式处理已有入口文档。
- 不默认引入 Web UI、复杂 TUI、外部数据库、daemon、真实 LLM SDK。
