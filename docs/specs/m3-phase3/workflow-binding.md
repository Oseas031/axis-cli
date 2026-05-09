# M3 Phase 3 Workflow Binding

## Upstream Workflows

本特性遵循项目既有工作流体系：

- **wf-doc-004** Meta-Workflow Management: 文档先行，显式依赖，HANDOVER 同步
- **wf-occams** Occam's Razor: 不引入外部依赖，Go stdlib only，只构建当前所需
- **wf-pr-check** PR Quality Check: 质量门禁 + 非阻塞文档上下文
- **wf-ci** Continuous Integration: build, format, race tests, ≥85% coverage
- **wf-doc-006** Document Audit: HANDOVER 一致性

## Phase 3 特殊工作流约束

- SLA 策略引擎和工具调用层可**并行开发**（无代码依赖），最后在 T10 汇合
- 两部分共享 `internal/types/types.go` 的修改（T1 + T5），需先合并类型变更
- BashTool 使用 `os/exec`，Go stdlib 内置，不引入外部依赖
- 多轮执行循环的 max turns 硬编码为 10，不引入配置层
