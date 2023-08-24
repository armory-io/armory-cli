package sourceControl

import (
	"context"
	"errors"
	"fmt"
	gh "github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"os"
	"strconv"
	"strings"
)

const (
	ghToken           = "GH_TOKEN"
	ghRepo            = "GITHUB_REPOSITORY"
	ghRefName         = "GITHUB_REF_NAME"
	ghActor           = "GITHUB_ACTOR"
	ghSha             = "GITHUB_SHA"
	ghEvent           = "GITHUB_EVENT_NAME"
	ghRef             = "GITHUB_REF"
	ghTriggeringActor = "GITHUB_TRIGGERING_ACTOR"
	ghRefType         = "GITHUB_REF_TYPE"
	ghServer          = "GITHUB_SERVER_URL"
	ghRunId           = "GITHUB_RUN_ID"
	ghWorkflow        = "GITHUB_WORKFLOW"
)

type GithubScmc struct {
	BaseScmc
	Github GithubData `json:"github,omitempty"`
}

type GithubData struct {
	RunId    string `json:"runId,omitempty"`
	Workflow string `json:"workflow,omitempty"`
}

func (scm GithubScmc) GetContext() (ScmContext, error) {
	var err error
	githubScmc := GithubScmc{
		BaseScmc: BaseScmc{
			Type:                github,
			Event:               Event(os.Getenv(ghEvent)),
			Reference:           Reference(os.Getenv(ghRefType)),
			ReferenceName:       os.Getenv(ghRefName),
			Principal:           os.Getenv(ghActor),
			TriggeringPrincipal: os.Getenv(ghTriggeringActor),
			Sha:                 os.Getenv(ghSha),
			Repository:          os.Getenv(ghRepo),
			Server:              os.Getenv(ghServer)},
		Github: GithubData{
			RunId:    os.Getenv(ghRunId),
			Workflow: os.Getenv(ghWorkflow),
		}}

	token, enabled := os.LookupEnv(ghToken)
	if !enabled || token == "" {
		return githubScmc, errors.New("scm is enabled and the GH_TOKEN is missing or empty")
	}

	client := getGithubClient(token)
	var pr gh.PullRequest
	pr, err = getPrInfo(client)

	if err != nil {
		return githubScmc, err
	}

	githubScmc.Source = pr.GetHead().GetRef()
	githubScmc.Target = pr.GetBase().GetRef()
	githubScmc.PrTitle = pr.GetTitle()
	githubScmc.PrUrl = getURL(pr.GetNumber())

	return githubScmc, err
}

func getURL(number int) string {
	url := fmt.Sprintf("%s/%s/pull/%d", os.Getenv(ghServer), os.Getenv(ghRepo), number)
	return url
}

func getPrInfo(client *gh.Client) (gh.PullRequest, error) {

	var pull *gh.PullRequest

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
		options := &gh.PullRequestListOptions{State: "closed"}
		pull, err = searchForPr(ctx, client, options)
		if pull == nil {
			options.State = "open"
			pull, err = searchForPr(ctx, client, options)
		}
	}

	return *pull, err
}

func searchForPr(ctx context.Context, client *gh.Client, options *gh.PullRequestListOptions) (*gh.PullRequest, error) {
	sha := os.Getenv(ghSha)
	repo := os.Getenv(ghRepo)
	splitRepo := strings.Split(repo, "/")
	owner, repoName := splitRepo[0], splitRepo[1]

	var pullRequests []*gh.PullRequest
	var err error

	pullRequests, _, err = client.PullRequests.List(ctx, owner, repoName, options)
	for _, pr := range pullRequests {
		if pr.GetMergeCommitSHA() == sha {
			return pr, nil
		}
	}
	return nil, err
}

func getGithubClient(token string) *gh.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	return gh.NewClient(tc)
}
