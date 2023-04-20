package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// mainIdnt defines the "main" literal.
const mainIdnt = "main"

// ExitMainAnalyzer defines an analyzer that checks os.Exit in the main function of the main package.
var ExitMainAnalyzer = &analysis.Analyzer{
	Name: "exitmain",
	Doc:  "detect os.Exit call in the main package, main function",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	expr := func(x *ast.ExprStmt) {
		if call, ok := x.X.(*ast.CallExpr); ok {
			if s, ok := call.Fun.(*ast.SelectorExpr); ok {
				if id, ok := s.X.(*ast.Ident); ok &&
					id.Name == "os" && s.Sel.Name == "Exit" {
					pass.Reportf(call.Pos(), "os.Exit call")
				}
			}
		}
	}

	for _, file := range pass.Files {
		// Validate package name.
		if file.Name.Name != mainIdnt {
			continue
		}
		for _, v := range file.Decls {
			// Get func main.
			if f, ok := v.(*ast.FuncDecl); ok && f.Name.Name == mainIdnt {
				for _, stmt := range f.Body.List {
					if x, ok := stmt.(*ast.ExprStmt); ok {
						expr(x)
					}
				}
			}
		}
	}
	return nil, nil
}
