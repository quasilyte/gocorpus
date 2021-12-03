package filters

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

func CompileExpr(s string) (*Expr, Info, error) {
	s = preprocess(s)
	if s == "" {
		return &Expr{Op: OpNop}, Info{}, nil
	}
	root, err := parser.ParseExpr(s)
	if err != nil {
		return nil, Info{}, err
	}

	var cl compiler
	cl.isTopLevel = true
	e, err := cl.CompileExpr(root)
	if err != nil {
		return nil, cl.info, err
	}
	e = cl.Optimize(e)
	return e, cl.info, nil
}

func preprocess(s string) string {
	s = strings.TrimSpace(s)
	return strings.ReplaceAll(s, "$", "__var_")
}

func isPatternVar(s string) bool {
	return strings.HasPrefix(s, "__var_")
}

func patternVarName(s string) string {
	return strings.TrimPrefix(s, "__var_")
}

type compiler struct {
	info Info

	isTopLevel bool
	isNegated  bool
}

func (cl *compiler) Optimize(e *Expr) *Expr {
	for i, arg := range e.Args {
		e.Args[i] = cl.Optimize(arg)
	}
	switch e.Op {
	case OpAnd:
		if e.Args[0].Op == OpNop {
			*e = *e.Args[1]
		} else if e.Args[1].Op == OpNop {
			*e = *e.Args[0]
		}
	case OpNot:
		if e.Args[0].Op == OpNop {
			e.Op = OpNop
			e.Args = nil
		}
	}
	return e
}

func (cl *compiler) CompileExpr(root ast.Expr) (*Expr, error) {
	switch root := root.(type) {
	case *ast.UnaryExpr:
		return cl.compileUnaryExpr(root)
	case *ast.BinaryExpr:
		return cl.compileBinaryExpr(root)
	case *ast.ParenExpr:
		return cl.CompileExpr(root.X)
	case *ast.CallExpr:
		return cl.compileCallExpr(root)
	default:
		return nil, fmt.Errorf("compile expr: unsupported %T", root)
	}
}

func (cl *compiler) compileCallExpr(root *ast.CallExpr) (*Expr, error) {
	if selector, ok := root.Fun.(*ast.SelectorExpr); ok {
		return cl.compileMethodCallExpr(root, selector)
	}
	return nil, fmt.Errorf("compile call expr: unsupported %T function", root.Fun)
}

func (cl *compiler) compileMethodCallExpr(root *ast.CallExpr, selector *ast.SelectorExpr) (*Expr, error) {
	var object string
	ident, ok := selector.X.(*ast.Ident)
	if ok {
		object = ident.Name
	}

	switch object {
	case "file":
		return cl.compileFileMethodCallExpr(root, selector.Sel)
	default:
		if isPatternVar(object) {
			return cl.compilePatternVarMethodCallExpr(root, patternVarName(object), selector.Sel)
		}
		return nil, fmt.Errorf("compile method expr: unsupported %T object", selector.X)
	}
}

func (cl *compiler) compilePatternVarMethodCallExpr(root *ast.CallExpr, varname string, method *ast.Ident) (*Expr, error) {
	switch method.Name {
	case "IsConst":
		return &Expr{Op: OpVarIsConst, Str: varname}, nil
	case "IsPure":
		return &Expr{Op: OpVarIsPure, Str: varname}, nil
	default:
		return nil, fmt.Errorf("compile %s method call: unsupported %s method", varname, method.Name)
	}
}

func (cl *compiler) compileFileMethodCallExpr(root *ast.CallExpr, method *ast.Ident) (*Expr, error) {
	if !cl.isTopLevel {
		return nil, fmt.Errorf("file filters can't be a part of || expression")
	}
	switch method.Name {
	case "IsTest":
		if !cl.info.TestFileCond.IsUnset() {
			return nil, fmt.Errorf("duplicated file.IsTest cond")
		}
		cl.info.TestFileCond.SetValue(!cl.isNegated)
		return &Expr{Op: OpNop}, nil
	case "IsAutogen":
		if !cl.info.AutogenFileCond.IsUnset() {
			return nil, fmt.Errorf("duplicated file.IsAutogen cond")
		}
		cl.info.AutogenFileCond.SetValue(!cl.isNegated)
		return &Expr{Op: OpNop}, nil
	case "IsMain":
		if !cl.info.MainFileCond.IsUnset() {
			return nil, fmt.Errorf("duplicated file.IsMain cond")
		}
		cl.info.MainFileCond.SetValue(!cl.isNegated)
		return &Expr{Op: OpNop}, nil
	default:
		return nil, fmt.Errorf("compile file method call: unsupported %s method", method.Name)
	}
}

func (cl *compiler) compileUnaryExpr(root *ast.UnaryExpr) (*Expr, error) {
	switch root.Op {
	case token.NOT:
		isNegated := cl.isNegated
		cl.isNegated = !isNegated
		x, err := cl.CompileExpr(root.X)
		cl.isNegated = isNegated
		if err != nil {
			return nil, err
		}
		return &Expr{Op: OpNot, Args: []*Expr{x}}, nil
	default:
		return nil, fmt.Errorf("compile unary expr: unsupported %s", root.Op)
	}
}

func (cl *compiler) compileBinaryExpr(root *ast.BinaryExpr) (*Expr, error) {
	switch root.Op {
	case token.LAND:
		lhs, err := cl.CompileExpr(root.X)
		if err != nil {
			return nil, err
		}
		rhs, err := cl.CompileExpr(root.Y)
		if err != nil {
			return nil, err
		}
		return &Expr{Op: OpAnd, Args: []*Expr{lhs, rhs}}, nil

	default:
		return nil, fmt.Errorf("compile binary expr: unsupported %s", root.Op)
	}
}
