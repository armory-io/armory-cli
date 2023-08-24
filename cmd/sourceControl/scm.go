package sourceControl

import (
	"context"
	"fmt"
	"github.com/fatih/color"
	"io"
)

type (
	Manager   string
	Reference string
	Event     string

	Context interface {
		GetContext() (Context, error)
	}

	ServiceProvider interface {
		GetGithubService() GithubService
	}

	DefaultServiceProvider struct {
		Ctx context.Context
	}

	BaseContext struct {
		Type                Manager   `json:"type,omitempty"`
		Event               Event     `json:"event,omitempty"`
		Reference           Reference `json:"reference,omitempty"`
		ReferenceName       string    `json:"referenceName,omitempty"`
		Source              string    `json:"source,omitempty"`
		Target              string    `json:"target,omitempty"`
		Principal           string    `json:"principal,omitempty"`
		TriggeringPrincipal string    `json:"triggeringPrincipal,omitempty"`
		PrTitle             string    `json:"prTitle,omitempty"`
		PrUrl               string    `json:"prUrl,omitempty"`
		Sha                 string    `json:"sha,omitempty"`
		Repository          string    `json:"repository,omitempty"`
		Server              string    `json:"server,omitempty"`
	}
)

const (
	github           Manager   = "github"
	bitbucket        Manager   = "bitbucket"
	branch           Reference = "branch"
	tagRef           Reference = "tag"
	pullRequest      Event     = "pull_request"
	push             Event     = "push"
	workflowDispatch Event     = "workflow_dispatch"
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
