package traceappend

import (
	"go/ast"
	"go/token"
	"io/ioutil"
	"reflect"
	"testing"
)

func TestAddImportToFile(t *testing.T) {
	type args struct {
		file string
	}
	dat, err := ioutil.ReadFile("./test/test.go")
	if err != nil {
		t.Errorf("Failed to open test file to initialize test.")
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			"Append import for ast",
			args{file: "test/test.go"},
			dat,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AddImportToFile(tt.args.file)
			t.Log(string(got))
			if (err != nil) != tt.wantErr {
				t.Errorf("AddImportToFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddImportToFile() = %v, want %v", got, tt.want)
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
