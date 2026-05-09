package contracts

import (
	"testing"

	"github.com/axis-cli/axis/internal/types"
)

func TestAnalyzeContract_Schema(t *testing.T) {
	c := AnalyzeContract()
	if c.ContractID != ContractIDAnalyze {
		t.Errorf("expected contract ID %q, got %q", ContractIDAnalyze, c.ContractID)
	}
	if len(c.InputSchema.Fields) != 3 {
		t.Errorf("expected 3 input fields, got %d", len(c.InputSchema.Fields))
	}
	if len(c.OutputSchema.Fields) != 4 {
		t.Errorf("expected 4 output fields, got %d", len(c.OutputSchema.Fields))
	}
}

func TestImplementContract_Schema(t *testing.T) {
	c := ImplementContract()
	if c.ContractID != ContractIDImplement {
		t.Errorf("expected contract ID %q, got %q", ContractIDImplement, c.ContractID)
	}
	if len(c.InputSchema.Fields) != 2 {
		t.Errorf("expected 2 input fields, got %d", len(c.InputSchema.Fields))
	}
}

func TestValidateContract_Schema(t *testing.T) {
	c := ValidateContract()
	if c.ContractID != ContractIDValidate {
		t.Errorf("expected contract ID %q, got %q", ContractIDValidate, c.ContractID)
	}
	if len(c.OutputSchema.Fields) != 5 {
		t.Errorf("expected 5 output fields, got %d", len(c.OutputSchema.Fields))
	}
}

func TestUpdateDocsContract_Schema(t *testing.T) {
	c := UpdateDocsContract()
	if c.ContractID != ContractIDUpdate {
		t.Errorf("expected contract ID %q, got %q", ContractIDUpdate, c.ContractID)
	}
}

func TestReviewContract_Schema(t *testing.T) {
	c := ReviewContract()
	if c.ContractID != ContractIDReview {
		t.Errorf("expected contract ID %q, got %q", ContractIDReview, c.ContractID)
	}
}

func TestSpawnContract_Schema(t *testing.T) {
	c := SpawnContract()
	if c.ContractID != ContractIDSpawn {
		t.Errorf("expected contract ID %q, got %q", ContractIDSpawn, c.ContractID)
	}
}

func TestAllContracts(t *testing.T) {
	contracts := AllContracts()
	if len(contracts) != 6 {
		t.Errorf("expected 6 contracts, got %d", len(contracts))
	}
}

func TestRegisterAll(t *testing.T) {
	var registered []*types.AgentContract
	RegisterAll(func(c *types.AgentContract) {
		registered = append(registered, c)
	})
	if len(registered) != 6 {
		t.Errorf("expected 6 registered contracts, got %d", len(registered))
	}
}
