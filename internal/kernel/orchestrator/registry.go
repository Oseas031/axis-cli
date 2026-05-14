package orchestrator

import (
	"os"

	"github.com/axis-cli/axis/internal/model/tool"
)

func defaultToolRegistry() *tool.Registry {
	root, err := os.Getwd()
	if err != nil {
		root = "."
	}
	return BuildToolRegistry(root, nil)
}
