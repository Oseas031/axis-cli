package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/axis-cli/axis/internal/contextpack"
	"github.com/axis-cli/axis/internal/control"
	"github.com/axis-cli/axis/internal/guarantee"
)

func newLocalHTTPServer(handler http.Handler) *http.Server {
	return &http.Server{
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
}

func runLocalRuntime(ctx context.Context, root string, out io.Writer, port int) error {
	return runLocalRuntimeWithUI(ctx, root, out, port, false)
}

func runLocalRuntimeWithUI(ctx context.Context, root string, out io.Writer, port int, withUI bool) error {
	return runLocalRuntimeWithOpts(ctx, root, out, port, withUI, nil)
}

func runLocalRuntimeWithOpts(ctx context.Context, root string, out io.Writer, port int, withUI bool, onReady func(addr string)) error {
	// Single-writer discipline: refuse when a healthy runtime already owns
	// this project root (design.md D2). Prevents mailbox double-write races.
	if err := ensureNoLiveRuntime(root); err != nil {
		return err
	}
	if err := contextpack.InitDefaultRegistry(root); err != nil {
		return fmt.Errorf("failed to init readiness registry: %w", err)
	}

	eventLog := control.NewTaskEventLog(root)

	// Human scheduling components must exist before initOrchestrator wires
	// them into the dispatcher (design.md D2).
	setupHumanScheduling(root, eventLog)
	defer closeAndClearHumanSched()

	initOrchestrator()
	runtimeOrch := orch

	// Verify guarantees before starting
	if defaultApp.guarantees != nil {
		violations := defaultApp.guarantees.Verify()
		for _, v := range violations {
			if v.Level == guarantee.LevelHard {
				return fmt.Errorf("hard guarantee violation [%s]: %v", v.GuaranteeID, v.Error)
			}
			if v.Level == guarantee.LevelSoft && out != nil {
				fmt.Fprintf(out, "Warning: soft guarantee violation [%s]: %v\n", v.GuaranteeID, v.Error)
			}
		}
	}

	if err := runtimeOrch.Start(ctx); err != nil {
		return fmt.Errorf("failed to start orchestrator: %w", err)
	}
	defer func() {
		closeAndClearHumanSched() // wake parked waiters so Shutdown can drain
		_ = runtimeOrch.Shutdown(context.Background())
	}()

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return fmt.Errorf("failed to start local control server: %w", err)
	}
	defer listener.Close()

	record := control.RuntimeRecord{
		PID:         os.Getpid(),
		Protocol:    "http",
		Address:     "http://" + listener.Addr().String(),
		StartedAt:   time.Now().UTC(),
		ProjectRoot: root,
	}
	locator := control.NewRuntimeLocator(root)
	if err := locator.Save(record); err != nil {
		return fmt.Errorf("failed to write runtime locator: %w", err)
	}
	defer func() { _ = locator.Delete() }()

	if out != nil {
		fmt.Fprintf(out, "Local Axis runtime started at %s\n", record.Address)
	}

	// Mark orphaned tasks from previous runtime as abandoned
	if orphaned, err := control.MarkOrphanedTasks(eventLog); err != nil {
		if out != nil {
			fmt.Fprintf(out, "Warning: failed to mark orphaned tasks: %v\n", err)
		}
	} else if orphaned > 0 && out != nil {
		fmt.Fprintf(out, "Marked %d orphaned task(s) as abandoned\n", orphaned)
	}

	// Startup scan: requests that expired while the runtime was down degrade
	// fail-closed now (design.md D3).
	degradeExpiredRequests(out)

	server := control.NewServerWithEventLog(runtimeOrch, record, eventLog)
	if g := getHumanGate(); g != nil {
		server.SetInboxAPI(inboxAPIBridge{g: g})
	}
	if withUI {
		server.SetUIEnabled(true)
		server.StartRoutines(ctx)
		if dump, err := buildConfigDump(defaultApp); err == nil && dump != nil {
			server.SetUIInfo(control.UIInfo{
				Provider:        dump.Provider,
				ProfileName:     dump.ProfileName,
				Model:           dump.Model,
				RuntimeProfile:  dump.RuntimeProfile,
				FallbackProfile: dump.FallbackProfile,
				HasAPIKey:       dump.HasAPIKey,
				ProjectRoot:     dump.ProjectRoot,
			})
		}
		if out != nil {
			fmt.Fprintf(out, "Observatory UI available at %s/ui/\n", record.Address)
		}
	}

	httpServer := newLocalHTTPServer(server.Handler())
	serveErr := make(chan error, 1)
	go func() {
		serveErr <- httpServer.Serve(listener)
	}()
	if onReady != nil {
		onReady(record.Address)
	}

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdownCtx)
		return nil
	case err := <-serveErr:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}
