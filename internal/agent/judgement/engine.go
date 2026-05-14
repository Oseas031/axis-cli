// Package judgement provides self-judgement capabilities for agent execution validation.
package judgement

import (
	"fmt"
	"log"
	"sync"

	"github.com/axis-cli/axis/internal/agent/judgement/strategies"
	"github.com/axis-cli/axis/internal/model/provider"
)

// Engine is the core self-judgement engine that coordinates all validation strategies.
type Engine struct {
	registry *strategies.StrategyRegistry
	logger   func(string, ...interface{})
	mu       sync.RWMutex
}

// NewEngine creates a new self-judgement engine with default strategies registered.
func NewEngine() *Engine {
	e := &Engine{
		registry: strategies.NewStrategyRegistry(),
		logger:   defaultLogger,
	}

	// Register default strategies
	e.registry.Register(strategies.JudgementTypeSyntax, strategies.NewSyntaxValidationStrategy())
	e.registry.Register(strategies.JudgementTypeTest, strategies.NewTestValidationStrategy())
	e.registry.Register(strategies.JudgementTypeCoverage, strategies.NewCoverageValidationStrategy())

	return e
}

// NewEngineWithProviders creates a new engine with strategies that require external providers.
func NewEngineWithProviders(
	modelProvider provider.ModelProvider,
	contractExecutor strategies.ContractExecutorProvider,
) *Engine {
	e := NewEngine()

	// Register semantic strategy with model provider
	if modelProvider != nil {
		e.registry.Register(strategies.JudgementTypeSemantic, strategies.NewSemanticValidationStrategy(modelProvider))
	}

	// Register contract strategy with contract executor
	if contractExecutor != nil {
		e.registry.Register(strategies.JudgementTypeContract, strategies.NewContractValidationStrategy(contractExecutor))
	}

	return e
}

// LoggerOption sets a custom logger for the engine.
func LoggerOption(logger func(string, ...interface{})) func(*Engine) {
	return func(e *Engine) {
		e.logger = logger
	}
}

// defaultLogger is the default logger implementation.
func defaultLogger(format string, args ...interface{}) {
	log.Printf("[Judgement] "+format, args...)
}

// Judge performs judgement on an execution result using the provided criteria.
// Input is passed through IsolateContext to strip intermediate transcript data
// and prevent Context Rot degradation in judgement accuracy.
func (e *Engine) Judge(
	input any,
	criteria []strategies.JudgementCriteria,
) (*JudgementResult, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Isolate context: extract only final artifacts to prevent Context Rot
	isolated := IsolateContext(input)

	result := NewJudgementResult()

	if len(criteria) == 0 {
		result.Passed = true
		result.Score = 1.0
		result.Confidence = 1.0
		result.Judgements = []strategies.JudgementItem{}
		return result, nil
	}

	for _, c := range criteria {
		if !c.Enabled {
			e.logger("Skipping disabled criteria: %s", c.Name)
			continue
		}

		strategy, ok := e.registry.Get(c.Type)
		if !ok {
			result.Judgements = append(result.Judgements, strategies.JudgementItem{
				CriteriaName: c.Name,
				Passed:       false,
				Score:        0,
				Details:      "",
				Error:        fmt.Sprintf("no strategy registered for type %s", c.Type),
			})
			continue
		}

		if !strategy.CanHandle(c) {
			result.Judgements = append(result.Judgements, strategies.JudgementItem{
				CriteriaName: c.Name,
				Passed:       false,
				Score:        0,
				Details:      "",
				Error:        fmt.Sprintf("strategy cannot handle criteria %s", c.Name),
			})
			continue
		}

		item, err := strategy.Validate(isolated, c)
		if err != nil {
			e.logger("Validation error for %s: %v", c.Name, err)
			item = &strategies.JudgementItem{
				CriteriaName: c.Name,
				Passed:       false,
				Score:        0,
				Details:      "",
				Error:        err.Error(),
			}
		}

		result.AddJudgement(*item)
	}

	return result, nil
}

// RegisterStrategy registers a validation strategy for a judgement type.
func (e *Engine) RegisterStrategy(judgementType strategies.JudgementType, strategy strategies.ValidationStrategy) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.registry.Register(judgementType, strategy)
}

// GetStrategy returns the strategy for a judgement type.
func (e *Engine) GetStrategy(judgementType strategies.JudgementType) (strategies.ValidationStrategy, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.registry.Get(judgementType)
}

// ListStrategies returns all registered judgement types.
func (e *Engine) ListStrategies() []strategies.JudgementType {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.registry.List()
}
