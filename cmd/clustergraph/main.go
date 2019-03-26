package main

import (
	"flag"
	"fmt"
	"path/filepath"

	"go.uber.org/zap"

	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/gremlin"
	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/serviceparser"
	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/utils"
	"github.com/tidwall/gjson"
)

var logger, _ = zap.NewDevelopment()
var sugarLogger = logger.Sugar()

func main() {
	servicedir := flag.String("servicedir", "", "The path to folder that contains the cluster_version.json file.")
	destdir := flag.String("destdir", "", "A folder where we can clone the repos of the service for analysis")

	flag.Parse()

	fmt.Println(servicedir)
	clusterInfo := gremlin.ReadJSON(filepath.Join(*servicedir, "cluster_version.json"))
	services := gjson.Get(clusterInfo, "references.spec.tags").Array()
	clusterVersion := gjson.Get(clusterInfo, "digest").String()
	sugarLogger.Infow("Cluster version is", "clusterVersion", clusterVersion)

	gremlin.CreateClusterVerisonNode(clusterVersion)

	for idx := range services {
		service := services[idx].Map()
		serviceName := service["name"].String()
		sugarLogger.Info("Parsing service ", serviceName)
		serviceDetails := service["annotations"].Map()
		serviceVersion := serviceDetails["io.openshift.build.commit.id"].String()

		gremlin.CreateNewServiceVersionNode(clusterVersion, serviceName, serviceVersion)

		// Git clone the repo
		serviceRoot := utils.RunCloneShell(serviceDetails["io.openshift.build.source-location"].String(), *destdir,
			serviceDetails["io.openshift.build.commit.ref"].String(), serviceDetails["io.openshift.build.commit.id"].String())
		serviceparser.ParseService(serviceName, serviceRoot, *destdir)

		gremlin.AddPackageFunctionNodesToGraph(serviceName, serviceVersion)

		serviceImports := serviceparser.AllPkgImports[serviceName]
		for _, imports := range serviceImports {
			imported, ok := imports.([]serviceparser.ImportContainer)
			if !ok {
				sugarLogger.Errorf("Imports are of wrong type: %T\n", imported)
			}
			gremlin.CreateDependencyNodes(serviceName, serviceVersion, imported)
		}
		gremlin.CreateCompileTimeFlows(serviceName, serviceVersion, serviceparser.AllCompileTimeFlows[serviceName])
		// This concludes the offline flow.
	}
}
