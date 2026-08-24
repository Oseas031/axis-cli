package contextpack

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// RuleEngine 配置化规则引擎
type RuleEngine struct {
	mu         sync.RWMutex
	rules      []Rule
	configPath string
}

// Rule 规则定义
type Rule struct {
	ID          string   `json:"id"`
	Keywords    []string `json:"keywords"`
	Relevance   float64  `json:"relevance"`
	Context     string   `json:"context"`
	Description string   `json:"description"`
}

// NewRuleEngine 创建规则引擎
func NewRuleEngine(configPath string) *RuleEngine {
	engine := &RuleEngine{
		configPath: configPath,
	}
	engine.Load()
	return engine
}

// Load 加载规则配置
func (e *RuleEngine) Load() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 读取配置文件
	data, err := os.ReadFile(e.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// 使用默认规则
			e.rules = defaultRules()
			return nil
		}
		return err
	}

	var config struct {
		Version string `json:"version"`
		Rules   []Rule `json:"rules"`
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	e.rules = config.Rules
	return nil
}

// Match 匹配文本，返回匹配的规则
func (e *RuleEngine) Match(text string) []Rule {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var matched []Rule
	text = strings.ToLower(text)

	for _, rule := range e.rules {
		for _, keyword := range rule.Keywords {
			if strings.Contains(text, strings.ToLower(keyword)) {
				matched = append(matched, rule)
				break
			}
		}
	}

	return matched
}

// GetRule 根据ID获取规则
func (e *RuleEngine) GetRule(id string) *Rule {
	e.mu.RLock()
	defer e.mu.RUnlock()

	for _, rule := range e.rules {
		if rule.ID == id {
			return &rule
		}
	}
	return nil
}

// GetAllRules 获取所有规则
func (e *RuleEngine) GetAllRules() []Rule {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make([]Rule, len(e.rules))
	copy(result, e.rules)
	return result
}

// AddRule 添加规则
func (e *RuleEngine) AddRule(rule Rule) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.rules = append(e.rules, rule)
}

// RemoveRule 移除规则
func (e *RuleEngine) RemoveRule(id string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	for i, rule := range e.rules {
		if rule.ID == id {
			e.rules = append(e.rules[:i], e.rules[i+1:]...)
			break
		}
	}
}

// Save 保存规则到配置文件
func (e *RuleEngine) Save() error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	config := struct {
		Version string `json:"version"`
		Rules   []Rule `json:"rules"`
	}{
		Version: "1.0",
		Rules:   e.rules,
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	// 确保目录存在
	dir := filepath.Dir(e.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(e.configPath, data, 0644)
}

// defaultRules 默认规则
func defaultRules() []Rule {
	return []Rule{
		{
			ID:          "natural-language-scheduling",
			Keywords:    []string{"ask", "natural language", "prompt", "intent", "scheduling", "自然语言", "意图"},
			Relevance:   0.92,
			Context:     "nl-scheduling-spec",
			Description: "Natural language scheduling specification",
		},
		{
			ID:          "model-provider",
			Keywords:    []string{"provider", "model", "deepseek", "minimax", "openai", "api key", "profile", "模型", "提供者"},
			Relevance:   0.90,
			Context:     "model-provider-spec",
			Description: "Model provider configuration",
		},
		{
			ID:          "adaptive-context-assembly",
			Keywords:    []string{"context", "assembly", "bundle", "packet", "adaptive", "上下文", "装配"},
			Relevance:   0.95,
			Context:     "adaptive-context-spec",
			Description: "Adaptive context assembly specification",
		},
		{
			ID:          "vigil-tracking",
			Keywords:    []string{"vigil", "todo", "task", "追踪", "进度", "work tracking"},
			Relevance:   0.88,
			Context:     "vigil-spec",
			Description: "Vigil work tracking system",
		},
		{
			ID:          "memory-system",
			Keywords:    []string{"memory", "记忆", "recall", "horizon", "working memory"},
			Relevance:   0.90,
			Context:     "memory-spec",
			Description: "Memory system specification",
		},
	}
}
