package traceappend

import (
	"os"
	"path/filepath"

	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/utils"

	"go.uber.org/zap"
)

var logger, _ = zap.NewDevelopment()
var slogger = logger.Sugar()

// PatchSource patches a source path to add tracing.
func PatchSource(sourcePath string) {
	importedTracey := make(map[string]bool)
	slogger.Infof("Patching sourcePath: %v\n", sourcePath)
	err := filepath.Walk(sourcePath, func(path string, f os.FileInfo, err error) error {
		// Don't patch vendor and .git for now.
		if f.IsDir() && (f.Name() == ".git" || f.Name() == "vendor") {
			return filepath.SkipDir
		}

		if filepath.Ext(path) == ".go" {
			slogger.Infof("Patching file: %v\n", path)
			dirName := filepath.Dir(path)
			_, hasImp := importedTracey[dirName]
			err = patchFile(path, !hasImp)
			if err != nil {
				return err
			}
			if !hasImp {
				importedTracey[dirName] = true
			}
		}

		if f.Name() == "Gopkg.toml" {
			utils.InstallDependency("Gopkg.toml")
		} else if f.Name() == "glide.yaml" {
			utils.InstallDependency("glide.yaml")
		} else if f.Name() == "Godeps.json" {
			utils.InstallDependency("Godeps.json")
		} else if f.Name() == "go.mod" {
			utils.InstallDependency("go.mod")
		}

		return nil
	})

	if err != nil {
		slogger.Errorf("Got error: %v\n", err)
	}
}

func patchFile(filePath string, patchImports bool) error {
	if patchImports {
		patched, err := AddImportToFile(filePath)

		if err != nil {
			return err
		}
		err = utils.WriteStringToFile(filePath, string(patched))

		if err != nil {
			return err
		}
	}

	patched, err := AppendExpr(filePath, patchImports)
	if err != nil {
		return err
	}
	err = utils.WriteStringToFile(filePath, string(patched))

	// If err is nil, nil will be returned.
	return err
}
