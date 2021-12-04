package filters

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
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
	case "IsStringLit":
		return &Expr{Op: OpVarIsStringLit, Str: varname}, nil
	case "IsRuneLit":
		return &Expr{Op: OpVarIsRuneLit, Str: varname}, nil
	case "IsIntLit":
		return &Expr{Op: OpVarIsIntLit, Str: varname}, nil
	case "IsFloatLit":
		return &Expr{Op: OpVarIsFloatLit, Str: varname}, nil
	case "IsComplexLit":
		return &Expr{Op: OpVarIsComplexLit, Str: varname}, nil
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
	if _, ok := root.X.(*ast.BasicLit); ok {
		if _, ok := root.Y.(*ast.BasicLit); !ok {
			switch root.Op {
			case token.GEQ:
				return cl.compileBinaryExprXY(token.LEQ, root.Y, root.X)
			case token.LEQ:
				return cl.compileBinaryExprXY(token.GEQ, root.Y, root.X)
			case token.GTR:
				return cl.compileBinaryExprXY(token.LSS, root.Y, root.X)
			case token.LSS:
				return cl.compileBinaryExprXY(token.GTR, root.Y, root.X)
			case token.EQL:
				return cl.compileBinaryExprXY(token.EQL, root.Y, root.X)
			case token.NEQ:
				return cl.compileBinaryExprXY(token.NEQ, root.Y, root.X)
			}
		}
	}

	return cl.compileBinaryExprXY(root.Op, root.X, root.Y)
}

func (cl *compiler) invertOp(op token.Token) token.Token {
	switch op {
	case token.EQL:
		return token.NEQ
	case token.NEQ:
		return token.EQL
	case token.LSS:
		return token.GEQ
	case token.GTR:
		return token.LEQ
	case token.LEQ:
		return token.GTR
	case token.GEQ:
		return token.LSS

	default:
		return token.ILLEGAL
	}
}

func (cl *compiler) compileBinaryExprXY(op token.Token, x, y ast.Expr) (*Expr, error) {
	fileProp := cl.unpackFileOperand(x)

	switch op {
	case token.LEQ, token.GEQ, token.LSS, token.GTR, token.EQL, token.NEQ:
		if fileProp != "" {
			if !cl.isTopLevel {
				return nil, fmt.Errorf("file filters can't be a part of || expression")
			}
			if cl.isNegated {
				op = cl.invertOp(op)
			}
			rhsValue, ok := cl.toInt(y)
			if op != token.ILLEGAL && ok {
				cl.info.FileMaxDepthOp = op
				cl.info.FileMaxDepth = int(rhsValue)
				return &Expr{Op: OpNop}, nil
			}
		}

	case token.LOR:
		isTopLevel := cl.isTopLevel
		cl.isTopLevel = false
		lhs, err := cl.CompileExpr(x)
		if err != nil {
			return nil, err
		}
		rhs, err := cl.CompileExpr(y)
		if err != nil {
			return nil, err
		}
		cl.isTopLevel = isTopLevel
		return &Expr{Op: OpOr, Args: []*Expr{lhs, rhs}}, nil

	case token.LAND:
		lhs, err := cl.CompileExpr(x)
		if err != nil {
			return nil, err
		}
		rhs, err := cl.CompileExpr(y)
		if err != nil {
			return nil, err
		}
		return &Expr{Op: OpAnd, Args: []*Expr{lhs, rhs}}, nil
	}

	return nil, fmt.Errorf("compile binary expr: unsupported %s", op)
}

func (cl *compiler) toInt(e ast.Expr) (int64, bool) {
	switch e := e.(type) {
	case *ast.BasicLit:
		if e.Kind != token.INT {
			return 0, false
		}
		v, err := strconv.ParseInt(e.Value, 0, 64)
		if err != nil {
			return 0, false
		}
		return int64(v), true

	case *ast.ParenExpr:
		return cl.toInt(e.X)

	default:
		return 0, false
	}
}

func (cl *compiler) unpackFileOperand(e ast.Expr) string {
	call, ok := e.(*ast.CallExpr)
	if !ok {
		return ""
	}
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return ""
	}
	object, ok := selector.X.(*ast.Ident)
	if !ok {
		return ""
	}
	if object.Name != "file" {
		return ""
	}
	return selector.Sel.Name
}
