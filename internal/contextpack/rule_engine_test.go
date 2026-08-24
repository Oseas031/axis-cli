package contextpack

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRuleEngine_Match(t *testing.T) {
	// 创建临时配置文件
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "rules.json")

	config := `{
        "version": "1.0",
        "rules": [
            {
                "id": "test-rule",
                "keywords": ["test", "测试"],
                "relevance": 0.9,
                "context": "test-context",
                "description": "Test rule"
            }
        ]
    }`

	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		t.Fatal(err)
	}

	engine := NewRuleEngine(configPath)

	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{"match english", "this is a test", 1},
		{"match chinese", "这是一个测试", 1},
		{"no match", "hello world", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched := engine.Match(tt.text)
			if len(matched) != tt.expected {
				t.Errorf("Match(%q) = %d rules, want %d", tt.text, len(matched), tt.expected)
			}
		})
	}
}

func TestRuleEngine_AddRemove(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "rules.json")

	// 创建空配置文件
	emptyConfig := `{
        "version": "1.0",
        "rules": []
    }`
	if err := os.WriteFile(configPath, []byte(emptyConfig), 0644); err != nil {
		t.Fatal(err)
	}

	engine := NewRuleEngine(configPath)

	// 添加规则
	engine.AddRule(Rule{
		ID:          "new-rule",
		Keywords:    []string{"new"},
		Relevance:   0.8,
		Context:     "new-context",
		Description: "New rule",
	})

	if len(engine.GetAllRules()) != 1 {
		t.Errorf("AddRule: got %d rules, want 1", len(engine.GetAllRules()))
	}

	// 移除规则
	engine.RemoveRule("new-rule")

	if len(engine.GetAllRules()) != 0 {
		t.Errorf("RemoveRule: got %d rules, want 0", len(engine.GetAllRules()))
	}
}

func TestRuleEngine_GetRule(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "rules.json")

	config := `{
        "version": "1.0",
        "rules": [
            {
                "id": "existing-rule",
                "keywords": ["existing"],
                "relevance": 0.7,
                "context": "existing-context",
                "description": "Existing rule"
            }
        ]
    }`

	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		t.Fatal(err)
	}

	engine := NewRuleEngine(configPath)

	// 获取存在的规则
	rule := engine.GetRule("existing-rule")
	if rule == nil {
		t.Fatal("GetRule(existing-rule) returned nil")
	}
	if rule.ID != "existing-rule" {
		t.Errorf("GetRule(existing-rule).ID = %q, want %q", rule.ID, "existing-rule")
	}

	// 获取不存在的规则
	rule = engine.GetRule("nonexistent")
	if rule != nil {
		t.Errorf("GetRule(nonexistent) = %+v, want nil", rule)
	}
}

func TestRuleEngine_Save(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "rules.json")

	// 创建空配置文件
	emptyConfig := `{
        "version": "1.0",
        "rules": []
    }`
	if err := os.WriteFile(configPath, []byte(emptyConfig), 0644); err != nil {
		t.Fatal(err)
	}

	engine := NewRuleEngine(configPath)

	// 添加规则
	engine.AddRule(Rule{
		ID:          "save-test",
		Keywords:    []string{"save"},
		Relevance:   0.6,
		Context:     "save-context",
		Description: "Save test rule",
	})

	// 保存
	if err := engine.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// 重新加载并验证
	engine2 := NewRuleEngine(configPath)
	rules := engine2.GetAllRules()
	if len(rules) != 1 {
		t.Fatalf("After Save+Load: got %d rules, want 1", len(rules))
	}
	if rules[0].ID != "save-test" {
		t.Errorf("After Save+Load: rule.ID = %q, want %q", rules[0].ID, "save-test")
	}
}

func TestRuleEngine_LoadDefault(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nonexistent-rules.json")

	engine := NewRuleEngine(configPath)

	// 应该使用默认规则
	rules := engine.GetAllRules()
	if len(rules) == 0 {
		t.Error("LoadDefault: expected default rules, got none")
	}
}

func TestRuleEngine_MatchMultiple(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "rules.json")

	config := `{
        "version": "1.0",
        "rules": [
            {
                "id": "rule-a",
                "keywords": ["alpha"],
                "relevance": 0.9,
                "context": "context-a",
                "description": "Rule A"
            },
            {
                "id": "rule-b",
                "keywords": ["beta"],
                "relevance": 0.8,
                "context": "context-b",
                "description": "Rule B"
            }
        ]
    }`

	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		t.Fatal(err)
	}

	engine := NewRuleEngine(configPath)

	// 匹配两个规则
	matched := engine.Match("alpha and beta")
	if len(matched) != 2 {
		t.Errorf("Match('alpha and beta') = %d rules, want 2", len(matched))
	}

	// 匹配一个规则
	matched = engine.Match("only alpha here")
	if len(matched) != 1 {
		t.Errorf("Match('only alpha here') = %d rules, want 1", len(matched))
	}
}

func TestRuleEngine_CaseInsensitive(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "rules.json")

	config := `{
        "version": "1.0",
        "rules": [
            {
                "id": "case-test",
                "keywords": ["CaseInsensitive"],
                "relevance": 0.9,
                "context": "case-context",
                "description": "Case insensitive test"
            }
        ]
    }`

	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		t.Fatal(err)
	}

	engine := NewRuleEngine(configPath)

	// 测试大小写不敏感
	matched := engine.Match("this is CASEINSENSITIVE test")
	if len(matched) != 1 {
		t.Errorf("Match('this is CASEINSENSITIVE test') = %d rules, want 1", len(matched))
	}

	matched = engine.Match("this is caseinsensitive test")
	if len(matched) != 1 {
		t.Errorf("Match('this is caseinsensitive test') = %d rules, want 1", len(matched))
	}
}
