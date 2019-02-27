package serviceparser

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/golang-collections/collections/stack"
)

// CodePath represents a source code flow.
type CodePath struct {
	From     string `json:"from"`
	To       string `json:"to"`
	PathType string `json:"type"`
	Selector string `json:"selector"`
}

func processCallExpression(expr *ast.CallExpr, fnStack *stack.Stack) {
	parseExpressionStmt(&ast.ExprStmt{X: expr.Fun}, fnStack)
}

func processSelectorExpr(expr *ast.SelectorExpr, fnStack *stack.Stack) {
	parseExpressionStmt(&ast.ExprStmt{X: expr.X}, fnStack)
	processIdentifier(expr.Sel, fnStack)
}

func processIdentifier(expr *ast.Ident, fnStack *stack.Stack) {
	//suf := ""
	//if expr.Obj != nil && expr.Obj.Kind == 5 {
	//	suf = "()"
	//}
	idN := expr.Name
	//if suf != "" {
	//	idN = idN + suf
	//}
	fnStack.Push(idN)
}

func parseExpressionStmt(expr *ast.ExprStmt, fnStack *stack.Stack) {
	switch exp := expr.X.(type) {
	case *ast.CallExpr:
		processCallExpression(exp, fnStack)
	case *ast.SelectorExpr:
		processSelectorExpr(exp, fnStack)
	case *ast.Ident:
		processIdentifier(exp, fnStack)
	}
}

func processWrapperFunction(e *ast.FuncDecl, allCompilePaths *[]CodePath) {
	// Save wrapper function name
	f := e.Name.Name
	fmt.Println("Wrapper function name: ", f)

	for _, expression := range e.Body.List {
		ast.Inspect(expression, func(n ast.Node) bool {
			var compilePaths = *allCompilePaths
			switch x := n.(type) {
			case *ast.CallExpr:
				fnStack := stack.New()
				processCallExpression(x, fnStack)
				fn, _ := fnStack.Pop().(string)
				sel := ""
				for el := fnStack.Pop(); el != nil; {
					selstr, _ := el.(string)
					sel = sel + string(selstr) + " "
					el = fnStack.Pop()
				}
				compilePaths = append(compilePaths, CodePath{From: f, To: fn, PathType: "compile", Selector: strings.TrimSpace(sel)})
				fmt.Printf("\n")
			}
			*allCompilePaths = compilePaths
			return true
		})
	}
}

// ParseTreePaths extracts all the compile time paths from the ast.
func ParseTreePaths(root ast.Node) []CodePath {
	var allCompilePaths []CodePath

	ast.Inspect(root, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			processWrapperFunction(x, &allCompilePaths)
		}
		return true
	})

}
