package runtimelogs

import (
	"fmt"
	"testing"
)

func TestRunE2ETestsInGoPath(t *testing.T) {
	type args struct {
		srcdir string
		gopath string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name:"Run tests in gopath",
			args:args{
				srcdir: "/tmp/src/github.com/openshift/machine-config-operator",
				gopath: "/tmp/",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Println(string(RunE2ETestsInGoPath(tt.args.srcdir, tt.args.gopath)))
		})
	}
}
