package main

import (
	"log"
	"os"
	"path/filepath"

	"./gremlin"
	"./serviceparser"
	"./utils"
	gjson "github.com/tidwall/gjson"
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

	for idx := range services {
		service := services[idx].Map()
		serviceName := service["name"].String()
		log.Print("Parsing service ", serviceName)
		serviceDetails := service["annotations"].Map()

		// Git clone the repo
		serviceRoot := utils.RunCloneShell(serviceDetails["io.openshift.build.source-location"].String(), destdir)
		serviceparser.ParseService(serviceName, serviceRoot, destdir)
		// serviceVersion := serviceDetails["io.openshift.build.commit.id"].String()
		break
	}
}
