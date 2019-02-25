package traceappend

import (
	"go/ast"
	"go/token"
	"reflect"
	"testing"
)

func TestAddImportToFile(t *testing.T) {
	type args struct {
		file string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			"Append import for ast",
			args{file: "testdata/test.go"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AddImportToFile(tt.args.file)
			t.Log(string(got))
			if got == nil {
				t.Errorf("AddImportToFile() got nil")
			}
			if err != nil {
				t.Errorf("%v\n", err)
			}
		})
	}
}

func TestGenerateFile(t *testing.T) {
	type args struct {
		fset *token.FileSet
		file *ast.File
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateFile(tt.args.fset, tt.args.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateFile() = %v, want %v", got, tt.want)
			}
		})
	}
}
