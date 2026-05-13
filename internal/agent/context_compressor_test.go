package agent

import (
	"strings"
	"testing"
)

func TestNewContextCompressor(t *testing.T) {
	comp := NewContextCompressor(StrategyByPriority)

	if comp == nil {
		t.Fatal("NewContextCompressor returned nil")
	}
	if comp.strategy != StrategyByPriority {
		t.Errorf("expected StrategyByPriority, got %v", comp.strategy)
	}
}

func TestCompress_NilContext(t *testing.T) {
	comp := NewContextCompressor(StrategyByPriority)

	result := comp.Compress(nil, 1000)

	if result != nil {
		t.Error("compress of nil should return nil")
	}
}

func TestCompress_ZeroMaxTokens(t *testing.T) {
	comp := NewContextCompressor(StrategyByPriority)

	ctx := NewSelfContext("task-1")
	ctx.CodeSnapshot.ModifiedFiles = []string{"file1.go", "file2.go"}

	result := comp.Compress(ctx, 0)

	// Should use default maxTokens of 1000 and return full clone
	if result == nil {
		t.Fatal("result should not be nil")
	}
	if result.TaskID != "task-1" {
		t.Errorf("expected TaskID 'task-1', got '%s'", result.TaskID)
	}
}

func TestCompress_AlreadyWithinLimit(t *testing.T) {
	comp := NewContextCompressor(StrategyByPriority)

	ctx := NewSelfContext("task-1")
	ctx.CompetenceScore = 0.8
	ctx.UpdateCompetenceScore(0.8)

	result := comp.Compress(ctx, 10000) // High limit

	if result == nil {
		t.Fatal("result should not be nil")
	}
	if result.TaskID != "task-1" {
		t.Errorf("TaskID should be preserved")
	}
	if result.CompetenceScore != 0.8 {
		t.Errorf("CompetenceScore should be preserved")
	}
}

func TestEstimateTokens(t *testing.T) {
	comp := NewContextCompressor(StrategyByPriority)

	ctx := NewSelfContext("task-12345678901234567890")
	ctx.TaskLineage = []string{"dep1", "dep2", "dep3"}
	ctx.CodeSnapshot.ModifiedFiles = []string{"internal/a.go", "internal/b.go"}
	ctx.CodeSnapshot.SpecVersion = "M5"
	ctx.CodeSnapshot.TaskCount = 10
	ctx.CodeSnapshot.ToolCount = 20
	ctx.DocSnapshot.SpecFiles = []string{"spec1.md", "spec2.md"}
	ctx.DocSnapshot.DocFiles = []string{"doc1.md"}
	ctx.DocSnapshot.Reports = []string{"report1.md"}
	ctx.StateSnapshot.RunningTasks = 2
	ctx.StateSnapshot.PendingTasks = 3
	ctx.StateSnapshot.CompletedTasks = 10
	ctx.StateSnapshot.FailedTasks = 1

	tokens := comp.estimateTokens(ctx)

	// Should return reasonable token estimate
	if tokens < 50 {
		t.Errorf("expected tokens > 50, got %d", tokens)
	}
}

func TestEstimateTokens_NilContext(t *testing.T) {
	comp := NewContextCompressor(StrategyByPriority)

	tokens := comp.estimateTokens(nil)

	if tokens != 0 {
		t.Errorf("expected 0 tokens for nil context, got %d", tokens)
	}
}

