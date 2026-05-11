package contextpack

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/axis-cli/axis/internal/types"
)

const (
	MetadataBundleID         = "context.bundle_id"
	MetadataAssemblyMode     = "context.assembly_mode"
	MetadataPacketCount      = "context.packet_count"
	MetadataTruncated        = "context.truncated"
	MetadataSourceDigest     = "context.source_digest"
	MetadataRequestedSources = "context.requested_sources"
)

type ReadinessArtifact struct {
	BundleID     string   `json:"bundle_id"`
	AssemblyMode string   `json:"assembly_mode"`
	TaskID       string   `json:"task_id"`
	PacketCount  int      `json:"packet_count"`
	Truncated    bool     `json:"truncated"`
	SourceDigest string   `json:"source_digest"`
	Sources      []string `json:"sources"`
}

func NewReadinessArtifact(bundle *ContextBundle) (ReadinessArtifact, error) {
	if bundle == nil {
		return ReadinessArtifact{}, fmt.Errorf("context bundle is required")
	}
	sources := bundleSources(bundle)
	digest := sourceDigest(sources)
	artifact := ReadinessArtifact{
		AssemblyMode: "rule_based",
		TaskID:       bundle.TaskID,
		PacketCount:  len(bundle.Packets),
		Truncated:    bundle.Budget.Truncated,
		SourceDigest: digest,
		Sources:      sources,
	}
	artifact.BundleID = bundleID(bundle, artifact)
	return artifact, nil
}

func AttachReadinessMetadata(task *types.AgentTask, artifact ReadinessArtifact) error {
	if task == nil {
		return fmt.Errorf("agent task is required")
	}
	if artifact.BundleID == "" {
		return fmt.Errorf("readiness artifact bundle id is required")
	}
	if task.Metadata == nil {
		task.Metadata = make(map[string]string)
	}
	task.Metadata[MetadataBundleID] = artifact.BundleID
	task.Metadata[MetadataAssemblyMode] = artifact.AssemblyMode
	task.Metadata[MetadataPacketCount] = strconv.Itoa(artifact.PacketCount)
	task.Metadata[MetadataTruncated] = strconv.FormatBool(artifact.Truncated)
	task.Metadata[MetadataSourceDigest] = artifact.SourceDigest
	return nil
}

func bundleSources(bundle *ContextBundle) []string {
	sources := make([]string, 0, len(bundle.Packets))
	seen := make(map[string]bool)
	for _, packet := range bundle.Packets {
		if packet.Source == "" || seen[packet.Source] {
			continue
		}
		seen[packet.Source] = true
		sources = append(sources, packet.Source)
	}
	sort.Strings(sources)
	return sources
}

func sourceDigest(sources []string) string {
	h := sha256.Sum256([]byte(strings.Join(sources, "\n")))
	return hex.EncodeToString(h[:])[:16]
}

func bundleID(bundle *ContextBundle, artifact ReadinessArtifact) string {
	basis := strings.Join([]string{
		bundle.TaskID,
		bundle.ContractID,
		bundle.Goal,
		artifact.AssemblyMode,
		artifact.SourceDigest,
		strconv.Itoa(artifact.PacketCount),
		strconv.FormatBool(artifact.Truncated),
	}, "\n")
	h := sha256.Sum256([]byte(basis))
	return "ctx-" + hex.EncodeToString(h[:])[:16]
}
