package rule

import (
	"fmt"
	"go/ast"

	"github.com/mgechev/revive/lint"
)

type GoRoutinesRule struct{}

// Apply applies the rule to given file.
func (r *GoRoutinesRule) Apply(file *lint.File, _ lint.Arguments) []lint.Failure {
	var failures []lint.Failure

	fileAst := file.AST
	walker := lintDivideByZero{
		file:    file,
		fileAst: fileAst,
		onFailure: func(failure lint.Failure) {
			failures = append(failures, failure)
		},
	}

	ast.Walk(walker, fileAst)

	return failures
}

// Name returns the rule name.
func (r *GoRoutinesRule) Name() string {
	return "go-routines"
}

type lintGoRoutines struct {
	file      *lint.File
	fileAst   *ast.File
	onFailure func(lint.Failure)
}

func (w *lintGoRoutines) Visit(node ast.Node) ast.Visitor {
	fmt.Println("1")
	switch n := node.(type) {
	case *ast.GoStmt:
		fmt.Println("2")
		w.onFailure(lint.Failure{
			Confidence: 1,
			Failure:    "should not use goroutines, will lead to non-deterministic behaviour",
			Node:       n,
			Category:   "goroutines",
		})
		//return w
	}
	fmt.Println("3")
	return w
}
