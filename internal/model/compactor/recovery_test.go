package compactor

import (
	"testing"
	"time"
)

func TestExtractRecoveryContext(t *testing.T) {
	tests := []struct {
		name        string
		history     []Message
		hasPlan     bool
		hasFeedback bool
		fileCount   int
	}{
		{
			name:        "empty history",
			history:     []Message{},
			hasPlan:     false,
			hasFeedback: false,
			fileCount:   0,
		},
		{
			name: "with plan",
			history: []Message{
				{Role: "assistant", Content: "## Plan\n1. Step 1\n2. Step 2"},
				{Role: "user", Content: "go ahead"},
			},
			hasPlan:     true,
			hasFeedback: true,
			fileCount:   0,
		},
		{
			name: "with file operations",
			history: []Message{
				{Role: "tool", Content: "file created", Path: "test.go"},
				{Role: "tool", Content: "file modified", Path: "test2.go"},
			},
			hasPlan:     false,
			hasFeedback: false,
			fileCount:   2,
		},
		{
			name: "long plan truncated",
			history: []Message{
				{Role: "assistant", Content: "This is a long assistant message without plan headers that exceeds five hundred characters in length to test the truncation logic that should kick in and return only the first five hundred characters followed by ellipsis to indicate the rest was truncated during extraction"},
			},
			hasPlan:     true,
			hasFeedback: false,
			fileCount:   0,
		},
		{
			name: "long feedback truncated",
			history: []Message{
				{Role: "user", Content: "This is a very long user message that exceeds one thousand characters and should be truncated during feedback extraction so that only the first thousand characters are kept and an ellipsis is appended to indicate the rest was dropped because it is too old or irrelevant to the current context"},
			},
			hasPlan:     false,
			hasFeedback: true,
			fileCount:   0,
		},
		{
			name: "tool message without path skipped",
			history: []Message{
				{Role: "tool", Content: "some result", Path: ""},
			},
			hasPlan:     false,
			hasFeedback: false,
			fileCount:   0,
		},
		{
			name: "delete operation detected",
			history: []Message{
				{Role: "tool", Content: "file deleted", Path: "old.go"},
			},
			hasPlan:     false,
			hasFeedback: false,
			fileCount:   1,
		},
		{
			name: "chinese create operation",
			history: []Message{
				{Role: "tool", Content: "文件已创建", Path: "new.go"},
			},
			hasPlan:     false,
			hasFeedback: false,
			fileCount:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := ExtractRecoveryContext(tt.history)

			if (ctx.ActivePlan != "") != tt.hasPlan {
				t.Errorf("ActivePlan: got %v, want %v", ctx.ActivePlan != "", tt.hasPlan)
			}
			if (ctx.Feedback != "") != tt.hasFeedback {
				t.Errorf("Feedback: got %v, want %v", ctx.Feedback != "", tt.hasFeedback)
			}
			if len(ctx.FileStates) != tt.fileCount {
				t.Errorf("FileStates: got %d, want %d", len(ctx.FileStates), tt.fileCount)
			}
			if ctx.ExtractedAt.IsZero() {
				t.Error("ExtractedAt should be set")
			}
		})
	}
}

func TestExtractActivePlan_WithHeaders(t *testing.T) {
	history := []Message{
		{Role: "assistant", Content: "Some intro text\n## Plan\n1. Do stuff\n2. Do more stuff"},
	}
	plan := extractActivePlan(history)
	if plan != "## Plan\n1. Do stuff\n2. Do more stuff" {
		t.Errorf("expected plan content, got %q", plan)
	}
}

func TestExtractActivePlan_ChineseHeader(t *testing.T) {
	history := []Message{
		{Role: "assistant", Content: "准备开始\n## 执行计划\n1. 实现功能\n2. 测试"},
	}
	plan := extractActivePlan(history)
	if plan != "## 执行计划\n1. 实现功能\n2. 测试" {
		t.Errorf("expected chinese plan content, got %q", plan)
	}
}

func TestExtractFeedback_RecentUserMessage(t *testing.T) {
	history := []Message{
		{Role: "assistant", Content: "doing stuff"},
		{Role: "user", Content: "please fix the bug"},
		{Role: "assistant", Content: "fixing"},
	}
	feedback := extractFeedback(history)
	if feedback != "please fix the bug" {
		t.Errorf("expected user feedback, got %q", feedback)
	}
}

