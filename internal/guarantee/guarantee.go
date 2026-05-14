package guarantee

// Level 表示保证的强度
type Level int

const (
	LevelHard Level = iota // 违反 = 系统 bug，必须修复
	LevelSoft             // 违反 = 降级，记录但不中断
)

// Guarantee 是一个可验证的系统承诺
type Guarantee struct {
	ID          string
	Description string
	Level       Level
	Check       func() error
}

// Violation 记录一次保证违反
type Violation struct {
	GuaranteeID string
	Level       Level
	Error       error
}

// Registry 管理所有已注册的保证
type Registry struct {
	guarantees []Guarantee
}

func NewRegistry() *Registry {
	return &Registry{}
}

func (r *Registry) Register(g Guarantee) {
	r.guarantees = append(r.guarantees, g)
}

func (r *Registry) Verify() []Violation {
	var violations []Violation
	for _, g := range r.guarantees {
		if err := g.Check(); err != nil {
			violations = append(violations, Violation{
				GuaranteeID: g.ID,
				Level:       g.Level,
				Error:       err,
			})
		}
	}
	return violations
}

func (r *Registry) List() []Guarantee {
	return r.guarantees
}
