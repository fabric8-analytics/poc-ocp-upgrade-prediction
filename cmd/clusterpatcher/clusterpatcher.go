// Clusterpatcher generates a utility binary that can be used to create a new cluster image with
// all the service sources patched.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/traceappend"
	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/utils"
	"github.com/tidwall/gjson"

	"go.uber.org/zap"
)

var logger, _ = zap.NewDevelopment()
var slogger = logger.Sugar()

func main() {
	payloadVersion := flag.String("cluster-version", "", "The payload version that needs to be patched.")
	repoSourceDir := flag.String("repo-src-dir", "",
		`The Openshift services folder where the clustergraph binary has cloned all the service repos,
		if this flag is present it is given precedence over the cluster version flag.`)

	flag.Parse()
	slogger.Debugf("Flags initialized.")

	if repoSourceDir != nil && *repoSourceDir != "" {
		srcDirList, err := ioutil.ReadDir(*repoSourceDir)
		if err != nil {
			slogger.Errorf("%v\n", err)
		}
		for _, fileInf := range srcDirList {
			traceappend.PatchSource(fileInf.Name())
		}
	} else if payloadVersion != nil && *payloadVersion != "" {
		slogger.Infof("Running flow for cluster version option.")
		cmd := exec.Command("oc", "adm", "release", "info", "--commits=true",
			fmt.Sprintf("registry.svc.ci.openshift.org/ocp/release:%s", *payloadVersion),
			"-o", "json")

		clusterJsonBin, err := cmd.CombinedOutput()
		if err != nil {
			slogger.Fatalf("%v\n", err)
		}
		slogger.Infof("Successfully retrieved ocp payload details, now cloning repositories")
		clusterServicesJSON := string(clusterJsonBin)
		if err != nil {
			slogger.Fatalf("Could not get payload details error:(%v);\n Output: %s\n", err, string(clusterServicesJSON))
		}

		services := gjson.Get(clusterServicesJSON, "references.spec.tags").Array()
		clusterVersion := gjson.Get(clusterServicesJSON, "digest").String()
		destdir, err := os.Getwd()
		if err != nil {
			slogger.Errorf("%v\n", err)
		}
		destdir = filepath.Join(destdir, clusterVersion)
		for idx := range services {
			service := services[idx].Map()
			serviceDetails := service["annotations"].Map()
			slogger.Debugf("Cloning repository: %s", serviceDetails["io.openshift.build.source-location"].String())

			// Git clone the repo
			serviceRoot, cloned := utils.RunCloneShell(serviceDetails["io.openshift.build.source-location"].String(), destdir,
				serviceDetails["io.openshift.build.commit.ref"].String(), serviceDetails["io.openshift.build.commit.id"].String())

			if cloned == false {
				continue
			}
			// Now run the source code patching script.
			traceappend.PatchSource(serviceRoot)
		}
	} else {
		fmt.Printf("No arguments provided, exiting gracefully.\n Usage: \n")
		flag.PrintDefaults()
	}
}
