package main

import (
	"fmt"
	"os"

	flag "github.com/spf13/pflag"

	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/traceappend"
)

func main() {
	sourceDirPtr := flag.String("source-dir", "", "The directory to patch")
	flag.Parse()

	if *sourceDirPtr == "" {
		fmt.Fprintf(os.Stderr, "Could not run binary, usage: patchsource --source-dir=[sourcedir]\n")
		os.Exit(1)
	}

	traceappend.PatchSource(*sourceDirPtr)
}
