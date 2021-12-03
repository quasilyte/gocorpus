package filters

import (
	"fmt"
	"strings"
)

func Sprint(e *Expr) string {
	parts := make([]string, 0, len(e.Args))
	if e.Str != "" {
		parts = append(parts, fmt.Sprintf("%q", e.Str))
	}
	for _, arg := range e.Args {
		parts = append(parts, Sprint(arg))
	}
	if len(parts) == 0 {
		return e.Op.String()
	}
	return "(" + e.Op.String() + " " + strings.Join(parts, " ") + ")"
}
