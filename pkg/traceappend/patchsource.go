package traceappend

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/utils"

	"go.uber.org/zap"
)

var logger, _ = zap.NewDevelopment()
var slogger = logger.Sugar()

// PatchSource patches a source path to add tracing.
func PatchSource(sourcePath string) {
	addedTracer := make(map[string]bool)
	slogger.Infof("Patching sourcePath: %v\n", sourcePath)
	err := filepath.Walk(sourcePath, func(path string, f os.FileInfo, err error) error {
		// Don't patch vendor and .git for now.
		if f.IsDir() && (f.Name() == ".git" || f.Name() == "vendor") {
			return filepath.SkipDir
		}

		if filepath.Ext(path) == ".go" {
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
		}
		return nil
	})

	if err != nil {
		slogger.Errorf("Got error: %v\n", err)
	}
}

func patchFile(filePath string, addFunc bool) error {
	if addFunc {
		patched, err := AddImportToFile(filePath)

		if err != nil {
			return err
		}

		// Write the imports to file.
		err = utils.WriteStringToFile(filePath, string(patched))
		if err != nil {
			return err
		}

		// Change REMOTE_SERVER_URL in the code we have to add.
		url, exists := os.LookupEnv("REMOTE_SERVER_URL")

		if !exists {
			slogger.Fatalf("REMOTE_SERVER_URL does not exist in environment.")
		}
		// This is ugly but go's url thing sucks.
		if !strings.HasSuffix(url, "/") {
			url += "/"
		}
		URLAddedCode := strings.ReplaceAll(string(codetoadd), "REMOTE_SERVER_URL", url)
		utils.WriteStringToFile(filePath, addFuncToSource(filePath, URLAddedCode))
	}

	patched, err := AppendExpr(filePath)
	if err != nil {
		return err
	}
	err = utils.WriteStringToFile(filePath, string(patched))

	// If err is nil, nil will be returned.
	return err
}
