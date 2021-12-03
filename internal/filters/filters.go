package filters

import "strings"

type Info struct {
	TestFileCond    Bool3
	AutogenFileCond Bool3
	MainFileCond    Bool3
}

func (i Info) String() string {
	var parts []string
	if !i.TestFileCond.IsUnset() {
		parts = append(parts, "TestFileCond="+i.TestFileCond.String())
	}
	if !i.AutogenFileCond.IsUnset() {
		parts = append(parts, "AutogenFileCond="+i.AutogenFileCond.String())
	}
	if !i.MainFileCond.IsUnset() {
		parts = append(parts, "MainFileCond="+i.MainFileCond.String())
	}
	return strings.Join(parts, " ")
}

type Expr struct {
	Op   Operation
	Args []*Expr
	Str  string
}

//go:generate stringer -type=Operation -trimprefix=Op
type Operation int

const (
	OpInvalid Operation = iota

	// OpNop = do nothing (should be optimized-away)
	OpNop

	// OpNot = !$Args[0]
	OpNot

	// OpAnd = $Args[0] && $Args[1]
	OpAnd

	// OpOr = $Args[0] || $Args[1]
	OpOr

	// OpVarIsConst = vars[$Str].IsConst()
	OpVarIsConst

	// OpVarIsPure = vars[$Str].IsPure()
	OpVarIsPure

	// OpVarIsStringLit = vars[$Str].IsStringLit()
	OpVarIsStringLit

	// OpVarIsRuneLit = vars[$Str].IsRuneLit()
	OpVarIsRuneLit

	// OpVarIsIntLit = vars[$Str].IsIntLit()
	OpVarIsIntLit

	// OpVarIsFloatLit = vars[$Str].IsFloatLit()
	OpVarIsFloatLit

	// OpVarIsComplexLit = vars[$Str].IsComplexLit()
	OpVarIsComplexLit
)
