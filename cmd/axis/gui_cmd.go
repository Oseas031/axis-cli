package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

func newGUICommand() *cobra.Command {
	var port int
	cmd := &cobra.Command{
		Use:   "gui",
		Short: "Launch the Axis GUI (observation dashboard)",
		RunE: func(cmd *cobra.Command, args []string) error {
			guiExe := findGUIExe()
			if guiExe == "" {
				return fmt.Errorf("axis-gui executable not found; build it with: cd tools/axis-gui && go build -o axis-gui.exe")
			}
			cwd, _ := os.Getwd()
			guiCmd := exec.Command(guiExe, "--port", fmt.Sprintf("%d", port), "--root", cwd)
			guiCmd.Stdout = os.Stdout
			guiCmd.Stderr = os.Stderr
			fmt.Printf("Starting axis-gui on port %d...\n", port)
			return guiCmd.Run()
		},
	}
	cmd.Flags().IntVar(&port, "port", 3000, "GUI server port")
	return cmd
}

func findGUIExe() string {
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	// Check relative to the axis binary
	exePath, _ := os.Executable()
	if exePath != "" {
		dir := filepath.Dir(exePath)
		candidate := filepath.Join(dir, "..", "tools", "axis-gui", "axis-gui"+ext)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	// Check relative to cwd
	cwd, _ := os.Getwd()
	candidate := filepath.Join(cwd, "tools", "axis-gui", "axis-gui"+ext)
	if _, err := os.Stat(candidate); err == nil {
		return candidate
	}
	// Check PATH
	if path, err := exec.LookPath("axis-gui" + ext); err == nil {
		return path
	}
	return ""
}
