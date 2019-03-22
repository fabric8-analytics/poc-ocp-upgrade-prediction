// Contains the logic for interaction with the PR on Github
package ghpr

import "testing"

func TestGetPRPayload(t *testing.T) {
	type args struct {
		repoStr string
		prId    int
		gopath string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Test PR retrieval",
			args: args{
				repoStr:"openshift/machine-config-operator",
				prId: 513,
				gopath:"/tmp",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			GetPRPayload(tt.args.repoStr, tt.args.prId, tt.args.gopath)
		})
	}
}
