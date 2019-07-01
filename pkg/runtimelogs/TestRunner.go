package runtimelogs

import (
	"os"
	"os/exec"

	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/utils"
)

// RunE2ETestsInGopath does exactly what's advertised- Run the component E2E tests and generate an logfile.
func RunE2ETestsInGoPath(srcdir, gopath string) string {
	cmd := exec.Command("dep", "ensure", "-v")
	cmd.Env = make([]string, 0)
	cmd.Env = append(cmd.Env, "GOPATH="+gopath+":/private/tmp")
	cmd.Env = append(cmd.Env, "PATH="+os.Getenv("PATH"))
	cmd.Dir = srcdir

	depOut, depErr := utils.RunCmdWithWait(cmd)

	slogger.Debug(depOut)
	if depErr != "" {
		slogger.Error(depErr)
	}
	makeTest := exec.Command("make", "-C", srcdir, "test-e2e")
	makeTest.Env = make([]string, 0)
	makeTest.Env = append(makeTest.Env, "GOPATH="+gopath)
	makeTest.Env = append(makeTest.Env, "PATH="+os.Getenv("PATH"))
	makeTest.Env = append(makeTest.Env, "KUBECONFIG="+os.Getenv("KUBECONFIG"))
	makeTest.Env = append(makeTest.Env, "KUBERNETES_SERVICE_PORT="+os.Getenv("KUBERNETES_SERVICE_PORT"))
	makeTest.Env = append(makeTest.Env, "KUBERNETES_SERVICE_HOST="+os.Getenv("KUBERNETES_SERVICE_HOST"))

	makeOut, makeErr := utils.RunCmdWithWait(makeTest)

	return makeOut + makeErr
}
