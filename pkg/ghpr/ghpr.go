// Contains the logic for interaction with the PR on Github
package ghpr

import (
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	gdf "sourcegraph.com/sourcegraph/go-diff/diff"

	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/serviceparser"
	"go.uber.org/zap"

	"github.com/google/go-github/v24/github"
	"golang.org/x/oauth2"
)

var logger, _ = zap.NewDevelopment()
var sugarLogger = logger.Sugar()

// MetaRepo contains all the fields that are required to clone something.
type MetaRepo struct {
	branch   string
	revision string
	URL      string
}

// GetPRPayload uses the GHPR API to get all the data for a pull request from Github.
func GetPRPayload(repoStr string, prId int, gopath string) ([][]*gdf.Hunk, []MetaRepo) {
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

	var allHunks [][]*gdf.Hunk
	for _, fileDiff := range fileDiffs {
		if !strings.HasSuffix(fileDiff.OrigName, ".go") {
			sugarLogger.Debugf("Not processing non go source %s\n", fileDiff.OrigName)
			continue
		}
		hunks := fileDiff.Hunks
		allHunks = append(allHunks, hunks)
	}

	// Get PR details for cloning.
	fork := MetaRepo{
		branch:   pr.Head.GetRef(),
		revision: pr.Head.GetSHA(),
		URL:      pr.Head.Repo.GetCloneURL(),
	}

	upstream := MetaRepo{
		branch:   pr.Base.GetRef(),
		revision: pr.Base.GetSHA(),
		URL:      pr.Base.Repo.GetCloneURL(),
	}

	// return the diffs and PR details.
	return allHunks, []MetaRepo{fork, upstream}
}
