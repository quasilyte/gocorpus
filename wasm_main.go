package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"syscall/js"

	"github.com/quasilyte/gocorpus/internal/filebits"
	"github.com/quasilyte/gocorpus/internal/filters"
	"github.com/quasilyte/gocorpus/internal/gogrep"
)

func main() {
	js.Global().Set("gogrep", js.FuncOf(jsGogrep))

	<-make(chan bool)
}

func isPureExpr(expr ast.Expr) bool {
	// This list switch is not comprehensive and uses
	// whitelist to be on the conservative side.
	// Can be extended as needed.

	if expr == nil {
		return true
	}

	switch expr := expr.(type) {
	case *ast.StarExpr:
		return isPureExpr(expr.X)
	case *ast.BinaryExpr:
		return isPureExpr(expr.X) &&
			isPureExpr(expr.Y)
	case *ast.UnaryExpr:
		return expr.Op != token.ARROW &&
			isPureExpr(expr.X)
	case *ast.BasicLit, *ast.Ident:
		return true
	case *ast.SliceExpr:
		return isPureExpr(expr.X) &&
			isPureExpr(expr.Low) &&
			isPureExpr(expr.High) &&
			isPureExpr(expr.Max)
	case *ast.IndexExpr:
		return isPureExpr(expr.X) &&
			isPureExpr(expr.Index)
	case *ast.SelectorExpr:
		return isPureExpr(expr.X)
	case *ast.ParenExpr:
		return isPureExpr(expr.X)
	case *ast.TypeAssertExpr:
		return isPureExpr(expr.X)
	case *ast.CompositeLit:
		return isPureExprList(expr.Elts)

	case *ast.CallExpr:
		ident, ok := expr.Fun.(*ast.Ident)
		if !ok {
			return false
		}
		switch ident.Name {
		case "len", "cap", "real", "imag":
			return isPureExprList(expr.Args)
		default:
			return false
		}

	default:
		return false
	}
}

func isPureExprList(list []ast.Expr) bool {
	for _, expr := range list {
		if !isPureExpr(expr) {
			return false
		}
	}
	return true
}

func isConstExpr(e ast.Expr) bool {
	switch e := e.(type) {
	case *ast.BasicLit:
		return true
	case *ast.UnaryExpr:
		return isConstExpr(e.X)
	case *ast.BinaryExpr:
		return isConstExpr(e.X) && isConstExpr(e.Y)
	default:
		return false
	}
}

var badExpr = &ast.BadExpr{}

func getMatchExpr(m gogrep.MatchData, name string) ast.Expr {
	n, ok := m.CapturedByName(name)
	if !ok {
		return badExpr
	}
	e, ok := n.(ast.Expr)
	if !ok {
		return badExpr
	}
	return e
}

func checkBasicLit(n ast.Expr, kind token.Token) bool {
	if lit, ok := n.(*ast.BasicLit); ok {
		return lit.Kind == kind
	}
	return false
}

func applyFilter(f *filters.Expr, n ast.Node, m gogrep.MatchData) bool {
	switch f.Op {
	case filters.OpNot:
		return !applyFilter(f.Args[0], n, m)

	case filters.OpAnd:
		return applyFilter(f.Args[0], n, m) && applyFilter(f.Args[1], n, m)

	case filters.OpOr:
		return applyFilter(f.Args[0], n, m) || applyFilter(f.Args[1], n, m)

	case filters.OpVarIsConst:
		v, ok := m.CapturedByName(f.Str)
		if !ok {
			return false
		}
		if e, ok := v.(ast.Expr); ok {
			return isConstExpr(e)
		}
		return false

	case filters.OpVarIsStringLit:
		return checkBasicLit(getMatchExpr(m, f.Str), token.STRING)
	case filters.OpVarIsRuneLit:
		return checkBasicLit(getMatchExpr(m, f.Str), token.CHAR)
	case filters.OpVarIsIntLit:
		return checkBasicLit(getMatchExpr(m, f.Str), token.INT)
	case filters.OpVarIsFloatLit:
		return checkBasicLit(getMatchExpr(m, f.Str), token.FLOAT)
	case filters.OpVarIsComplexLit:
		return checkBasicLit(getMatchExpr(m, f.Str), token.IMAG)

	case filters.OpVarIsPure:
		v, ok := m.CapturedByName(f.Str)
		if !ok {
			return false
		}
		if e, ok := v.(ast.Expr); ok {
			return isPureExpr(e)
		}
		return false

	default:
		fmt.Fprintf(os.Stderr, "can't handle %s\n", filters.Sprint(f))
	}

	return true
}

func checkFileDepth(op token.Token, fileDepth, limit int) bool {
	if op == token.ILLEGAL {
		return true
	}
	switch op {
	case token.EQL:
		return fileDepth == limit
	case token.NEQ:
		return fileDepth != limit
	case token.LSS:
		return fileDepth < limit
	case token.GTR:
		return fileDepth > limit
	case token.LEQ:
		return fileDepth <= limit
	case token.GEQ:
		return fileDepth >= limit

	default:
		return true
	}
}

func canSkipFile(cond filters.Bool3, flags, mask int) bool {
	if cond.IsTrue() && !filebits.Check(flags, mask) {
		return true
	}
	if cond.IsFalse() && filebits.Check(flags, mask) {
		return true
	}
	return false
}

var skipFileResult = map[string]interface{}{
	"matches": []interface{}{},
	"skipped": true,
}

func jsGogrep(this js.Value, args []js.Value) interface{} {
	argsObject := args[0]
	patString := argsObject.Get("pattern").String()
	filterString := argsObject.Get("filter").String()
	fileFlags := argsObject.Get("fileFlags").Int()
	fileMaxDepth := argsObject.Get("fileMaxDepth").Int()
	targetName := argsObject.Get("targetName").String()
	targetSrc := argsObject.Get("targetSrc").String()

	filterExpr, filterInfo, err := filters.CompileExpr(filterString)
	if err != nil {
		return map[string]interface{}{"err": "filter: " + err.Error()}
	}

	// Check whether we can skip this file without parsing it.
	if !checkFileDepth(filterInfo.FileMaxDepthOp, fileMaxDepth, filterInfo.FileMaxDepth) {
		return skipFileResult
	}
	if canSkipFile(filterInfo.TestFileCond, fileFlags, filebits.IsTest) {
		return skipFileResult
	}
	if canSkipFile(filterInfo.MainFileCond, fileFlags, filebits.IsMain) {
		return skipFileResult
	}
	if canSkipFile(filterInfo.AutogenFileCond, fileFlags, filebits.IsAutogen) {
		return skipFileResult
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, targetName, targetSrc, 0)
	if err != nil {
		return map[string]interface{}{"err": "parse Go: " + err.Error()}
	}
	pat, _, err := gogrep.Compile(fset, patString, false)
	if err != nil {
		return map[string]interface{}{"err": "parse pattern: " + err.Error()}
	}

	var matches []interface{}
	state := gogrep.NewMatcherState()
	ast.Inspect(f, func(n ast.Node) bool {
		pat.MatchNode(&state, n, func(m gogrep.MatchData) {
			if filterExpr.Op == filters.OpNop || applyFilter(filterExpr, m.Node, m) {
				begin := fset.Position(m.Node.Pos()).Offset
				end := fset.Position(m.Node.End()).Offset
				matches = append(matches, targetSrc[begin:end])
			}
		})
		return true
	})
	return map[string]interface{}{
		"matches": matches,
		"skipped": false,
	}
}
