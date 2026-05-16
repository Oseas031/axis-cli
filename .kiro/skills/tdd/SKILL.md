---
name: tdd
description: Test-driven development with red-green-refactor loop. Vertical slices, one test at a time. Use when user wants TDD, mentions "red-green-refactor", wants test-first, or says "写测试"/"测试驱动".
tags: [engineering, testing, vertical-slices]
source: mattpocock/skills
source_version: 2026-05-15
---

# TDD

> 来源：mattpocock/skills (MIT)，适配 Axis 语境。

## 哲学

**核心原则**：测试验证行为（通过公共接口），不验证实现细节。代码可以完全重写；测试不应该断。

**好测试** = 集成风格，走真实代码路径，通过公共 API。读起来像 spec。
**坏测试** = 耦合实现，mock 内部协作者，测私有方法。重构时断但行为没变 = 坏测试。

## 反模式：水平切片

**禁止先写所有测试再写所有实现。**

```
错误（水平）：
RED: test1, test2, test3, test4, test5
GREEN: impl1, impl2, impl3, impl4, impl5

正确（纵向）：
RED→GREEN: test1→impl1
RED→GREEN: test2→impl2
RED→GREEN: test3→impl3
```

## 工作流

### 1. 规划

- [ ] 确认用户需要什么接口变更
- [ ] 确认测试哪些行为（优先级排序）
- [ ] 识别 deep module 机会（小接口，深实现）
- [ ] 列出要测试的行为（不是实现步骤）
- [ ] 用户批准计划

问："公共接口应该长什么样？哪些行为最重要？"

### 2. Tracer Bullet

写 ONE 测试确认 ONE 件事：

```
RED: 写第一个行为的测试 → 测试失败
GREEN: 写最小代码通过 → 测试通过
```

### 3. 增量循环

每个剩余行为：

```
RED: 写下一个测试 → 失败
GREEN: 最小代码通过 → 通过
```

规则：
- 一次一个测试
- 只写够通过当前测试的代码
- 不预判未来测试

### 4. 重构

所有测试通过后：
- [ ] 提取重复
- [ ] 深化模块（复杂度藏在简单接口后面）
- [ ] 自然地应用 SOLID
- [ ] 每次重构后跑测试

**永远不在 RED 时重构。先到 GREEN。**

## 每轮检查

```
[ ] 测试描述行为，不是实现
[ ] 测试只用公共接口
[ ] 测试能扛住内部重构
[ ] 代码对当前测试是最小的
[ ] 没有投机性功能
```

## Axis 特定

- 测试文件放在同包 `_test.go`
- 用 `go test -race ./...` 验证
- 遵循 CLAUDE.md §9：围绕风险路径设计测试
- 遵循 CLAUDE.md §10：永远不做真实外部网络调用（用 `httptest`）
