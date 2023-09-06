package sourceControl

import (
	"context"
	"fmt"
	de "github.com/armory-io/deploy-engine/pkg/api"
	"github.com/fatih/color"
	"io"
)

type (
	Context interface {
		GetContext() (Context, error)
	}

	ServiceProvider interface {
		GetGithubService() GithubService
	}

	DefaultServiceProvider struct {
		Ctx context.Context
	}
)

const (
	github           de.Manager   = "github"
	bitbucket        de.Manager   = "bitbucket"
	branch           de.Reference = "branch"
	tagRef           de.Reference = "tag"
	pullRequest      de.Event     = "pull_request"
	push             de.Event     = "push"
	workflowDispatch de.Event     = "workflow_dispatch"
)

func RetrieveContext(out io.Writer, provider ServiceProvider) (Context, error) {
	var scmc Context
	var err error

	scmc, err = GithubContext{service: provider.GetGithubService()}.GetContext()

	if err != nil {
		msg := color.New(color.FgYellow, color.Bold).Sprint("scm error: ")
		fmt.Fprintf(out, "%s %s\n\n", msg, err)
	}

	return scmc, err
}

func (d DefaultServiceProvider) GetGithubService() GithubService {
	return DefaultGithubService{client: &DefaultGithubClient{ctx: context.Background()}}
}
