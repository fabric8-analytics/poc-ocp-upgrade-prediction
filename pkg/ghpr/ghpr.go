// Contains the logic for interaction with the PR on Github
package ghpr

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/v24/github"
	"golang.org/x/oauth2"
)

// Get a list of all the commits made in a PR
func GetPRCommits() {
	ghPrToken := os.Getenv("GH_TOKEN")

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: ghPrToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	// list all repositories for the authenticated user
	// TODO
	repos, _, _ := client.Repositories.List(ctx, "", nil)
	fmt.Printf("%v\n", repos)
}
