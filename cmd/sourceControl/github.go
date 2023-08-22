package sourceControl

import (
	"context"
	"fmt"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"os"
	"strconv"
	"strings"
)

const (
	GhToken           = "GH_TOKEN"
	ghRepo            = "GITHUB_REPOSITORY"
	ghRefName         = "GITHUB_REF_NAME"
	ghActor           = "GITHUB_ACTOR"
	ghSha             = "GITHUB_SHA"
	ghEvent           = "GITHUB_EVENT_NAME"
	ghRef             = "GITHUB_REF"
	ghTriggeringActor = "GITHUB_TRIGGERING_ACTOR"
	ghRefType         = "GITHUB_REF_TYPE"
	ghServer          = "GITHUB_SERVER_URL"
)

func GetGhContext(token string) (ScmContext, error) {

	client := getGithubClient(token)
	pr, err := getPrInfo(client)

	scmc := ScmContext{
		scm:             gitHub,
		event:           Event(os.Getenv(ghEvent)),
		reference:       Reference(os.Getenv(ghRefType)),
		referenceName:   os.Getenv(ghRefName),
		source:          pr.GetBase().GetRef(),
		target:          pr.GetHead().GetRef(),
		actor:           os.Getenv(ghActor),
		triggeringActor: os.Getenv(ghTriggeringActor),
		sha:             os.Getenv(ghSha),
		repository:      os.Getenv(ghRepo),
		prTitle:         pr.GetTitle(),
		prUrl:           getURL(pr.GetNumber())}

	return scmc, err
}
func getURL(number int) string {
	url := fmt.Sprintf("%s/%s/pull/%d", os.Getenv(ghServer), os.Getenv(ghRepo), number)
	return url
}

func getPrInfo(client *github.Client) (github.PullRequest, error) {

	var pull *github.PullRequest

	var prNumber int
	var err error

	ctx := context.Background()

	reference := os.Getenv(ghRef)
	event := Event(os.Getenv(ghEvent))

	repo := os.Getenv(ghRepo)
	splitRepo := strings.Split(repo, "/")
	owner, repoName := splitRepo[0], splitRepo[1]

	if event == pullRequest {
		prNumber, err = strconv.Atoi(strings.Split(reference, "/")[2])
		pull, _, err = client.PullRequests.Get(ctx, owner, repoName, prNumber)
	} else {
		options := &github.PullRequestListOptions{State: "closed"}
		pull, err = searchForPr(ctx, client, options)
		if pull == nil {
			options.State = "open"
			pull, err = searchForPr(ctx, client, options)
		}
	}

	return *pull, err
}

func searchForPr(ctx context.Context, client *github.Client, options *github.PullRequestListOptions) (*github.PullRequest, error) {
	sha := os.Getenv(ghSha)
	repo := os.Getenv(ghRepo)
	splitRepo := strings.Split(repo, "/")
	owner, repoName := splitRepo[0], splitRepo[1]

	var pullRequests []*github.PullRequest
	var err error

	pullRequests, _, err = client.PullRequests.List(ctx, owner, repoName, options)
	for _, pr := range pullRequests {
		if pr.GetMergeCommitSHA() == sha {
			return pr, nil
		}
	}
	return nil, err
}

func getGithubClient(token string) *github.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}
