package rule

import (
	"go/ast"
	"go/token"

	"github.com/mgechev/revive/lint"
)

type GlobalVariablesRule struct{}

// Apply applies the rule to given file.
func (r *GlobalVariablesRule) Apply(file *lint.File, _ lint.Arguments) []lint.Failure {
	var failures []lint.Failure

	fileAst := file.AST
	walker := lintGlobalVariables{
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
func (r *GlobalVariablesRule) Name() string {
	return "global-variables"
}

type lintGlobalVariables struct {
	file      *lint.File
	fileAst   *ast.File
	onFailure func(lint.Failure)
}

func (w lintGlobalVariables) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.GenDecl:
		if n.Tok == token.VAR {
			ast.Inspect(node, func(x ast.Node) bool {
				switch x := x.(type) {
				case *ast.FuncDecl:
					if n.TokPos < x.Body.Lbrace || n.TokPos > x.Body.Rbrace {
						w.onFailure(lint.Failure{
							Confidence: 1,
							Failure:    "should not use global variables, will lead to non-deterministic behaviour",
							Node:       n,
							Category:   "variables",
						})
					}
				}
				return true
			})

		}
	}
	return w
}
