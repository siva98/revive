package rule

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"log"

	"github.com/mgechev/revive/lint"
)

type RangeOverMapRule struct{}

// Apply applies the rule to given file.
func (r *RangeOverMapRule) Apply(file *lint.File, _ lint.Arguments) []lint.Failure {
	var failures []lint.Failure

	fileAst := file.AST
	walker := lintRangeOverMap{
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
func (r *RangeOverMapRule) Name() string {
	return "range-over-map"
}

type lintRangeOverMap struct {
	file      *lint.File
	fileAst   *ast.File
	onFailure func(lint.Failure)
}

var rangeOverMapName string

func (w lintRangeOverMap) Visit(node ast.Node) ast.Visitor {
	src := node
	var fset = token.NewFileSet()
	nodes, err := parser.ParseFile(fset, "src.go", src, 0)
	if err != nil {
		log.Fatal(err) // parse error
	}

	conf := types.Config{Importer: importer.Default()}
	info := &types.Info{Types: make(map[ast.Expr]types.TypeAndValue),
		Defs: make(map[*ast.Ident]types.Object)}

	if _, err := conf.Check("cmd/hello", fset, []*ast.File{nodes}, info); err != nil {
		log.Fatal(err) // type error
	}

	rangeOverMapName = ""

	switch n := node.(type) {
	case *ast.RangeStmt:
		rangeOverMapName = types.ExprString(n.X)

		ast.Inspect(node, func(x ast.Node) bool {
			if expr, ok := x.(ast.Expr); ok {
				if tv, ok := info.Types[expr]; ok {
					mapString := tv.Type.String()[0:3]

					if rangeOverMapName == nodeStringRangeOverMap(expr) && n.X.Pos() == expr.Pos() && mapString == "map" {
						fmt.Println("RANGE OVER MAP DETECTED")
					}
				}
			}
			return true
		})
	}
	return w
}

func nodeStringRangeOverMap(n ast.Node) string {
	var fset = token.NewFileSet()
	var buf bytes.Buffer
	format.Node(&buf, fset, n)
	return buf.String()
}
