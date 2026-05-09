package contracts

import "github.com/axis-cli/axis/internal/types"

const (
	ContractIDAnalyze   = "self/analyze-change-request"
	ContractIDImplement = "self/implement-change"
	ContractIDValidate  = "self/run-validation"
	ContractIDUpdate    = "self/update-docs"
	ContractIDReview    = "self/review-result"
	ContractIDSpawn     = "self/spawn-followup"
)

func AllContracts() []*types.AgentContract {
	return []*types.AgentContract{
		AnalyzeContract(),
		ImplementContract(),
		ValidateContract(),
		UpdateDocsContract(),
		ReviewContract(),
		SpawnContract(),
	}
}

func RegisterAll(registry func(*types.AgentContract)) {
	for _, c := range AllContracts() {
		registry(c)
	}
}

func AnalyzeContract() *types.AgentContract {
	return &types.AgentContract{
		ContractID: ContractIDAnalyze,
		InputSchema: &types.InputSchema{Fields: []types.FieldDef{
			{Name: "change_description", Type: types.FieldTypeString, Required: true},
			{Name: "target_files", Type: types.FieldTypeArray, Required: false},
			{Name: "motivation", Type: types.FieldTypeString, Required: false},
		}},
		OutputSchema: &types.OutputSchema{Fields: []types.FieldDef{
			{Name: "impact_scope", Type: types.FieldTypeArray, Required: true},
			{Name: "risk_level", Type: types.FieldTypeString, Required: true},
			{Name: "suggested_order", Type: types.FieldTypeArray, Required: false},
			{Name: "confidence", Type: types.FieldTypeFloat, Required: true},
		}},
	}
}

func ImplementContract() *types.AgentContract {
	return &types.AgentContract{
		ContractID: ContractIDImplement,
		InputSchema: &types.InputSchema{Fields: []types.FieldDef{
			{Name: "analysis_result", Type: types.FieldTypeObject, Required: true},
			{Name: "implementation_plan", Type: types.FieldTypeArray, Required: true},
		}},
		OutputSchema: &types.OutputSchema{Fields: []types.FieldDef{
			{Name: "modified_files", Type: types.FieldTypeArray, Required: true},
			{Name: "new_contracts", Type: types.FieldTypeArray, Required: false},
			{Name: "implementation_notes", Type: types.FieldTypeString, Required: false},
		}},
	}
}

func ValidateContract() *types.AgentContract {
	return &types.AgentContract{
		ContractID: ContractIDValidate,
		InputSchema: &types.InputSchema{Fields: []types.FieldDef{
			{Name: "modified_files", Type: types.FieldTypeArray, Required: true},
			{Name: "test_scope", Type: types.FieldTypeString, Required: false},
		}},
		OutputSchema: &types.OutputSchema{Fields: []types.FieldDef{
			{Name: "tests_passed", Type: types.FieldTypeInt, Required: true},
			{Name: "tests_failed", Type: types.FieldTypeInt, Required: true},
			{Name: "coverage", Type: types.FieldTypeFloat, Required: true},
			{Name: "is_acceptable", Type: types.FieldTypeBool, Required: true},
			{Name: "blocking_issues", Type: types.FieldTypeArray, Required: false},
		}},
	}
}

func UpdateDocsContract() *types.AgentContract {
	return &types.AgentContract{
		ContractID: ContractIDUpdate,
		InputSchema: &types.InputSchema{Fields: []types.FieldDef{
			{Name: "changed_files", Type: types.FieldTypeArray, Required: true},
			{Name: "validation_summary", Type: types.FieldTypeObject, Required: false},
		}},
		OutputSchema: &types.OutputSchema{Fields: []types.FieldDef{
			{Name: "updated_docs", Type: types.FieldTypeArray, Required: true},
			{Name: "new_docs", Type: types.FieldTypeArray, Required: false},
			{Name: "doc_quality_score", Type: types.FieldTypeFloat, Required: true},
		}},
	}
}

func ReviewContract() *types.AgentContract {
	return &types.AgentContract{
		ContractID: ContractIDReview,
		InputSchema: &types.InputSchema{Fields: []types.FieldDef{
			{Name: "implementation_result", Type: types.FieldTypeObject, Required: true},
			{Name: "validation_result", Type: types.FieldTypeObject, Required: true},
			{Name: "doc_result", Type: types.FieldTypeObject, Required: false},
		}},
		OutputSchema: &types.OutputSchema{Fields: []types.FieldDef{
			{Name: "approval_status", Type: types.FieldTypeString, Required: true},
			{Name: "review_notes", Type: types.FieldTypeString, Required: false},
			{Name: "suggested_followups", Type: types.FieldTypeArray, Required: false},
		}},
	}
}

func SpawnContract() *types.AgentContract {
	return &types.AgentContract{
		ContractID: ContractIDSpawn,
		InputSchema: &types.InputSchema{Fields: []types.FieldDef{
			{Name: "review_result", Type: types.FieldTypeObject, Required: true},
			{Name: "current_task_id", Type: types.FieldTypeString, Required: true},
		}},
		OutputSchema: &types.OutputSchema{Fields: []types.FieldDef{
			{Name: "new_tasks", Type: types.FieldTypeArray, Required: true},
			{Name: "loop_count", Type: types.FieldTypeInt, Required: true},
			{Name: "termination_reason", Type: types.FieldTypeString, Required: false},
		}},
	}
}
