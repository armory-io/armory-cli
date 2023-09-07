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
				Type:                de.Github,
				Event:               de.Event(os.Getenv(de.GithubEvent)),
				Reference:           de.Reference(os.Getenv(de.GithubRefType)),
				ReferenceName:       os.Getenv(de.GithubRefName),
				Principal:           os.Getenv(de.GithubActor),
				TriggeringPrincipal: os.Getenv(de.GithubTriggeringActor),
				SHA:                 os.Getenv(de.GithubSHA),
				Repository:          os.Getenv(de.GithubRepo),
				Server:              os.Getenv(de.GithubServer)},
			GithubData: de.GithubData{
				RunId:    os.Getenv(de.GithubRunID),
				Workflow: os.Getenv(de.GithubWorkflow),
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
	url := fmt.Sprintf("%s/%s/pull/%d", os.Getenv(de.GithubServer), os.Getenv(de.GithubRepo), number)
	return url
}

func (gp DefaultGithubService) GetPR() (gh.PullRequest, error) {

	var pull *gh.PullRequest
	var prNumber int
	var err error

	token, enabled := os.LookupEnv(de.GithubToken)
	if !enabled || token == "" {
		return gh.PullRequest{}, errors.New("scm is enabled and the GH_TOKEN is missing or empty")
	}

	gp.client.init(token)

	reference := os.Getenv(de.GithubRef)
	event := de.Event(os.Getenv(de.GithubEvent))

	repo := os.Getenv(de.GithubRepo)
	err = checkNotEmpty(reference, repo)
	if err != nil {
		return gh.PullRequest{}, err
	}

	splitRepo := strings.Split(repo, "/")
	owner, repoName := splitRepo[0], splitRepo[1]

	if event == de.PullRequest {
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
	if pull == nil {
		pull = &gh.PullRequest{}
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
	var pullRequests []*gh.PullRequest
	var err error

	sha := os.Getenv(de.GithubSHA)
	repo := os.Getenv(de.GithubRepo)

	err = checkNotEmpty(sha, repo)
	if err != nil {
		return &gh.PullRequest{}, err
	}
	splitRepo := strings.Split(repo, "/")
	owner, repoName := splitRepo[0], splitRepo[1]

	pullRequests, _, err = d.client.PullRequests.List(d.ctx, owner, repoName, options)
	for _, pr := range pullRequests {
		if pr.GetMergeCommitSHA() == sha {
			return pr, nil
		}
	}
	return nil, err
}

func checkNotEmpty(values ...string) error {
	for _, value := range values {
		if value == "" {
			return errors.New("missing environment properties")
		}
	}
	return nil
}
