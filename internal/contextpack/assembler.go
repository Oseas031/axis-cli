package contextpack

import (
	"fmt"
	"sort"
	"strings"

	"github.com/axis-cli/axis/internal/types"
)

type Assembler struct {
	budget ContextBudget
	index  *TFIDFIndex
}

type Option func(*Assembler)

func WithBudget(budget ContextBudget) Option {
	return func(a *Assembler) {
		a.budget = budget
	}
}

func WithIndex(index *TFIDFIndex) Option {
	return func(a *Assembler) {
		a.index = index
	}
}

func NewAssembler(opts ...Option) *Assembler {
	a := &Assembler{budget: DefaultBudget()}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func (a *Assembler) Assemble(task *types.AgentTask) (*ContextBundle, error) {
	if task == nil {
		return nil, fmt.Errorf("agent task is required")
	}
	goal := taskGoal(task)
	if strings.TrimSpace(goal) == "" {
		return nil, fmt.Errorf("task goal is required for context assembly")
	}
	candidates := a.candidates(goal, task.ContractID)
	traceNotes := []string{"rule-based preview only; execution path unchanged"}
	if a.index != nil && len(a.index.Chunks) > 0 {
		traceNotes = []string{"hybrid mode: rule + retrieval; execution path unchanged"}
	} else {
		traceNotes = append(traceNotes, "rule-only fallback: index not found or empty")
	}
	bundle := &ContextBundle{
		TaskID:     task.TaskID,
		ContractID: task.ContractID,
		Goal:       goal,
		Budget:     a.budget,
		Trace: AssemblyTrace{
			Goal:  goal,
			Notes: traceNotes,
		},
	}
	for _, packet := range candidates {
		packet.Bytes = packetSize(packet)
		if err := packet.Validate(); err != nil {
			return nil, err
		}
		if len(bundle.Packets) >= bundle.Budget.MaxPackets {
			bundle.Budget.Truncated = true
			bundle.Trace.Excluded = append(bundle.Trace.Excluded, traceItem(packet, "excluded by packet count budget"))
			continue
		}
		if bundle.Budget.UsedBytes+packet.Bytes > bundle.Budget.MaxBytes {
			remaining := bundle.Budget.MaxBytes - bundle.Budget.UsedBytes
			truncated, ok := tryTruncatePacket(packet, remaining)
			if ok {
				truncated.Bytes = packetSize(truncated)
				bundle.Packets = append(bundle.Packets, truncated)
				bundle.Budget.UsedBytes += truncated.Bytes
				bundle.Trace.Selected = append(bundle.Trace.Selected, traceItem(truncated, fmt.Sprintf("content truncated to %d bytes", truncated.TruncatedAt)))
				continue
			}
			bundle.Budget.Truncated = true
			bundle.Trace.Excluded = append(bundle.Trace.Excluded, traceItem(packet, "excluded by context budget"))
			continue
		}
		bundle.Packets = append(bundle.Packets, packet)
		bundle.Budget.UsedBytes += packet.Bytes
		bundle.Trace.Selected = append(bundle.Trace.Selected, traceItem(packet, packet.Reason))
	}
	if len(bundle.Packets) == 0 {
		bundle.Trace.Notes = append(bundle.Trace.Notes, "no rule matched; returned an empty preview bundle")
	}
	return bundle, nil
}

func (a *Assembler) candidates(goal string, contractID string) []ContextPacket {
	var packets []ContextPacket
	// Rule-based deterministic recall
	for _, r := range candidateRules() {
		if matchesAny(goal, r.keywords) || matchesAny(contractID, r.keywords) {
			packets = append(packets, r.packet)
		}
	}

	// Retrieval rerank when index is present and healthy
	if a.index != nil && len(a.index.Chunks) > 0 {
		results := a.index.Query(goal, 10)
		existingSources := make(map[string]int)
		for i, p := range packets {
			existingSources[p.Source] = i
		}
		for _, res := range results {
			if res.Score <= 0 {
				continue
			}
			// Deduplication: boost existing rule packet if source matches or is prefix.
			// Longer sources are checked first for deterministic priority.
			var sources []string
			for source := range existingSources {
				sources = append(sources, source)
			}
			sort.Slice(sources, func(i, j int) bool {
				return len(sources[i]) > len(sources[j])
			})
			boosted := false
			for _, source := range sources {
				idx := existingSources[source]
				if strings.HasPrefix(res.Chunk.Source, source) || res.Chunk.Source == source {
					if res.Score > packets[idx].Relevance {
						packets[idx].Relevance = res.Score
					}
					boosted = true
					break
				}
			}
			if boosted {
				continue
			}
			pktType := PacketTypeDoc
			if res.Chunk.DocType == "code" {
				pktType = PacketTypeCode
			}
			packets = append(packets, ContextPacket{
				ID:        "retrieval:" + res.Chunk.Source,
				Type:      pktType,
				Source:    res.Chunk.Source,
				Summary:   res.Chunk.Content,
				Reason:    "retrieval: tf-idf cosine similarity",
				Relevance: res.Score,
			})
		}
	}

	sort.SliceStable(packets, func(i, j int) bool {
		return packets[i].Relevance > packets[j].Relevance
	})
	return packets
}

func taskGoal(task *types.AgentTask) string {
	if value, ok := task.Input["goal"].(string); ok && strings.TrimSpace(value) != "" {
		return value
	}
	if value, ok := task.Input["message"].(string); ok && strings.TrimSpace(value) != "" {
		return value
	}
	return ""
}

func packetSize(packet ContextPacket) int {
	return len(packet.ID) + len(string(packet.Type)) + len(packet.Source) + len(packet.Summary) + len(packet.Content) + len(packet.Reason)
}

// tryTruncatePacket attempts to fit packet into remainingBytes by truncating
// its Content (and then Summary) at semantic boundaries.
// It returns the modified packet and true if truncation succeeded.
func tryTruncatePacket(packet ContextPacket, remainingBytes int) (ContextPacket, bool) {
	if remainingBytes <= 0 {
		return packet, false
	}
	// Fixed overhead: fields that must never be truncated.
	fixed := len(packet.ID) + len(string(packet.Type)) + len(packet.Source) + len(packet.Reason)
	if fixed >= remainingBytes {
		return packet, false
	}
	available := remainingBytes - fixed

	// Strategy: truncate Content first (usually longest), then Summary.
	if packet.Content != "" {
		contentBudget := available - len(packet.Summary)
		if contentBudget > 0 {
			truncatedContent, pos := truncateAtSemanticBoundary(packet.Content, contentBudget)
			if pos > 0 {
				packet.Content = truncatedContent
				packet.IsPartial = true
				packet.TruncatedAt = pos
				return packet, true
			}
		}
	}
	if packet.Summary != "" {
		summaryBudget := available - len(packet.Content)
		if summaryBudget > 0 {
			truncatedSummary, pos := truncateAtSemanticBoundary(packet.Summary, summaryBudget)
			if pos > 0 {
				packet.Summary = truncatedSummary
				packet.IsPartial = true
				packet.TruncatedAt = pos
				return packet, true
			}
		}
	}
	return packet, false
}

func traceItem(packet ContextPacket, reason string) TraceItem {
	return TraceItem{PacketID: packet.ID, Source: packet.Source, Reason: reason, Relevance: packet.Relevance, Bytes: packet.Bytes}
}
