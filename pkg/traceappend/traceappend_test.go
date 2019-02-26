package traceappend

import (
	"testing"
)

func TestAddImportToFile(t *testing.T) {
	type args struct {
		file string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			"Append import for ast",
			args{file: "testdata/test.go"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AddImportToFile(tt.args.file)
			if got == nil {
				t.Errorf("AddImportToFile() got nil")
			}
			if err != nil {
				t.Errorf("%v\n", err)
			}
		})
	}
}
