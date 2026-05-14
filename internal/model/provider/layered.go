package provider

import "context"

// LayeredProvider routes requests to different backends based on request type.
type LayeredProvider struct {
	primary ModelProvider
	utility ModelProvider
}

// NewLayeredProvider creates a provider that routes utility requests to a cheaper backend.
func NewLayeredProvider(primary, utility ModelProvider) *LayeredProvider {
	return &LayeredProvider{primary: primary, utility: utility}
}

func (p *LayeredProvider) Execute(ctx context.Context, req *ModelRequest) (*ModelResponse, error) {
	if isUtilityRequest(req) {
		return p.utility.Execute(ctx, req)
	}
	return p.primary.Execute(ctx, req)
}

// isUtilityRequest checks if the request is a utility call based on metadata.
func isUtilityRequest(req *ModelRequest) bool {
	if req.Metadata == nil {
		return false
	}
	reqType, _ := req.Metadata["request_type"].(string)
	return reqType == "utility" || reqType == "summarize" || reqType == "classify"
}
