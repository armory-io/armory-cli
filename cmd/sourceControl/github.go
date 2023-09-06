package sourceControl

import (
	"context"
	"errors"
	"fmt"
	de "github.com/armory-io/deploy-engine/pkg/api"
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

type (
	GithubContext struct {
		de.GithubSCM
		service GithubService
	}

	GithubService interface {
		GetPR() (gh.PullRequest, error)
	}
	GithubClient interface {
		getPR(owner string, repo string, number int) (*gh.PullRequest, error)
		searchForPr(options *gh.PullRequestListOptions) (*gh.PullRequest, error)
		init(token string)
	}
	DefaultGithubService struct {
		client GithubClient
	}

	DefaultGithubClient struct {
		client gh.Client
		ctx    context.Context
	}
)

func (gc GithubContext) GetContext() (Context, error) {
	githubContext := GithubContext{
		GithubSCM: de.GithubSCM{
			SCM: de.SCM{
				Type:                github,
				Event:               de.Event(os.Getenv(ghEvent)),
				Reference:           de.Reference(os.Getenv(ghRefType)),
				ReferenceName:       os.Getenv(ghRefName),
				Principal:           os.Getenv(ghActor),
				TriggeringPrincipal: os.Getenv(ghTriggeringActor),
				SHA:                 os.Getenv(ghSha),
				Repository:          os.Getenv(ghRepo),
				Server:              os.Getenv(ghServer)},
			GithubData: de.GithubData{
				RunId:    os.Getenv(ghRunId),
				Workflow: os.Getenv(ghWorkflow),
			}}}

	pr, err := gc.service.GetPR()

	if err != nil {
		return githubContext, err
	}

	githubContext.Source = pr.GetHead().GetRef()
	githubContext.Target = pr.GetBase().GetRef()
	githubContext.PRTitle = pr.GetTitle()
	githubContext.PRUrl = getURL(pr.GetNumber())

	return githubContext, err
}

func getURL(number int) string {
	url := fmt.Sprintf("%s/%s/pull/%d", os.Getenv(ghServer), os.Getenv(ghRepo), number)
	return url
}

func (gp DefaultGithubService) GetPR() (gh.PullRequest, error) {

	var pull *gh.PullRequest
	var prNumber int
	var err error

	token, enabled := os.LookupEnv(ghToken)
	if !enabled || token == "" {
		return gh.PullRequest{}, errors.New("scm is enabled and the GH_TOKEN is missing or empty")
	}

	gp.client.init(token)

	reference := os.Getenv(ghRef)
	event := de.Event(os.Getenv(ghEvent))

	repo := os.Getenv(ghRepo)
	splitRepo := strings.Split(repo, "/")
	owner, repoName := splitRepo[0], splitRepo[1]

	if event == pullRequest {
		prNumber, _ = strconv.Atoi(strings.Split(reference, "/")[2])
		pull, err = gp.client.getPR(owner, repoName, prNumber)
	} else {
		options := &gh.PullRequestListOptions{State: "closed"}
		pull, err = gp.client.searchForPr(options)
		if pull == nil {
			options.State = "open"
			pull, err = gp.client.searchForPr(options)
		}
	}

	return *pull, err
}

func (d *DefaultGithubClient) init(token string) {
	ctx := d.ctx
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	d.client = *gh.NewClient(tc)
}

func (d *DefaultGithubClient) getPR(owner string, repo string, number int) (*gh.PullRequest, error) {
	pr, _, err := d.client.PullRequests.Get(d.ctx, owner, repo, number)
	return pr, err
}

func (d *DefaultGithubClient) searchForPr(options *gh.PullRequestListOptions) (*gh.PullRequest, error) {
	sha := os.Getenv(ghSha)
	repo := os.Getenv(ghRepo)
	splitRepo := strings.Split(repo, "/")
	owner, repoName := splitRepo[0], splitRepo[1]

	var pullRequests []*gh.PullRequest
	var err error

	pullRequests, _, err = d.client.PullRequests.List(d.ctx, owner, repoName, options)
	for _, pr := range pullRequests {
		if pr.GetMergeCommitSHA() == sha {
			return pr, nil
		}
	}
	return nil, err
}
