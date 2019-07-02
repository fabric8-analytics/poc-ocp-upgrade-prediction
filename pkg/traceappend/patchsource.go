package traceappend

import (
	"fmt"
	"io/ioutil"
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
			err = patchFile(path, appendFuncPath, prependStatementsPath, appendFuncPath != "" && !hasTracer, true, prependStatementsPath != "")
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

func patchFile(filePath, appendFuncPath, prependStatementPath string, addFunc bool, addImport bool, addExpressions bool) error {

	if addFunc {
		// Get the functions to be appended.
		funcAppendContent, err := ioutil.ReadFile(appendFuncPath)
		if err != nil {
			return err
		}
		funcAppendString := string(funcAppendContent)
		patched := AddFuncToSource(filePath, funcAppendString)
		err = utils.WriteStringToFile(filePath, string(patched))
		if err != nil {
			return err
		}
	}

	if addExpressions {
		// Get the expressions to be appended.
		exprAppendContent, err := ioutil.ReadFile(prependStatementPath)
		if err != nil {
			return err
		}
		exprAppendString := string(exprAppendContent)

		patched, err := AppendExpr(filePath, exprAppendString)
		if err != nil {
			return err
		}
		err = utils.WriteStringToFile(filePath, string(patched))
		if err != nil {
			return err
		}
	}

	if addImport {
		importsListPrintParent := map[string]string{
			"godefaultruntime": "runtime",
			"goformat":         "fmt",
			"gotime":           "time",
			"goos":             "os",
		}
		patched, err := AddImportToFile(filePath, importsListPrintParent)

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