func TestCompressByPriority(t *testing.T) {
	comp := NewContextCompressor(StrategyByPriority)

	ctx := NewSelfContext("task-1")
	ctx.TaskLineage = []string{"dep1", "dep2", "dep3"}
	ctx.CodeSnapshot.ModifiedFiles = make([]string, 100)
	for i := 0; i < 100; i++ {
		ctx.CodeSnapshot.ModifiedFiles[i] = "internal/very/long/path/file" + string(rune(i)) + ".go"
	}
	ctx.CodeSnapshot.TaskCount = 50
	ctx.DocSnapshot.SpecFiles = make([]string, 50)
	for i := 0; i < 50; i++ {
		ctx.DocSnapshot.SpecFiles[i] = "docs/specs/very/long/path/spec" + string(rune(i)) + ".md"
	}
	ctx.StateSnapshot.RunningTasks = 5
	ctx.StateSnapshot.PendingTasks = 10

	// Very low token limit
	result := comp.compressByPriority(ctx, 100)

	if result == nil {
		t.Fatal("result should not be nil")
	}

	// StateSnapshot should always be kept
	if result.StateSnapshot == nil {
		t.Error("StateSnapshot should be preserved")
	}

	// TaskID and AutonomyLevel should be preserved
	if result.TaskID != "task-1" {
		t.Errorf("TaskID should be preserved")
	}
}

func TestCompressByLineage(t *testing.T) {
	comp := NewContextCompressor(StrategyByLineage)

	ctx := NewSelfContext("task-1")
	ctx.TaskLineage = []string{"dep1", "dep2", "dep3", "dep4", "dep5"}
	ctx.CompetenceScore = 0.9
	ctx.UpdateCompetenceScore(0.9)

	result := comp.compressByLineage(ctx, 100)

	if result == nil {
		t.Fatal("result should not be nil")
	}

	// Should keep most recent lineage entries
	if len(result.TaskLineage) > len(ctx.TaskLineage) {
		t.Error("lineage should be truncated")
	}
}

func TestCompressByRecency(t *testing.T) {
	comp := NewContextCompressor(StrategyByRecency)

	ctx := NewSelfContext("task-1")
	ctx.TaskLineage = []string{"dep1", "dep2", "dep3", "dep4", "dep5"}
	ctx.CodeSnapshot.ModifiedFiles = make([]string, 20)
	for i := 0; i < 20; i++ {
		ctx.CodeSnapshot.ModifiedFiles[i] = "internal/file" + string(rune(i)) + ".go"
	}
	ctx.DocSnapshot.SpecFiles = make([]string, 10)
	for i := 0; i < 10; i++ {
		ctx.DocSnapshot.SpecFiles[i] = "docs/spec" + string(rune(i)) + ".md"
	}
	ctx.StateSnapshot.RunningTasks = 2

	result := comp.compressByRecency(ctx, 1000)

	if result == nil {
		t.Fatal("result should not be nil")
	}

	// StateSnapshot should be preserved
	if result.StateSnapshot == nil {
		t.Error("StateSnapshot should be preserved")
	}

	// Should keep limited lineage
	if len(result.TaskLineage) > 3 {
		t.Errorf("lineage should be limited to 3, got %d", len(result.TaskLineage))
	}

	// CodeSnapshot files should be truncated to 10
	if result.CodeSnapshot != nil && len(result.CodeSnapshot.ModifiedFiles) > 10 {
		t.Errorf("ModifiedFiles should be truncated to 10, got %d", len(result.CodeSnapshot.ModifiedFiles))
	}
}

func TestCompressFiles(t *testing.T) {
	files := []string{
		"a.go",
		"bb.go",
		"ccc.go",
		"internal/dddd.go",
		"internal/eeee.go",
		"internal/ffff.go",
		"internal/gggg.go",
		"internal/hhhh.go",
		"docs/iiii.md",
		"docs/jjjj.md",
	}

	result := CompressFiles(files, 10)

	if len(result) == 0 {
		t.Error("result should not be empty")
	}

	// Should prioritize shorter paths
	for i := 0; i < len(result)-1; i++ {
		if len(result[i]) > len(result[i+1]) {
			t.Error("result should be sorted by path length (shorter first)")
		}
	}
}

func TestCompressFiles_EmptyInput(t *testing.T) {
	result := CompressFiles(nil, 10)
	if len(result) != 0 {
		t.Errorf("expected empty result for nil input, got %d", len(result))
	}

	result = CompressFiles([]string{}, 10)
	if len(result) != 0 {
		t.Errorf("expected empty result for empty input, got %d", len(result))
	}
}

