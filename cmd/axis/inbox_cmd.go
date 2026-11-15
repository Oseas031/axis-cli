package main

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/axis-cli/axis/internal/comm"
	"github.com/axis-cli/axis/internal/control"
	humangate "github.com/axis-cli/axis/internal/human/gate"
	"github.com/spf13/cobra"
)

// inboxAccess routes inbox operations to the live runtime's control plane
// when one exists (single-writer discipline, design.md D2), otherwise to the
// local .axis/comm files (safe because no runtime process owns them).
type inboxAccess struct {
	client *control.Client
	local  *comm.Inbox
}

func newInboxCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inbox",
		Short: "Human inbox for agent requests (list/show/respond)",
	}
	cmd.AddCommand(newInboxListCmd(), newInboxShowCmd(), newInboxRespondCmd())
	return cmd
}

func resolveInboxAccess(root string) (*inboxAccess, error) {
	locator := control.NewRuntimeLocator(root)
	if record, err := locator.Load(); err == nil {
		client := control.NewClient(locator, control.LocalHTTPClient())
		ctx, cancel := context.WithTimeout(context.Background(), runtimeProbeTimeout)
		if _, herr := client.Health(ctx); herr == nil {
			cancel()
			return &inboxAccess{client: client}, nil
		}
		cancel()
		_ = record // stale locator: fall through to local access
	}
	return &inboxAccess{local: comm.NewInbox(filepath.Join(root, ".axis", "comm"))}, nil
}

// list returns all records regardless of access path.
func (a *inboxAccess) list() ([]comm.InboxRecord, int, error) {
	if a.client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), fallbackInboxTimeout)
		defer cancel()
		resp, err := a.client.InboxList(ctx)
		if err != nil {
			return nil, 0, err
		}
		return resp.Records, resp.CorruptedLines, nil
	}
	return a.local.List(humangate.ActorID)
}

// respond answers callID (or the only pending request when empty).
func (a *inboxAccess) respond(callID, choice, text string) (*comm.InboxRecord, error) {
	if a.client != nil {
		if callID == "" {
			return nil, fmt.Errorf("a runtime is running: --call-id is required (see `axis inbox list`)")
		}
		ctx, cancel := context.WithTimeout(context.Background(), fallbackInboxTimeout)
		defer cancel()
		if _, err := a.client.InboxRespond(ctx, callID, choice, text); err != nil {
			return nil, err
		}
		return &comm.InboxRecord{Message: comm.Message{ID: callID}}, nil
	}
	if callID == "" {
		id, _, err := singlePendingID(a)
		if err != nil {
			return nil, err
		}
		callID = id
	}
	err := a.local.MutatePending(humangate.ActorID, callID, func(r *comm.InboxRecord) {
		r.Status = comm.InboxStatusCompleted
		r.Output = map[string]any{"choice": choice, "text": text}
	})
	if err != nil {
		return nil, err
	}
	return a.local.Get(humangate.ActorID, callID)
}

func singlePendingID(a *inboxAccess) (string, []comm.InboxRecord, error) {
	recs, _, err := a.list()
	if err != nil {
		return "", nil, err
	}
	var pending []comm.InboxRecord
	for _, r := range recs {
		if r.Pending() {
			pending = append(pending, r)
		}
	}
	switch len(pending) {
	case 1:
		return pending[0].Message.ID, pending, nil
	case 0:
		return "", pending, fmt.Errorf("no pending requests in the human inbox")
	default:
		ids := ""
		for i, r := range pending {
			if i > 0 {
				ids += ", "
			}
			ids += r.Message.ID
		}
		return "", pending, fmt.Errorf("%d pending requests; specify --call-id (one of: %s)", len(pending), ids)
	}
}

// fallbackInboxTimeout bounds CLI HTTP calls when the caller passed no ctx.
const fallbackInboxTimeout = 10 * time.Second

