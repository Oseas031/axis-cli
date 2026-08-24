package contextpack

import (
	"path/filepath"
	"strings"

	"github.com/axis-cli/axis/internal/project"
)

type rule struct {
	id       string
	packet   ContextPacket
	keywords []string
}

// defaultRuleEnginePath 返回默认的规则配置文件路径
func defaultRuleEnginePath() string {
	root := project.MustResolveRoot()
	return filepath.Join(root, ".axis", "context", "rules.json")
}

// candidateRules 使用配置化规则引擎返回候选规则
func candidateRules() []rule {
	engine := NewRuleEngine(defaultRuleEnginePath())
	configRules := engine.GetAllRules()

	var rules []rule
	for _, configRule := range configRules {
		// 根据规则ID构建ContextPacket
		packet := buildPacketFromRule(configRule)
		rules = append(rules, rule{
			id:       configRule.ID,
			packet:   packet,
			keywords: configRule.Keywords,
		})
	}

	return rules
}

// buildPacketFromRule 根据配置规则构建ContextPacket
func buildPacketFromRule(r Rule) ContextPacket {
	// 根据规则ID确定packet类型和源路径
	packetType := PacketTypeSpec
	source := "docs/specs/" + r.Context + "/"

	// 特殊处理某些规则
	switch r.ID {
	case "natural-language-scheduling":
		source = "docs/specs/natural-language-scheduling/"
	case "model-provider":
		source = "docs/specs/model-provider/"
	case "interactive-shell":
		source = "docs/specs/interactive-shell/"
	case "adaptive-context-assembly":
		source = "docs/specs/adaptive-context-assembly/"
	case "dag-scheduling":
		source = "docs/architecture/dag-scheduling.md"
	case "axis-up":
		packetType = PacketTypeTool
		source = "tools/axis-up/"
	case "vigil-tracking":
		packetType = PacketTypeTool
		source = "internal/vigil/"
	case "memory-system":
		source = "docs/specs/memory/"
	case "skills-system":
		packetType = PacketTypeTool
		source = "internal/skills/"
	case "evolution-protocol":
		source = "docs/specs/evolution/"
	}

	// 使用更详细的summary和reason，与原始硬编码规则保持一致
	summary := r.Description
	reason := "Task matches rule: " + r.Description

	// 根据规则ID提供更详细的summary和reason
	switch r.ID {
	case "natural-language-scheduling":
		summary = "Natural language task scheduling requirements, design, and tasks."
		reason = "Task mentions natural language, ask, prompt, or intent-to-task scheduling."
	case "model-provider":
		summary = "Model provider configuration and provider profile behavior."
		reason = "Task mentions provider, model, DeepSeek, MiniMax, OpenAI-compatible setup, or API profiles."
	case "interactive-shell":
		summary = "Interactive shell requirements and command behavior."
		reason = "Task mentions shell, interactive command flow, or shell command behavior."
	case "adaptive-context-assembly":
		summary = "Adaptive Context Assembly requirements, design, task plan, and safety boundaries."
		reason = "Task mentions context assembly, context bundle, context packet, or adaptive context."
	case "dag-scheduling":
		summary = "DAG scheduling architecture and dependency behavior."
		reason = "Task mentions scheduler, DAG, dependencies, or readiness."
	case "axis-up":
		summary = "External onboarding helper design and command behavior."
		reason = "Task mentions axis-up, onboarding, first run, start, check, demo, or fix helper flow."
	}

	return ContextPacket{
		ID:        string(packetType) + ":" + r.ID,
		Type:      packetType,
		Source:    source,
		Summary:   summary,
		Reason:    reason,
		Relevance: r.Relevance,
	}
}

func matchesAny(text string, keywords []string) bool {
	text = strings.ToLower(text)
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}
