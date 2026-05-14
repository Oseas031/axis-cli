package orchestrator

import (
	"os"
	"path/filepath"

	contractexec "github.com/axis-cli/axis/internal/contract/executor"
	"github.com/axis-cli/axis/internal/memory/horizon"
	"github.com/axis-cli/axis/internal/model/provider"
	"github.com/axis-cli/axis/internal/model/tool"
	"github.com/axis-cli/axis/internal/project"
	"github.com/axis-cli/axis/internal/skills"
)

// ContractExecutorDeps holds the dependencies for building a ContractExecutor.
type ContractExecutorDeps struct {
	Provider     provider.ModelProvider
	ToolRegistry *tool.Registry
	Root         string
}

// BuildToolRegistry creates a tool registry rooted at the given directory.
// If filter is non-empty, only tools whose names appear in filter are registered.
func BuildToolRegistry(root string, filter []string) *tool.Registry {
	registry := tool.NewRegistry()
	allowedDirs := []string{root}
	if homeDir, err := os.UserHomeDir(); err == nil {
		desktopDir := filepath.Join(homeDir, "Desktop")
		if _, err := os.Stat(desktopDir); err == nil {
			allowedDirs = append(allowedDirs, desktopDir)
		}
	}

	// Build filter set for O(1) lookup
	allowed := make(map[string]bool, len(filter))
	for _, name := range filter {
		allowed[name] = true
	}
	shouldRegister := func(name string) bool {
		return len(filter) == 0 || allowed[name]
	}

	type entry struct {
		tool   tool.Tool
		scopes []string
	}
	all := []entry{
		{tool.NewBashTool(), []string{string(tool.ScopeSubprocess)}},
		{tool.NewVerifyBashTool(), []string{string(tool.ScopeSubprocess)}},
		{tool.NewFileReadTool(allowedDirs), []string{string(tool.ScopeFilesystemRead)}},
		{tool.NewFileWriteTool(allowedDirs), []string{string(tool.ScopeFilesystemWrite)}},
		{tool.NewHTTPClientTool([]string{"localhost", "127.0.0.1"}), []string{string(tool.ScopeNetwork)}},
	}

	skillsLoader := skills.NewLoader(project.SkillsDir(root))
	all = append(all, entry{tool.NewLoadSkillTool(skillsLoader), []string{string(tool.ScopeFilesystemRead)}})

	memoryStore := horizon.NewStore(project.MemoryDir(root))
	_ = memoryStore.Init()
	all = append(all,
		entry{tool.NewRecallMemoryTool(memoryStore), []string{string(tool.ScopeFilesystemRead)}},
		entry{tool.NewStoreMemoryTool(memoryStore), []string{string(tool.ScopeFilesystemWrite)}},
		entry{tool.NewCompactTool(), nil},
		entry{tool.NewYieldTool(), nil},
		entry{tool.NewCheckpointTool(), nil},
		entry{tool.NewSpawnTool(), nil},
	)

	for _, e := range all {
		if shouldRegister(e.tool.Name()) {
			_ = registry.Register(e.tool, e.scopes)
		}
	}

	return registry
}

// BuildContractExecutor creates a ContractExecutorImpl with the given dependencies.
// Note: creates its own skillsLoader and principlesStore instances.
// This is safe because both are stateless readers (no file locks).
func BuildContractExecutor(cfg ContractExecutorDeps) *contractexec.ContractExecutorImpl {
	skillsPromptLoader := skills.NewLoader(project.SkillsDir(cfg.Root))
	principlesStore := horizon.NewStore(project.MemoryDir(cfg.Root))

	return contractexec.NewContractExecutorWithConfig(contractexec.ExecutorConfig{
		Provider:     cfg.Provider,
		ToolRegistry: cfg.ToolRegistry,
		SkillsLoader: skillsPromptLoader,
		PrinciplesLoader: principlesStore,
		CompactionPipeline: &contractexec.ThreeLayerCompaction{
			Micro:  &contractexec.ToolResultCompaction{KeepRecent: 3},
			Auto:   &contractexec.SummarizationCompaction{Provider: cfg.Provider, KeepRecent: 4},
			Budget: 32000,
		},
	})
}
