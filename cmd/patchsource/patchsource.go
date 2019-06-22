package main

import (
	"fmt"
	"os"

	flag "github.com/spf13/pflag"

	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/traceappend"
)

func main() {
	sourceDirPtr := flag.String("source-dir", "", "The directory to patch")
	appendFunctionsPtr := flag.String("append-functions", "", "Path to a go file containing the functions to append to every package.")
	PrePendStatementsPtr := flag.String("prepend-statements", "", "Path to file containing the statements to prepend to each and every function's statement body.")
	flag.Parse()

	if *sourceDirPtr == "" || *appendFunctionsPtr == "" || *PrePendStatementsPtr == "" {
		fmt.Fprintf(os.Stderr, "Could not run binary, usage: patchsource --source-dir=[sourcedir] --append-functions [filepath] --prepend-statements [filepath]\n")
		os.Exit(1)
	}

	traceappend.PatchSource(*sourceDirPtr, *appendFunctionsPtr, *PrePendStatementsPtr)
}
