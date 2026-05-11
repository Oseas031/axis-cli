package contextpack

import "testing"

func TestContextPacketValidateRequiresProvenance(t *testing.T) {
	packet := ContextPacket{ID: "p1", Source: "docs/specs/example/", Reason: "matched example", Relevance: 0.7}
	if err := packet.Validate(); err != nil {
		t.Fatalf("valid packet should pass validation: %v", err)
	}
}

func TestContextPacketValidateRejectsMissingSource(t *testing.T) {
	packet := ContextPacket{ID: "p1", Reason: "matched example", Relevance: 0.7}
	if err := packet.Validate(); err == nil {
		t.Fatal("expected missing source to fail validation")
	}
}

func TestContextPacketValidateRejectsMissingReason(t *testing.T) {
	packet := ContextPacket{ID: "p1", Source: "docs/specs/example/", Relevance: 0.7}
	if err := packet.Validate(); err == nil {
		t.Fatal("expected missing reason to fail validation")
	}
}
