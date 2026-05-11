# SWE1.6 Renormalization Report

**Date**: 2026-05-10
**Guide**: docs/architecture/swe1-6-renormalization-guide.md

## Scope

Audit and normalize existing Axis code against the current convention suite, focusing on P0 priority targets:
- CLI and shell normalization
- Ask command normalization
- Provider command safety
- Metadata key alignment
- External tool boundary audit
- Spec status alignment
- Evolution boundary alignment

## Convention Applied

- metadata-key-conventions.md (namespaced metadata keys)
- module-and-naming-conventions.md (module structure and naming)
- external-tool-boundaries.md (external tool isolation)
- spec-lifecycle-conventions.md (spec status tracking)
- semantic-boundaries.md (semantic ownership)
- secret-handling.md (API key redaction)

## Evolution Boundary Check

No evolution-boundary issues identified. All changes are ordinary renormalization that preserve user-visible behavior without modifying system structure, control logic, or promotion/discard semantics.

## Files Changed

1. **internal/intent/parser.go**
   - Added namespaced metadata keys (`intent.source`, `intent.original_prompt`, `intent.parser_mode`)
   - Preserved legacy un-namespaced keys for backward compatibility during transition period
   - Category: metadata

2. **internal/intent/parser_test.go**
   - Updated tests to verify new namespaced keys
   - Added backward compatibility checks for legacy keys
   - Category: test-gap

3. **docs/architecture/module-and-naming-conventions.md**
   - Updated "Current Known Alignment Gaps" section to reflect intent metadata normalization
   - Category: doc-drift

## Behavior Preserved

- Intent parser continues to write metadata with the same values
- Legacy un-namespaced keys remain present for backward compatibility
- CLI output unchanged
- Task submission behavior unchanged
- All existing tests pass with added compatibility checks

## Tests Run

```bash
go test ./internal/intent/...  # PASS
go test ./...                  # PASS (all packages)
```

Specific test coverage:
- TestDeterministicParser_ParseDefaultTask - now checks both namespaced and legacy keys
- TestDeterministicParser_ParseExplicitContractAndTaskID - unchanged
- TestDeterministicParser_ParseEmptyPrompt - unchanged

## Remaining Gaps

None identified for P0 targets. All P0 priorities are now compliant:

- ✅ CLI and shell normalization: Shell logic already extracted to shell_cmd.go
- ✅ Ask command normalization: Shared rendering helper renderTaskProposal already exists
- ✅ Provider command safety: API keys are not printed in list/status commands
- ✅ Metadata key alignment: Intent metadata now uses namespaced keys with legacy compatibility
- ✅ External tool boundary: tools/axis-up does not import internal packages (verified via grep)
- ✅ Spec status alignment: All checked specs have proper status tracking (Completed/Planned)
- ✅ Evolution boundary alignment: evolution.* namespace reserved in conventions

## Category: metadata

**Issue**: Intent metadata keys used un-namespaced form (`source`, `original_prompt`, `parser_mode`)

**Resolution**: Added namespaced keys (`intent.source`, `intent.original_prompt`, `intent.parser_mode`) while preserving legacy keys for backward compatibility per metadata-key-conventions.md compatibility rule.

**Future cleanup**: After a migration window, legacy keys can be removed following the deprecation documentation in metadata-key-conventions.md.

## Category: test-gap

**Issue**: Tests only checked legacy un-namespaced metadata keys

**Resolution**: Updated parser_test.go to verify new namespaced keys while maintaining backward compatibility checks for legacy keys during transition.

## Category: doc-drift

**Issue**: module-and-naming-conventions.md listed intent metadata as a known alignment gap

**Resolution**: Updated the document to reflect that intent metadata normalization is now complete per SWE1.6.

## Acceptance Criteria

- ✅ Changes limited to declared scope (metadata key normalization)
- ✅ Behavior preserved (backward compatibility maintained)
- ✅ Tests pass (go test ./... passes)
- ✅ No new feature scope introduced
- ✅ Relevant docs/specs synchronized (module-and-naming-conventions.md updated)
- ✅ Remaining gaps listed clearly (none for P0 targets)
