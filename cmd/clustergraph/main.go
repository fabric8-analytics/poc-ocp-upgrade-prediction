package main

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"

	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/gremlin"
	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/serviceparser"
	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/utils"
	"github.com/tidwall/gjson"
)

var logger, _ = zap.NewProduction()
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

	var gremlinQuery string
	gremlinQuery += gremlin.CreateClusterVerisonNode(clusterVersion)

	for idx := range services {
		service := services[idx].Map()
		serviceName := service["name"].String()
		sugarLogger.Info("Parsing service ", serviceName)
		serviceDetails := service["annotations"].Map()
		serviceVersion := serviceDetails["io.openshift.build.commit.id"].String()

		gremlinQuery += gremlin.CreateNewServiceVersionNode(serviceName, serviceVersion)

		// Git clone the repo
		serviceRoot := utils.RunCloneShell(serviceDetails["io.openshift.build.source-location"].String(), destdir)
		serviceparser.ParseService(serviceName, serviceRoot, destdir)
		addPackageFunctionNodesToGraph(serviceName, gremlinQuery)
	}
}

func addPackageFunctionNodesToGraph(serviceName string, gremlinQuery string) {
	for pkg, functions := range serviceparser.AllPkgFunc[serviceName] {
		gremlinQuery += gremlin.CreateNewPackageNode(pkg)
		gremlinQuery += gremlin.CreateFunctionNodes(functions)
	}
	sugarLogger.Info("Executing gremlin query for service: ", serviceName)
	gremlinResponse := gremlin.RunQuery(gremlinQuery)
	sugarLogger.Info(gremlinResponse)
}
