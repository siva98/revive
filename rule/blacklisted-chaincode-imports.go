package rule

import (
	"fmt"
	"go/ast"

	"github.com/mgechev/revive/lint"
)

type BlacklistedChaincodeImportsRule struct{}

// Apply applies the rule to given file.
func (r *BlacklistedChaincodeImportsRule) Apply(file *lint.File, _ lint.Arguments) []lint.Failure {
	var failures []lint.Failure

	fileAst := file.AST
	walker := lintBlacklistedChaincodeImports{
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
func (r *BlacklistedChaincodeImportsRule) Name() string {
	return "blacklisted-chaincode-imports"
}

type lintBlacklistedChaincodeImports struct {
	file      *lint.File
	fileAst   *ast.File
	onFailure func(lint.Failure)
}

var blacklistedImports = []string{"time"}

func (w lintBlacklistedChaincodeImports) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.ImportSpec:
		for _, blacklistedImport := range blacklistedImports {
			blacklistedImport = "\"" + blacklistedImport + "\""
			if n.Path.Value == blacklistedImport {
				w.onFailure(lint.Failure{
					Confidence: 1,
					Failure:    fmt.Sprintf("should not use the following blacklisted import: %s", n.Path.Value),
					Node:       n,
					Category:   "imports",
				})
				return w
			}
		}
	}
	return w

}
