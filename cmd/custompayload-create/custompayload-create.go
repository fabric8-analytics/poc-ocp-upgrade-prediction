// Clusterpatcher generates a utility binary that can be used to create a new cluster image with
// all the service sources patched.
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	flag "github.com/spf13/pflag"

	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/imageutils"

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
	noImages := flag.Bool("no-images", false, "Whether container images need to be built")
	username := flag.String("user-name", "", "The Github user name of the user whose token is set for GH_TOKEN")
	destdirPtr := flag.String("destdir", "", "The directory where the repositories need to be cloned.")
	// noClone = flag.String("no-clone", false, "Whether clones already exist in [DESTDIR]")

	flag.Parse()
	if os.Getenv("GH_TOKEN") == "" {
		slogger.Fatalf("Need a Github token for running this script to go around rate limits.")
	}
	if *username == "" {
		slogger.Fatalf("Need a username for using the supplied GH_TOKEN.")
	}
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

		clusterJSONBin, err := cmd.CombinedOutput()
		if err != nil {
			slogger.Fatalf("%v\n", err)
		}
		slogger.Infof("Successfully retrieved ocp payload details, now cloning repositories")
		clusterServicesJSON := string(clusterJSONBin)
		if err != nil {
			slogger.Fatalf("Could not get payload details error:(%v);\n Output: %s\n", err, string(clusterServicesJSON))
		}

		services := gjson.Get(clusterServicesJSON, "references.spec.tags").Array()
		clusterVersion := gjson.Get(clusterServicesJSON, "digest").String()

		var destdir string
		if *destdirPtr != "" {
			destdir = *destdirPtr
		} else {
			destdir, err = os.Getwd()
			if err != nil {
				slogger.Errorf("%v\n", err)
			}
		}
		destdir = filepath.Join(destdir, clusterVersion)
		for idx := range services {
			service := services[idx].Map()
			serviceName := service["name"].String()
			serviceDetails := service["annotations"].Map()
			slogger.Debugf("Cloning repository: %s", serviceDetails["io.openshift.build.source-location"].String())
			slogger.Debugf("%v\n", serviceDetails)

			// Git clone the repo, with a token
			ghToken := os.Getenv("GH_TOKEN")
			if serviceDetails["io.openshift.build.source-location"].String() == "" {
				// This shouldn't be required but there's payloads that don't have this in the CI.
				continue
			}
			cloneURLParts := strings.Split(serviceDetails["io.openshift.build.source-location"].String(), "://")
			serviceRoot, _ := utils.RunCloneShell(fmt.Sprintf("%s://%s:%s@%s", cloneURLParts[0], *username, ghToken, cloneURLParts[1]), destdir,
				serviceDetails["io.openshift.build.commit.ref"].String(), serviceDetails["io.openshift.build.commit.id"].String())

			// Now run the source code patching script.
			traceappend.PatchSource(serviceRoot)

			if *noImages {
				continue
			}
			// Get all the Dockerfiles in the service
			matches, err := filepath.Glob(filepath.Join(serviceRoot, "Dockerfile*"))
			if err != nil {
				slogger.Errorf("%v\n", err)
			}

			created := 0

			for _, match := range matches {
				// Now creeate the docker image of the patched source code
				if !strings.Contains(match, "rhel") && strings.Contains(match, serviceName) {
					imageutils.CreateImage("quay.io", serviceName, match)
					created = 1
					break
				}
			}

			if created == 0 {
				_, err := os.Stat(filepath.Join(serviceRoot, "Dockerfile"))
				if err == nil {
					imageutils.CreateImage("quay.io", serviceName, filepath.Join(serviceRoot, "Dockerfile"))
				} else {
					slogger.Errorf("Not creating any image since we did not find a suitable dockerfile.")
				}
			}
		}
	} else {
		fmt.Printf("No arguments provided, exiting gracefully.\n Usage: \n")
		flag.PrintDefaults()
	}
}
