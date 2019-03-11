package traceappend

import (
	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/utils"
	"testing"
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
			_, err := AppendExpr(tt.args.file)
			if err != nil {
				t.Errorf("AppendExpr() error = %v", err)
				return
			}
		})
	}
}
