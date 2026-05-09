package provider

import "fmt"

// NewProvider creates a ModelProvider by name. Supported: "mock", "echo".
func NewProvider(name string) (ModelProvider, error) {
	switch name {
	case "mock":
		return NewMockModelProvider(), nil
	case "echo":
		return NewEchoModelProvider(), nil
	default:
		return nil, fmt.Errorf("unknown provider %q: supported values are \"mock\" and \"echo\"", name)
	}
}
