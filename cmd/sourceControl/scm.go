package sourceControl

import (
	"context"
	"fmt"
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
