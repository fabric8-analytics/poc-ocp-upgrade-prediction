package utils

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RunCloneShell runs a git clone inside a child shell and clones it as a subdir of destdir
func RunCloneShell(repo string, destdir string) string {
	_, repodir := filepath.Split(repo)
	repodir = strings.Split(repodir, ".git")[0]

	if _, err := os.Stat(destdir); os.IsNotExist(err) {
		os.Mkdir(destdir, os.ModePerm)
	}

	cmdRun := exec.Command("git", "clone", repo, filepath.Join(destdir, repodir))
	stdout, err := cmdRun.CombinedOutput()

	if err != nil {
		log.Print(string(stdout))
		log.Print(err)
	}

	return filepath.Join(destdir, repodir)
}
