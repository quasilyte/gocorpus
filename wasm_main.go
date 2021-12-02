package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"syscall/js"

	"github.com/quasilyte/gocorpus/internal/gogrep"
)

func main() {
	js.Global().Set("gogrep", js.FuncOf(jsGogrep))

	<-make(chan bool)
}

func jsGogrep(this js.Value, args []js.Value) interface{} {
	patString := args[0].String()
	targetName := args[1].String()
	targetSrc := args[2].String()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, targetName, targetSrc, 0)
	if err != nil {
		return map[string]interface{}{"err": err.Error()}
	}
	pat, _, err := gogrep.Compile(fset, patString, false)
	if err != nil {
		return map[string]interface{}{"err": err.Error()}
	}
	var matches []interface{}
	state := gogrep.NewMatcherState()
	ast.Inspect(f, func(n ast.Node) bool {
		pat.MatchNode(&state, n, func(m gogrep.MatchData) {
			begin := fset.Position(m.Node.Pos()).Offset
			end := fset.Position(m.Node.End()).Offset
			matches = append(matches, targetSrc[begin:end])
		})
		return true
	})
	return map[string]interface{}{"matches": matches}
}
