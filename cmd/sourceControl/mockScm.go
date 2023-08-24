package sourceControl

import gh "github.com/google/go-github/github"

type (
	MockSmc struct {
		BaseContext
	}

	MockServiceProvider struct {
		pullRequest gh.PullRequest
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

func GetEmptyMockScmc() Context {
	scmc, _ := MockSmc{}.GetContext()
	return scmc
}

func (mock MockServiceProvider) GetGithubService() GithubService {
	return DefaultGithubService{client: MockGithubClient{pullRequest: mock.pullRequest}}
}

func (m MockGithubClient) init(token string) {

}

func (m MockGithubClient) getPR(owner string, repo string, number int) (*gh.PullRequest, error) {
	return &m.pullRequest, nil
}

func (m MockGithubClient) searchForPr(options *gh.PullRequestListOptions) (*gh.PullRequest, error) {
	return &m.pullRequest, nil
}