func TestExtractFileStates_Deduplication(t *testing.T) {
	history := []Message{
		{Role: "tool", Content: "file created", Path: "a.go"},
		{Role: "tool", Content: "file modified", Path: "a.go"},
		{Role: "tool", Content: "file deleted", Path: "b.go"},
	}
	states := extractFileStates(history)
	if len(states) != 2 {
		t.Errorf("expected 2 unique file states, got %d", len(states))
	}
	if states["a.go"].Operation != "modified" {
		t.Errorf("expected last operation for a.go to be 'modified', got %q", states["a.go"].Operation)
	}
	if states["b.go"].Operation != "deleted" {
		t.Errorf("expected b.go operation to be 'deleted', got %q", states["b.go"].Operation)
	}
}

func TestMergeRecoveryContexts(t *testing.T) {
	ctx1 := &RecoveryContext{
		ActivePlan: "Plan 1",
		Feedback:   "Feedback 1",
		FileStates: map[string]FileState{
			"a.go": {Path: "a.go", Operation: "created"},
		},
		ExtractedAt: time.Now(),
	}

	ctx2 := &RecoveryContext{
		ActivePlan: "Plan 2",
		Feedback:   "Feedback 2",
		FileStates: map[string]FileState{
			"b.go": {Path: "b.go", Operation: "modified"},
		},
		ExtractedAt: time.Now().Add(time.Minute),
	}

	merged := MergeRecoveryContexts(ctx1, ctx2)

	if merged.ActivePlan != "Plan 2" {
		t.Errorf("ActivePlan: got %q, want %q", merged.ActivePlan, "Plan 2")
	}
	if merged.Feedback != "Feedback 2" {
		t.Errorf("Feedback: got %q, want %q", merged.Feedback, "Feedback 2")
	}
	if len(merged.FileStates) != 2 {
		t.Errorf("FileStates: got %d, want 2", len(merged.FileStates))
	}
}

func TestMergeRecoveryContexts_NilSafe(t *testing.T) {
	merged := MergeRecoveryContexts(nil, nil)
	if merged == nil {
		t.Fatal("expected non-nil merged context")
	}
	if len(merged.FileStates) != 0 {
		t.Errorf("expected empty file states, got %d", len(merged.FileStates))
	}
}

func TestMergeRecoveryContexts_Overlap(t *testing.T) {
	ctx1 := &RecoveryContext{
		ActivePlan: "Plan 1",
		FileStates: map[string]FileState{
			"a.go": {Path: "a.go", Operation: "created"},
		},
	}
	ctx2 := &RecoveryContext{
		ActivePlan: "Plan 2",
		FileStates: map[string]FileState{
			"a.go": {Path: "a.go", Operation: "modified"},
		},
	}

	merged := MergeRecoveryContexts(ctx1, ctx2)
	if merged.ActivePlan != "Plan 2" {
		t.Errorf("ActivePlan: got %q, want %q", merged.ActivePlan, "Plan 2")
	}
	if merged.FileStates["a.go"].Operation != "modified" {
		t.Errorf("FileStates overlap: got %q, want 'modified'", merged.FileStates["a.go"].Operation)
	}
}

func TestBuildRecoveryMessage(t *testing.T) {
	ctx := &RecoveryContext{
		ActivePlan: "fix the login bug",
		Feedback:   "use bcrypt not md5",
		FileStates: map[string]FileState{
			"auth.go": {Path: "auth.go", Operation: "modified"},
		},
	}

	msg := ctx.BuildRecoveryMessage()
	if msg == "" {
		t.Fatal("expected non-empty message")
	}
	if !containsSubstring(msg, "[compact_boundary]") {
		t.Error("message should contain compact_boundary marker")
	}
	if !containsSubstring(msg, "[active_plan]") {
		t.Error("message should contain active_plan marker")
	}
	if !containsSubstring(msg, "[feedback]") {
		t.Error("message should contain feedback marker")
	}
}

func TestBuildRecoveryMessage_Nil(t *testing.T) {
	var ctx *RecoveryContext
	if msg := ctx.BuildRecoveryMessage(); msg != "" {
		t.Errorf("expected empty message for nil context, got %q", msg)
	}
}

func containsSubstring(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && findSubstring(s, sub))
}

func findSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
