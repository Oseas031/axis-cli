package provider

// modelPricing holds approximate per-1M-token pricing in USD.
// This is a lightweight lookup table for cost estimation, not exact billing.
var modelPricing = map[string]struct {
	InputPrice  float64
	OutputPrice float64
}{
	"claude-3-5-sonnet-20241022": {InputPrice: 3.0, OutputPrice: 15.0},
	"gpt-4o-mini":                {InputPrice: 0.15, OutputPrice: 0.60},
	"deepseek-v4-flash":          {InputPrice: 0.10, OutputPrice: 0.50},
	"MiniMax-M2.7":               {InputPrice: 0.20, OutputPrice: 0.80},
}

// estimateCost computes a rough USD cost from token counts and model name.
func estimateCost(model string, inputTokens, outputTokens int) float64 {
	p, ok := modelPricing[model]
	if !ok {
		return 0
	}
	inputCost := float64(inputTokens) * p.InputPrice / 1e6
	outputCost := float64(outputTokens) * p.OutputPrice / 1e6
	return inputCost + outputCost
}
