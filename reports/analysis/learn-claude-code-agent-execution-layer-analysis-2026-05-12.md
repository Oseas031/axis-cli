# learn-claude-code Agent 执行能力层设计分析

**日期**: 2026-05-12
**来源**: [shareAI-lab/learn-claude-code](https://github.com/shareAI-lab/learn-claude-code)
**目的**: 为 Axis 提供执行能力层设计参考

---

## 1. 核心设计哲学

### 1.1 Agent = Model + Harness

```
┌─────────────────────────────────────────────────────────────┐
│                        Agent                                 │
│  ┌─────────────┐                    ┌───────────────────┐   │
│  │    Model    │                    │     Harness       │   │
│  │  (Agency)   │                    │ (Tools/Knowledge/ │   │
│  │             │                    │  Observation/     │   │
│  │  感知/推理  │                    │  Action/Permission)│   │
│  │  /行动能力  │                    │                   │   │
│  └─────────────┘                    └───────────────────┘   │
│        │                                      │             │
│        └─────────── 决策者 ◄─── 执行者 ───────┘             │
└─────────────────────────────────────────────────────────────┘
```

**关键洞察**:
- Agency（感知、推理、行动能力）来自模型训练，不是外部代码编排
- Harness 是载具，Model 是驾驶者
- Harness 不试图成为 Agent 本身

### 1.2 核心循环

```python
def agent_loop(messages):
    while True:
        response = client.messages.create(model=MODEL, messages=messages, tools=TOOLS)
        messages.append({"role": "assistant", "content": response.content})
        
        if response.stop_reason != "tool_use":
            return
        
        results = []
        for block in response.content:
            if block.type == "tool_use":
                output = TOOL_HANDLERS[block.name](**block.input)
                results.append({"type": "tool_result", "tool_use_id": block.id, "content": output})
        messages.append({"role": "user", "content": results})
```

**特点**:
- 循环本身属于 Agent，机制属于 Harness
- 每个课程在循环之上叠加一个 harness 机制，循环本身始终不变

---

## 2. 12 个递进式课程架构

```
第一阶段: 循环                       第二阶段: 规划与知识
==================                   ==============================
s01  Agent Loop              [1]     s03  TodoWrite               [5]
s02  Tool Use                [4]     s04  Subagent               [5]
                                     s05  Skills                 [5]
                                     s06  Context Compact        [5]

第三阶段: 持久化                     第四阶段: 团队
==================                   =====================
s07  Task System             [8]     s09  Agent Teams            [9]
s08  Background Tasks        [6]     s10  Team Protocols        [12]
                                     s11  Autonomous Agents     [14]
                                     s12  Worktree Isolation    [16]

[N] = 工具数量
```

---

## 3. 执行能力层详细设计

### 3.1 s01: Agent Loop - 最小循环

**格言**: *"One loop & Bash is all you need"*

**架构**:
```
User --> messages[] --> LLM --> response
                              |
                    stop_reason == "tool_use"?
                   /                        \
                 yes                         no
                  |                           |
            execute tools                 return text
            append results
            loop back -------------> messages[]
```

**工具**:
| 工具 | 功能 | 安全措施 |
|------|------|----------|
| `bash` | 执行 shell 命令 | 危险命令黑名单、120s 超时、50000 字符截断 |

**安全设计**:
```python
def run_bash(command: str) -> str:
    dangerous = ["rm -rf /", "sudo", "shutdown", "reboot", "> /dev/"]
    if any(d in command for d in dangerous):
        return "Error: Dangerous command blocked"
    # ... timeout + truncation
```

---

### 3.2 s02: Tool Use - 工具分发

**格言**: *"加一个工具, 只加一个 handler"*

**架构**:
```
+----------+      +-------+      +------------------+
|   User   | ---> |  LLM  | ---> | Tool Dispatch    |
|  prompt  |      |       |      | {                |
+----------+      +---+---+      |   bash: run_bash |
                      ^          |   read: run_read |
                      |          |   write: run_wr  |
                      +----------+   edit: run_edit |
                      tool_result| }                |
                                 +------------------+
```

**核心设计**:
```python
TOOL_HANDLERS = {
    "bash":       lambda **kw: run_bash(kw["command"]),
    "read_file":  lambda **kw: run_read(kw["path"], kw.get("limit")),
    "write_file": lambda **kw: run_write(kw["path"], kw["content"]),
    "edit_file":  lambda **kw: run_edit(kw["path"], kw["old_text"], kw["new_text"]),
}
```

**工具集**:
| 工具 | 功能 | 安全措施 |
|------|------|----------|
| `bash` | Shell 命令执行 | 危险命令黑名单 |
| `read_file` | 文件读取 | 路径逃逸检查、行数限制 |
| `write_file` | 文件写入 | 路径逃逸检查、父目录自动创建 |
| `edit_file` | 文件编辑 | 精确文本替换、路径逃逸检查 |

**路径安全**:
```python
def safe_path(p: str) -> Path:
    path = (WORKDIR / p).resolve()
    if not path.is_relative_to(WORKDIR):
        raise ValueError(f"Path escapes workspace: {p}")
    return path
```

---

### 3.3 s03: TodoWrite - 任务跟踪

**格言**: *"没有计划的 agent 走哪算哪"*

**架构**:
```
+----------+      +-------+      +---------+
|   User   | ---> |  LLM  | ---> | Tools   |
|  prompt  |      |       |      | + todo  |
+----------+      +---+---+      +----+----+
                      ^               |
                      |   tool_result |
                      +---------------+
                            |
                +-----------+-----------+
                | TodoManager state     |
                | [ ] task A            |
                | [>] task B <- doing   |
                | [x] task C            |
                +-----------------------+
                            |
                if rounds_since_todo >= 3:
                  inject <reminder>
```

**设计要点**:
- 结构化状态管理（pending / in_progress / completed）
- 单一 in_progress 约束
- 自动 nag 提醒（3 轮未更新则注入提醒）
- 最多 20 个 todos

---

### 3.4 s04: Subagent - 上下文隔离

**格言**: *"大任务拆小, 每个小任务干净的上下文"*

**架构**:
```
Parent agent                     Subagent
+------------------+             +------------------+
| messages=[...]   |             | messages=[]      |  <-- fresh
|                  |  dispatch   |                  |
| tool: task       | ---------->| while tool_use:  |
|   prompt="..."   |            |   call tools     |
|   description="" |            |   append results |
|                  |  summary   |                  |
|   result = "..." | <--------- | return last text |
+------------------+             +------------------+
          |
Parent context stays clean.
Subagent context is discarded.
```

**核心设计**:
```python
def run_subagent(prompt: str) -> str:
    sub_messages = [{"role": "user", "content": prompt}]  # fresh context
    for _ in range(30):  # safety limit
        response = client.messages.create(model=MODEL, messages=sub_messages, tools=CHILD_TOOLS)
        # ... tool execution loop
    # Only the final text returns to the parent -- child context is discarded
    return "".join(b.text for b in response.content if hasattr(b, "text"))
```

**关键点**:
- Subagent 用独立 `messages=[]`，不污染主对话
- 子 Agent 工具子集（无递归 spawning）
- 仅返回摘要给父 Agent

---

### 3.5 s05: Skills - 按需知识加载

**格言**: *"用到什么知识, 临时加载什么知识"*

**两层注入架构**:
```
Layer 1 (cheap): skill names in system prompt (~100 tokens/skill)
Layer 2 (on demand): full skill body in tool_result

skills/
  pdf/
    SKILL.md          <-- frontmatter (name, description) + body
  code-review/
    SKILL.md

System prompt:
+--------------------------------------+
| You are a coding agent.              |
| Skills available:                    |
|   - pdf: Process PDF files...        |  <-- Layer 1: metadata only
|   - code-review: Review code...      |
+--------------------------------------+

When model calls load_skill("pdf"):
+--------------------------------------+
| tool_result:                         |
| <skill>                              |
|   Full PDF processing instructions   |  <-- Layer 2: full body
|   Step 1: ...                        |
|   Step 2: ...                        |
| </skill>                             |
+--------------------------------------+
```

**SKILL.md 结构**:
```yaml
---
name: pdf
description: Process PDF files - extract text, create PDFs, merge documents.
tags: document, pdf
---

# PDF Processing Skill

You now have expertise in PDF manipulation...

## Reading PDFs
...

## Creating PDFs
...
```

---

### 3.6 s06: Context Compact - 三层压缩

**格言**: *"上下文总会满, 要有办法腾地方"*

**三层压缩管道**:
```
Every turn:
+------------------+
| Tool call result |
+------------------+
        |
        v
[Layer 1: micro_compact]        (silent, every turn)
  Replace non-read_file tool_result content older than last 3
  with "[Previous: used {tool_name}]"
        |
        v
[Check: tokens > 50000?]
   |               |
   no              yes
   |               |
   v               v
continue    [Layer 2: auto_compact]
              Save full transcript to .transcripts/
              Ask LLM to summarize conversation.
              Replace all messages with [summary].
                    |
                    v
            [Layer 3: compact tool]
              Model calls compact -> immediate summarization.
```

**关键设计**:
- Layer 1: 每轮自动执行，保留最近 3 个 tool_result
- Layer 2: 阈值触发，保存完整 transcript 到磁盘
- Layer 3: 手动触发，Agent 主动压缩
- 保留 `read_file` 结果（避免重复读取）

---

### 3.7 s07: Task System - 持久化任务图

**格言**: *"大目标要拆成小任务, 排好序, 记在磁盘上"*

**架构**:
```
.tasks/
  task_1.json  {"id":1, "subject":"...", "status":"completed", ...}
  task_2.json  {"id":2, "blockedBy":[1], "status":"pending", ...}
  task_3.json  {"id":3, "blockedBy":[2], ...}

Dependency resolution:
+----------+     +----------+     +----------+
| task 1   | --> | task 2   | --> | task 3   |
| complete |     | blocked  |     | blocked  |
+----------+     +----------+     +----------+
     |                ^
     +--- completing task 1 removes it from task 2's blockedBy
```

**任务模型**:
```python
task = {
    "id": 1,
    "subject": "Implement auth",
    "description": "...",
    "status": "pending",  # pending | in_progress | completed
    "owner": "",
    "worktree": "",
    "blockedBy": [],
    "created_at": timestamp,
    "updated_at": timestamp,
}
```

**工具集**:
| 工具 | 功能 |
|------|------|
| `task_create` | 创建新任务 |
| `task_get` | 获取任务详情 |
| `task_update` | 更新任务状态/依赖 |
| `task_list` | 列出所有任务 |

---

### 3.8 s08: Background Tasks - 后台执行

**格言**: *"慢操作丢后台, agent 继续想下一步"*

**架构**:
```
Main thread                Background thread
+-----------------+        +-----------------+
| agent loop      |        | task executes   |
| ...             |        | ...             |
| [LLM call] <---+------- | enqueue(result) |
|  ^drain queue   |        +-----------------+
+-----------------+

Timeline:
Agent ----[spawn A]----[spawn B]----[other work]----
             |              |
             v              v
          [A runs]      [B runs]        (parallel)
             |              |
             +-- notification queue --> [results injected]
```

**核心实现**:
```python
class BackgroundManager:
    def run(self, command: str) -> str:
        task_id = str(uuid.uuid4())[:8]
        self.tasks[task_id] = {"status": "running", ...}
        threading.Thread(target=self._execute, args=(task_id, command), daemon=True).start()
        return f"Background task {task_id} started"
    
    def drain_notifications(self) -> list:
        """Return and clear all pending completion notifications."""
        with self._lock:
            notifs = list(self._notification_queue)
            self._notification_queue.clear()
        return notifs
```

**工具集**:
| 工具 | 功能 |
|------|------|
| `background_run` | 启动后台任务 |
| `check_background` | 检查任务状态 |

---

### 3.9 s09: Agent Teams - 多 Agent 协作

**格言**: *"任务太大一个人干不完, 要能分给队友"*

**架构**:
```
.team/config.json                   .team/inbox/
+----------------------------+      +------------------+
| {"team_name": "default",   |      | alice.jsonl      |
|  "members": [              |      | bob.jsonl        |
|    {"name":"alice",        |      | lead.jsonl       |
|     "role":"coder",        |      +------------------+
|     "status":"idle"}       |
|  ]}                        |      send_message("alice", "fix bug"):
+----------------------------+        open("alice.jsonl", "a").write(msg)

Thread: alice             Thread: bob
+------------------+      +------------------+
| agent_loop       |      | agent_loop       |
| status: working  |      | status: idle     |
+------------------+      +------------------+
```

**消息类型**:
| 类型 | 用途 |
|------|------|
| `message` | 普通文本消息 |
| `broadcast` | 广播给所有队友 |
| `shutdown_request` | 请求优雅关闭 |
| `shutdown_response` | 关闭确认 |
| `plan_approval_response` | 计划审批响应 |

**工具集**:
| 工具 | 功能 |
|------|------|
| `spawn_teammate` | 生成持久化队友 |
| `list_teammates` | 列出所有队友 |
| `send_message` | 发送消息到队友邮箱 |
| `read_inbox` | 读取并清空自己的邮箱 |
| `broadcast` | 广播消息 |

---

### 3.10 s10: Team Protocols - 团队协议

**格言**: *"队友之间要有统一的沟通规矩"*

**Shutdown FSM**:
```
Lead                              Teammate
+---------------------+          +---------------------+
| shutdown_request     |          |                     |
| {                    | -------> | receives request    |
|   request_id: abc    |          | decides: approve?   |
| }                    |          |                     |
+---------------------+          +---------------------+
                                         |
+---------------------+          +-------v-------------+
| shutdown_response    | <------- | shutdown_response   |
| {                    |          | {                   |
|   request_id: abc    |          |   request_id: abc   |
|   approve: true      |          |   approve: true     |
| }                    |          | }                   |
+---------------------+          +---------------------+
        |
        v
status -> "shutdown", thread stops
```

**Plan Approval FSM**:
```
Teammate                          Lead
+---------------------+          +---------------------+
| plan_approval        |          |                     |
| submit: {plan:"..."}| -------> | reviews plan text   |
+---------------------+          | approve/reject?     |
                                 +---------------------+
                                         |
+---------------------+          +-------v-------------+
| plan_approval_resp   | <------- | plan_approval       |
| {approve: true}      |          | review: {req_id,    |
+---------------------+          |   approve: true}     |
                                 +---------------------+
```

**request_id 关联模式**:
```python
shutdown_requests = {}  # {request_id: {"target": name, "status": "pending|..."}}
plan_requests = {}      # {request_id: {"from": name, "plan": "...", "status": "..."}}
```

---

### 3.11 s11: Autonomous Agents - 自主 Agent

**格言**: *"队友自己看看板, 有活就认领"*

**生命周期**:
```
+-------+
| spawn |
+---+---+
    |
    v
+-------+  tool_use    +-------+
| WORK  | <----------- |  LLM  |
+---+---+              +-------+
    |
    | stop_reason != tool_use
    v
+--------+
| IDLE   | poll every 5s for up to 60s
+---+----+
    |
    +---> check inbox -> message? -> resume WORK
    |
    +---> scan .tasks/ -> unclaimed? -> claim -> resume WORK
    |
    +---> timeout (60s) -> shutdown
```

**Identity Re-injection**:
```python
def make_identity_block(name: str, role: str, team_name: str) -> dict:
    return {
        "role": "user",
        "content": f"<identity>You are '{name}', role: {role}, team: {team_name}.</identity>",
    }

# After compression, re-inject identity
if len(messages) <= 3:
    messages.insert(0, make_identity_block(name, role, team_name))
    messages.insert(1, {"role": "assistant", "content": f"I am {name}. Continuing."})
```

**自动认领**:
```python
def scan_unclaimed_tasks() -> list:
    unclaimed = []
    for f in TASKS_DIR.glob("task_*.json"):
        task = json.loads(f.read_text())
        if (task.get("status") == "pending"
                and not task.get("owner")
                and not task.get("blockedBy")):
            unclaimed.append(task)
    return unclaimed
```

---

### 3.12 s12: Worktree Isolation - 目录隔离

**格言**: *"各干各的目录, 互不干扰"*

**架构**:
```
.tasks/task_12.json
  {
    "id": 12,
    "subject": "Implement auth refactor",
    "status": "in_progress",
    "worktree": "auth-refactor"
  }

.worktrees/index.json
  {
    "worktrees": [
      {
        "name": "auth-refactor",
        "path": ".../.worktrees/auth-refactor",
        "branch": "wt/auth-refactor",
        "task_id": 12,
        "status": "active"
      }
    ]
  }
```

**EventBus**:
```python
class EventBus:
    def emit(self, event: str, task: dict = None, worktree: dict = None, error: str = None):
        payload = {"event": event, "ts": time.time(), "task": task or {}, "worktree": worktree or {}}
        if error:
            payload["error"] = error
        with self.path.open("a") as f:
            f.write(json.dumps(payload) + "\n")
```

**工具集**:
| 工具 | 功能 |
|------|------|
| `worktree_create` | 创建 git worktree 并绑定任务 |
| `worktree_list` | 列出所有 worktree |
| `worktree_status` | 查看 worktree git 状态 |
| `worktree_run` | 在指定 worktree 中执行命令 |
| `worktree_keep` | 保留 worktree（不删除） |
| `worktree_remove` | 移除 worktree 并可选完成任务 |
| `worktree_events` | 查看生命周期事件 |

---

## 4. 执行能力层对比

### 4.1 learn-claude-code vs Axis

| 维度 | learn-claude-code | Axis |
|------|-------------------|------|
| **工具定义** | dispatch map + JSON Schema | Tool interface + ToolRegistry |
| **权限模型** | 危险命令黑名单 + 路径逃逸检查 | Contract + Permission ladder |
| **上下文管理** | 三层压缩 (micro/auto/manual) | contextpack + ReadinessRegistry |
| **扩展方式** | Skills (SKILL.md) | Spec-RDT + Evolution Protocol |
| **任务系统** | JSON 文件持久化 + 依赖图 | AgentTask + Scheduler |
| **多 Agent** | Thread + JSONL 邮箱 | 无（单 Agent） |
| **隔离机制** | Git worktree | Sandboxed Evolution |
| **事件追踪** | EventBus (append-only JSONL) | .axis/events/tasks.jsonl |

### 4.2 Axis 可借鉴的设计点

| 来源 | 设计点 | Axis 应用建议 |
|------|--------|---------------|
| s02 | `safe_path` 路径逃逸检查 | 已实现，继续强化 |
| s03 | TodoWrite + nag 提醒 | 可作为 AgentTask 的轻量级跟踪层 |
| s04 | Subagent 上下文隔离 | 可用于复杂任务的并行分解 |
| s05 | Skills 两层加载 | 可替代部分 Spec-RDT 的功能 |
| s06 | 三层上下文压缩 | 可集成到 contextpack |
| s07 | 文件持久化任务图 | 与现有 Scheduler 互补 |
| s08 | 后台任务 + 通知队列 | 可用于长时间运行的任务 |
| s09 | JSONL 邮箱协议 | 可用于多 Agent 协作 |
| s10 | request_id 关联模式 | 可用于分布式事务 |
| s11 | Identity re-injection | 压缩后身份恢复 |
| s11 | 自动认领 unclaimed tasks | Agent 自主性增强 |
| s12 | worktree 隔离 | 与 Sandboxed Evolution 结合 |

---

## 5. 具体借鉴建议

### 5.1 短期 (P0)

1. **Skills 系统**: 引入 SKILL.md 格式，作为轻量级知识注入机制
2. **三层压缩**: 集成到 contextpack，实现无限会话
3. **TodoWrite**: 作为 AgentTask 的轻量级跟踪层

### 5.2 中期 (P1)

1. **Subagent 隔离**: 用于复杂任务的并行分解
2. **Background Tasks**: 长时间运行任务的异步执行
3. **EventBus**: 统一事件追踪格式

### 5.3 长期 (P2)

1. **多 Agent 协作**: JSONL 邮箱协议 + TeammateManager
2. **自动认领**: Agent 自主扫描任务看板
3. **Worktree 隔离**: 与 Sandboxed Evolution 深度集成

---

## 6. 关键代码模式

### 6.1 工具分发模式

```python
TOOL_HANDLERS = {
    "bash": lambda **kw: run_bash(kw["command"]),
    "read_file": lambda **kw: run_read(kw["path"], kw.get("limit")),
    # ... more tools
}

# In agent loop
for block in response.content:
    if block.type == "tool_use":
        handler = TOOL_HANDLERS.get(block.name)
        output = handler(**block.input) if handler else f"Unknown tool: {block.name}"
        results.append({"type": "tool_result", "tool_use_id": block.id, "content": str(output)})
```

### 6.2 安全路径检查模式

```python
def safe_path(p: str) -> Path:
    path = (WORKDIR / p).resolve()
    if not path.is_relative_to(WORKDIR):
        raise ValueError(f"Path escapes workspace: {p}")
    return path
```

### 6.3 JSONL 邮箱模式

```python
class MessageBus:
    def send(self, sender: str, to: str, content: str, msg_type: str = "message"):
        msg = {"type": msg_type, "from": sender, "content": content, "timestamp": time.time()}
        with open(INBOX_DIR / f"{to}.jsonl", "a") as f:
            f.write(json.dumps(msg) + "\n")
    
    def read_inbox(self, name: str) -> list:
        path = INBOX_DIR / f"{name}.jsonl"
        messages = [json.loads(l) for l in path.read_text().strip().splitlines() if l]
        path.write_text("")  # drain after read
        return messages
```

### 6.4 request_id 关联模式

```python
# Trackers
shutdown_requests = {}  # {request_id: {"target": name, "status": "pending|..."}}

# Request
def handle_shutdown_request(teammate: str) -> str:
    req_id = str(uuid.uuid4())[:8]
    shutdown_requests[req_id] = {"target": teammate, "status": "pending"}
    BUS.send("lead", teammate, "Please shut down.", "shutdown_request", {"request_id": req_id})
    return f"Shutdown request {req_id} sent to '{teammate}'"

# Response
def handle_shutdown_response(req_id: str, approve: bool):
    shutdown_requests[req_id]["status"] = "approved" if approve else "rejected"
```

---

## 7. 总结

learn-claude-code 提供了一个完整的、渐进式的 Agent Harness 工程教程。其核心贡献在于：

1. **清晰的架构分层**: Model 与 Harness 的明确分离
2. **渐进式复杂度**: 从 1 个工具到 16 个工具的递进
3. **实用的设计模式**: 安全检查、上下文隔离、按需加载、压缩策略
4. **可扩展的协作机制**: JSONL 邮箱、request_id 关联、worktree 隔离

这些设计模式可以显著提升 Axis 的执行能力层，特别是在 Agent 自主性、多 Agent 协作、以及长时间运行任务管理方面。

---

**参考链接**: https://github.com/shareAI-lab/learn-claude-code
