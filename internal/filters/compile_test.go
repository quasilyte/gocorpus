package filters

import (
	"fmt"
	"testing"
)

func TestCompile(t *testing.T) {
	tests := []struct {
		input string
		expr  string
		info  string
	}{
		{
			input: `file.IsTest()`,
			expr:  `Nop`,
			info:  `TestFileCond=true`,
		},
		{
			input: `!file.IsTest()`,
			expr:  `Nop`,
			info:  `TestFileCond=false`,
		},
		{
			input: `!(!file.IsTest())`,
			expr:  `Nop`,
			info:  `TestFileCond=true`,
		},

		{
			input: `file.IsAutogen()`,
			expr:  `Nop`,
			info:  `AutogenFileCond=true`,
		},
		{
			input: `!file.IsAutogen()`,
			expr:  `Nop`,
			info:  `AutogenFileCond=false`,
		},

		{
			input: `file.IsMain()`,
			expr:  `Nop`,
			info:  `MainFileCond=true`,
		},
		{
			input: `!file.IsMain()`,
			expr:  `Nop`,
			info:  `MainFileCond=false`,
		},

		{
			input: `$x.IsPure()`,
			expr:  `(VarIsPure "x")`,
		},
		{
			input: `$x.IsPure() || $y.IsPure()`,
			expr:  `(Or (VarIsPure "x") (VarIsPure "y"))`,
		},

		{
			input: `$x.IsStringLit()`,
			expr:  `(VarIsStringLit "x")`,
		},
		{
			input: `$x.IsRuneLit()`,
			expr:  `(VarIsRuneLit "x")`,
		},
		{
			input: `$x.IsIntLit()`,
			expr:  `(VarIsIntLit "x")`,
		},
		{
			input: `$x.IsFloatLit()`,
			expr:  `(VarIsFloatLit "x")`,
		},
		{
			input: `$x.IsComplexLit()`,
			expr:  `(VarIsComplexLit "x")`,
		},

		{
			input: `!file.IsAutogen() && (!$x.IsPure() || !$y.IsPure())`,
			expr:  `(Or (Not (VarIsPure "x")) (Not (VarIsPure "y")))`,
			info:  `AutogenFileCond=false`,
		},

		{
			input: `(file.IsAutogen()) && !file.IsTest()`,
			expr:  `Nop`,
			info:  `TestFileCond=false AutogenFileCond=true`,
		},

		{
			input: `file.IsTest() && $c.IsConst()`,
			expr:  `(VarIsConst "c")`,
			info:  `TestFileCond=true`,
		},
		{
			input: `$c.IsConst() && !file.IsTest()`,
			expr:  `(VarIsConst "c")`,
			info:  `TestFileCond=false`,
		},
		{
			input: `$x.IsConst() && !($y.IsConst() && file.IsTest())`,
			expr:  `(And (VarIsConst "x") (Not (VarIsConst "y")))`,
			info:  `TestFileCond=false`,
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(fmt.Sprintf("test%d", i), func(t *testing.T) {
			compiled, info, err := CompileExpr(test.input)
			if err != nil {
				t.Fatalf("compile %q: %v", test.input, err)
			}
			if test.info != info.String() {
				t.Fatalf("info mismatch for %q:\nhave: %s\nwant: %s", test.input, info, test.info)
			}
			have := Sprint(compiled)
			if test.expr != have {
				t.Fatalf("result mismatch for %q:\nhave: %s\nwant: %s", test.input, have, test.expr)
			}
		})
	}
}
