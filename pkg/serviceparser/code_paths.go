package serviceparser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"

	"github.com/golang-collections/collections/stack"
	"go.uber.org/zap"
)

// CodePath represents a source code flow.
type CodePath struct {
	From     string `json:"from"`
	To       string `json:"to"`
	PathType string `json:"type"`
}

// Allpaths contains all the identified compile time flows.
var Allpaths map[string][]CodePath

// parseTreePaths extracts all the compile time paths from the ast.
func (cp *CodePath) parseTreePaths(node *ast.Node) CodePath {
	paths := CodePath{From: "", To: "", PathType: "compile"}
	return paths
}

func processCallExpression(expr *ast.CallExpr, fnStack *stack.Stack) {
	// fmt.Printf("%#v\n", expr)
	parseExpressionStmt(&ast.ExprStmt{X: expr.Fun}, fnStack)
}

func processSelectorExpr(expr *ast.SelectorExpr, fnStack *stack.Stack) {
	// fmt.Printf("%#v\n", expr)
	parseExpressionStmt(&ast.ExprStmt{X: expr.X}, fnStack)
	fnStack.Push(expr.Sel.Name)
}

func processIdentifier(expr *ast.Ident, fnStack *stack.Stack) {
	// fmt.Printf("%#v\n", expr.Name)
	fnStack.Push(expr.Name)
}

func parseExpressionStmt(expr *ast.ExprStmt, fnStack *stack.Stack) ([]string, error) {
	switch exp := expr.X.(type) {
	case *ast.CallExpr:
		processCallExpression(exp, fnStack)
	case *ast.SelectorExpr:
		processSelectorExpr(exp, fnStack)
	case *ast.Ident:
		processIdentifier(exp, fnStack)
	}
	return nil, nil
}

func processWrapperFunction(e *ast.FuncDecl) {
	// Save wrapper function name
	f := e.Name.Name
	fmt.Println("Wrapper function name: ", f)

	for _, expression := range e.Body.List {
		ast.Inspect(expression, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.CallExpr:
				fnStack := stack.New()
				processCallExpression(x, fnStack)
				for el := fnStack.Pop(); el != nil; {
					fmt.Printf("%v ", el)
					el = fnStack.Pop()
				}
				fmt.Printf("\n")
			}
			return true
		})
	}
}

func main() {
	logger, _ := zap.NewProduction()
	sugarLogger := logger.Sugar()
	set := token.NewFileSet()
	packs, err := parser.ParseFile(set, "/Users/avgupta/golang/ocp-upgrade-repos/cluster-api-provider-aws/cmd/aws-actuator/main.go", nil, 0)
	if err != nil {
		sugarLogger.Errorf("Failed to parse package: %v", err)
		os.Exit(1)
	}

	ast.Inspect(packs, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			processWrapperFunction(x)
		}
		return true
	})

}
