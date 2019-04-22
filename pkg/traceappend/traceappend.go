package traceappend

// Parent: https://gist.github.com/josephspurrier/19fb8096099bfff5556742072680d061

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"

	"go.uber.org/zap"
	"golang.org/x/tools/go/ast/astutil"
)

var loggertra, _ = zap.NewDevelopment()
var sugarLogger = loggertra.Sugar()

// AddImportToFile will be used to import G, O objects for logging.
func AddImportToFile(file string) ([]byte, error) {
	// Create the AST by parsing src
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, nil, 0)

	// This never fails, because its failure means that a module is already imported.

	astutil.AddImport(fset, f, "runtime")
	astutil.AddImport(fset, f, "net/http")
	astutil.AddImport(fset, f, "bytes")
	// Generate the code
	src, err := generateFile(fset, f)
	if err != nil {
		sugarLogger.Error(err)
		return nil, err
	}

	fo, err := os.OpenFile(file, os.O_WRONLY, 0644)
	if err != nil {
		sugarLogger.Errorf("%v\n", err)
	}

	_, err = fo.Write(src)
	if err != nil {
		sugarLogger.Errorf("%v\n", err)
	}
	// Don't care for any closing errors.
	fo.Close()
	return src, err
}

// GenerateFile creates a new file with the new code appended and returns its contents.
func generateFile(fset *token.FileSet, file *ast.File) ([]byte, error) {
	var output []byte
	buffer := bytes.NewBuffer(output)
	if err := printer.Fprint(buffer, fset, file); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// AppendExpr modifies an AST by adding an expr at the start of its body.
func AppendExpr(file string, patchImports bool) ([]byte, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, nil, 0)
	if err != nil {
		sugarLogger.Errorf("%v\n", err)
	}
	deferNode := createNewNodes()

	count := 0
	fset = token.NewFileSet()

	astutil.Apply(f, func(c *astutil.Cursor) bool {
		_, ok := c.Parent().(*ast.FuncDecl)
		if ok {
			bodyList, ok := c.Node().(*ast.BlockStmt)
			if ok {
				count++
				bodyList.List = append(bodyList.List, deferNode...)
			}
		}
		return true
	}, nil)
	sugarLogger.Infof("Total tracers appended: %d\n", count)
	// Generate the code
	src, err := generateFile(fset, f)
	if err != nil {
		sugarLogger.Error(err)
		return nil, err
	}

	return src, err
}

// createNewNodes creates Append statements.
func createNewNodes() []ast.Stmt {
	expr, err := parser.ParseExpr("func() {_logClusterCodePath();defer _logClusterCodePath();}")

	if err != nil {
		sugarLogger.Errorf("%v\n", err)
	}
	// This cannot error, it's literally hardcoded.
	return expr.(*ast.FuncLit).Body.List
}

func addFuncToSource(appendCode, filepath string) string {
	fset1 := token.NewFileSet()
	fset2 := token.NewFileSet()
	cf1, err := parser.ParseFile(fset1, "code1.go", appendCode, parser.ParseComments)
	if err != nil {
		fmt.Println(err)
	}
	cf2, err := parser.ParseFile(fset2, filepath, nil, parser.ParseComments)
	if err != nil {
		fmt.Println(err)
	}

	cf2.Decls = append(cf2.Decls, cf1.Decls...)
	content, _ := generateFile(fset2, cf2)
	return string(content)
}
