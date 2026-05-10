// Package contracts provides self-iteration contracts.
package contracts

import "github.com/axis-cli/axis/internal/types"

// ContractIDJudge is the contract ID for the self-judgement execution.
const ContractIDJudge = "self/judge-execution"

// JudgeContract returns the self-judgement execution contract.
func JudgeContract() *types.AgentContract {
	return &types.AgentContract{
		ContractID: ContractIDJudge,
		InputSchema: &types.InputSchema{Fields: []types.FieldDef{
			{Name: "execution_result", Type: types.FieldTypeObject, Required: true},
			{Name: "criteria", Type: types.FieldTypeArray, Required: false},
			{Name: "context", Type: types.FieldTypeObject, Required: false},
		}},
		OutputSchema: &types.OutputSchema{Fields: []types.FieldDef{
			{Name: "judgement", Type: types.FieldTypeObject, Required: true},
			{Name: "confidence", Type: types.FieldTypeFloat, Required: true},
			{Name: "suggested_fixes", Type: types.FieldTypeArray, Required: false},
		}},
	}
}
