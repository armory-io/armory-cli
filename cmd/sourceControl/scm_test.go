package sourceControl

import (
	"bytes"
	"fmt"
	de "github.com/armory-io/deploy-engine/pkg/api"
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
			os.Setenv(de.GithubToken, "")
		},
		provider:          DefaultServiceProvider{},
		expectErrContains: "GH_TOKEN",
		expectedSCMC:      getGithubMockContext},
		{
			name:         "Happy path",
			setup:        setGithubEnv,
			provider:     MockServiceProvider{service: DefaultGithubService{client: MockGithubClient{pullRequest: testPullRequest}}},
			expectedSCMC: getGithubMockContextWithPR},
		{
			name: "Missing repo",
			setup: func() {
				setGithubEnv()
				os.Setenv(de.GithubRepo, "")
			},
			provider:          MockServiceProvider{service: DefaultGithubService{client: MockGithubClient{pullRequest: testPullRequest}}},
			expectErrContains: "missing",
			expectedSCMC:      getGithubMockContextNoRepo}}

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
	token = setOrRetrieveEnv(de.GithubToken, token)
	repo = setOrRetrieveEnv(de.GithubRepo, repo)
	refName = setOrRetrieveEnv(de.GithubRefName, refName)
	actor = setOrRetrieveEnv(de.GithubActor, actor)
	sha = setOrRetrieveEnv(de.GithubSHA, sha)
	event = setOrRetrieveEnv(de.GithubEvent, event)
	ref = setOrRetrieveEnv(de.GithubRef, ref)
	triggeringActor = setOrRetrieveEnv(de.GithubTriggeringActor, triggeringActor)
	refType = setOrRetrieveEnv(de.GithubRefType, refType)
	server = setOrRetrieveEnv(de.GithubServer, server)
	runId = setOrRetrieveEnv(de.GithubRunID, runId)
	workflow = setOrRetrieveEnv(de.GithubWorkflow, workflow)
}

func getBaseContext() de.SCM {
	return de.SCM{
		Type:                de.Github,
		Event:               de.Event(event),
		Reference:           de.Reference(refType),
		ReferenceName:       refName,
		Principal:           actor,
		TriggeringPrincipal: triggeringActor,
		SHA:                 sha,
		Repository:          repo,
		Server:              server,
	}
}

func getGithubMockContext() Context {
	return GithubContext{
		GithubSCM: de.GithubSCM{
			SCM: getBaseContext(),
			GithubData: de.GithubData{
				RunId:    runId,
				Workflow: workflow,
			}}}
}

func getGithubMockContextWithPR() Context {
	context := GithubContext{
		GithubSCM: de.GithubSCM{
			SCM: getBaseContext(),
			GithubData: de.GithubData{
				RunId:    runId,
				Workflow: workflow,
			}}}
	context.Source = sourceBranch
	context.Target = targetBranch
	context.PRTitle = title
	context.PRUrl = fmt.Sprintf("%s/%s/pull/%d", server, repo, number)
	return context
}

func getGithubMockContextNoRepo() Context {
	base := getBaseContext()
	base.Repository = ""
	context := GithubContext{
		GithubSCM: de.GithubSCM{
			SCM: base,
			GithubData: de.GithubData{
				RunId:    runId,
				Workflow: workflow,
			}}}
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
