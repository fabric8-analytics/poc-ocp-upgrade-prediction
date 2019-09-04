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

	if len(flag.Args()) > 0 {
		for _, path := range flag.Args() {
			
			serviceName := ""
			// Hardcoded for kube
			if strings.HasSuffix(path, "vendor/k8s.io/kubernetes") {
				serviceName = "hyperkube"
			} else {
				serviceName = ServicePackageMap[filepath.Base(path)]
			}
			
			serviceVersion := utils.GetServiceVersion(path)
			components := serviceparser.NewServiceComponents(serviceName)
			gremlin.CreateNewServiceVersionNode(clusterVersion, serviceName, serviceVersion)

			// Add the imports, packages, functions to graph.
			components.ParseService(serviceName, path)
			gremlin.AddPackageFunctionNodesToGraph(serviceVersion, components)
			parseImportPushGremlin(serviceName, serviceVersion, components)
			edges, err := serviceparser.GetCompileTimeCalls(path, []string{"./cmd/" + serviceName}, *gopathCompilePtr)
			sugarLogger.Infof("silly %s",  len(edges))


			if err != nil {
				sugarLogger.Errorf("Got error: %v, cannot build graph for %s", err, serviceName)
			}
			// Now create the compile time paths
			sugarLogger.Infof("GOING to create compile time edges %s",  len(edges))
			gremlin.CreateCompileTimePaths(edges, serviceName, serviceVersion)
			
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


