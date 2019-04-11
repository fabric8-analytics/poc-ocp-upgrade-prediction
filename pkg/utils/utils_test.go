package utils

import (
	"os"
	"testing"
)

func TestRunCloneShell(t *testing.T) {
	type args struct {
		repo     string
		destdir  string
		branch   string
		revision string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test git clone at revision",
			args: args{
				repo:     "https://github.com/openshift/machine-config-operator",
				destdir:  "/tmp",
				branch:   "master",
				revision: "287504634d7a52a605d2c7f7c46f93f281368915",
			},
			want: "/tmp/src/github.com/openshift/machine-config-operator",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, cloned := RunCloneShell(tt.args.repo, tt.args.destdir, tt.args.branch, tt.args.revision)
			_, err := os.Stat(got)
			if err != nil  || !cloned {
				t.Errorf("Clone failed.")
			}
			if got != tt.want {
				t.Errorf("Wanted: %v\n, Got: %v\n", tt.want, got)
			}
			_ = os.Remove(got)
		})
	}
}
