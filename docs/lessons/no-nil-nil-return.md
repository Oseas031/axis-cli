# no-nil-nil-return

触发条件: 任何返回 `(T, error)` 的函数
来源: CLAUDE.md §10 防御性编程

## 教训

永远不返回 `(nil, nil)`。调用者无法区分"成功但无结果"和"出错了"。使用 sentinel error 或类型化零值。

## 反例

```go
// 错误：调用者无法判断是否成功
func FindUser(id string) (*User, error) {
    if id == "" {
        return nil, nil  // 这是成功还是失败？
    }
}

// 正确：明确语义
var ErrUserNotFound = errors.New("user not found")

func FindUser(id string) (*User, error) {
    if id == "" {
        return nil, ErrUserNotFound
    }
}
```

## 验证

```bash
grep -rn "return nil, nil" internal/ cmd/ --include="*.go" | grep -v "_test.go"
```

期望: 0 行