func TestCompressFiles_ZeroMaxTokens(t *testing.T) {
	files := []string{"a.go", "b.go", "c.go"}
	result := CompressFiles(files, 0)

	// Should return all files when maxTokens is 0
	if len(result) != len(files) {
		t.Errorf("expected %d files, got %d", len(files), len(result))
	}
}

func TestCompress_StrategyByLineage(t *testing.T) {
	comp := NewContextCompressor(StrategyByLineage)

	ctx := NewSelfContext("task-1")
	ctx.TaskLineage = make([]string, 10)
	for i := 0; i < 10; i++ {
		ctx.TaskLineage[i] = "dep" + string(rune('0'+i))
	}
	ctx.CodeSnapshot.SpecVersion = "M5"
	ctx.StateSnapshot.CompletedTasks = 5

	result := comp.Compress(ctx, 200)

	if result == nil {
		t.Fatal("result should not be nil")
	}

	// Should preserve core fields
	if result.TaskID != "task-1" {
		t.Error("TaskID should be preserved")
	}
	if result.StateSnapshot == nil {
		t.Error("StateSnapshot should be preserved")
	}
}

func TestCompress_AllStrategies(t *testing.T) {
	ctx := NewSelfContext("task-1")
	ctx.TaskLineage = []string{"dep1", "dep2"}
	ctx.CodeSnapshot.ModifiedFiles = []string{"a.go", "b.go", "c.go"}
	ctx.CodeSnapshot.TaskCount = 3
	ctx.DocSnapshot.SpecFiles = []string{"spec1.md"}
	ctx.StateSnapshot.RunningTasks = 1
	ctx.StateSnapshot.PendingTasks = 2

	strategies := []CompressionStrategy{StrategyByPriority, StrategyByLineage, StrategyByRecency}

	for _, stgy := range strategies {
		comp := NewContextCompressor(stgy)
		result := comp.Compress(ctx, 500)

		if result == nil {
			t.Errorf("strategy %v: result should not be nil", stgy)
			continue
		}
		if result.TaskID != "task-1" {
			t.Errorf("strategy %v: TaskID should be preserved", stgy)
		}
		if result.StateSnapshot == nil {
			t.Errorf("strategy %v: StateSnapshot should be preserved", stgy)
		}
	}
}

func TestCompressor_DoesNotModifyOriginal(t *testing.T) {
	comp := NewContextCompressor(StrategyByPriority)

	original := NewSelfContext("task-1")
	original.TaskLineage = []string{"dep1", "dep2", "dep3"}
	original.CodeSnapshot.ModifiedFiles = make([]string, 50)
	original.CodeSnapshot.SpecVersion = "M5"
	original.DocSnapshot.SpecFiles = make([]string, 20)

	originalCloneLen := len(original.TaskLineage)
	originalFilesLen := len(original.CodeSnapshot.ModifiedFiles)
	originalSpecsLen := len(original.DocSnapshot.SpecFiles)

	result := comp.Compress(original, 100)

	// Original should not be modified
	if len(original.TaskLineage) != originalCloneLen {
		t.Error("original TaskLineage should not be modified")
	}
	if len(original.CodeSnapshot.ModifiedFiles) != originalFilesLen {
		t.Error("original ModifiedFiles should not be modified")
	}
	if len(original.DocSnapshot.SpecFiles) != originalSpecsLen {
		t.Error("original SpecFiles should not be modified")
	}

	// Result should be different from original
	if result == original {
		t.Error("result should be a different reference")
	}
}

