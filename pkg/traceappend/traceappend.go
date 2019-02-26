package traceappend

// Parent: https://gist.github.com/josephspurrier/19fb8096099bfff5556742072680d061

import (
	"bytes"
	"errors"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"

	"go.uber.org/zap"
	"golang.org/x/tools/go/ast/astutil"
)

var logger, _ = zap.NewProduction()
var sugarLogger = logger.Sugar()

// AddImportToFile will be used to import G, O objects for logging.
func AddImportToFile(file string) ([]byte, error) {
	// Create the AST by parsing src
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, nil, 0)

	done := astutil.AddImport(fset, f, "go/ast")

	if !done {
		return nil, errors.New("Unable to add import to AST")
	}
	// Generate the code
	src, err := GenerateFile(fset, f)
	if err != nil {
		sugarLogger.Error(err)
		return nil, err
	}

	fo, err := os.OpenFile(file, os.O_WRONLY, 0644)
	if err != nil {
		sugarLogger.Errorf("%v\n", err)
	}

	sugarLogger.Infof(string(src))
	_, err = fo.Write(src)
	if err != nil {
		sugarLogger.Errorf("%v\n", err)
	}
	fo.Close()
	return src, err
}

// GenerateFile creates a new file with the new code appended and returns its contents.
func GenerateFile(fset *token.FileSet, file *ast.File) ([]byte, error) {
	var output []byte
	buffer := bytes.NewBuffer(output)
	if err := printer.Fprint(buffer, fset, file); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
