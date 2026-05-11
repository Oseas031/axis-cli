# 每日复盘：2026-05-11

**日期**: 2026-05-11
**工作焦点**: axis-gui 连接修复、TDD 优化 busy-poll + stale running 崩溃修复、前端实时化与体验优化
**核心原则**: 非侵入 axis 核心，GUI 仅作为独立工具存在

---

## 一、工作内容分类

### 类别 1：axis-gui 运行时连接修复

| 工作项 | 文件 | 说明 |
|---|---|---|
| 修复 axis 二进制路径解析 | `tools/axis-gui/server.go` | `findAxisBinary()` 返回绝对路径，解决 Go 1.19+ exec 安全限制（"cannot run executable found relative to current directory"） |
| 修复代理响应头顺序 | `tools/axis-gui/server.go` | `proxyToAxis` 先写 body 再写 header 导致 HTTP 500，调整为正确的 header-first 顺序 |
| 修复字体 CDN | `tools/axis-gui/frontend/index.html` | 替换失效的 fontsource CDN URL |
| 支持 contract_id 字段 | `tools/axis-gui/frontend/src/lib/api.ts` | `submitTask` API 增加 `contract_id`，解决 "contract ID is empty" 提交失败 |
| 增强前端错误展示 | `tools/axis-gui/frontend/src/lib/api.ts` | 从后端 error response 提取 `message` 字段准确展示 |

### 类别 2：TDD 修复 scheduler/orchestrator（符合 axis 核心设计哲学）

| 工作项 | 文件 | 说明 |
|---|---|---|
| T1: stale running task reset 测试 | `internal/kernel/scheduler/scheduler_test.go` | `TestScheduler_StaleRunningTasksReset`：模拟崩溃后重启 scheduler，断言 `Running` → `Failed` |
| T1: 实现 crash recovery | `internal/kernel/scheduler/scheduler.go` | `NewScheduler` 从 `stateStore.ListAll()` 恢复任务，重置 stale `Running` 为 `Failed` |
| T1: StateStore 接口扩展 | `internal/kernel/sharedlayer/state_store.go` | 新增 `ListAll()` 方法及 `MemoryStateStore` 实现 |
| T2: busy-poll 移除测试准备 | `internal/kernel/orchestrator/orchestrator_test.go` | 验证现有测试在 cond var 替换后仍通过 |
| T2: cond var 通知实现 | `internal/kernel/orchestrator/orchestrator.go` | worker 释放时 `select { case o.taskSubmitted <- struct{}{}: default: }`，`runTaskLoop` 两处 `time.After(100ms)` 替换为 `<-o.taskSubmitted` |

**设计决策**：
- 不添加 `context.Cancel` 到 worker goroutine（避免侵入核心调度语义）
- 使用已有 `taskSubmitted` channel 做信号通知，不引入新锁或 cond var 对象
- 保留 `default:` 非阻塞发送，防止 channel 满时阻塞 worker 清理

### 类别 3：前端实时化与体验优化（纯 axis-gui 范围）

| 工作项 | 文件 | 说明 |
|---|---|---|
| T5: WebSocket 集成 TasksPage | `tools/axis-gui/frontend/src/pages/TasksPage.tsx` | 新增 WebSocket 监听 `/ws/events`，事件到达即时更新列表；poll fallback 从 5s 降至 30s |
| T5: 连接状态可视化 | `tools/axis-gui/frontend/src/pages/TasksPage.tsx` | live badge 动态反映 `wsConnected`：绿色 pulse = 实时，红色 = 断开 |
| T6: 任务时间线聚合 | `tools/axis-gui/frontend/src/pages/TasksPage.tsx` | 按 `task_id` group events，行内展示最新状态；展开显示完整事件时间线（event_type → status → message → timestamp） |
| T6: i18n 补充 | `tools/axis-gui/frontend/src/i18n/en.json`, `zh.json` | 新增 `disconnected`、`timeline`、`refreshEvery`(30s) 键 |
| T7: 暗色模式系统偏好监听 | `tools/axis-gui/frontend/src/components/ThemeToggle.tsx` | 未手动设置时自动跟随 `prefers-color-scheme` 变化 |

