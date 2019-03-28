package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/ghpr"
	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/gremlin"
	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/runtimelogs"
	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/serviceparser"
	"go.uber.org/zap"
)

var logger, _ = zap.NewDevelopment()
var sugar = logger.Sugar()

type PRPayload struct {
	PrID    int    `json:"pr_id"`
	RepoURL string `json:"repo_url"`
}

func processPR(w http.ResponseWriter, r *http.Request) {
	// Read body
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Unmarshal
	var pr PRPayload
	err = json.Unmarshal(b, &pr)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Get the PR hunks, details of base and fork and the clonePath where the fork has been cloned.
	hunks, branchDetails, clonePath := ghpr.GetPRPayload(pr.RepoURL, pr.PrID, "/tmp")

	// ParseService called to parse and populate all the arrays in serviceparser.
	serviceparser.ParseService("machine-config-controller", clonePath)

	// Run the E2E tests on the cloned fork and write results to file.
	logFileE2E := runtimelogs.RunE2ETestsInGoPath(clonePath, "/tmp")

	// Parse the file to generate condepaths and add the corresponding results to graph.
	// TODO: Map service name from git path back to name
	gremlin.AddRuntimePathsToGraph("machine-config-controller",
		branchDetails[1].Revision, runtimelogs.CreateRuntimePaths(strings.Split(logFileE2E, "\n")))

	touchPoints := serviceparser.GetTouchPointsOfPR(hunks, branchDetails)
	response := gremlin.GetTouchPointCoverage(touchPoints)
	output, err := json.Marshal(response)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.Write(output)
}

func main() {
	http.HandleFunc("/", processPR)
	address := ":8080"
	log.Println("Starting server on address", address)
	err := http.ListenAndServe(address, nil)
	if err != nil {
		panic(err)
	}
}
