package main

import (
	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/traceappend"
	"os"
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
	if len(os.Args) < 2 {
		sugarLogger.Fatal("Usage: main servicedir [destdir]")
	}

	servicedir := os.Args[1]
	var destdir string

	if len(os.Args) > 2 {
		destdir = os.Args[2]
	} else {
		destdir = os.Args[1]
	}

	clusterInfo := gremlin.ReadJSON(filepath.Join(servicedir, "cluster_version.json"))
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

		gremlin.CreateNewServiceVersionNode(serviceName, serviceVersion)

		// Git clone the repo
		serviceRoot := utils.RunCloneShell(serviceDetails["io.openshift.build.source-location"].String(), destdir)
		serviceparser.ParseService(serviceName, serviceRoot, destdir)

		gremlin.AddPackageFunctionNodesToGraph(serviceName, serviceVersion)

		serviceImports := serviceparser.AllPkgImports[serviceName]
		for _, imports := range serviceImports {
			imported, ok := imports.([]serviceparser.ImportContainer)
			if !ok {
				sugarLogger.Errorf("Imports are of wrong type: %T\n", imported)
			}
			gremlin.CreateDependencyNodes(clusterVersion, serviceName, serviceVersion, imported)
		}
		gremlin.CreateCompileTimeFlows(clusterVersion, serviceName, serviceVersion, serviceparser.AllCompileTimeFlows[serviceName])

		// Append markers for runtime flow deduction
		sugarLogger.Info("Now patching source.")
		traceappend.PatchSource(serviceRoot)
	}
}