func newInboxListCmd() *cobra.Command {
	var jsonOut bool
	var all bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List requests waiting for a human response",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := defaultApp.resolvedRoot()
			access, err := resolveInboxAccess(root)
			if err != nil {
				return err
			}
			records, corrupted, err := access.list()
			if err != nil {
				return err
			}
			var shown []comm.InboxRecord
			for _, r := range records {
				if all || r.Pending() {
					shown = append(shown, r)
				}
			}
			if jsonOut {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{
					"records":         shown,
					"corrupted_lines": corrupted,
				})
			}
			pendingCount := 0
			for _, r := range records {
				if r.Pending() {
					pendingCount++
				}
			}
			fmt.Fprintf(cmd.OutOrStdout(), "inbox: %d pending / %d total\n", pendingCount, len(records))
			for _, r := range shown {
				fmt.Fprintf(cmd.OutOrStdout(), "%s [%s]", r.Message.ID, r.Status)
				if req, err := decodeRequest(r); err == nil {
					fmt.Fprintf(cmd.OutOrStdout(), " %s %s", req.Kind, req.Prompt)
				}
				if r.Nudges > 0 {
					fmt.Fprintf(cmd.OutOrStdout(), " (nudged %d)", r.Nudges)
				}
				fmt.Fprintln(cmd.OutOrStdout())
			}
			if pendingCount > 0 && !all {
				fmt.Fprintf(cmd.OutOrStdout(), "Next: axis inbox show <call-id>, then axis inbox respond <call-id> --choice <option>\n")
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "machine-readable JSON output")
	cmd.Flags().BoolVar(&all, "all", false, "include resolved and degraded requests")
	return cmd
}

func newInboxShowCmd() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "show <call-id>",
		Short: "Show one inbox request with full context",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := defaultApp.resolvedRoot()
			access, err := resolveInboxAccess(root)
			if err != nil {
				return err
			}
			records, _, err := access.list()
			if err != nil {
				return err
			}
			for _, r := range records {
				if r.Message.ID == args[0] {
					if jsonOut {
						return json.NewEncoder(cmd.OutOrStdout()).Encode(r)
					}
					req, derr := decodeRequest(r)
					if derr != nil {
						return derr
					}
					out := cmd.OutOrStdout()
					fmt.Fprintf(out, "call:   %s\nstatus: %s\nkind:   %s\ntask:   %s\n", r.Message.ID, r.Status, req.Kind, req.TaskID)
					fmt.Fprintf(out, "prompt: %s\n", req.Prompt)
					if len(req.Options) > 0 {
						fmt.Fprintf(out, "options: %v\n", req.Options)
					}
					if !req.Deadline.IsZero() {
						fmt.Fprintf(out, "deadline: %s\n", req.Deadline.Format(time.RFC3339))
					}
					if len(req.Context) > 0 {
						fmt.Fprintf(out, "context: %v\n", req.Context)
					}
					fmt.Fprintf(out, "Next: axis inbox respond %s --choice <option>\n", r.Message.ID)
					return nil
				}
			}
			return fmt.Errorf("call %s not found; run `axis inbox list --all`", args[0])
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "machine-readable JSON output")
	return cmd
}

func newInboxRespondCmd() *cobra.Command {
	var callID, choice, text string
	cmd := &cobra.Command{
		Use:   "respond [call-id]",
		Short: "Answer a pending request (resume or degrade the parked task)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				callID = args[0]
			}
			if choice == "" && text == "" {
				return fmt.Errorf("nothing to answer: pass --choice <option> or --text \"<answer>\"")
			}
			root := defaultApp.resolvedRoot()
			access, err := resolveInboxAccess(root)
			if err != nil {
				return err
			}
			rec, err := access.respond(callID, choice, text)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Responded to %s\n", rec.Message.ID)
			fmt.Fprintf(out, "Next: axis status <task-id> to see the resumed task\n")
			return nil
		},
	}
	cmd.Flags().StringVar(&callID, "call-id", "", "request ID (optional when exactly one request is pending)")
	cmd.Flags().StringVar(&choice, "choice", "", "selected option for decision requests")
	cmd.Flags().StringVar(&text, "text", "", "free-text answer for info requests")
	return cmd
}

func decodeRequest(r comm.InboxRecord) (*humangateRequestView, error) {
	raw, ok := r.Message.Payload["request"]
	if !ok {
		return nil, fmt.Errorf("record %s carries no request payload", r.Message.ID)
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}
	var view humangateRequestView
	if err := json.Unmarshal(data, &view); err != nil {
		return nil, err
	}
	return &view, nil
}

// humangateRequestView mirrors types.HumanCallRequest for CLI rendering.
type humangateRequestView struct {
	CallID   string            `json:"call_id"`
	TaskID   string            `json:"task_id"`
	Kind     string            `json:"kind,omitempty"`
	Prompt   string            `json:"prompt,omitempty"`
	Options  []string          `json:"options,omitempty"`
	Context  map[string]any    `json:"context,omitempty"`
	Input    map[string]any    `json:"input"`
	Metadata map[string]string `json:"metadata,omitempty"`
	Deadline time.Time         `json:"deadline,omitempty"`
}
