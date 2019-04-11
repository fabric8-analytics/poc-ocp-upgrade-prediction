// Contains the logic for interaction with the PR on Github
package ghpr

import (
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/utils"

	gdf "sourcegraph.com/sourcegraph/go-diff/diff"

	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/serviceparser"
	"go.uber.org/zap"

	"github.com/google/go-github/v24/github"
	"golang.org/x/oauth2"
)

var logger, _ = zap.NewDevelopment()
var sugarLogger = logger.Sugar()

// GetPRPayload uses the GHPR API to get all the data for a pull request from Github.
func GetPRPayload(repoStr string, prId int, gopath string) ([]*gdf.FileDiff, []serviceparser.MetaRepo, string) {
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

	sugarLogger.Info(*pr.DiffURL)
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

	// In the interest of time, just use more space.
	var allDiffs []*gdf.FileDiff
	for _, fileDiff := range fileDiffs {
		if !strings.HasSuffix(fileDiff.OrigName, ".go") &&
			!(fileDiff.OrigName != "/dev/null" && strings.HasSuffix(fileDiff.NewName, ".go")) &&
			!(fileDiff.OrigName == "Gopkg.toml" || fileDiff.OrigName == "Godeps.json" || fileDiff.OrigName == "glide.yaml" || fileDiff.OrigName == "go.mod") {
			sugarLogger.Debugf("Not processing non go source %s -> %s\n", fileDiff.OrigName, fileDiff.NewName)
			continue
		}

		allDiffs = append(allDiffs, fileDiff)
	}

	// Get PR details for cloning.
	fork := serviceparser.MetaRepo{
		Branch:    pr.Head.GetRef(),
		Revision:  pr.Head.GetSHA(),
		URL:       pr.Head.Repo.GetCloneURL(),
		LocalPath: "",
	}

	upstream := serviceparser.MetaRepo{
		Branch:   pr.Base.GetRef(),
		Revision: pr.Base.GetSHA(),
		URL:      pr.Base.Repo.GetCloneURL(),
	}

	// Clone the fork
	fork.LocalPath, _ = utils.RunCloneShell(fork.URL, gopath, fork.Branch, fork.Revision)
	// return the diffs and PR details.
	return allDiffs, []serviceparser.MetaRepo{fork, upstream}, fork.LocalPath
}