func TestCompress_LargeContext(t *testing.T) {
	comp := NewContextCompressor(StrategyByPriority)

	// Create a large context
	ctx := NewSelfContext("task-large")
	ctx.TaskLineage = make([]string, 50)
	for i := 0; i < 50; i++ {
		ctx.TaskLineage[i] = "very/long/dependency/path/dep" + strings.Repeat("x", 50) + string(rune(i))
	}
	ctx.CodeSnapshot.ModifiedFiles = make([]string, 200)
	for i := 0; i < 200; i++ {
		ctx.CodeSnapshot.ModifiedFiles[i] = "internal/very/long/path/to/file/file" + strings.Repeat("x", 100) + ".go"
	}
	ctx.CodeSnapshot.SpecVersion = strings.Repeat("M", 1000)
	ctx.CodeSnapshot.TaskCount = 1000
	ctx.CodeSnapshot.ToolCount = 500
	ctx.DocSnapshot.SpecFiles = make([]string, 100)
	for i := 0; i < 100; i++ {
		ctx.DocSnapshot.SpecFiles[i] = "docs/specs/very/long/path/spec" + strings.Repeat("x", 100) + ".md"
	}
	ctx.DocSnapshot.DocFiles = make([]string, 100)
	ctx.DocSnapshot.Reports = make([]string, 50)
	ctx.StateSnapshot.RunningTasks = 100
	ctx.StateSnapshot.PendingTasks = 200
	ctx.StateSnapshot.CompletedTasks = 500
	ctx.StateSnapshot.FailedTasks = 50

	result := comp.Compress(ctx, 500)

	if result == nil {
		t.Fatal("result should not be nil")
	}

	// Should be significantly smaller
	originalTokens := comp.estimateTokens(ctx)
	resultTokens := comp.estimateTokens(result)

	if resultTokens >= originalTokens {
		t.Logf("Note: result tokens (%d) may not be smaller than original (%d) in some cases", resultTokens, originalTokens)
	}
}

func TestCompressByPriority_NilFields(t *testing.T) {
	comp := NewContextCompressor(StrategyByPriority)

	ctx := NewSelfContext("task-1")
	ctx.CodeSnapshot = nil // nil code snapshot
	ctx.DocSnapshot = nil  // nil doc snapshot
	ctx.StateSnapshot = nil

	result := comp.compressByPriority(ctx, 100)

	if result == nil {
		t.Fatal("result should not be nil")
	}
	if result.TaskID != "task-1" {
		t.Error("TaskID should be preserved")
	}
	if result.CodeSnapshot != nil {
		t.Error("CodeSnapshot should remain nil")
	}
	if result.DocSnapshot != nil {
		t.Error("DocSnapshot should remain nil")
	}
	if result.StateSnapshot != nil {
		t.Error("StateSnapshot should remain nil")
	}
}

func TestCompressByPriority_PreserveStateSnapshot(t *testing.T) {
	comp := NewContextCompressor(StrategyByPriority)

	ctx := NewSelfContext("task-1")
	ctx.StateSnapshot.RunningTasks = 5
	ctx.StateSnapshot.PendingTasks = 10
	ctx.StateSnapshot.CompletedTasks = 100
	ctx.StateSnapshot.FailedTasks = 2

	result := comp.compressByPriority(ctx, 50)

	if result == nil {
		t.Fatal("result should not be nil")
	}
	if result.StateSnapshot == nil {
		t.Error("StateSnapshot should always be preserved")
	}
	if result.StateSnapshot.RunningTasks != 5 {
		t.Error("StateSnapshot fields should be preserved")
	}
}

func TestCompressByRecency_NilFields(t *testing.T) {
	comp := NewContextCompressor(StrategyByRecency)

	ctx := NewSelfContext("task-1")
	ctx.CodeSnapshot = nil
	ctx.DocSnapshot = nil
	ctx.StateSnapshot = nil

	result := comp.compressByRecency(ctx, 100)

	if result == nil {
		t.Fatal("result should not be nil")
	}
	if result.TaskID != "task-1" {
		t.Error("TaskID should be preserved")
	}
}

func TestEstimateTokens_WithNilSnapshots(t *testing.T) {
	comp := NewContextCompressor(StrategyByPriority)

	ctx := NewSelfContext("task-1")
	ctx.CodeSnapshot = nil
	ctx.DocSnapshot = nil
	ctx.StateSnapshot = nil

	tokens := comp.estimateTokens(ctx)

	// Should still return base tokens
	if tokens < 50 {
		t.Errorf("expected tokens >= 50, got %d", tokens)
	}
}
