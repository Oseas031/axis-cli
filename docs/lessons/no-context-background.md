# no-context-background

触发条件: I/O 操作、长时间运行操作、可取消操作
来源: CLAUDE.md §10 Context 传播

## 教训

`context.Background()` 只允许在 `main()` 或测试顶层使用。业务逻辑中必须接受 `ctx context.Context` 作为第一参数并传递。

## 反例

```go
// 错误：业务函数中创建新 context
func (s *Service) DoWork() error {
    ctx := context.Background()
    return s.provider.Call(ctx, req)
}

// 正确：接受并传递 context
func (s *Service) DoWork(ctx context.Context) error {
    return s.provider.Call(ctx, req)
}
```

## 验证

```bash
grep -rn "context\.Background()" internal/ --include="*.go" | grep -v "_test.go" | grep -v "main.go"
```

期望: 0 行
