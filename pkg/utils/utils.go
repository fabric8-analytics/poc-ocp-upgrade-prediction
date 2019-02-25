package utils

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

var logger, _ = zap.NewProduction()
var sugarLogger = logger.Sugar()

// RunCloneShell runs a git clone inside a child shell and clones it as a subdir of destdir
func RunCloneShell(repo string, destdir string) string {
	_, repodir := filepath.Split(repo)
	repodir = strings.Split(repodir, ".git")[0]

	if _, err := os.Stat(destdir); os.IsNotExist(err) {
		errdir := os.Mkdir(destdir, os.ModePerm)
		if errdir != nil {
			sugarLogger.Fatal(errdir)
		}
	}

	cmdRun := exec.Command("git", "clone", repo, filepath.Join(destdir, repodir))
	stdout, err := cmdRun.CombinedOutput()

	if err != nil {
		sugarLogger.Error(string(stdout))
		sugarLogger.Error(err)
	}

	return filepath.Join(destdir, repodir)
}
