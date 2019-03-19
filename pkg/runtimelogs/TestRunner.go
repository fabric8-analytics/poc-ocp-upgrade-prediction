package runtimelogs

import (
	"os"
	"os/exec"
)

// RunE2ETestsInGopath does exactly what's advertised- Run the component E2E tests and generate an logfile.
func RunE2ETestsInGoPath(gopath string) string {
	makeTest := exec.Command("make", "-C", "", "test-e2e")
	makeTest.Env = os.Environ()
	makeTest.Env = append(makeTest.Env, "GOPATH=" + gopath)
	makeTest.Env = append(makeTest.Env, "KUBECONFIG=" + os.Getenv("KUBECONFIG"))
	makeTest.Env = append(makeTest.Env, "KUBERNETES_SERVICE_PORT=" + os.Getenv("KUBERNETES_SERVICE_PORT"))
	makeTest.Env = append(makeTest.Env, "KUBERNETES_SERVICE_HOST=" + os.Getenv("KUBERNETES_SERVICE_HOST"))
	err := makeTest.Start()
	if err != nil {
		slogger.Errorf("%v\n", err)
	}
	err = makeTest.Wait()
	if err != nil {
		slogger.Errorf("%v\n", err)
	}

	var stdout []byte
	var stderr []byte
	_, err = makeTest.Stdout.Write(stdout)

	if err != nil {
		slogger.Errorf("%v\n", err)
	}

	_, err = makeTest.Stderr.Write(stdout)

	if err != nil {
		slogger.Errorf("%v\n", err)
	}

	return string(append(stdout, stderr...))
}