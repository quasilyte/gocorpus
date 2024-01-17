package main

import (
	"bytes"
	"go/ast"
	"go/token"

	"github.com/go-toolsmith/minformat"
)

func minifyGo(fset *token.FileSet, f *ast.File) []byte {
	var buf bytes.Buffer
	if err := minformat.Node(&buf, fset, f); err != nil {
		panic(err)
	}
	return buf.Bytes()
}
