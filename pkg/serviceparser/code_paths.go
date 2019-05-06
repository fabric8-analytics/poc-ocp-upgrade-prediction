package serviceparser

import (
	"go/ast"
	"strings"

	"github.com/golang-collections/collections/stack"
)

// CodePath represents a source code flow.
type CodePath struct {
	From                   string            `json:"from"`
	To                     string            `json:"to"`
	PathType               string            `json:"type"`
	SelectorCallee         string            `json:"selector_callee"`
	ContainerPackage       string            `json:"container_package"`
	ContainerPackageCaller string            `json:"container_package_caller"`
	PathAttrs              map[string]string `json:"path_attrs"`
}

func processCallExpression(expr *ast.CallExpr, fnStack *stack.Stack) {
	parseExpressionStmt(&ast.ExprStmt{X: expr.Fun}, fnStack)
}

func processSelectorExpr(expr *ast.SelectorExpr, fnStack *stack.Stack) {
	parseExpressionStmt(&ast.ExprStmt{X: expr.X}, fnStack)
	processIdentifier(expr.Sel, fnStack)
}

func processIdentifier(expr *ast.Ident, fnStack *stack.Stack) {
	idN := expr.Name
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

func processWrapperFunction(e *ast.FuncDecl, allCompilePaths *[]CodePath, pkg string) {
	// Save wrapper function name
	f := e.Name.Name
	sugarLogger.Info("Wrapper function name: ", f)

	for _, expression := range e.Body.List {
		ast.Inspect(expression, func(n ast.Node) bool {
			var compilePaths = *allCompilePaths
			switch x := n.(type) {
			case *ast.CallExpr:
				fnStack := stack.New()
				processCallExpression(x, fnStack)
				fn, _ := fnStack.Pop().(string)
				if fn == "" {
					// Do not add empty string!
					break
				}
				sel := ""
				for el := fnStack.Pop(); el != nil; {
					selstr, _ := el.(string)
					sel = sel + strings.Trim(string(selstr), ". ") + ","
					el = fnStack.Pop()
				}
				// Remove the last comma
				sel = strings.TrimRight(sel, ",")
				// If one of the builtins then ignore.
				if _, ok := Builtins[fn]; ok {
					break
				}
				// The caller will never have a selector, since it's one of the functions defined in this service.
				compilePaths = append(compilePaths, CodePath{From: f, To: strings.Trim(fn, ". "), PathType: "compile", SelectorCallee: sel, ContainerPackage: pkg})
			}
			*allCompilePaths = compilePaths
			return true
		})
	}
}

// ParseTreePaths extracts all the compile time paths from the ast.
func ParseTreePaths(pkg string, root ast.Node) []CodePath {
	var allCompilePaths []CodePath

	ast.Inspect(root, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			processWrapperFunction(x, &allCompilePaths, pkg)
		}
		return true
	})
	return allCompilePaths
}

// A map of all the builtins of the go programming language.
var Builtins = map[string]bool{
	"append":      true,
	"cap":         true,
	"close":       true,
	"complex":     true,
	"copy":        true,
	"delete":      true,
	"imag":        true,
	"len":         true,
	"make":        true,
	"new":         true,
	"panic":       true,
	"print":       true,
	"println":     true,
	"real":        true,
	"recover":     true,
	"ComplexType": true,
	"FloatType":   true,
	"IntegerType": true,
	"Type":        true,
	"Type1":       true,
	"bool":        true,
	"byte":        true,
	"complex128":  true,
	"complex64":   true,
	"error":       true,
	"float32":     true,
	"float64":     true,
	"int":         true,
	"int16":       true,
	"int32":       true,
	"int64":       true,
	"int8":        true,
	"rune":        true,
	"string":      true,
	"uint":        true,
	"uint16":      true,
	"uint32":      true,
	"uint64":      true,
	"uint8":       true,
	"uintptr":     true,
}
