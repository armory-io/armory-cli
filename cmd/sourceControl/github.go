package sourceControl

import (
	"context"
	"fmt"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"os"
)

const (
	GhToken  = "GH_TOKEN"
	ghRepo   = "GITHUB_REPOSITORY"
	ghServer = "GITHUB_SERVER_URL"
	ghActor  = "GITHUB_ACTOR"
	ghSha    = "GITHUB_SHA"
	ghEvent  = "GITHUB_EVENT_NAME"
	ghRef    = "GITHUB_REF"
)

func GetGhContext(token string) ScmContext {
	var scmc ScmContext
	client := getGithubClient(token)

	scmc.scm = gitHub
	scmc.actor = os.Getenv(ghActor)
	scmc.sha = os.Getenv(ghSha)
	scmc.prTitle, scmc.prUrl = getPrInfo(client)

	return scmc
}
func getPrInfo(client *github.Client) (string, string) {

	var url string
	var title string

	server := os.Getenv(ghServer)
	repo := os.Getenv(ghRepo)
	reference := os.Getenv(ghRef)

	event := Event(os.Getenv(ghActor))

	if event == pullRequest {
		// we star index at 4 to remove the refs word
		url = fmt.Sprintf("%s/%s%s", server, repo, reference[4:])
	}

	return title, url
}
func getGithubClient(token string) *github.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}
