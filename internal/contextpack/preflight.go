package contextpack

import (
	"fmt"
	"strconv"

	"github.com/axis-cli/axis/internal/types"
)

type PreflightStatus string

const (
	PreflightStatusReady       PreflightStatus = "ready"
	PreflightStatusMissing     PreflightStatus = "missing"
	PreflightStatusUntraceable PreflightStatus = "untraceable"
)

type PreflightResult struct {
	TaskID       string          `json:"task_id"`
	Status       PreflightStatus `json:"status"`
	Reason       string          `json:"reason,omitempty"`
	BundleID     string          `json:"bundle_id,omitempty"`
	PacketCount  int             `json:"packet_count,omitempty"`
	Truncated    bool            `json:"truncated"`
	SourceDigest string          `json:"source_digest,omitempty"`
}

func Preflight(task *types.AgentTask, registry *ReadinessRegistry) PreflightResult {
	if task == nil {
		return PreflightResult{Status: PreflightStatusMissing, Reason: "task is required"}
	}
	result := PreflightResult{TaskID: task.TaskID}
	if task.Metadata == nil {
		result.Status = PreflightStatusMissing
		result.Reason = "task has no context readiness metadata"
		return result
	}
	bundleID := task.Metadata[MetadataBundleID]
	if bundleID == "" {
		result.Status = PreflightStatusMissing
		result.Reason = "task has no context.bundle_id"
		return result
	}
	result.BundleID = bundleID
	result.SourceDigest = task.Metadata[MetadataSourceDigest]
	if v := task.Metadata[MetadataPacketCount]; v != "" {
		result.PacketCount, _ = strconv.Atoi(v)
	}
	if v := task.Metadata[MetadataTruncated]; v != "" {
		result.Truncated, _ = strconv.ParseBool(v)
	}
	if registry == nil {
		result.Status = PreflightStatusUntraceable
		result.Reason = "readiness registry is not available"
		return result
	}
	record, err := registry.Inspect(bundleID)
	if err != nil {
		result.Status = PreflightStatusUntraceable
		result.Reason = fmt.Sprintf("context readiness record is not inspectable: %v", err)
		return result
	}
	if record.Artifact.SourceDigest != result.SourceDigest {
		result.Status = PreflightStatusUntraceable
		result.Reason = "context source digest does not match readiness record"
		return result
	}
	if result.PacketCount <= 0 {
		result.Status = PreflightStatusMissing
		result.Reason = "context readiness has no selected packets"
		return result
	}
	result.Status = PreflightStatusReady
	return result
}
