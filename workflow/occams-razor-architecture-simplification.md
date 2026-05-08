# 奥卡姆剃刀架构简化工作流

## 目的

防止 Axis 在宏大设计、复杂自动化或真实 Agent runtime 的诱惑下过早扩张。奥卡姆剃刀不是削弱愿景，而是保护愿景有正确的发生顺序。

## 三个判断

### 1. 当前是否必须做

只实现当前 milestone 已确认的最小能力。  
如果一个想法属于后续 bootstrap-loop / autogenesis-loop，就先写 spec 或报告，不混入当前实现。

### 2. 现有轻量方案是否足够

新增以下复杂度前，必须说明为什么现有轻量方案不够：

- UI：为什么 CLI / Shell 不够
- 模型 Provider：为什么 MockProvider 不够
- workflow：为什么 `workflow/entry.md` 现有路由不够
- 自动化：为什么非阻塞提醒不够
- 依赖：为什么标准库或现有依赖不够
- 持久化 / daemon：为什么当前进程内语义或 file-backed state 不够

无法说明时，默认不新增。

### 3. 是否破坏 Scaffold-to-Self

workflow、contract、permission rule、spec 是过渡性结构。  
新增规则不能把临时脚手架伪装成永久控制。

## 设计哲学校正规则

当发现实现与 **More Context, More Action, Zero Control**、**Bash is All You Need** 或 **Competence earns autonomy** 不一致时：

1. 先判断是否可通过错误语义、文档或测试修正。
2. 必要时插入最小校正任务并写入对应 `tasks.md`。
3. 不借校正任务引入 Web UI、复杂 TUI、外部数据库、daemon 或真实 LLM SDK。
4. 确需新增复杂度时，先创建独立 spec。

## 注意事项

- 不做破坏性编辑。
- 保持里程碑边界。
- 废弃内容移动到 `docs/deprecated/`，不要抹除历史。
- 入口文档可以更新方向，实现任务必须遵守当前 scope。
