package gate

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/axis-cli/axis/internal/comm"
	"github.com/axis-cli/axis/internal/types"
)

// requestFromRecord decodes the HumanCallRequest carried in a record payload.
func requestFromRecord(r *comm.InboxRecord) (*types.HumanCallRequest, error) {
	if r == nil {
		return nil, fmt.Errorf("gate: nil record")
	}
	raw, ok := r.Message.Payload["request"]
	if !ok {
		return nil, fmt.Errorf("gate: record %s carries no request payload", r.Message.ID)
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("gate: re-encode payload: %w", err)
	}
	var req types.HumanCallRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("gate: decode payload: %w", err)
	}
	return &req, nil
}

// parkUntil computes the next wake-up for the parked caller:
// the earlier of the degrade deadline and the next nudge tick.
func parkUntil(req *types.HumanCallRequest, nudgesDone int, cfg Config, now time.Time) time.Time {
	nextNudge := now.Add(cfg.NudgeInterval)
	if req.Deadline.Before(nextNudge) {
		return req.Deadline
	}
	return nextNudge
}

// shouldDegrade reports whether the SLA budget is spent (deadline reached or
// nudge allowance consumed). Fail-closed: one of these always fires before an
// indefinite wait.
func shouldDegrade(req *types.HumanCallRequest, nudgesDone int, cfg Config, now time.Time) bool {
	if !req.Deadline.After(now) {
		return true
	}
	return nudgesDone >= cfg.MaxNudges
}
