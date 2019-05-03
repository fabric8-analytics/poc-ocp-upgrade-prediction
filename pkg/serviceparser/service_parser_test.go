package serviceparser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func Test_parseServiceAST(t *testing.T) {
	type args struct {
		node ast.Node
		fset *token.FileSet
		pkg  string
	}
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "./testdata/main.go", nil, parser.AllErrors)
	if err != nil {
		t.Errorf("Got error: %v", err)
	}
	tests := []struct {
		name string
		args args
	}{
		{
			"Import parsing",
			args{
				node: node,
				fset: fset,
				pkg:  "main",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := parseServiceAST(tt.args.node.(*ast.Package), tt.args.fset, tt.args.pkg)
			if got == nil {
				t.Errorf("parseServiceAST() got nil for functions.")
			}
			if got1 == nil {
				t.Errorf("parserServiceAST() got nil for imports.")
			}
		})
	}
}
