package utils

import (
	"log"
	"os/exec"
)

// RunShell runs a shell command
func RunShell(cmd string) {
	cmdRun := exec.Command(cmd)
	err := cmdRun.Run()

	if err != nil {
		log.Fatal(err)
	}
}
