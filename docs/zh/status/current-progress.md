# 褰撳墠宸ヤ綔杩涘害

**鏇存柊鏃堕棿**: 2026-05-11
**褰撳墠閲岀▼纰?*: Milestone 1 鉁?| Milestone 2 鉁?| Milestone 3 Phase 1-3 鉁?| Milestone 4 鉁?| Milestone 5 鉁?宸插畬鎴?

## 褰撳墠璁捐瀹氫綅

Axis 褰撳墠涓嶅啀鍙瀹氫箟涓烘櫘閫?Agent 璋冨害骞冲彴锛岃€屾槸 Agent 鑷洜鍖栫殑鏃╂湡鎵ц搴曞骇銆?

鏍稿績鍒ゆ柇锛?

- 鑷妇璧风偣宸茬粡鍙戠敓锛氬閮?Agent 姝ｅ湪鍚?Axis 娉ㄥ叆鍙鍥哄寲銆佹墽琛屻€佸弽鎬濆拰婕斿寲鐨勬€濇兂
- M2 涓嶆槸鏅€氬苟琛岃皟搴﹂噷绋嬬锛岃€屾槸鏈潵 Autogenesis Loop 鐨勬墽琛屽簳搴?
- workflow 鏄复鏃惰剼鎵嬫灦锛宑ontract 鏄垚闀胯竟鐣岋紝permission rule 鏄€掕繘鑷富鏉冩満鍒讹紝spec 鏄瀛?
- M2 宸插叏閮ㄥ畬鎴?
- M3 Phase 1 宸插畬鎴愶細ModelProvider 鎵ц璺緞鎵撻€氥€佽鐩栫巼 88.8%銆丏AG/SLA 琛ュ叏

