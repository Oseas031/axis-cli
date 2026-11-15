package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/spf13/cobra"
)

func newGUICommand() *cobra.Command {
	var port int
	var noOpen bool
	cmd := &cobra.Command{
		Use:     "gui",
		Aliases: []string{"web"},
		Short:   "Start Axis runtime and built-in Observatory UI together",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
			defer signal.Stop(sigChan)

			go func() {
				<-sigChan
				fmt.Fprintln(cmd.ErrOrStderr(), "\nShutting down gracefully...")
				cancel()
			}()

			onReady := func(addr string) {
				uiURL := addr + "/ui/"
				if !noOpen {
					if err := openBrowser(uiURL); err != nil {
						fmt.Fprintf(cmd.ErrOrStderr(), "Warning: could not open browser automatically: %v\n", err)
					}
				}
			}

			return runLocalRuntimeWithOpts(ctx, defaultApp.resolvedRoot(), cmd.OutOrStdout(), port, true, onReady)
		},
	}
	cmd.Flags().IntVar(&port, "port", 0, "Fixed port for runtime and UI (0 = random loopback port)")
	cmd.Flags().BoolVar(&noOpen, "no-open", false, "Do not automatically open the browser")
	return cmd
}

func openBrowser(targetURL string) error {
	switch runtime.GOOS {
	case "windows":
		// #nosec G204
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", targetURL).Start()
	case "darwin":
		// #nosec G204
		return exec.Command("open", targetURL).Start()
	default:
		// #nosec G204
		return exec.Command("xdg-open", targetURL).Start()
	}
}
