package sourceControl

import gh "github.com/google/go-github/github"

type (
	MockSmc struct {
		BaseContext
	}

	MockServiceProvider struct {
		service GithubService
	}

	MockGithubClient struct {
		pullRequest gh.PullRequest
	}
)

func (mock MockSmc) GetContext() (Context, error) {
	scmc := MockSmc{
		BaseContext: BaseContext{
			Type: "mock"}}
	return scmc, nil
}

func GetEmptyMockSCMC() Context {
	scmc, _ := MockSmc{}.GetContext()
	return scmc
}

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
