# channel-close-once

触发条件: goroutine 间通信、channel 使用
来源: CLAUDE.md §10 并发安全

## 教训

永远不从多个 goroutine close 同一个 channel。只有发送方 close，且只 close 一次。用 `sync.Once` 保护 close 操作，或设计为只有一个 owner。

## 反例

```go
// 错误：多个 goroutine 可能 close 同一个 channel
for i := 0; i < n; i++ {
    go func() {
        defer close(done)  // panic: close of closed channel
        work()
    }()
}

// 正确：用 WaitGroup + 单点 close
var wg sync.WaitGroup
for i := 0; i < n; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        work()
    }()
}
go func() { wg.Wait(); close(done) }()
```

## 验证

```bash
grep -rn "defer close(" internal/ cmd/ --include="*.go" | grep -v "_test.go"
```

期望: 每个匹配项需人工确认只有一个 goroutine 执行该 close（无法完全自动化，但可标记审查点）
