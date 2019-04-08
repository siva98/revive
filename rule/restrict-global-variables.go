package rule

import (
	"fmt"
	"go/ast"

	"github.com/mgechev/revive/lint"
)

type RestrictGlobalVariablesRule struct{}

// Apply applies the rule to given file.
func (r *RestrictGlobalVariablesRule) Apply(file *lint.File, _ lint.Arguments) []lint.Failure {
	var failures []lint.Failure

	fileAst := file.AST
	walker := lintRestrictGlobalVariables{
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
func (r *RestrictGlobalVariablesRule) Name() string {
	return "restrict-global-variables"
}

type lintRestrictGlobalVariables struct {
	file      *lint.File
	fileAst   *ast.File
	onFailure func(lint.Failure)
}

func (w lintRestrictGlobalVariables) Visit(node ast.Node) ast.Visitor {
	globalVar := true
	ast.Inspect(node, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.ValueSpec:
			fmt.Println("Variable name: ")
			for _, name := range n.Names {
				if globalVar == true {
					w.onFailure(lint.Failure{
						Failure:    fmt.Sprintf("global variable found: %s, should not use global variable as they are not tracked on the ledger", name.Name),
						Confidence: 1,
						Node:       n,
						Category:   "variable scope",
					})
				}
			}
		case *ast.FuncDecl:
			globalVar = false
		}
		return true
	})
	return w

}
