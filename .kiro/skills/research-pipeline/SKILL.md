---
name: research-pipeline
description: End-to-end research pipeline — from MindMagnifier discovery to vigil action items. Use when user says "找论文"/"研究"/"查文献".
tags: [workflow, research, automation]
---

# 研究管线：放大镜 → 研究 → 待办

## 完整链路

```
1. 发现  →  2. 筛选  →  3. 深入  →  4. 扬弃  →  5. 落地
amp sync     AI 判断     读论文     提取精华     vigil add
amp list     相关度      写报告     丢弃外壳     commit
```

## Step 1: 发现（MindMagnifier）

> 详细命令参考：`mind-magnifier` skill

```bash
# 同步 + 按分类查看
C:\Users\ASUS\Desktop\MindMagnifier\amp.exe sync
C:\Users\ASUS\Desktop\MindMagnifier\amp.exe list --category llm --since 48h --json --limit 30
```

## Step 2: 筛选（AI 判断相关度）

对每篇论文问：**它与 Axis 当前待办的哪个方向相关？**

相关度判断标准：
- 🔴 高：直接映射 Axis 已有模块或 vigil 中的 P0/P1 项
- 🟠 中：验证 Axis 设计方向或提供可借鉴的工程模式
- ⚪ 低：有趣但不可行动

只深入 🔴 和 🟠。⚪ 跳过。

## Step 3: 深入（写研究报告）

输出到 `docs/research/<topic>-<date>.md`，固定结构：

```markdown
# <论文标题>：对 Axis 的启示

> 基于 arXiv:XXXX.XXXXX — "<title>"

## 1. 核心机制（≤10 行）
## 2. 关键数据（数字说话）
## 3. 对 Axis 的启示
   - 当前状态（Axis 有什么）
   - 论文做了什么不同的
   - 可借鉴 / 不能借鉴
## 4. 可行动建议（带优先级和模块）
```

## Step 4: 扬弃（提取精华）

从研究报告的 §4 中提取：
- **P0/P1 行动**：直接创建 vigil 待办
- **P2 方向验证**：记录但不立即行动
- **丢弃**：论文的形式化/特定领域细节

## Step 5: 落地（vigil add）

每个可行动建议转为 vigil 待办：

```bash
axis vigil add "<行动标题>" \
  --priority <P0/P1/P2> \
  --tag research \
  --tag <相关模块> \
  --notes "arXiv:XXXX.XXXXX: <一句话说明价值>"
```

## 完整示例

```bash
# 1. 发现
C:\Users\ASUS\Desktop\MindMagnifier\amp.exe list --category llm --since 48h

# 2. 筛选 → 选出 "Context Rot" 论文（🔴 高相关：影响 Judge 安全）

# 3. 深入 → 写 docs/research/context-rot-safety-warning.md

# 4. 扬弃 → 提取：Judge 必须用独立短 context

# 5. 落地
axis vigil add "Judge must use independent short context" \
  --priority P0 --tag security --tag judge \
  --notes "arXiv:2605.12366: 800K token后漏检率2x-30x"

# 6. 提交
git add docs/research/... .axis/vigil/items.json
git commit -m "docs(research): ... vigil:vigil-xxx"
```

## 触发条件

- 用户说"找论文"/"查文献"/"研究"/"用放大镜" → 执行此管线
- 每次新会话 `vigil resume` 后如果无紧急 P0 → 可主动提议"要查最新论文吗？"

## 不做的事

- 不自动同步（用户触发）
- 不自动深入所有论文（AI 筛选后只深入高相关的）
- 不替用户判断优先级（建议优先级，用户确认）
