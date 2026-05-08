package provider

import "context"

// MockModelProvider returns an echo of the input for testing.
type MockModelProvider struct{}

func NewMockModelProvider() *MockModelProvider {
	return &MockModelProvider{}
}

func (m *MockModelProvider) Execute(ctx context.Context, req *ModelRequest) (*ModelResponse, error) {
	output := make(map[string]any)
	for k, v := range req.Input {
		output[k] = v
	}
	output["contract_id"] = req.ContractID
	output["status"] = "completed"
	output["provider"] = "mock"
	return &ModelResponse{Output: output}, nil
}
