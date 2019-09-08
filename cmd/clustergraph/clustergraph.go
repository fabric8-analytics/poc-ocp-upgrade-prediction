package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	flag "github.com/spf13/pflag"

	"go.uber.org/zap"

	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/gremlin"
	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/serviceparser"
	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/utils"
	"github.com/tidwall/gjson"
)

var logger, _ = zap.NewDevelopment()
var sugarLogger = logger.Sugar()

func main() {
	clusterversion := flag.String("cluster-version", "", "A release version of OCP")
	destdir := flag.String("destdir", "./", "A folder where we can clone the repos of the service for analysis")
	fmt.Printf(" destdir are %s", destdir)

	gopathCompilePtr := flag.String("gopath", os.Getenv("GOPATH"), "GOPATH for compile time path builds. Defaults to system GOPATH")
	fmt.Printf(" gopathCompilePtr are %s", *gopathCompilePtr)

	flag.Parse()
	fmt.Printf(" Args are %s", flag.Args())

	payloadInfo, err := exec.Command("oc", "adm", "release", "info", "--commits=true",
		fmt.Sprintf("quay.io/openshift-release-dev/ocp-release:%s", *clusterversion), "-o", "json").CombinedOutput()
	if err != nil {
		sugarLogger.Errorf("(%v): %s", err, string(payloadInfo))
	}

	clusterInfo := string(payloadInfo)
	// services := gjson.Get(clusterInfo, "references.spec.tags").Array()
	clusterVersion := gjson.Get(clusterInfo, "digest").String()
	sugarLogger.Infow("Cluster version is", "clusterVersion", clusterVersion)

	gremlin.CreateClusterVerisonNode(clusterVersion)
	serviceVersion := "sha256:155ef40a64608c946ca9ca0310bbf88f5a4664b2925502b3acac86847bc158e6"
	numberofservices := len(flag.Args())
	compileTimePaths := make([]ServiceCompileTimeCalls, numberofservices)

	if len(flag.Args()) > 0 {
		for index, path := range flag.Args() {

			serviceName := ""
			// Hardcoded for kube
			if strings.HasSuffix(path, "vendor/k8s.io/kubernetes") {
				serviceName = "hyperkube"
			} else {
				serviceName = ServicePackageMap[filepath.Base(path)].servicename
				serviceVersion = utils.GetServiceVersion(path)
			}
			components := serviceparser.NewServiceComponents(serviceName)
			gremlin.CreateNewServiceVersionNode(clusterVersion, serviceName, serviceVersion)

			// Add the imports, packages, functions to graph.
			components.ParseService(serviceName, path)
			gremlin.AddPackageFunctionNodesToGraph(serviceName, serviceVersion, components)
			parseImportPushGremlin(serviceName, serviceVersion, components)
			edges, err := serviceparser.GetCompileTimeCalls(path, ServicePackageMap[filepath.Base(path)].cmdname, *gopathCompilePtr)
			compileTimePaths[index] = ServiceCompileTimeCalls{servicename: serviceName, edges: edges}
			sugarLogger.Infof(" Number of compile time paths for service %s are  %s", serviceName, len(edges))

			if err != nil {
				sugarLogger.Errorf("Got error: %v, in calculating compile time paths %s", err, serviceName)
			}
		}
		// Create compile time paths for all the functions observed under each package of each known service
		for index, _ := range compileTimePaths {
			sugarLogger.Infof("For service %s GOING to create compile time edges %s ", compileTimePaths[index].servicename, len(compileTimePaths[index].edges))
			gremlin.CreateCompileTimePaths(compileTimePaths[index].edges, compileTimePaths[index].servicename)
		}
	}
}

func filterImports(imports []serviceparser.ImportContainer, serviceName string) []serviceparser.ImportContainer {
	var filtered []serviceparser.ImportContainer
	unique := make(map[string]bool)
	for _, imported := range imports {
		if len(strings.Split(imported.ImportPath, "/")) > 2 && !strings.Contains(imported.ImportPath, serviceName) {
			if !unique[imported.ImportPath] {
				filtered = append(filtered, imported)
				unique[imported.ImportPath] = true
			}
		}
	}
	return filtered
}

func parseImportPushGremlin(serviceName, serviceVersion string, components *serviceparser.ServiceComponents) {
	serviceImports := components.AllPkgImports
	for _, imports := range serviceImports {
		imported, ok := imports.([]serviceparser.ImportContainer)
		if !ok {
			sugarLogger.Errorf("Imports are of wrong type: %T\n", imported)
		}
		imported = filterImports(imported, serviceName)
		gremlin.CreateDependencyNodes(serviceName, serviceVersion, imported)
	}
}
