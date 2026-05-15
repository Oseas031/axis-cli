# filepath-not-slash

触发条件: 任何涉及文件路径拼接的代码
来源: CLAUDE.md §10 跨平台安全

## 教训

永远用 `filepath.Join()` 处理路径拼接。不硬编码 `/` 或 `\\`。不用 `path.Join()`（它是 URL 路径用的）。

## 反例

```go
// 错误
filePath := dir + "/" + filename
filePath := path.Join(dir, filename)

// 正确
filePath := filepath.Join(dir, filename)
```

## 验证

```bash
grep -rn '"/"' internal/ cmd/ --include="*.go" | grep -v "_test.go" | grep -v "// url" | grep -v "filepath" | grep -v "path/" | grep -v "strings\." | grep -v "http"
```

期望: 0 行（排除 URL 路径、测试文件、filepath 相关上下文）
