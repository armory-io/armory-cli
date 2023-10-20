package sourceControl

import (
	de "github.com/armory-io/deploy-engine/pkg/api"
	gh "github.com/google/go-github/github"
)

type (
	MockSmc struct {
		de.SCM
	}

	MockServiceProvider struct {
		service GithubService
	}

	MockGithubClient struct {
		pullRequest gh.PullRequest
	}
)

func (m MockServiceProvider) GetGithubService() GithubService {
	return m.service
}

func (m MockGithubClient) init(token string) {

}

func (m MockGithubClient) getPR(owner string, repo string, number int) (*gh.PullRequest, error) {
	return &m.pullRequest, nil
}

func (m MockGithubClient) searchForPr(options *gh.PullRequestListOptions) (*gh.PullRequest, error) {
	return &m.pullRequest, nil
}
