# M3 Phase 3 Tasks

## Related Documents

- [requirements.md](requirements.md)
- [design.md](design.md)
- [workflow-binding.md](workflow-binding.md)

## Progress Tracking

| Task | Status | Depends On |
|---|---|---|
| T1: SLA types & constants | Completed | 鈥?|
| T2: Failure class routing + backoff | Completed | T1 |
| T3: Priority sorting in scheduler | Completed | T1 |
| T4: SLA admission extension | Completed | T1 |
| T5: Tool types & registry | Completed | 鈥?|
| T6: BashTool | Completed | T5 |
| T7: Extended ModelProvider (multi-turn) | Completed | T5 |
| T8: MockModelProvider tool-aware | Completed | T6, T7 |
| T9: Multi-turn execution loop | Completed | T5, T7 |
| T10: Orchestrator wiring | Completed | T2, T3, T4, T8, T9 |
| T11: Tests & coverage | Completed | T10 |
| T12: CLI/docs update | Completed | T11 |

---

## T1: SLA types & constants

**Goal**: Add failure class, priority, backoff constants to types package.

**Files**: `internal/types/types.go`

**Acceptance Criteria**:
- `FailureClassRetryable`, `FailureClassFatal`, `FailureClassDegradable` 甯搁噺瀛樺湪
- `SLAKeyPriority`, `SLAKeyBackoff` metadata keys 瀛樺湪
- `BackoffFixed`, `BackoffLinear`, `BackoffExponential` 甯搁噺瀛樺湪
- 缂栬瘧閫氳繃

---

## T2: Failure class routing + backoff

