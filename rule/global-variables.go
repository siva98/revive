package rule

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
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

var localVars []string
var globalVar bool

func (w lintGlobalVariables) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.FuncDecl:
		ast.Inspect(n.Body, func(x ast.Node) bool {
			switch x := x.(type) {
			case *ast.ValueSpec:
				for _, name := range x.Names {
					localVars = append(localVars, nodeStringGlobalVars(name))
				}
			}
			return true
		})
	case *ast.ValueSpec:
		for _, name := range n.Names {
			globalVar = true
			for _, localVar := range localVars {
				if nodeStringGlobalVars(name) == localVar {
					globalVar = false
				}
			}
			if globalVar == true {
				w.onFailure(lint.Failure{
					Confidence: 1,
					Failure:    fmt.Sprintf("global variable detected: %s; should not use global variables, will lead to non-deterministic behaviour", name),
					Node:       n,
					Category:   "variables",
				})
				return w
			}
		}
	}

	return w
}

func nodeStringGlobalVars(n ast.Node) string {
	var fset = token.NewFileSet()
	var buf bytes.Buffer
	format.Node(&buf, fset, n)
	return buf.String()
}
