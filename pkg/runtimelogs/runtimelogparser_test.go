package runtimelogs

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/utils"
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
		name string
		args args
	}{
		{
			"Test log parsing",
			args{
				testLog: testData,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseComponentE2ELogs(tt.args.testLog)
			fmt.Printf("Got: %v\n", got)
			if err != nil {
				t.Errorf("ParseComponentE2ELogs() error = %v", err)
				return
			}
			if reflect.TypeOf(got) != reflect.TypeOf(map[string][]RuntimeLogEntry{}) {
				t.Errorf("ParseComponentE2ELogs() = %v", got)
			}
		})
	}
}

func TestCreateRuntimePaths(t *testing.T) {
	type args struct {
		logPath string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			"Test runtime path creation",
			args{"./testdata/e2e_log.txt"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			CreateRuntimePaths(tt.args.logPath)
		})
	}
}
