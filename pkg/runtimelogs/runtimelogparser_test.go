package runtimelogs

import (
	"fmt"
	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/utils"
	"reflect"
	"testing"
)

func TestParseComponentE2ELogs(t *testing.T) {
	type args struct {
		testLog []string
	}
	testData, err := utils.ReadFileLines("./testdata/e2e_log.txt")
	if err != nil {
		t.Errorf("Got error: %v\n", err)
	}
	tests := []struct {
		name    string
		args    args
		want    *Runtimelog
		wantErr bool
	}{
		{
			"Test log parsing",
			args{
				testLog: testData,
			},
			&Runtimelog{},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseComponentE2ELogs(tt.args.testLog)
			fmt.Printf("Got: %v\n", got)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseComponentE2ELogs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("ParseComponentE2ELogs() = %v, want %v", got, tt.want)
			}
		})
	}
}
