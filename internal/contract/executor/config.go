package executor

import (
	"context"

	"github.com/axis-cli/axis/internal/model/provider"
	"github.com/axis-cli/axis/internal/model/tool"
	"github.com/axis-cli/axis/internal/types"
)

// ExecutorConfig holds all dependencies for ContractExecutorImpl.
type ExecutorConfig struct {
	Provider           provider.ModelProvider
	ToolRegistry       *tool.Registry
	SkillsLoader       interface{ BuildSkillsPromptSection(context.Context) string }
	PrinciplesLoader   interface{ BuildPrinciplesPromptSection() string }
	CompactionPipeline Compactor
	AllowedScopes      []string

	// CircuitBreakerThreshold is the number of consecutive tool errors before aborting. Default: 5.
	CircuitBreakerThreshold int
	// MaxTurns is the maximum number of multi-turn iterations. Default: 10.
	MaxTurns int
}

// NewContractExecutorWithConfig creates a new contract executor with all dependencies provided upfront.
func NewContractExecutorWithConfig(cfg ExecutorConfig) *ContractExecutorImpl {
	cbThreshold := cfg.CircuitBreakerThreshold
	if cbThreshold == 0 {
		cbThreshold = 5
	}
	maxTurns := cfg.MaxTurns
	if maxTurns == 0 {
		maxTurns = 10
	}
	return &ContractExecutorImpl{
		contracts:              make(map[string]*types.AgentContract),
		provider:               cfg.Provider,
		toolRegistry:           cfg.ToolRegistry,
		skillsLoader:           cfg.SkillsLoader,
		principlesLoader:       cfg.PrinciplesLoader,
		compactionPipeline:     cfg.CompactionPipeline,
		circuitBreakerThreshold: cbThreshold,
		maxTurns:               maxTurns,
	}
}
