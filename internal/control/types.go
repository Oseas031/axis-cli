package control

import (
	"github.com/axis-cli/axis/internal/comm"
	"github.com/axis-cli/axis/internal/types"
)

type SubmitTaskRequest struct {
	Task *types.AgentTask `json:"task"`
}

type SubmitTaskResponse struct {
	TaskID string           `json:"task_id"`
	Status types.TaskStatus `json:"status"`
}

type StatusResponse struct {
	TaskID string           `json:"task_id"`
	Status types.TaskStatus `json:"status"`
	Output map[string]any   `json:"output,omitempty"`
	Error  string           `json:"error,omitempty"`
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Hint    string `json:"hint,omitempty"`
}

type HealthResponse struct {
	Status      string `json:"status"`
	Protocol    string `json:"protocol"`
	Address     string `json:"address"`
	ProjectRoot string `json:"project_root"`
}

// InboxListResponse is the GET /v1/inbox payload (design.md D7).
type InboxListResponse struct {
	Records        []comm.InboxRecord `json:"records"`
	CorruptedLines int                `json:"corrupted_lines,omitempty"`
}

// InboxRespondRequest is the POST /v1/inbox/{call-id}/respond body.
type InboxRespondRequest struct {
	Choice string `json:"choice,omitempty"`
	Text   string `json:"text,omitempty"`
}

// InboxRespondResponse is the respond reply.
type InboxRespondResponse struct {
	CallID string `json:"call_id"`
	Status string `json:"status"`
}
