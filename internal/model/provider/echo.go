package provider

import "context"

// EchoModelProvider returns an exact echo of the input without mock markers.
type EchoModelProvider struct{}

func NewEchoModelProvider() *EchoModelProvider {
	return &EchoModelProvider{}
}

func (e *EchoModelProvider) Execute(ctx context.Context, req *ModelRequest) (*ModelResponse, error) {
	output := make(map[string]any)
	for k, v := range req.Input {
		output[k] = v
	}
	output["contract_id"] = req.ContractID
	output["status"] = "ok"
	output["provider"] = "echo"
	return &ModelResponse{Output: output, InputTokens: 100, OutputTokens: 50}, nil
}
