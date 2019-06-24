package traceappend

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/utils"
)

func TestAppendExpr(t *testing.T) {
	type args struct {
		file string
	}
	err := utils.CopyFile("./testdata/testexprappend.go", "./testdata/testexprappend_bkp.go")
	if err != nil {
		t.Errorf("%v\n", err)
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "Test expression append",
			args: args{
				file: "./testdata/testexprappend_bkp.go",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AppendExpr(tt.args.file)
			if err != nil {
				t.Errorf("AppendExpr() error = %v", err)
			}
			if !strings.Contains(string(got), "_logClusterCodePath") {
				t.Errorf("Did not append expr to code, got: %v\n", string(got))
			}
		})
	}
	os.Remove("./testdata/testexprappend_bkp.go")
}

func Test_addContextArgumentToFunction(t *testing.T) {
	file, err := ioutil.TempFile("/tmp", "prefix")
	if err != nil {
		log.Fatal(err)
	}
	testProgram := `package main

	func main() {
		go func() {}()
		someLibFunctThatAcceptsHOF(func() {})
	}

	func alreadyHasContext(ctx context.Context) {}

	func regularFunc() {}
	`
	file.Write([]byte(testProgram))
	defer os.Remove(file.Name())

	type args struct {
		filePath string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "All the corner cases",
			args: args{
				filePath: file.Name(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addContextArgumentToFuncDecl(tt.args.filePath)
			buf, err := ioutil.ReadFile(file.Name())
			if err != nil {
				fmt.Printf("Got error while reading output file, failing test. Error: %v\n", err)
			}
			fmt.Printf("%v\n", string(buf))
		})
	}
}

func TestAddContextToCallExpressions(t *testing.T) {
	file, err := ioutil.TempFile("/tmp", "prefix")
	if err != nil {
		log.Fatal(err)
	}
	testProgram := `package main

	func main() {
		go func() {}()
		var t string
		someLibFunctThatAcceptsHOF(func() {})
		somefunction(t)
	}

	func alreadyHasContext(ctx context.Context) {}

	func regularFunc() {}
	`
	file.Write([]byte(testProgram))
	defer os.Remove(file.Name())

	type args struct {
		filePath string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Test if code is added to function argument.",
			args: args{
				filePath: file.Name(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AddContextToCallExpressions(tt.args.filePath)
			buf, err := ioutil.ReadFile(file.Name())
			if err != nil {
				fmt.Printf("Got error while reading output file, failing test. Error: %v\n", err)
			}
			fmt.Printf("%v\n", string(buf))
		})
	}
}
