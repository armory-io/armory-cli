package sourceControl

import (
	"bytes"
	"fmt"
	gh "github.com/google/go-github/github"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var (
	sourceBranch    = "featureBranch"
	targetBranch    = "mainBranch"
	title           = "test pull_request"
	number          = 1
	principal       = "Armory"
	repo            = "Armory/repo"
	server          = "https://github.com"
	testPullRequest = gh.PullRequest{
		Base:   &gh.PullRequestBranch{Ref: &targetBranch},
		Head:   &gh.PullRequestBranch{Ref: &sourceBranch},
		Title:  &title,
		Number: &number}
)

func TestScm(t *testing.T) {
	cases := []struct {
		name              string
		setup             func()
		provider          ServiceProvider
		expectErrContains string
		expectedScmc      Context
	}{{
		name: "Missing GH_TOKEN",
		setup: func() {
			os.Setenv(ghActor, principal)
		},
		provider:          DefaultServiceProvider{},
		expectErrContains: "GH_TOKEN",
		expectedScmc: GithubContext{
			BaseContext: BaseContext{
				Type:      github,
				Principal: principal}}},
		{
			name: "Happy path",
			setup: func() {
				os.Setenv(ghToken, "gho_mocktoken")
				os.Setenv(ghActor, principal)
				os.Setenv(ghServer, server)
				os.Setenv(ghRepo, repo)
			},
			provider: MockServiceProvider{
				pullRequest: testPullRequest},
			expectedScmc: GithubContext{
				BaseContext: BaseContext{
					Type:       github,
					Principal:  principal,
					Repository: repo,
					Server:     server,
					Source:     sourceBranch,
					Target:     targetBranch,
					PrTitle:    title,
					PrUrl:      fmt.Sprintf("%s/%s/pull/%d", server, repo, number),
				}}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			writer := bytes.NewBufferString("")
			c.setup()
			scmc, actualError := RetrieveContext(writer, c.provider)
			if c.expectErrContains != "" {
				assert.ErrorContains(t, actualError, c.expectErrContains)
			} else {
				assert.NoError(t, actualError)
			}
			assert.Equal(t, c.expectedScmc, scmc)
		})
	}

}
