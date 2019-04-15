package rule

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/token"
	"strings"

	"github.com/mgechev/revive/lint"
)

type ReadAfterWriteRule struct{}

// Apply applies the rule to given file.
func (r *ReadAfterWriteRule) Apply(file *lint.File, _ lint.Arguments) []lint.Failure {
	var failures []lint.Failure

	fileAst := file.AST
	walker := lintReadAfterWrite{
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
func (r *ReadAfterWriteRule) Name() string {
	return "read-after-write"
}

type lintReadAfterWrite struct {
	file      *lint.File
	fileAst   *ast.File
	onFailure func(lint.Failure)
}

var writeKey string
var readKey string

func (w lintReadAfterWrite) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.FuncDecl:
		writeKey = ""
		readKey = ""

		ast.Inspect(n.Body, func(x ast.Node) bool {
			switch x := x.(type) {
			case *ast.CallExpr:
				callExpr := nodeString(x.Fun)

				if strings.Contains(callExpr, ".") {
					putState := strings.Split(callExpr, ".")

					if putState[1] == "PutState" {
						writeKey = nodeString(x.Args[0])

						ast.Inspect(n.Body, func(y ast.Node) bool {
							switch y := y.(type) {
							case *ast.CallExpr:
								callExpr = nodeString(y.Fun)

								if strings.Contains(callExpr, ".") {
									putState := strings.Split(callExpr, ".")

									if putState[1] == "GetState" {
										readKey = nodeString(y.Args[0])

										if y.Pos() > x.Pos() && readKey == writeKey {
											w.onFailure(lint.Failure{
												Confidence: 1,
												Failure:    "should not read after write: write",
												Node:       x,
												Category:   "control flow",
											})
											w.onFailure(lint.Failure{
												Confidence: 1,
												Failure:    "should not read after write: read",
												Node:       y,
												Category:   "control flow",
											})

											return true
										}
									}
								}
							}
							return true
						})
					}
				}
			}
			return true
		})

	}
	return w
}

func nodeString(n ast.Node) string {
	var fset = token.NewFileSet()
	var buf bytes.Buffer
	format.Node(&buf, fset, n)
	return buf.String()
}
