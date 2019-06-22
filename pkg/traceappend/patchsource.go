package traceappend

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/utils"

	"go.uber.org/zap"
)

var logger, _ = zap.NewDevelopment()
var slogger = logger.Sugar()

// PatchSource patches a source path to add tracing.
func PatchSource(sourcePath, appendFuncPath, prependStatementsPath string) {
	addedTracer := make(map[string]bool)
	slogger.Infof("Patching sourcePath: %v\n", sourcePath)
	err := filepath.Walk(sourcePath, func(path string, f os.FileInfo, err error) error {
		// Don't patch vendor and .git for now.
		fmt.Printf("%v %v\n", f.Name(), path)
		if f.IsDir() && !utils.IsRestrictedDir(f.Name()) {
			return filepath.SkipDir
		}

		excludeVendor, set := os.LookupEnv("EXCLUDE_VENDOR")
		if !set || (set && excludeVendor != "true") {
			if f.IsDir() && f.Name() == "vendor" {
				return filepath.SkipDir
			}
		}
		if !utils.IsIgnoredFile(path) {
			slogger.Infof("Patching file: %v\n", path)
			dirName := filepath.Dir(path)
			_, hasTracer := addedTracer[dirName]
			err = patchFile(path, !hasTracer)
			if err != nil {
				return err
			}
			if !hasTracer {
				addedTracer[dirName] = true
			}
		} else {
			slogger.Infof("Skipping ignored file: %s\n", path)
		}
		return nil
	})

	if err != nil {
		slogger.Errorf("Got error: %v\n", err)
	}
}

func patchFile(filePath string, addFunc bool) error {
	if addFunc {
		patched, err := AddOpenTracingImportToFile(filePath)

		if err != nil {
			return err
		}

		// Write the imports to file.
		err = utils.WriteStringToFile(filePath, string(patched))
		if err != nil {
			return err
		}
	}

	patched, err := AppendExpr(filePath)
	if err != nil {
		return err
	}
	err = utils.WriteStringToFile(filePath, string(patched))

	// Add a context parameter to all functions.
	addContextArgumentToFuncDecl(filePath)

	// Add a context parameter as the first argument to all function calls.
	AddContextToCallExpressions(filePath)

	// If err is nil, nil will be returned.
	return err
}
