package traceappend

// Parent: https://gist.github.com/josephspurrier/19fb8096099bfff5556742072680d061

import (
	"bytes"
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
	_ = astutil.AddImport(fset, f, "github.com/rootAvish/go-tracey")

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
	_ = fo.Close()
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

// AppendExpr modifies an AST by adding an expr at the start of its body. Also adds the tracey decl to its genDecls.
func AppendExpr(file string, patchImports bool) ([]byte, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, nil, 0)
	if err != nil {
		sugarLogger.Errorf("%v\n", err)
	}
	declNode, deferNode := createNewNodes()
	if err != nil {
		sugarLogger.Errorf("%v\n", err)
	}
	count := 0
	fset = token.NewFileSet()

	// Only add Enter, Exit again if not declared in package previously
	if patchImports {
		declSt, _ := declNode.(*ast.DeclStmt)
		f.Decls = append(f.Decls, declSt.Decl)
	}
	astutil.Apply(f, func(c *astutil.Cursor) bool {
		_, ok := c.Parent().(*ast.FuncDecl)
		if ok {
			bodyList, ok := c.Node().(*ast.BlockStmt)
			if ok {
				count++
				bodyList.List = append([]ast.Stmt{deferNode}, bodyList.List...)
			}
		}
		return true
	}, nil)
	sugarLogger.Infof("Total functions appended: %d\n", count)
	// Generate the code
	src, err := generateFile(fset, f)
	if err != nil {
		sugarLogger.Error(err)
		return nil, err
	}

	return src, err
}

// createNewNodes creates Append statements.
func createNewNodes() (ast.Stmt, ast.Stmt) {
	expr, err := parser.ParseExpr("func() {var Exit, Enter = tracey.New(&tracey.Options{DisableNesting: true}); defer Exit(Enter())}")

	if err != nil {
		sugarLogger.Errorf("%v\n", err)
	}

	if st, ok := expr.(*ast.FuncLit); ok {
		declStmt := st.Body.List[0]
		deferStmt := st.Body.List[1]
		return declStmt, deferStmt
	}

	sugarLogger.Fatalf("Could not create new nodes.")
	return nil, nil
}
