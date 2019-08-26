package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/ghpr"
	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/gremlin"
	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/serviceparser"

	"go.uber.org/zap"
)

var logger, _ = zap.NewDevelopment()
var sugar = logger.Sugar()

func processPR(w http.ResponseWriter, r *http.Request) {
	// Read body
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Unmarshal
	var pr gremlin.PRPayload
	err = json.Unmarshal(b, &pr)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Get the PR diffs, details of base and fork and the clonePath where the fork has been cloned.
	diffs, branchDetails, _ := ghpr.GetPRPayload(pr.RepoURL, pr.PrID)
	sugar.Infof("%v\n%v\n", diffs, branchDetails)
	// components := serviceparser.NewServiceComponents("machine-config-controller")
	// ParseService called to parse and populate all the arrays in serviceparser.
	// components.ParseService("machine-config-controller", clonePath)

	// Run the E2E tests on the cloned fork and write results to file.
	// logFileE2E := runtimelogs.RunE2ETestsInGoPath(clonePath, "/tmp")

	// Parse the file to generate condepaths and add the corresponding results to graph.
	// TODO: Map service name from git path back to name
	// gremlin.AddComponentRuntimePathsToGraph("machine-config-controller",
	// 	branchDetails[1].Revision, runtimelogs.CreateRuntimePaths(strings.Split(logFileE2E, "\n"), components))

	// touchPoints := serviceparser.GetTouchPointsOfPR(diffs, branchDetails)
	response := map[string]string{
		"status": "processing",
	}
	output, err := json.Marshal(response)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.Write(output)
}

func prConfidenceScore(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	var pr gremlin.PRPayload
	err = json.Unmarshal(b, &pr)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	diffs, branchDetails, prTitle := ghpr.GetPRPayload(pr.RepoURL, pr.PrID)
	touchPoints := serviceparser.GetTouchPointsOfPR(diffs, branchDetails)

	response := gremlin.GetPRConfidenceScore(pr)
	response.PrTitle = prTitle
	response.TouchPoints = *touchPoints
	output, err := json.Marshal(response)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.Write(output)
}

func main() {
	http.HandleFunc("/api/v1/createprnode", processPR)
	http.HandleFunc("/api/v1/getprconfidence", prConfidenceScore)
	address := ":8080"
	log.Println("Starting server on address", address)
	err := http.ListenAndServe(address, nil)
	if err != nil {
		panic(err)
	}
}
