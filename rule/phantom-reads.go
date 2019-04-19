package rule

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/token"
	"strings"

	"github.com/mgechev/revive/lint"
)

type PhantomReadsRule struct{}

// Apply applies the rule to given file.
func (r *PhantomReadsRule) Apply(file *lint.File, _ lint.Arguments) []lint.Failure {
	var failures []lint.Failure

	fileAst := file.AST
	walker := lintPhantomReads{
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
func (r *PhantomReadsRule) Name() string {
	return "phantom-reads"
}

type lintPhantomReads struct {
	file      *lint.File
	fileAst   *ast.File
	onFailure func(lint.Failure)
}

var writeQueryOrKey string
var getQueryOrKey string

func (w lintPhantomReads) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.FuncDecl:
		writeKey = ""
		getQueryOrKey = ""

    ast.Inspect(n.Body, func(x ast.Node) bool {
      switch x := x.(type) {
      case *ast.CallExpr:
        callExpr := nodeStringPhantomReads(x.Fun)

        if strings.Contains(callExpr, ".") {
          putState := strings.Split(callExpr, ".")

          if putState[1] == "GetHistoryForKey" || putState[1] == "GetQueryResult" {
            getQueryOrKey = nodeStringPhantomReads(x.Args[0])

            ast.Inspect(n.Body, func(y ast.Node) bool {
              switch y := y.(type) {
              case *ast.CallExpr:
                callExpr = nodeStringPhantomReads(y.Fun)

                if strings.Contains(callExpr, ".") {
                  putState := strings.Split(callExpr, ".")

                  if putState[1] == "PutState" {
                    writeQueryOrKey = nodeStringPhantomReads(y.Args[0])

                    if y.Pos() > x.Pos() && writeQueryOrKey == getQueryOrKey {
                      w.onFailure(lint.Failure{
												Confidence: 1,
												Failure:    "data obtained from phantom reads should not be used to write new data or update data on the ledger: write",
												Node:       x,
												Category:   "control flow",
											})
											w.onFailure(lint.Failure{
												Confidence: 1,
												Failure:    "data obtained from phantom reads should not be used to write new data or update data on the ledger: read",
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

func nodeStringPhantomReads(n ast.Node) string {
	var fset = token.NewFileSet()
	var buf bytes.Buffer
	format.Node(&buf, fset, n)
	return buf.String()
}
