package main

import (
	"fmt"
	"os"
	"strconv"

	flag "github.com/spf13/pflag"

	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/traceappend"
)

func main() {
	sourceDirPtr := flag.String("source-dir", "", "The directory to patch")
	configYamlPtr := flag.String("code-config-yaml", "", "A yaml containing the code that will be used to modify the original source. See documentation for more details.")
	includeVendor := flag.Bool("include-vendor", false, "Whether the vendor folder should be included in the patching.")
	flag.Parse()

	os.Setenv("EXCLUDE_VENDOR", strconv.FormatBool(*includeVendor))
	if *sourceDirPtr == "" {
		fmt.Fprintf(os.Stderr, "Could not run binary, usage: patchsource --source-dir=[sourcedir] --append-functions [filepath] --prepend-statements [filepath]\n")
		os.Exit(1)
	}

	traceappend.PatchSource(*sourceDirPtr, *configYamlPtr)
}