**Goal**: `parseSLA` 杩斿洖 failure class 鍜?backoff 绛栫暐锛沗executeTask` 鎸?failure class 鍒嗘敮澶勭悊銆?

**Files**: `internal/kernel/orchestrator/orchestrator.go`

**Acceptance Criteria**:
- `parseSLA` 杩斿洖 `(timeout, retries, failureClass, backoff string)`
- `fatal` 绫诲瀷涓嶉噸璇曪紝鐩存帴鏍囪 failed
- `degradable` 绫诲瀷鍦ㄤ緷璧栨湭灏辩华鏃朵笉闃诲锛堣烦杩囦緷璧栨鏌ワ級
- `retryable` + 鏈缃繚鎸佸綋鍓嶉噸璇曡涓?
- backoff 绛栫暐鍦ㄩ噸璇曚箣闂寸敓鏁?

---

## T3: Priority sorting in scheduler

**Goal**: `GetReadyTasks` 鎸?`sla.priority` 闄嶅簭杩斿洖 ready tasks銆?

**Files**: `internal/kernel/scheduler/scheduler.go`

**Acceptance Criteria**:
- 瑙ｆ瀽姣忎釜 task 鐨?priority metadata锛堥粯璁?128锛?
- `GetReadyTasks` 杩斿洖鎸変紭鍏堢骇闄嶅簭鎺掑垪鐨勪换鍔?
- 鍚屼紭鍏堢骇淇濇寔 FIFO
- 涓嶅奖鍝?Submit/GetStatus/Cancel 琛屼负

---

## T4: SLA admission extension

**Goal**: Admission 楠岃瘉 priority 鍜?backoff 瀛楁銆?

**Files**: `internal/contract/admission/admission.go`

**Acceptance Criteria**:
- `sla.priority`: 蹇呴』鏄?0-255 鏁存暟
- `sla.backoff`: 蹇呴』鏄?"fixed" | "linear" | "exponential"
- `sla.failure_class`: 蹇呴』鏄?"retryable" | "fatal" | "degradable"
- 鏃犳晥鍊兼椂 admission 鎷掔粷

---

## T5: Tool types & registry

**Goal**: 瀹氫箟 Tool 鎺ュ彛銆乀oolRegistry銆乀oolDefinition/ToolCall/ToolResult 绫诲瀷銆?

**Files**:
- `internal/types/types.go` 鈥?ToolCall, ToolResult, ToolDefinition, ModelMessage
- `internal/model/tool/tool.go` 鈥?Tool interface + ToolRegistry

**Acceptance Criteria**:
- `Tool` 鎺ュ彛锛歂ame(), Schema(), Execute(ctx, input) (output, error)
- `ToolRegistry`锛歊egister, Get, List 鏂规硶
- 缂栬瘧鍣ㄩ€氳繃

---

## T6: BashTool

**Goal**: 瀹炵幇 Bash 宸ュ叿锛岄€氳繃 `os/exec` 鎵ц鍛戒护銆?

**Files**: `internal/model/tool/bash.go`

**Acceptance Criteria**:
- 璇诲彇 input["command"] 浣滀负 bash 鍛戒护
- 30 绉掕秴鏃讹紙context.WithTimeout锛?
- 杩斿洖 stdout, stderr, exit_code
- 鍛戒护澶辫触鏃朵笉杩斿洖 error锛坋xit_code 闈為浂鏄甯哥粨鏋滐級锛屽彧瀵圭郴缁熼敊璇繑鍥?error

---

## T7: Extended ModelProvider (multi-turn)

**Goal**: 鎵╁睍 ModelRequest/ModelResponse/ModelProvider 鏀寔澶氳疆鍜?tool calls銆?

**Files**: `internal/model/provider/provider.go`

**Acceptance Criteria**:
- `ModelRequest` 澧炲姞 `Tools []ToolDefinition` 鍜?`History []ModelMessage` 瀛楁
- `ModelResponse` 澧炲姞 `ToolCalls []ToolCall` 瀛楁
- `ModelProvider` 鎺ュ彛涓嶅彉锛圗xecute 绛惧悕涓嶅彉锛?
- 鍚戝悗鍏煎锛堢幇鏈夎皟鐢ㄦ柟鏃犻渶淇敼锛?

---

## T8: MockModelProvider tool-aware

**Goal**: Mock provider 鑳芥ā鎷?tool-use 澶氳疆浜や簰銆?

**Files**: `internal/model/provider/mock.go`

**Acceptance Criteria**:
- 褰?input 鍖呭惈 `"tool"` key 鏃讹細杩斿洖 tool_call锛堥潪 Output锛?
- 褰?History 鏈€鍚庢槸 tool result 鏃讹細杩斿洖 final output
- 鏃?tool 鍦烘櫙琛屼负涓嶅彉

---

## T9: Multi-turn execution loop

**Goal**: ContractExecutor 鏀寔 provider 鈫?tool 鈫?provider 寰幆銆?

**Files**: `internal/contract/executor/executor.go`

**Acceptance Criteria**:
- 濡傛灉 request 鍖呭惈 tools锛岃繘鍏?multi-turn 妯″紡
- Provider 杩斿洖 tool_calls 鈫?execute tools 鈫?feed back 鈫?repeat
- 鏈€澶?10 杞?
- 濡傛灉 request 娌℃湁 tools锛岃涓轰笉鍙橈紙鍗曡疆锛?
- Tool execution 閿欒璁板綍鍦?ToolResult.Error 涓?

---

## T10: Orchestrator wiring

**Goal**: 鍦?NewOrchestrator 涓粍瑁?ToolRegistry + BashTool + SLA 绛栫暐銆?

**Files**: `internal/kernel/orchestrator/orchestrator.go`

**Acceptance Criteria**:
- ToolRegistry 鍒涘缓骞舵敞鍐?BashTool
- ToolRegistry 娉ㄥ叆 ContractExecutor
- SLA 绛栫暐閫氳繃 orchestrator 鐢熸晥
- `go build -o axis-dev.exe cmd/axis/main.go` 鎴愬姛

---

## T11: Tests & coverage

**Goal**: 鎵€鏈夋柊浠ｇ爜鏈夋祴璇曡鐩栵紝瑕嗙洊鐜?鈮?85%銆?

**Files**:
- `internal/kernel/orchestrator/orchestrator_test.go` 鈥?SLA routing tests
- `internal/kernel/scheduler/scheduler_test.go` 鈥?priority ordering tests
- `internal/contract/admission/admission_test.go` 鈥?SLA validation tests
- `internal/model/tool/tool_test.go` 鈥?registry tests
- `internal/model/tool/bash_test.go` 鈥?bash execution tests
- `internal/model/provider/mock_test.go` 鈥?tool-aware mock tests
- `internal/contract/executor/executor_test.go` 鈥?multi-turn tests

**Acceptance Criteria**:
- `go test -race ./...` 閫氳繃
- 瑕嗙洊鐜?鈮?85%

---

## T12: CLI/docs update

**Goal**: 鏇存柊鏂囨。鍙嶆槧 M3 Phase 3 鍙樻洿銆?

**Files**:
- `docs/status/current-progress.md`
- `HANDOVER.md`
- `docs/guides/QUICKSTART.md`

**Acceptance Criteria**:
- 杩涘害鏂囨。鏍囪 Phase 3 瀹屾垚
- HANDOVER 璁板綍鏂拌兘鍔?
- 瑕嗙洊鐜囨暟鎹洿鏂?

