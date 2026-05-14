package tool

import (
	"context"
	"testing"
)

func TestCompactTool_Name(t *testing.T) {
	ct := NewCompactTool()
	if ct.Name() != "compact" {
		t.Errorf("Name() = %q, want compact", ct.Name())
	}
}

func TestCompactTool_Schema(t *testing.T) {
	ct := NewCompactTool()
	schema := ct.Schema()
	if schema.Name != "compact" {
		t.Errorf("schema name = %q", schema.Name)
	}
	if len(schema.Parameters) != 0 {
		t.Error("compact tool should have no parameters")
	}
}

func TestCompactTool_Execute(t *testing.T) {
	ct := NewCompactTool()
	result, err := ct.Execute(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if result["status"] != "compaction_requested" {
		t.Errorf("status = %v, want compaction_requested", result["status"])
	}
	if result["message"] != "Context compaction will be applied on next turn." {
		t.Errorf("message = %v", result["message"])
	}
}
