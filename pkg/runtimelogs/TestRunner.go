package runtimelogs

import (
	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/traceappend"
	"os"
	"os/exec"
)

// RunE2ETestsInGopath does exactly what's advertised- Run the component E2E tests and generate an logfile.
func RunE2ETestsInGoPath(srcdir, gopath string) []byte {
	cmd:= exec.Command("dep", "ensure", "-v")
	cmd.Env = make([]string, 0)
	cmd.Env = append(cmd.Env, "GOPATH=" + gopath)
	cmd.Env = append(cmd.Env, "PATH=" + os.Getenv("PATH"))
	cmd.Dir = srcdir
	_, err := cmd.Output()

	traceappend.PatchSource(srcdir)
	slogger.Debugf("%#v\n", cmd)
	if err != nil {
		slogger.Errorf("%v\n", err)
	}
	makeTest := exec.Command("make", "-C", srcdir, "test-e2e")
	makeTest.Env = make([]string, 0)
	makeTest.Env = append(makeTest.Env, "GOPATH=" + gopath)
	makeTest.Env = append(makeTest.Env, "PATH=" + os.Getenv("PATH"))
	makeTest.Env = append(makeTest.Env, "KUBECONFIG=" + os.Getenv("KUBECONFIG"))
	makeTest.Env = append(makeTest.Env, "KUBERNETES_SERVICE_PORT=" + os.Getenv("KUBERNETES_SERVICE_PORT"))
	makeTest.Env = append(makeTest.Env, "KUBERNETES_SERVICE_HOST=" + os.Getenv("KUBERNETES_SERVICE_HOST"))
	slogger.Debugf("%#v\n", makeTest)
	output, err := makeTest.CombinedOutput()
	if err != nil {
		slogger.Errorf("%v\n", err)
	}

	//var stdout []byte
	//var stderr []byte
	//_, err = makeTest.Stdout.Write(stdout)
	//
	//if err != nil {
	//	slogger.Errorf("%v\n", err)
	//}
	//
	//_, err = makeTest.Stderr.Write(stdout)
	//
	//if err != nil {
	//	slogger.Errorf("%v\n", err)
	//}

	//return append(stdout, stderr...)
	return output
}