---

## 二、验证结果

### 测试
```
go test ./...          # 全部通过（含 scheduler/orchestrator 回归测试）
```

### 冒烟测试
```
axis-dev.exe start                          # 运行时启动于 :56445
axis-gui.exe --port 8090 --axis-root .      # GUI 启动于 :8090
POST /api/tasks -> task_id: gui-smoke-final  # 提交成功，返回 pending
GET  /api/tasks/gui-plan-smoke/status        # 返回 completed
```

---

## 三、经验萃取

### 保留

| 经验 | 理由 |
|---|---|
| TDD 三步：先写失败测试 → 最小实现 → 全量回归 | 保证改动不引入回归，stale-running 和 busy-poll 均经此验证 |
| 用已有 channel 做信号而非引入新原语 | orchestrator 的 `taskSubmitted` 已存在，复用比新增 cond var 更符合最小改动原则 |
| axis-gui 纯代理模式 | 不修改 axis 核心路由/接口，所有 frontend 功能通过已有 `/v1/*` 和 `/ws/events` 消费 |
| 前端 group-by 替代后端聚合 | 事件日志已有 `task_id`，前端 `useMemo` 聚合零后端成本 |

### 修正

| 经验 | 问题 | 修正方向 |
|---|---|---|
| Windows 进程 kill 后端口释放延迟 | `Stop-Process` 后 `TIME_WAIT` 导致 axis-gui 启动失败 | 增加 `Start-Sleep -Seconds 2` 等待；长期应使用 graceful shutdown |

### 移除

| 经验 | 理由 |
|---|---|
| T6 原计划 "任务取消 UI" | 需 axis 核心新增 `CancelTask` 接口，违背 "Zero Control" 轻量原则；已替换为 "任务时间线聚合" |

---

## 四、对应工作流完善

### wf-dev：补充 axis-gui 开发验证步骤

```text
axis-gui 修改后验证清单：
1. cd tools/axis-gui/frontend && npm run build
2. cd tools/axis-gui && go build -o axis-gui.exe .
3. 确认旧进程已终止（tasklist / netstat）
4. .\axis-gui.exe --port 8090 --axis-root <root>
5. POST /api/tasks 冒烟测试
```

### wf-doc-006：进度文档同步范围补充

```text
axis-gui 相关改动需同步记录到：
- docs/status/current-progress.md（axis-gui 章节）
- reports/daily/daily-retrospective-YYYY-MM-DD.md
```

---

## 五、待处理任务

| 优先级 | 任务 | 状态 | 阻塞原因 |
|---|---|---|---|
| 高 | 完整开发计划（plan + worktree + subagent 并行） | 待启动 | 等待用户确认方向 |
| 中 | Provider 健康状态详细展示（ProvidersPage） | 待排期 | 纯前端，利用已有 `/api/runtime/status` |
| 低 | axis-gui 进程优雅退出（避免端口占用） | 待排期 | 需捕获信号 + http.Server.Shutdown |

---

## 六、总结

**完成度**: 100%（T1–T7 全部闭环）
- **T1–T4**（TDD + 冒烟）✅
- **T5**（WebSocket 实时集成）✅
- **T6**（任务时间线聚合）✅ — 原 "取消" 改为 "时间线"，零核心侵入
- **T7**（暗色模式增强）✅

**主要成果**：
- axis-gui ↔ axis 运行时端到端链路稳定
- scheduler crash recovery 机制落地
- orchestrator 从 busy-poll 升级为事件驱动
- TasksPage 从 5s 轮询升级为 WebSocket 实时 + 30s fallback
- 任务按 ID 聚合展示时间线，提升可观测性

**下一步**: 按用户授权制定完整开发计划，依托 worktree 分派 subagent 并行开发
