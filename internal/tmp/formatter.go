package tmp

import "strings"

type Formatter interface {
	Format(s string) string
}

type UpperFormatter struct{}
type TrimFormatter struct{}

func (f UpperFormatter) Format(s string) string {
	return strings.ToUpper(s)
}

func (f TrimFormatter) Format(s string) string {
	return strings.TrimSpace(s)
}

// ChainFormat applies all formatters in order.
func ChainFormat(s string, formatters ...Formatter) string {
	for _, formatter := range formatters {
		s = formatter.Format(s)
	}
	return s
}
