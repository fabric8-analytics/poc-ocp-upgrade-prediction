// Contains the logic for interaction with the PR on Github
package ghpr

import (
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/serviceparser"
	"go.uber.org/zap"

	"github.com/google/go-github/v24/github"
	"golang.org/x/oauth2"
)

var logger, _ = zap.NewDevelopment()
var sugarLogger = logger.Sugar()

// GetPRPayload uses the GHPR API to get all the data for a pull request from Github.
func GetPRPayload(repoStr string, prId int) {
	ghPrToken := os.Getenv("GH_TOKEN")

	if ghPrToken == "" {
		sugarLogger.Fatalf("Cannot connect to Github API without token.")
	}
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: ghPrToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	ownerRepo := strings.Split(repoStr, "/")
	pr, _, err := client.PullRequests.Get(ctx, ownerRepo[0], ownerRepo[1], prId)

	if err != nil {
		sugarLogger.Fatalf("%v\n", err)
	}
	diffStr, err := http.Get(*pr.DiffURL)

	if err != nil {
		sugarLogger.Error("Could not get diff, for repo: %s, PR: %d\n", repoStr, prId)
	}

	diff, err := ioutil.ReadAll(diffStr.Body)
	if err != nil {
		sugarLogger.Fatalf("Could not parse diff for PR: %d\n", prId)
	}

	fileDiffs, err := serviceparser.ParseDiff(string(diff))

	if err != nil {
		sugarLogger.Errorf("Unable to parse diff, got error: %v\n", err)
	}

	for _, fileDiff := range fileDiffs {
		if !strings.HasSuffix(fileDiff.OrigName, ".go") {
			sugarLogger.Debugf("Not processing non go source %s\n", fileDiff.OrigName)
			continue
		}
		hunks := fileDiff.Hunks

		for _, hunk := range hunks {
			// Do something with the fileDiffs.
			sugarLogger.Debugf("%v\n", hunk)
		}
	}
}
