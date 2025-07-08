// Пакет osexit запрещает использовать прямой вызов os.Exit в функции main пакета main.
package osexit

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// Analyzer запрещает использовать прямой вызов os.Exit в функции main пакета main.
var Analyzer = &analysis.Analyzer{
	Name: "osexit",
	Doc:  "it is forbidden to use os.Exit directly in main.main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for i := range pass.Files {
		if pass.Pkg.Name() != "main" {
			continue
		}

		ast.Inspect(pass.Files[i], func(n ast.Node) bool {
			callExpr, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			ident, ok := selExpr.X.(*ast.Ident)
			if !ok {
				return true
			}

			if ident.Name == "os" && selExpr.Sel.Name == "Exit" {
				for _, decl := range pass.Files[i].Decls {
					if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == "main" {
						pass.Reportf(callExpr.Lparen, "it is forbidden to use os.Exit directly in main.main")
					}
				}
			}

			return true
		})
	}

	//nolint:nilnil // Skipping the check for analyzer
	return nil, nil
}
