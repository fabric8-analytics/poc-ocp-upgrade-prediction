package traceappend

// Parent: https://gist.github.com/josephspurrier/19fb8096099bfff5556742072680d061

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"

	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/serviceparser"

	"go.uber.org/zap"
	"golang.org/x/tools/go/ast/astutil"
)

var loggertra, _ = zap.NewDevelopment()
var sugarLogger = loggertra.Sugar()

// AddImportToFile will be used to import G, O objects for logging.
func AddImportToFile(file string) ([]byte, error) {
	// Create the AST by parsing src
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, nil, parser.ParseComments)

	// This never fails, because its failure means that a module is already imported.
	astutil.AddNamedImport(fset, f, "godefaultruntime", "runtime")
	astutil.AddNamedImport(fset, f, "goformat", "fmt")
	astutil.AddNamedImport(fset, f, "gotime", "time")
	astutil.AddNamedImport(fset, f, "goos", "os")

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
	if err := format.Node(buffer, fset, file); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// AppendExpr modifies an AST by adding an expr at the start of its body.
func AppendExpr(file string) ([]byte, error) {
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
				tempBody := append(deferNode, bodyList.List...)
				bodyList.List = tempBody
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
	expr, err := parser.ParseExpr(`func() {	if ctx == nil {
		ctx = context.Background()
	}
	pc := make([]uintptr, 10) // at least 1 entry needed
	runtime.Callers(2, pc)
	fn := runtime.FuncForPC(pc[0])
	span, ctx := opentracing.StartSpanFromContext(ctx, fn.Name())

	defer span.Finish()
	span.LogFields(
		log.String("event", "entered function"),
		log.String("value", fn.Name()),
	)
	}`)

	if err != nil {
		sugarLogger.Errorf("%v\n", err)
	}
	// This cannot error, it's literally hardcoded.
	return expr.(*ast.FuncLit).Body.List
}

func addFuncToSource(filePath, appendCode string) string {
	fset1 := token.NewFileSet()
	fset2 := token.NewFileSet()

	cf1, err := parser.ParseFile(fset1, "code1.go", appendCode, parser.ParseComments)
	if err != nil {
		sugarLogger.Error(err)
	}
	cf2, err := parser.ParseFile(fset2, filePath, nil, parser.ParseComments)
	if err != nil {
		sugarLogger.Error(err)
	}

	cf2.Decls = append(cf2.Decls, cf1.Decls...)
	content, _ := generateFile(fset2, cf2)
	return string(content)
}

var codetoadd = `
package dummy

func _logClusterCodePath(op string) {
    // Skip this function, and fetch the PC and file for its parent
    pc, _, _, _ := godefaultruntime.Caller(1);
    goformat.Fprintf(goos.Stderr, "[%v][ANALYTICS] %s%s\n", gotime.Now().UTC(), op, godefaultruntime.FuncForPC(pc).Name())
}
`

// addContextArgumentToFunction function adds a new context parameter to all functions that are not anonymous, self executing functions
// unless there is already a context parameter passed in which case the variable name of the context parameter is returned.
// This is required for tracing intra-process context with opentracing.
func addContextArgumentToFunction(filePath string) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, nil, 0)
	if err != nil {
		slogger.Errorf("Got error: %v\n, failed to patch source file: %s", err, filePath)
	}

	astutil.Apply(f, func(c *astutil.Cursor) bool {
		switch t := c.Node().(type) {

		case *ast.FuncDecl:
			// Do not patch any main/init functions.
			params := t.Type.Params
			contextSelectorExpr := ast.SelectorExpr{
				X: &ast.Ident{
					Name: "context",
				},
				Sel: &ast.Ident{
					Name: "Context",
				},
			}
			contextArgument := ast.Field{
				Names: []*ast.Ident{
					&ast.Ident{
						Name: "ctx",
						Obj: &ast.Object{
							Kind: ast.Var,
							Name: "ctx",
						},
					},
				},
				Type: &contextSelectorExpr,
			}
			// TODO: check if context argument already present.
			if len(params.List) > 0 {
				c.Node().(*ast.FuncDecl).Type.Params.List = append([]*ast.Field{
					&contextArgument,
				}, params.List...)
			} else {
				c.Node().(*ast.FuncDecl).Type.Params.List = []*ast.Field{&contextArgument}
			}
		}
		return true
	}, nil)

	// Generate the code
	src, err := generateFile(fset, f)
	if err != nil {
		sugarLogger.Error(err)
	}

	fo, err := os.OpenFile(filePath, os.O_WRONLY, 0644)
	if err != nil {
		sugarLogger.Errorf("%v\n", err)
	}

	_, err = fo.Write(src)
	if err != nil {
		sugarLogger.Errorf("%v\n", err)
	}
	// Don't care for any closing errors.
	fo.Close()
}

// AddOpenTracingImportToFile will be used to import opentracing objects for runtime path logging.
func AddOpenTracingImportToFile(file string) ([]byte, error) {
	// Create the AST by parsing src
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, nil, parser.ParseComments)

	// This never fails, because its failure means that a module is already imported.
	astutil.AddNamedImport(fset, f, "opentracing", "github.com/opentracing/opentracing-go")
	astutil.AddImport(fset, f, "github.com/opentracing/opentracing-go/log")
	astutil.AddImport(fset, f, "context")
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

func getExprForObject() ast.Expr {
	expr, err := parser.ParseExpr("ctx")
	if err != nil {
		slogger.Errorf("Got error: %v\n", err)
	}
	return expr
}

// AddContextToCallExpressions adds our context argument as the first parameter in the function call.
func AddContextToCallExpressions(filePath string) {
	// Create the AST by parsing src
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)

	astutil.Apply(f, func(c *astutil.Cursor) bool {
		switch t := c.Node().(type) {
		case *ast.CallExpr:
			// Do not patch any main/init functions.
			if t.Fun.(*ast.Ident).Name == "main" || t.Fun.(*ast.Ident).Name == "init" {
				return true
			}
			// Don't patch any builtins.
			if _, exists := serviceparser.Builtins[t.Fun.(*ast.Ident).Name]; exists {
				return true
			}
			// Don't patch any library functions.

			contextArgument := getExprForObject()
			// TODO: check if context argument already present.
			if len(t.Args) > 0 {
				c.Node().(*ast.CallExpr).Args = append([]ast.Expr{
					contextArgument,
				}, t.Args...)
			} else {
				c.Node().(*ast.CallExpr).Args = []ast.Expr{contextArgument}
			}
		}
		return true
	}, nil)

	// Generate the code
	src, err := generateFile(fset, f)
	if err != nil {
		sugarLogger.Error(err)
	}

	fo, err := os.OpenFile(filePath, os.O_WRONLY, 0644)
	if err != nil {
		sugarLogger.Errorf("%v\n", err)
	}

	_, err = fo.Write(src)
	if err != nil {
		sugarLogger.Errorf("%v\n", err)
	}
	// Don't care for any closing errors.
	fo.Close()
}
