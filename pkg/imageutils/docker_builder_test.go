package imageutils

import "testing"

func Test_createImage(t *testing.T) {
	type args struct {
		imageRegistry  string
		imageName      string
		dockerfilePath string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Create machine-config-controller image",
			args: args{
				dockerfilePath: "/Users/avgupta/golang/src/github.com/openshift/machine-config-operator/Dockerfile.machine-config-controller",
				imageRegistry:  "quay.io",
				imageName:      "rootavish/machine-config-controller:latest",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createImage(tt.args.imageRegistry, tt.args.imageName, tt.args.dockerfilePath)
		})
	}
}