## 宸插畬鎴愪换鍔?
- [x] 淇 staticcheck ST1003 閿欒锛坰hared_layer 鈫?sharedlayer锛?
- [x] 淇濂戠害鎵ц鍣ㄦ灇涓鹃獙璇侀€昏緫锛堟敮鎸?int 绫诲瀷锛?
- [x] 淇 CI 宸ヤ綔娴?godoc -html 搴熷純鍙傛暟
- [x] 鍒涘缓宸ヤ綔娴佹敼杩涜鍒掑苟淇楂樹紭鍏堢骇闂
- [x] 鏂囨。瀹℃煡鍜屾竻鐞嗭紙绉诲姩 4 涓繃鏃舵枃妗ｅ埌 deprecated锛?
- [x] 鍒涘缓鏂囨。瀹℃煡宸ヤ綔娴侊紙document-audit.yml锛?
- [x] 鍒涘缓 Claude Code 宸ヤ綔娴佽鎺ユ寚鍗?
- [x] 宸ヤ綔娴佹暣鐞嗭紙鏇存柊宸ヤ綔娴佹敞鍐岃〃锛屽垱寤哄伐浣滄祦绱㈠紩锛?
- [x] 鏂囦欢澶归噸缁勶紙鍒涘缓 reports/ 鍜?docs/deprecated/workflows/锛?
- [x] 宸ヤ綔娴佸簾寮冨唴瀹规鏌ュ拰椋庨櫓璇勪及
- [x] 鍒犻櫎鏈娇鐢ㄧ殑 docs job
- [x] 宸ヤ綔娴佺粡楠屾€荤粨涓庡畬鍠勶紙鍒涘缓 registry-validator.yml锛?
- [x] 姣忔棩澶嶇洏鎶ュ憡
- [x] 閲岀▼纰?楠屾敹閫氳繃
- [x] 鐢熸垚閲岀▼纰?楠屾敹鎶ュ憡
- [x] 鍒涘缓閲岀▼纰?瑙勬牸鏂囨。楠ㄦ灦锛圖AG骞惰璋冨害銆佸绾﹀噯鍏ヨ鍒欍€丼LA銆侀敊璇爜锛?
- [x] 琛ラ綈閲岀▼纰? workflow binding锛堢粦瀹?wf-doc-004 + wf-pr-check + wf-ci + wf-doc-006锛?
- [x] 閲岀▼纰? workflow binding 宸茬‘璁?
- [x] T1 鍩虹嚎楠岃瘉瀹屾垚锛堟湰鍦?CI 绛変环瑕嗙洊鐜?62.8%锛岃秴杩?60% 闂ㄧ锛?
- [x] T2 scheduler ready-set API 瀹屾垚锛坄GetReadyTasks(limit int)`锛孋I 绛変环瑕嗙洊鐜?63.6%锛?
- [x] 瀹夎骞惰繍琛?GitHub Actions 绛変环宸ュ叿锛歚staticcheck`銆乣gosec`銆乣govulncheck`銆乣markdownlint`
- [x] T2.5 鏅€?CLI Bash-first 璇箟淇瀹屾垚锛圕I 绛変环瑕嗙洊鐜?67.3%锛?
- [x] T3 濂戠害鍑嗗叆灞傚疄鐜板畬鎴愶紙CI 绛変环瑕嗙洊鐜?68.1%锛?
- [x] T4 SLA parsing 涓庢墽琛岃秴鏃跺疄鐜板畬鎴愶紙CI 绛変环瑕嗙洊鐜?69.2%锛?
- [x] T5 orchestrator 骞惰鎵ц寰幆瀹炵幇瀹屾垚锛圕I 绛変环瑕嗙洊鐜?69.3%锛?
- [x] T6 error code 鍩虹瀹炵幇瀹屾垚
- [x] T7 CLI/docs 鏇存柊涓庨獙鏀跺畬鎴?
- [x] 娴嬭瘯瑕嗙洊鐜囨彁鍗囪嚦 75.7%
- [x] **Milestone 2 鍏ㄩ儴瀹屾垚**
- [x] 浠婃棩宸ヤ綔鎸夊敮涓€涓婃父 workflow 鍏ㄩ噺褰掔被銆佺粡楠岃瘎瀹″苟鍥哄寲鍥炲伐浣滄祦瑙勫垯
- [x] 鏍稿績鏂囨。鎸夎嚜鍥犲寲 / Autogenesis 璁捐鎬濇兂閲嶅啓鍏ュ彛瀹氫綅
- [x] 鍒涘缓 CLAUDE.md 鐢ㄤ簬 Claude Code 闆嗘垚锛堝寘鍚畬鏁撮」鐩笂涓嬫枃銆佹瀯寤哄懡浠ゃ€佹灦鏋勬瑕侊級
- [x] GitHub CLI (gh v2.92.0) 瀹夎骞惰璇佷负 Oseas031
- [x] Pre-commit hook 淇锛歐indows Python 鍏煎锛坆ash 鍖呰鍣級銆佹敞鍐岃〃璺緞鏇存柊銆乁nicode 瀹夊叏杈撳嚭
- [x] Registry 淇锛氭敞鍐?wf-entry銆佷慨澶?wf-release 鏂囦欢寮曠敤銆佷緷璧栭摼涓€鑷存€?
- [x] CI Workflow 淇锛歳egistry-validator scope bug銆乧i.yml 姝绘潯浠躲€乨ocument-audit M2 璇箟銆丆ODING_STANDARDS 鏇存
- [x] PR Quality Check 淇锛歞ocumentation-check git diff 娴呭厠闅嗗け璐?鈫?鍏ㄩ儴 4 涓?job 閫氳繃
- [x] Monitoring 鏁呴殰璇婃柇锛? 涓?job 澶辫触鏍瑰洜瀹氫綅锛屼慨澶嶅湪 milestone1-acceptance 鍒嗘敮灏辩华
- [x] lmh-harness-v1 宸ョ▼鏂规硶璁烘帴鍏?
- [x] 椤圭洰璁板繂绯荤粺鍒濆鍖栵紙GitHub CLI first 鍋忓ソ锛?
- [x] M3 Phase 1: ModelProvider 鎺ュ彛 + MockModelProvider锛坧rovider.go, mock.go锛?
- [x] M3 Phase 1: Dispatcher 鈫?ContractExecutor 鈫?ModelProvider 鎵ц璺緞鎵撻€?
- [x] M3 Phase 1: `ErrDependencyNotReady` 閿欒鐮?+ `sla.failure_class` SLA 甯搁噺
- [x] M3 Phase 1: 澶辫触渚濊禆澶勭悊锛坒ailed = done锛屼笉鍐嶆案涔呴樆濉炰笅娓革級
- [x] M3 Phase 1: `types_test.go` 瑕嗙洊 AgentError/ErrorCode/SLA/FieldType/TaskStatus
- [x] M3 Phase 1: orchestrator 閲嶈瘯鑰楀敖娴嬭瘯锛堣緭鍑洪獙璇佸け璐ヨЕ鍙?鈫?閲嶈瘯 鈫?鑰楀敖 鈫?Failed锛?
- [x] M3 Phase 1: cmd/axis shell stdin 妯℃嫙娴嬭瘯锛坔elp/exit/unknown/run/status/empty/quit锛?
- [x] M3 Phase 1: dispatcher 鐖?context 鍙栨秷 + errChan 璺緞娴嬭瘯
- [x] M3 Phase 1: executor SetProvider + Execute with provider + ValidateOutput 娴嬭瘯
- [x] M3 Phase 1: admission 绌?鏈夋晥 failure_class SLA 娴嬭瘯
- [x] 娴嬭瘯瑕嗙洊鐜囨彁鍗囪嚦 88.8%锛堣秴杩?85% 鐩爣锛?
- [x] Worktree 闅旂鏈哄埗缂洪櫡璋冩煡锛圗nterWorktree 鍩轰簬榛樿鍒嗘敮 main HEAD锛岄潪褰撳墠鍒嗘敮 HEAD锛?
- [x] 鎵嬪姩 worktree 骞惰寮€鍙戞柟妗?B 楠岃瘉锛坓it worktree add -b + EnterWorktree --path锛?

## 宸插畬鎴愪换鍔★紙M3 Phase 2锛?
- [x] ModelProvider 鍙厤缃寲锛團unctional Options Pattern: WithModelProvider锛?
- [x] EchoModelProvider 鏂板锛堝尯鍒簬 MockModelProvider锛?
- [x] NewProvider 宸ュ巶鍑芥暟锛堟敮鎸?"mock"銆?echo"锛?
- [x] DAG 澧炲己锛欸etAllTasks銆丟etDependencyGraph锛坰cheduler + orchestrator锛?
- [x] Shell dag 鍛戒护锛堝彲璇讳緷璧栧浘杈撳嚭锛?
- [x] HumanExecutor 璺敱锛歍askMetadataKeyExecutor + dispatcher executeHumanTask
- [x] HumanExecutor 杞绛夊緟 + 瓒呮椂鏈哄埗
- [x] Orchestrator ResolveCall 鏆撮湶 + Shell resolve 鍛戒护
- [x] 娴嬭瘯瑕嗙洊锛歱rovider registry銆乻cheduler DAG銆乨ispatcher human routing銆乻hell dag/resolve
- [x] 瑕嗙洊鐜囦繚鎸?86.8%锛堣秴杩?85% 鐩爣锛?

## 宸插畬鎴愪换鍔★紙M3 Phase 3锛?
- [x] SLA 绫诲瀷甯搁噺锛歠ailure_class (retryable/fatal/degradable)銆乸riority銆乥ackoff
- [x] Failure class 璺敱 + 閫€閬跨瓥鐣ワ紙orchestrator: parseSLA 鎵╁睍銆乥ackoffDelay銆乫atal/degradable 鍒嗘敮锛?
- [x] 浼樺厛绾ф帓搴忥紙scheduler: GetReadyTasks 鎸?priority 闄嶅簭锛屽悓浼樺厛绾?FIFO锛?
- [x] SLA admission 楠岃瘉鎵╁睍锛坧riority 0-255銆乥ackoff enum銆乫ailure_class enum锛?
- [x] Tool 鎺ュ彛 + ToolRegistry锛堝彲鎻掓嫈宸ュ叿娉ㄥ唽锛?
- [x] BashTool锛坥s/exec 鎵ц锛?0s 瓒呮椂锛岃繑鍥?stdout/exit_code锛?
- [x] ModelRequest/ModelResponse 鎵╁睍锛圱ools + History + ToolCalls 瀛楁锛?
- [x] MockModelProvider tool-aware锛堟ā鎷熷杞?tool-use锛?
- [x] Multi-turn 鎵ц寰幆锛圕ontractExecutor: provider 鈫?tool 鈫?provider锛宮ax 10 turns锛?
- [x] Orchestrator 缁勮锛圱oolRegistry + BashTool 娉ㄥ叆锛?
- [x] 娴嬭瘯瑕嗙洊 24+ 鏂板鐢ㄤ緥锛岃鐩栫巼 87.1%
- [x] axis-dev.exe 缂栬瘧閫氳繃

## 宸插畬鎴愪换鍔★紙M5 - Bootstrap Loop锛?
- [x] AgentExecutor 鎺ュ彛 + MockAgentExecutor 瀹炵幇
- [x] AgentRuntimeAdapter锛堝閮?Agent CLI 鏀寔锛?
- [x] Orchestrator AgentExecutor 娉ㄥ叆锛圵ithAgentExecutor option锛?
- [x] SelfContext 鏁版嵁缁撴瀯锛圱askID, TaskLineage, CodeSnapshot, DocSnapshot, StateSnapshot锛?
- [x] ContextBuilder 瀹炵幇锛圔uildSelfContext 鏂规硶锛?
- [x] ContextCompressor锛堜笂涓嬫枃鍘嬬缉锛?绉嶇瓥鐣ワ級
- [x] Self-iteration Contracts锛坅nalyze/implement/validate/update/review/spawn锛?
- [x] BootstrapOrchestrator锛堣嚜寰幆浠诲姟璋冨害 + loop tracking锛?
- [x] FollowUpTaskGenerator锛堜粠鎵ц缁撴灉鐢熸垚鍚庣画浠诲姟锛?
- [x] AutonomyTransition 鏁版嵁妯″瀷锛?绾?autonomy level锛?
- [x] RuleEngine锛堝熀浜?competence evidence 鐨勮鍒欏紩鎿庯級
- [x] 闆嗘垚娴嬭瘯锛團ull DAG workflow, concurrent tracking锛?
- [x] M5 鏂囨。鏇存柊锛坮equirements.md, design.md, tasks.md 鏍囪涓?Complete锛?
- [x] Phase 2: Sandboxed Evolution T1-T10 全部完成 (数据模型/存储/工作空间/账本/验证/inspect/promote/discard/测试/文档)
- [x] T2: 数据模型定义 (`EvolutionIntent`, `EvolutionRun`, `EvolutionStep`, `VerificationRecord`, `EvolutionDecision`)
- [x] T3: 项目本地存储实现 (`Store` 原子 JSON 写入，CreateRun/ReadRun/ListRuns)
- [x] T4: 隔离工作空间 (`Workspace` 创建/复制/写入/读取/列表/晋升)
- [x] T5: 追踪账本 (`Ledger` append-only JSONL，ReadSteps 容错 + strict 模式)
- [x] T6: 验证记录捕获 (`Verifier.Run` 命令执行，stdout/stderr/exitCode 捕获，Windows 兼容修复)
- [x] T7: inspect 命令 (`axis evolve inspect <run-id>` JSON 输出全量 run 信息)
- [x] T8: promote/discard 门控 (`DecisionGate.CanPromote/CanDiscard/Promote/Discard`，显式晋升/丢弃)
- [x] T9: 安全与回归测试 (decision 阻断测试、重复晋升/丢弃拒绝、命令格式修复)
- [x] T10: 文档更新 (tasks.md 标记全部 Completed)
- [x] Code Review 修复: 移除调试日志、Command 格式改为 strings.Join
- [x] 架构加固: decision.json → decisions.jsonl append-only，阻止非法状态反转
- [x] **Agent Context Query Model T8 完成**: `context.requested_sources` 元数据键、`ExecutionContextSummary` 三字段扩展、`AgentExecutionRequest.RequestedSources`、dispatcher 消除重复解析、review 修复（补充零值断言/untraceable+requested 测试/设计意图注释）

## 杩涜涓换鍔?
- [x] M6 Phase 6.4-6.5: Integration + Testing + Documentation (T14-T18 全部完成)
- [x] Phase 3: Agent 记忆系统评估与增强

## 杩涜涓换鍔?（axis-gui 工具链）
- [x] axis-gui 连接修复：绝对路径解析解决 Go 1.19+ exec 安全限制
- [x] axis-gui 代理修复：header-first 写入顺序修复 HTTP 500
- [x] axis-gui 字体 CDN 修复：替换失效 fontsource URL
- [x] axis-gui 合同 ID 支持：前端 submitTask 增加 contract_id 字段
- [x] axis-gui 错误增强：从后端 response 提取 message 字段准确展示
- [x] T1: scheduler crash recovery 测试 + 实现（stale Running → Failed）
- [x] T2: orchestrator busy-poll 移除（time.After 替换为 taskSubmitted channel 信号）
- [x] T3: 全量回归测试通过 go test ./...
- [x] T5: TasksPage WebSocket 实时集成（5s 轮询 → WebSocket 驱动 + 30s fallback）
- [x] T5: WebSocket 连接状态可视化（live badge 动态反映连接状态）
- [x] T6: 任务时间线聚合（按 task_id group events，展开显示完整事件时间线）
- [x] T7: 暗色模式系统偏好监听（未手动设置时自动跟随 prefers-color-scheme）

## 寰呭鐞嗕换鍔?
- [ ] M4: 鏇村宸ュ叿锛堟枃浠惰鍐欍€丠TTP client 绛夛級
- [ ] M4: 鐪熷疄 LLM 闆嗘垚
- [ ] M4: 瀹夊叿娌欑

## 閬囧埌鐨勯棶棰?
- 鉁?staticcheck ST1003 - 宸蹭慨澶嶏紙commit 1d9aaef, 37f23c0锛?
- 鉁?godoc -html 搴熷純鍙傛暟 - 宸蹭慨澶嶏紙commit 457b30a锛?
- 鉁?鏋氫妇楠岃瘉涓嶆敮鎸?int 绫诲瀷 - 宸蹭慨澶嶏紙commit 5c4231f锛?
- 鉁?鏂囨。杩囨椂闂 - 宸叉竻鐞嗭紙commit b323b7d锛?
- 鉁?缂哄皯鏂囨。瀹℃煡宸ヤ綔娴?- 宸插垱寤猴紙commit bb2045f锛?
- 鉁?宸ヤ綔娴佹敞鍐岃〃涓嶄竴鑷?- 宸叉暣鐞嗭紙commit f1fde53锛?
- 鉁?鏈娇鐢ㄥ唴瀹?- 宸查儴鍒嗕慨澶嶏紙docs job 鍒犻櫎锛宑ommit 27b94c5锛?
- 鉁?release.yml 涓?cd-workflow 閲嶅 - 宸插鐞嗭紙release.yml 宸插垹闄わ紝registry 鏍囪 deprecated锛?
- 鈿狅笍 sign-artifacts job 鏈娇鐢?- 寰呭鐞嗭紙閲岀▼纰?鍚庯級
- 鉁?T1 GitHub CI 绛変环瑕嗙洊鐜囬棬绂佸凡杈炬爣锛氭€昏鐩栫巼 62.8%
- 鉁?T2 鍚?GitHub CI 绛変环瑕嗙洊鐜囬棬绂佷粛杈炬爣锛氭€昏鐩栫巼 63.6%
- 鉁?`staticcheck ./...` 鏈湴閫氳繃
- 鉁?`gosec ./...` 鏈湴閫氳繃锛孖ssues: 0
- 鉁?`govulncheck ./...` 鏈湴閫氳繃
- 鉁?T2.5 鍚?GitHub CI 绛変环瑕嗙洊鐜囬棬绂佷粛杈炬爣锛氭€昏鐩栫巼 67.3%
- 鉁?T3 鍚?GitHub CI 绛変环瑕嗙洊鐜囬棬绂佷粛杈炬爣锛氭€昏鐩栫巼 68.1%
- 鉁?T5 鍚?GitHub CI 绛変环瑕嗙洊鐜囬棬绂佷粛杈炬爣锛氭€昏鐩栫巼 69.3%
- 鉁?娴嬭瘯瑕嗙洊鐜囨彁鍗囪嚦 75.7%锛堣秴杩?75% 鐩爣锛?
- 鉁?娴嬭瘯瑕嗙洊鐜囪繘涓€姝ユ彁鍗囪嚦 88.8%锛堣秴杩?85% 鐩爣锛?
- 鈿狅笍 Isolation worktree 鍩轰簬鏃?commit锛坢ain HEAD锛夎€岄潪褰撳墠鍒嗘敮 HEAD 鈫?宸茶皟鏌ユ牴鍥狅紝閲囩敤鎵嬪姩 worktree 鏂规 B 瑙勯伩
- 鈿狅笍 Windows 涓嶆敮鎸佺▼搴忓寲淇″彿鍙戦€?鈫?SIGINT 鐩稿叧娴嬭瘯鏃犳硶鍦?Windows 杩愯锛屽凡绉婚櫎
- 鈿狅笍 `markdownlint "**/*.md"` 鏈湴鍙戠幇鏃㈡湁 Markdown 椋庢牸闂锛涗笌 `document-audit.yml` 涓€鑷达紝璇ユ鏌ュ綋鍓嶄负闈為樆濉炲璁￠」
- 鉁?宸ヤ綔娴佸鐩樺凡杩藉姞鍒?`reports/daily/workflow-system-retrospective-2026-05-08.md`
- 鉁?澶嶇洏缁忛獙宸插浐鍖栧埌 `workflow/entry.md`銆乣workflow/meta-workflow-management.md`銆乣workflow/occams-razor-architecture-simplification.md`
- 鉁?PR Quality Check git diff 娴呭厠闅嗗け璐?- 宸蹭慨澶嶏紙commit f9962de锛屾坊鍔?fetch-depth:0 + || true锛?
- 鉁?Monitoring 3 涓?job 澶辫触 - 宸插湪 milestone1-acceptance 鍒嗘敮淇锛岀瓑 PR 鍚堝苟鍒?main 鐢熸晥

## 涓嬩竴姝ヨ鍔?
1. 鎻愪氦 M3 Phase 3 浠ｇ爜
2. 鍒涘缓 PR 鍒?main 瑙﹀彂 CI 楠岃瘉
3. 瑙勫垝 M4锛堢湡瀹?LLM 闆嗘垚銆佸畨鍏ㄦ矙绠憋級

## 閲嶈鎻愰啋
- Milestone 1 鉁?| Milestone 2 鉁?| Milestone 3 Phase 1 鉁?| Phase 2 鉁?| Phase 3 鉁?宸插畬鎴?
- 瑕嗙洊鐜?87.1%锛岃秴杩?85% 鐩爣
- SLA 绛栫暐寮曟搸锛氭敮鎸?failure_class 璺敱锛坮etryable/fatal/degradable锛夈€侀€€閬跨瓥鐣ャ€佷紭鍏堢骇鎺掑簭
- 宸ュ叿璋冪敤灞傦細Tool 鎺ュ彛 + BashTool + 澶氳疆鎵ц寰幆锛岄涓伐鍏峰嵆 Bash
- 閬靛惊濂ュ崱濮嗗墐鍒€鍘熷垯
- 缁х画淇濇寔 CLI-first / shell-native锛屼笉寮曞叆 Web UI 鎴栭噸鍨?TUI
- 鎵€鏈夊伐浣滆繘搴﹀繀椤昏褰曞湪鏂囨。涓?
- 浜ゆ帴鍓嶅繀椤诲畬鎴愪氦鎺ユ鏌ユ竻鍗?
- worktree 闅旂鏈夊凡鐭ョ己闄凤紙鍩轰簬 main HEAD锛夛紝骞惰寮€鍙戜娇鐢ㄦ墜鍔?worktree锛堟柟妗?B锛?

## 鏈€杩戞彁浜?
- (latest) - feat: Agent Context Query Model T8 — `context.requested_sources` + `ExecutionContextSummary` 三字段 + `AgentExecutionRequest.RequestedSources` + review 修复
- (latest) - refactor: dispatcher 消除 `parseRequestedSources` 重复解析，改为从 summary 复用
- (latest) - test: 补充 untraceable+requested 测试、零值断言、注释说明设计意图
- 85f9877 - merge: worktree B 鈥?dispatcher (95.5%), executor (94.3%), admission (100%)
t- cd63a28 - feat: M3 Phase 2 鈥?ModelProvider configurable, HumanExecutor routing, DAG enhancement
- 44e4f7c - test: raise dispatcher to 95.5%, executor to 94.3%, admission to 100%
- a73ef20 - test: raise overall coverage to 86.2% (cmd/axis 68%, orchestrator 87%)
- 3a9da92 - test: add types_test.go covering AgentError, ErrorCode, SLA keys, core types
- a2ea1e2 - feat: add ModelProvider, ErrDependencyNotReady, sla.failure_class (M3 Phase 1)
- 4d9af2d - feat: add structured error codes (T6) and update docs (T7)

## 褰撳墠瑙勬牸鏂囨。
- Milestone 2 Requirements: `docs/specs/milestone2/requirements.md`
- Milestone 2 Design: `docs/specs/milestone2/design.md`
- Milestone 2 Tasks: `docs/specs/milestone2/tasks.md`
- Milestone 2 Workflow Binding: `docs/specs/milestone2/workflow-binding.md`

## Architecture Diagnosis & Strategic Direction (2026-05-11)

**Full analysis**: `reports/strategy/agent-native-scenario-principles-analysis-2026-05-11.md`
**First principles reference**: `docs/architecture/agent-native-first-principles.md`

### Top 8 Core Gaps Identified

| # | Gap | Severity | Impact on Scenarios | Status |
|---|---|---|---|---|
| A | Cross-process context fracture: ReadinessRegistry is in-process only; no local-control-plane awareness | **Critical** | Breaks multi-terminal workflow, cluster ops, enterprise audit | Partial: Local Control Plane T1-T8 completed |
| B | Orchestrator is pseudo-parallel single-thread loop; no inter-Agent collaboration primitives | **Critical** | Blocks AI-native startup pipelines, digital workforce chaining | Open |
| C | Event log is append-only but lacks structured query API or feedback loop | **High** | Prevents competence-based autonomy, organizational intelligence | Open |
| D | ~~Sandboxed Evolution is spec-only; zero implementation~~ | **Critical** | ~~"Controllable Evolution" = "No Evolution"~~ | **RESOLVED** — T2-T10 fully implemented (2026-05-11) |
| E | Tool boundaries are static fences, not dynamic ladders | **High** | "Competence earns autonomy" remains unimplemented | Open |
| F | Model routing is manual gearbox; no latency/cost-aware dynamic scheduling or fallback | **High** | Token cost失控, no auto-degradation on provider failure | Open |
| G | No Agent identity or capability profile; only AgentTask exists | **High** | Cannot route tasks to best-fit Agent; "Capability is decision right" needs identity | Open |
| H | Execution feedback loop is fully broken; no quality assessment or system improvement from results | **High** | Same errors repeat; system is open-loop, not closed-loop | Open |

### Design Philosophy Assessment

- **Still fully applicable**: "More Context, More Action, Zero Control", "bash is all you need", "Interface is existence", "Contract is structure"
- **Partially resolved**: "Query is context" — T8 implemented: Agent declares needs via `context.requested_sources`, system resolves against readiness registry and reports satisfied/missing. Full Agent-driven demand (P1+) remains planned.
- **Needs refinement**: "Ladder is boundary" (static fences vs dynamic ladders)
- **Resolved**: "Controllable Evolution" — Sandboxed Evolution P0 fully implemented with isolated workspaces, atomic steps, verification gates, and explicit promote/discard decisions
- **Fundamental challenge**: Autonomy-reliability tension requires graduated autonomy (P0 high-control, P1 partial, P2 full-in-sandbox)

### Recommended Next Priority Order

1. ~~**Sandboxed Evolution P0 implementation**~~ — **COMPLETED** (2026-05-11)
2. **Cross-process state persistence**: Make ReadinessRegistry local-control-plane-aware
3. **Agent identity & capability profile**: Introduce Agent registry and behavioral scoring
4. **Event log structured query**: Add `axis audit` or equivalent for log consumption
5. **Dynamic model routing**: Cost/latency-aware provider selection with fallback chains
6. **Execution feedback loop**: Result quality scoring feeding back into intent/context assembly

