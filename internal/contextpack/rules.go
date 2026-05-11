package contextpack

import "strings"

type rule struct {
	id       string
	packet   ContextPacket
	keywords []string
}

func candidateRules() []rule {
	return []rule{
		{
			id: "natural-language-scheduling",
			packet: ContextPacket{
				ID:        "spec:natural-language-scheduling",
				Type:      PacketTypeSpec,
				Source:    "docs/specs/natural-language-scheduling/",
				Summary:   "Natural language task scheduling requirements, design, and tasks.",
				Reason:    "Task mentions natural language, ask, prompt, or intent-to-task scheduling.",
				Relevance: 0.92,
			},
			keywords: []string{"ask", "natural language", "prompt", "intent", "scheduling"},
		},
		{
			id: "model-provider",
			packet: ContextPacket{
				ID:        "spec:model-provider",
				Type:      PacketTypeSpec,
				Source:    "docs/specs/model-provider/",
				Summary:   "Model provider configuration and provider profile behavior.",
				Reason:    "Task mentions provider, model, DeepSeek, MiniMax, OpenAI-compatible setup, or API profiles.",
				Relevance: 0.9,
			},
			keywords: []string{"provider", "model", "deepseek", "minimax", "openai", "api key", "profile"},
		},
		{
			id: "interactive-shell",
			packet: ContextPacket{
				ID:        "spec:interactive-shell",
				Type:      PacketTypeSpec,
				Source:    "docs/specs/interactive-shell/",
				Summary:   "Interactive shell requirements and command behavior.",
				Reason:    "Task mentions shell, interactive command flow, or shell command behavior.",
				Relevance: 0.88,
			},
			keywords: []string{"shell", "interactive", "repl"},
		},
		{
			id: "adaptive-context-assembly",
			packet: ContextPacket{
				ID:        "spec:adaptive-context-assembly",
				Type:      PacketTypeSpec,
				Source:    "docs/specs/adaptive-context-assembly/",
				Summary:   "Adaptive Context Assembly requirements, design, task plan, and safety boundaries.",
				Reason:    "Task mentions context assembly, context bundle, context packet, or adaptive context.",
				Relevance: 0.95,
			},
			keywords: []string{"context", "assembly", "bundle", "packet", "adaptive"},
		},
		{
			id: "dag-scheduling",
			packet: ContextPacket{
				ID:        "architecture:dag-scheduling",
				Type:      PacketTypeSpec,
				Source:    "docs/architecture/dag-scheduling.md",
				Summary:   "DAG scheduling architecture and dependency behavior.",
				Reason:    "Task mentions scheduler, DAG, dependencies, or readiness.",
				Relevance: 0.86,
			},
			keywords: []string{"scheduler", "dag", "dependency", "dependencies", "ready"},
		},
		{
			id: "axis-up",
			packet: ContextPacket{
				ID:        "tool:axis-up",
				Type:      PacketTypeTool,
				Source:    "tools/axis-up/",
				Summary:   "External onboarding helper design and command behavior.",
				Reason:    "Task mentions axis-up, onboarding, first run, start, check, demo, or fix helper flow.",
				Relevance: 0.84,
			},
			keywords: []string{"axis-up", "onboarding", "first run", "start", "check", "demo", "fix"},
		},
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
