# Token-Level Hallucination Detection：对 Axis Judge 系统的启示

> 基于 arXiv:2605.12384 — Min et al., 2026-05-12

## 1. 核心机制（≤10行）

TokenHD 是一个训练 token 级幻觉检测器的完整 pipeline：
1. **可扩展数据引擎**：自动合成大规模 token 级幻觉标注数据，无需人工标注或预定义步骤分割
2. **重要性加权训练**：对关键 token 赋予更高权重，使小模型聚焦于真正影响正确性的位置
3. **直接操作自由文本**：检测器直接在原始输出上逐 token 判断，不依赖 step segmentation 或文本重格式化
4. **跨域泛化**：通过特定策略增强检测器在未见领域上的迁移能力

核心洞察：幻觉检测不需要理解全部语义，只需识别 token 级的异常模式——这使小模型成为可能。

## 2. 关键数据

| 指标 | 数值 |
|------|------|
| 最小有效检测器 | **0.6B 参数** |
| 超越的大模型基线 | QwQ-32B（推理专用模型） |
| 模型规模范围 | 0.6B → 8B，性能随规模一致提升 |
| 关键优势 | 无需 step segmentation，直接处理 free-form text |
| 泛化能力 | 跨多种实际场景验证有效 |

## 3. 对 Axis 的启示

- **现状痛点**：Axis SelfJudgementEngine 使用 5 种验证策略（Syntax/Semantic/Contract/Test/Coverage），Semantic 判断依赖大模型全量推理，成本高、延迟大
- **TokenHD 证明**：0.6B 小模型经过针对性训练后，幻觉检测能力超越 QwQ-32B，说明判断任务不需要通用智能
- **映射到 Quality-Gated Escalation**：轻量检测器（小模型/规则）做初筛 pass/fail，仅 fail 或 uncertain 的结果升级到大模型深度审查
- **可借鉴**：
  - 分层 judge 架构（cheap first → expensive on demand）
  - importance-weighted 思想：对关键输出位置加权验证，非均匀审查
  - 跨域泛化设计：一个 judge 配置覆盖多种任务类型
- **不能借鉴**：Axis 不训练模型，无法复现 TokenHD 的数据引擎和训练流程；但可以调用已发布的 TokenHD 检测器作为外部工具

## 4. 可行动建议

1. **引入 LightweightJudge 策略**：在 SelfJudgementEngine 中新增一个 `Lightweight` 验证层，使用小模型（如未来开源的 TokenHD 0.6B）做快速初筛
2. **实现 Escalation Gate**：初筛通过 → 直接 pass；初筛标记异常 → 触发现有 Semantic 大模型验证，预计减少 70%+ 的大模型调用
3. **importance-weighted 审查**：对 Agent 输出中的关键段落（代码变更、配置修改）加权审查，非关键段落（日志、注释）降低审查力度
4. **Tool 集成路径**：将 TokenHD 检测器封装为 `HallucinationDetectorTool`，纳入 tool permission scope 管理
