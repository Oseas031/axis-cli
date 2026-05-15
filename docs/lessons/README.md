# Lessons Index

> 结构化教训。每条带可执行验证命令。Agent 触碰相关模块时按触发条件匹配。

## 使用方式

- 新会话不需要全读——只在触碰相关模块时查阅
- 每条教训的 `验证` 字段是可执行命令，期望输出为空或 0 行
- `axis lint --lessons` 批量执行所有验证命令（待实现）

## 索引

| 文件 | 触发条件 | 一句话 |
|------|----------|--------|
| [filepath-not-slash](filepath-not-slash.md) | 文件路径拼接 | 用 `filepath.Join()`，不用 `"/"` |
| [no-context-background](no-context-background.md) | I/O 或可取消操作 | 禁止业务逻辑中用 `context.Background()` |
| [no-nil-nil-return](no-nil-nil-return.md) | 函数返回 `(T, error)` | 永远不返回 `(nil, nil)` |
| [channel-close-once](channel-close-once.md) | goroutine + channel | 只从一个 goroutine close channel |
| [snapshot-read-shared-files](snapshot-read-shared-files.md) | 读可能被其他进程写的文件 | 用 `os.ReadFile` 快照读，不用 streaming scanner |
