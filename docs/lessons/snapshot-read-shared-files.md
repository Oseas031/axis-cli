# snapshot-read-shared-files

触发条件: 读取可能被其他进程写入的文件（.axis/ 下的 JSON、JSONL、runtime.json）
来源: CLAUDE.md §10 跨平台安全

## 教训

可能被其他进程写入的文件用 `os.ReadFile`（原子快照读），不用 `os.Open` + `bufio.Scanner`（streaming 读可能读到半写状态）。

## 反例

```go
// 错误：streaming 读可能读到不完整的 JSON
f, _ := os.Open(".axis/runtime.json")
scanner := bufio.NewScanner(f)

// 正确：快照读，要么读到完整内容要么失败
data, err := os.ReadFile(".axis/runtime.json")
```

## 验证

```bash
grep -rn "os\.Open.*\.axis" internal/ cmd/ --include="*.go" | grep -v "_test.go"
```

期望: 0 行（.axis/ 下的文件应全部用 os.ReadFile）
