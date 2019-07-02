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
func PatchSource(sourcePath, configYamlPath string) {
	addedTracer := make(map[string]bool)
	slogger.Infof("Patching sourcePath: %v\n", sourcePath)
	err := filepath.Walk(sourcePath, func(path string, f os.FileInfo, err error) error {
		// Don't patch vendor and .git for now.
		fmt.Printf("%v %v\n", f.Name(), path)
		if f.IsDir() && utils.IsRestrictedDir(f.Name()) {
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
			err = patchFile(path, configYamlPath)
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

func patchFile(filePath, configYamlPath string) error {

	yamlComponents := utils.ReadCodeFromYaml(configYamlPath)
	if yamlComponents.FuncAdd != "" {
		patched := AddFuncToSource(filePath, yamlComponents.FuncAdd)
		err := utils.WriteStringToFile(filePath, string(patched))
		if err != nil {
			return err
		}
	}

	if yamlComponents.PrependBody != "" {
		patched, err := AppendExpr(filePath, yamlComponents.PrependBody)
		if err != nil {
			return err
		}
		err = utils.WriteStringToFile(filePath, string(patched))
		if err != nil {
			return err
		}
	}

	if len(yamlComponents.Imports) > 0 {
		patched, err := AddImportToFile(filePath, yamlComponents.Imports)

		if err != nil {
			return err
		}

		// Write the imports to file.
		err = utils.WriteStringToFile(filePath, string(patched))
		if err != nil {
			return err
		}
	}
	return nil
}
