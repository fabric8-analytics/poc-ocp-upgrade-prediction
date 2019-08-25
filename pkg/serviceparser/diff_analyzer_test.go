package serviceparser

import (
	"reflect"
	"testing"
)

func Test_getAddedFunctions(t *testing.T) {
	type args struct {
		diffContent string
	}
	tests := []struct {
		name string
		args args
		want []SimpleFunctionRepresentation
	}{
		{
			name: "Test function finder regex.",
			args: args{
				diffContent: `
							"fmt"
							"strings"

					-	"k8s.io/api/core/v1"
					+	v1 "k8s.io/api/core/v1"
							metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
							utilfeature "k8s.io/apiserver/pkg/util/feature"
							kubeapi "k8s.io/kubernetes/pkg/apis/core" // or equal to SystemCriticalPriority. Both the default scheduler and the kubelet use this function
						// to make admission and scheduling decisions.
						func IsCriticalPod(pod *v1.Pod) bool {
					+	if IsStaticPod(pod) {
					+		return true
					+	}
							if utilfeature.DefaultFeatureGate.Enabled(features.PodPriority) {
								if pod.Spec.Priority != nil && IsCriticalPodBasedOnPriority(*pod.Spec.Priority) {
									return true 	}
							return false
						}
					+
					+// IsStaticPod returns true if the pod is a static pod.
					+func IsStaticPod(pod *v1.Pod) bool {
					+	source, err := GetPodSource(pod)
					+	return err == nil && source != ApiserverSource
					+}
						`,
			},
			want: []SimpleFunctionRepresentation{
				SimpleFunctionRepresentation{
					Fun: "IsStaticPod",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getAddedFunctions(tt.args.diffContent); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getAddedFunctions() = %v, want %v", got, tt.want)
			}
		})
	}
}
