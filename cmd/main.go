package cmd

import (
	"log"
	"os"
	"path/filepath"

	"../pkg/gremlin"
	"../pkg/serviceparser"
	"../pkg/utils"
	"github.com/tidwall/gjson"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: main servicedir [destdir]")
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
	log.Print("Cluster version: ", clusterVersion)

	var gremlinQuery string
	gremlinQuery += gremlin.CreateClusterVerisonNode(clusterVersion)

	for idx := range services {
		service := services[idx].Map()
		serviceName := service["name"].String()
		log.Print("Parsing service ", serviceName)
		serviceDetails := service["annotations"].Map()
		serviceVersion := serviceDetails["io.openshift.build.commit.id"].String()

		gremlinQuery += gremlin.CreateNewServiceVersionNode(serviceName, serviceVersion)

		// Git clone the repo
		serviceRoot := utils.RunCloneShell(serviceDetails["io.openshift.build.source-location"].String(), destdir)
		serviceparser.ParseService(serviceName, serviceRoot, destdir)
		addPackageFunctionNodes(serviceName, gremlinQuery)
	}
}

func addPackageFunctionNodes(serviceName string, gremlinQuery string) {
	for pkg, functions := range serviceparser.AllPkgFunc[serviceName] {
		gremlinQuery += gremlin.CreateNewPackageNode(pkg)
		gremlinQuery += gremlin.CreateFunctionNodes(functions)
	}
	log.Print("Executing gremlin query for service: ", serviceName)
	gremlinResponse := gremlin.RunQuery(gremlinQuery)
	log.Print(gremlinResponse)
}
