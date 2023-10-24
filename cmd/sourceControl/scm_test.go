package sourceControl

import (
	"bytes"
	"fmt"
	de "github.com/armory-io/deploy-engine/pkg/api"
	gh "github.com/google/go-github/github"
	"github.com/stretchr/testify/assert"
	"io"
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
		expectedSCMC      func() de.SCM
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

func TestFileScm(t *testing.T) {
	cases := []struct {
		name              string
		setup             func(file string) string
		test              func(writer io.Writer) (de.GenericSCM, error)
		expectErrContains string
		expectedSCMC      de.SCM
	}{
		{
			name:              "Invalid path",
			expectErrContains: "no such file",
			test: func(writer io.Writer) (de.GenericSCM, error) {
				return GetContextFromFile(writer, "./fakePath.yaml")
			},
			expectedSCMC: de.SCM{},
		},
		{
			name: "SCM file only",
			test: func(writer io.Writer) (de.GenericSCM, error) {
				return Unmarshall([]byte(testGenericSCMFile))
			},
			expectedSCMC: getFileMockContext(),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			writer := bytes.NewBufferString("")
			scmc := de.SCM{}
			var actualError error
			scmc.GenericSCM, actualError = c.test(writer)
			if c.expectErrContains != "" {
				assert.ErrorContains(t, actualError, c.expectErrContains)
			} else {
				assert.NoError(t, actualError)
			}

			assert.Equal(t, c.expectedSCMC, scmc)
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
		GithubSCM: de.GithubSCM{
			Type:                de.Github,
			Event:               de.Event(event),
			Reference:           de.Reference(refType),
			ReferenceName:       refName,
			Principal:           actor,
			TriggeringPrincipal: triggeringActor,
			SHA:                 sha,
			Repository:          repo,
			Server:              server,
		},
	}
}

func getGithubMockContext() de.SCM {

	scmc := getBaseContext()
	scmc.GithubData = de.GithubData{
		RunId:    runId,
		Workflow: workflow,
	}
	return scmc
}

func getGithubMockContextWithPR() de.SCM {
	scmc := getGithubMockContext()
	scmc.GithubSCM.Source = sourceBranch
	scmc.GithubSCM.Target = targetBranch
	scmc.GithubSCM.PRTitle = title
	scmc.GithubSCM.PRUrl = fmt.Sprintf("%s/%s/pull/%d", server, repo, number)
	return scmc
}

func getGithubMockContextNoRepo() de.SCM {
	scmc := getGithubMockContext()
	scmc.GithubSCM.Repository = ""
	return scmc
}

func getFileMockContext() de.SCM {
	scmc := de.SCM{
		GenericSCM: de.GenericSCM{
			Type:          "jenkins",
			Event:         "push",
			Reference:     "refs/heads/main",
			ReferenceName: "5/merge",
			Source:        "feat/something-cool",
			SourceUrl:     "http://urlto/feat/something-cool",
			Target:        "main",
			TargetUrl:     "http://urlto/main",
			Principal:     "donquixote",
			PrincipalUrl:  "donquixote",
			PrTitle:       "feat: modified configuration - use / as ui root",
			PrUrl:         "https://urlto/pr/8",
			Sha:           "fe3540e4de2ac32312d86dd1b7a8e0d10d7b810b",
			ShaUrl:        "https://urlto/fe3540e4de2ac32312d86dd1b7a8e0d10d7b810b",
			Workflow:      "Workflow Run",
			WorkflowUrl:   "http://urlto/someworkflow",
			Tag:           "sometag",
			TagUrl:        "http://urlto/sometag",
		},
	}
	return scmc
}

func setOrRetrieveEnv(key string, value string) string {
	found, exists := os.LookupEnv(key)
	if exists && found != "" {
		return found
	}
	os.Setenv(key, value)
	return value
}

const testGenericSCMFile = `
type:          "jenkins"
event:         "push"
reference:     "refs/heads/main"
referenceName: "5/merge"
source:        "feat/something-cool"
sourceUrl:     "http://urlto/feat/something-cool"
target:        "main"
targetUrl:     "http://urlto/main"
principal:     "donquixote"
principalUrl:  "donquixote"
prTitle:       "feat: modified configuration - use / as ui root"
prUrl:         "https://urlto/pr/8"
sha:           "fe3540e4de2ac32312d86dd1b7a8e0d10d7b810b"
shaUrl:        "https://urlto/fe3540e4de2ac32312d86dd1b7a8e0d10d7b810b"
workflow:      "Workflow Run"
workflowUrl:   "http://urlto/someworkflow"
tag:           "sometag"
tagUrl:        "http://urlto/sometag"`
