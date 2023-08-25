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
	token           = "gho_mocktoken"
	sourceBranch    = "featureBranch"
	targetBranch    = "mainBranch"
	title           = "test pull_request"
	number          = 1
	repo            = "Armory/repo"
	server          = "https://github.com"
	refName         = "RefName"
	actor           = "Principal"
	sha             = "d0e3572f5462150287865ec29a7ad3a9953dd509"
	event           = "mock_request"
	ref             = "Ref"
	triggeringActor = "triggeringPrincipal"
	refType         = "RefType"
	runId           = "000001"
	workflow        = "Workflow"
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
		expectedSCMC      func() Context
	}{{
		name: "Missing GH_TOKEN",
		setup: func() {
			setGithubEnv()
			os.Setenv(ghToken, "")
		},
		provider:          DefaultServiceProvider{},
		expectErrContains: "GH_TOKEN",
		expectedSCMC:      getGithubMockContext},
		{
			name:         "Happy path",
			setup:        setGithubEnv,
			provider:     MockServiceProvider{service: DefaultGithubService{client: MockGithubClient{pullRequest: testPullRequest}}},
			expectedSCMC: getGithubMockContextWithPR},
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

			assert.Equal(t, c.expectedSCMC(), scmc)
		})
	}

}
func setGithubEnv() {
	token = setOrRetrieveEnv(ghToken, token)
	repo = setOrRetrieveEnv(ghRepo, repo)
	refName = setOrRetrieveEnv(ghRefName, refName)
	actor = setOrRetrieveEnv(ghActor, actor)
	sha = setOrRetrieveEnv(ghSha, sha)
	event = setOrRetrieveEnv(ghEvent, event)
	ref = setOrRetrieveEnv(ghRef, ref)
	triggeringActor = setOrRetrieveEnv(ghTriggeringActor, triggeringActor)
	refType = setOrRetrieveEnv(ghRefType, refType)
	server = setOrRetrieveEnv(ghServer, server)
	runId = setOrRetrieveEnv(ghRunId, runId)
	workflow = setOrRetrieveEnv(ghWorkflow, workflow)
}

func getBaseContext() BaseContext {
	return BaseContext{
		Type:                github,
		Event:               Event(event),
		Reference:           Reference(refType),
		ReferenceName:       refName,
		Principal:           actor,
		TriggeringPrincipal: triggeringActor,
		Sha:                 sha,
		Repository:          repo,
		Server:              server,
	}
}

func getGithubMockContext() Context {
	return GithubContext{
		BaseContext: getBaseContext(),
		Github: GithubData{
			RunId:    runId,
			Workflow: workflow,
		}}
}

func getGithubMockContextWithPR() Context {
	context := GithubContext{
		BaseContext: getBaseContext(),
		Github: GithubData{
			RunId:    runId,
			Workflow: workflow,
		}}
	context.Source = sourceBranch
	context.Target = targetBranch
	context.PrTitle = title
	context.PrUrl = fmt.Sprintf("%s/%s/pull/%d", server, repo, number)
	return context
}

func setOrRetrieveEnv(key string, value string) string {
	found, exists := os.LookupEnv(key)
	if exists && found != "" {
		return found
	}
	os.Setenv(key, value)
	return value
}
