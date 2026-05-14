package orchestrator

import (
	"os"
	"path/filepath"

	"github.com/axis-cli/axis/internal/memory/horizon"
	"github.com/axis-cli/axis/internal/model/tool"
	"github.com/axis-cli/axis/internal/project"
	"github.com/axis-cli/axis/internal/skills"
)

func defaultToolRegistry() *tool.Registry {
	registry := tool.NewRegistry()
	allowedDirs := []string{}
	allowedDir, err := os.Getwd()
	if err != nil {
		allowedDir = "."
	}
	allowedDirs = append(allowedDirs, allowedDir)
	if homeDir, err := os.UserHomeDir(); err == nil {
		desktopDir := filepath.Join(homeDir, "Desktop")
		if _, err := os.Stat(desktopDir); err == nil {
			allowedDirs = append(allowedDirs, desktopDir)
		}
	}
	_ = registry.Register(tool.NewBashTool(), []string{string(tool.ScopeSubprocess)})
	_ = registry.Register(tool.NewVerifyBashTool(), []string{string(tool.ScopeSubprocess)})
	_ = registry.Register(tool.NewFileReadTool(allowedDirs), []string{string(tool.ScopeFilesystemRead)})
	_ = registry.Register(tool.NewFileWriteTool(allowedDirs), []string{string(tool.ScopeFilesystemWrite)})
	_ = registry.Register(tool.NewHTTPClientTool([]string{"localhost", "127.0.0.1"}), []string{string(tool.ScopeNetwork)})

	// Skills: load_skill tool
	skillsDir := project.SkillsDir(allowedDir)
	skillsLoader := skills.NewLoader(skillsDir)
	_ = registry.Register(tool.NewLoadSkillTool(skillsLoader), []string{string(tool.ScopeFilesystemRead)})

	// Memory: recall_memory + store_memory tools
	memoryDir := project.MemoryDir(allowedDir)
	memoryStore := horizon.NewStore(memoryDir)
	_ = memoryStore.Init()
	_ = registry.Register(tool.NewRecallMemoryTool(memoryStore), []string{string(tool.ScopeFilesystemRead)})
	_ = registry.Register(tool.NewStoreMemoryTool(memoryStore), []string{string(tool.ScopeFilesystemWrite)})
	_ = registry.Register(tool.NewCompactTool(), nil)
	_ = registry.Register(tool.NewYieldTool(), nil)
	_ = registry.Register(tool.NewCheckpointTool(), nil)
	_ = registry.Register(tool.NewSpawnTool(), nil)

	return registry
}
