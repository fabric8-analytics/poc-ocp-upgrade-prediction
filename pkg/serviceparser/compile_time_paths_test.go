// Original Copyright: 2014 The Go Authors. All rights reserved.
package serviceparser

import (
	"testing"
)

func TestGetCompileTimeCalls(t *testing.T) {

	type args struct {
		dir    string
		args   []string
		gopath string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test for hypershift",
			args: args{
				"/Users/avgupta/golang/src/github.com/openshift/origin",
				[]string{"./cmd/hypershift"},
				"/Users/avgupta/golang",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := GetCompileTimeCalls(tt.args.dir, tt.args.args, tt.args.gopath); (err != nil) != tt.wantErr {
				t.Errorf("GetCompileTimeCalls() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
