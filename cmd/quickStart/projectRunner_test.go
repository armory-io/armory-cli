package quickStart

import (
	"errors"
	"github.com/armory/armory-cli/pkg/config"
	"github.com/armory/armory-cli/pkg/org"
	"github.com/hashicorp/go-multierror"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"net/url"
	"strings"
	"testing"
)

type MockConfiguration struct {
}

func (c *MockConfiguration) GetArmoryCloudEnv() config.ArmoryCloudEnv {
	return 0
}
func (c *MockConfiguration) GetAuthToken() string {
	return "abc_xyz_token"
}
func (c *MockConfiguration) GetCustomerEnvironmentId() string {
	return ""
}
func (c *MockConfiguration) GetArmoryCloudAddr() *url.URL {
	addr, _ := url.Parse("api.dev.cloud.armory.io")
	return addr
}
func (c *MockConfiguration) GetArmoryCloudEnvironmentConfiguration() *config.ArmoryCloudEnvironmentConfiguration {
	return &config.ArmoryCloudEnvironmentConfiguration{
		CloudConsoleBaseUrl: "console.dev.cloud.armory.io",
	}
}

func TestProjectRunnerTestSuite(t *testing.T) {
	suite.Run(t, new(ProjectRunnerSuite))
}

type ProjectRunnerSuite struct {
	suite.Suite
}

func (suite *ProjectRunnerSuite) SetupSuite() {

}

func (suite *ProjectRunnerSuite) SetupTest() {

}

func (suite *ProjectRunnerSuite) TearDownSuite() {
}

type Mock struct {
	Calls      int
	CalledWith string
}

func NewMock() Mock {
	return Mock{Calls: 0, CalledWith: ""}
}

func (m *Mock) Inc() {
	m.Calls = m.Calls + 1
}
func (m *Mock) Errors() error {
	m.Calls = m.Calls + 1
	return errors.New("some error")
}

func (m *Mock) ErrorsWith(s string) error {
	m.CalledWith = m.CalledWith + ";" + s
	return errors.New("some error: " + s)
}

func (m *Mock) Updates() error {
	m.Calls = m.Calls + 1
	return nil
}

func (m *Mock) UpdatesWith(s string) error {
	m.CalledWith = m.CalledWith + ";" + s
	return nil
}

func (suite *ProjectRunnerSuite) TestRunnerHasError() {
	runner := NewProjectRunner(&MockConfiguration{})
	suite.False(runner.HasErrors(), "Runner should have no errors")
	runner.Errors = &multierror.Error{
		Errors: []error{errors.New("some error")},
	}
	suite.True(runner.HasErrors(), "Runner should have errors")
}

func (suite *ProjectRunnerSuite) TestAppendError() {
	runner := NewProjectRunner(&MockConfiguration{})
	runner.AppendError(errors.New("some error"))

	suite.True(runner.HasErrors(), "Runner should have errors")
}

func (suite *ProjectRunnerSuite) TestExec() {
	runner := NewProjectRunner(&MockConfiguration{})
	mock := NewMock()

	runner.Exec(mock.Updates)
	suite.Equal(1, mock.Calls)

	runner.Exec(mock.Errors)
	suite.Equal(2, mock.Calls)

	//After an error occurs, Updates will not be called again
	runner.Exec(mock.Updates)
	suite.Equal(2, mock.Calls)
}

func (suite *ProjectRunnerSuite) TestExecWith() {
	runner := NewProjectRunner(&MockConfiguration{})
	mock := NewMock()

	runner.ExecWith(mock.UpdatesWith, "update")
	runner.ExecWith(mock.ErrorsWith, "error")
	//After an error occurs, Updates will not be called again
	runner.ExecWith(mock.UpdatesWith, "update")
	suite.Equal(";update;error", mock.CalledWith)
}

func (suite *ProjectRunnerSuite) TestWontFailOnErrorWithoutOne() {
	runner := NewProjectRunner(&MockConfiguration{})
	runner.FailOnError()
}

func (suite *ProjectRunnerSuite) TestFailOnError() {
	runner := NewProjectRunner(&MockConfiguration{})
	defer func() {
		log.StandardLogger().ExitFunc = nil
	}()

	var fatal bool
	log.StandardLogger().ExitFunc = func(int) {
		fatal = true
	}

	runner.AppendError(errors.New("test error"))
	runner.FailOnError()
	suite.True(fatal, "Given there is an error with the runner, a fatal exit is expected")
}

func (suite *ProjectRunnerSuite) TestPopulateAgentsSkips() {
	calls := 0
	og := orgGetAgents
	defer func() { orgGetAgents = og }()

	orgGetAgents = func(ArmoryCloudAddr *url.URL, accessToken string) ([]org.Agent, error) {
		calls = calls + 1
		return []org.Agent{
			{AgentIdentifier: "abc"},
			{AgentIdentifier: "xyz"},
		}, nil
	}

	runner := NewProjectRunner(&MockConfiguration{})
	runner.AppendError(NoAgentsFoundError{})
	runner.PopulateAgents()
	suite.Equal(0, calls, "Expected no calls to get agents if the runner has an error")
}

func (suite *ProjectRunnerSuite) TestPopulateAgentsGetAgentsError() {
	calls := 0
	og := orgGetAgents
	defer func() { orgGetAgents = og }()

	orgGetAgents = func(ArmoryCloudAddr *url.URL, accessToken string) ([]org.Agent, error) {
		calls = calls + 1
		return []org.Agent{
			{AgentIdentifier: "abc"},
			{AgentIdentifier: "xyz"},
		}, errors.New("some API error")
	}

	runner := NewProjectRunner(&MockConfiguration{})
	runner.PopulateAgents()
	suite.True(runner.HasErrors(), "error expected when returned from org.GetAgents")
	suite.Equal(1, calls, "Expected a call to orgGetAgents")
}

func (suite *ProjectRunnerSuite) TestPopulateAgentsSuccess() {
	calls := 0
	og := orgGetAgents
	defer func() { orgGetAgents = og }()

	orgGetAgents = func(ArmoryCloudAddr *url.URL, accessToken string) ([]org.Agent, error) {
		calls = calls + 1
		return []org.Agent{
			{AgentIdentifier: "abc"},
			{AgentIdentifier: "xyz"},
		}, nil
	}

	runner := NewProjectRunner(&MockConfiguration{})
	runner.PopulateAgents()
	suite.Equal(1, calls, "Expected a call to orgGetAgents")
	suite.Equal(2, len(*runner.AgentIdentifiers), "Should have two agent identifiers")
	suite.Equal("abc;xyz", strings.Join(*runner.AgentIdentifiers, ";"), "Should have two agent identifiers")
}
