package traceappend

import (
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
			got, err := AppendExpr(tt.args.file, true)
			if err != nil {
				t.Errorf("AppendExpr() error = %v", err)
			}
			if !strings.Contains(string(got), "defer Exit(Enter())") {
				t.Errorf("Did not append expr to code, got: %v\n", got)
			}
		})
	}
	os.Remove("./testdata/testexprappend_bkp.go")
}
