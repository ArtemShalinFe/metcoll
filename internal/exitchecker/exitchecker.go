// Package exitchecker checks a direct call to `os.Exit` in the main function of the main package
package exitchecker

import (
	"flag"
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var flagSet flag.FlagSet

func NewAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:  "exitchecker",
		Doc:   "checks a direct call to `os.Exit` in the main function of the main package",
		Run:   run,
		Flags: flagSet,
	}
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, f := range pass.Files {
		ast.Inspect(f, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.FuncDecl:
				if x.Name.Name == "main" {
					return true
				} else {
					return false
				}
			case *ast.SelectorExpr:
				i, ok := x.X.(*ast.Ident)
				if !ok {
					return false
				}
				osCall := i.Name == "os"
				exitCall := x.Sel.Name == "Exit"

				if osCall && exitCall {
					pass.Reportf(x.Pos(), "called os.Exit in the main function")
					return false
				}
				return false
			default:
				return true
			}
		})
	}
	return nil, nil
}
