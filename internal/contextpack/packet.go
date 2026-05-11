package contextpack

import "fmt"

type PacketType string

const (
	PacketTypeSpec      PacketType = "spec"
	PacketTypeCode      PacketType = "code"
	PacketTypeDoc       PacketType = "doc"
	PacketTypePrinciple PacketType = "principle"
	PacketTypeTool      PacketType = "tool"
)

type ContextPacket struct {
	ID          string     `json:"id"`
	Type        PacketType `json:"type"`
	Source      string     `json:"source"`
	Summary     string     `json:"summary,omitempty"`
	Content     string     `json:"content,omitempty"`
	Reason      string     `json:"reason"`
	Relevance   float64    `json:"relevance"`
	Bytes       int        `json:"bytes"`
	IsPartial   bool       `json:"is_partial,omitempty"`
	TruncatedAt int        `json:"truncated_at,omitempty"`
}

type ContextBundle struct {
	TaskID     string          `json:"task_id"`
	ContractID string          `json:"contract_id"`
	Goal       string          `json:"goal"`
	Packets    []ContextPacket `json:"packets"`
	Trace      AssemblyTrace   `json:"trace"`
	Budget     ContextBudget   `json:"budget"`
}

type AssemblyTrace struct {
	Goal     string      `json:"goal"`
	Selected []TraceItem `json:"selected"`
	Excluded []TraceItem `json:"excluded"`
	Notes    []string    `json:"notes,omitempty"`
}

type TraceItem struct {
	PacketID  string  `json:"packet_id"`
	Source    string  `json:"source"`
	Reason    string  `json:"reason"`
	Relevance float64 `json:"relevance"`
	Bytes     int     `json:"bytes"`
}

type ContextBudget struct {
	MaxPackets int  `json:"max_packets"`
	MaxBytes   int  `json:"max_bytes"`
	UsedBytes  int  `json:"used_bytes"`
	Truncated  bool `json:"truncated"`
}

func DefaultBudget() ContextBudget {
	return ContextBudget{MaxPackets: 5, MaxBytes: 8192}
}

func (p ContextPacket) Validate() error {
	if p.ID == "" {
		return fmt.Errorf("context packet id is required")
	}
	if p.Source == "" {
		return fmt.Errorf("context packet source is required")
	}
	if p.Reason == "" {
		return fmt.Errorf("context packet reason is required")
	}
	return nil
}
