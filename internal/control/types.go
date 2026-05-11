package control

import "github.com/axis-cli/axis/internal/types"

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
