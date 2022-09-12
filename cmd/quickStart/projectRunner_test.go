package quickStart

import (
	"errors"
	"github.com/armory/armory-cli/pkg/config"
	"github.com/armory/armory-cli/pkg/org"
	"github.com/armory/armory-cli/pkg/util"
	"github.com/chzyer/test"
	"github.com/hashicorp/go-multierror"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"net/url"
	"strings"
	"testing"
)

func TestProjectRunnerTestSuite(t *testing.T) {
	suite.Run(t, new(ProjectRunnerSuite))
}

var mockAgents = []org.Agent{
	{AgentIdentifier: "abc"},
	{AgentIdentifier: "abc"},
	{AgentIdentifier: "xyz"},
	{AgentIdentifier: "xyz"},
}

var runner *ProjectRunner

type ProjectRunnerSuite struct {
	suite.Suite
}

func (suite *ProjectRunnerSuite) SetupSuite() {

}

func (suite *ProjectRunnerSuite) SetupTest() {
	runner = NewProjectRunner(&config.MockConfiguration{})
}

func (suite *ProjectRunnerSuite) TearDownSuite() {
}

func (suite *ProjectRunnerSuite) TestNoAgentsFoundErrorMsg() {
	test.Equal("No armory agents found. Test1", NoAgentsFoundError{msg: "Test1"}.Error())
}

func (suite *ProjectRunnerSuite) TestSelectedAgentErrorMsg() {
	test.Equal("Unable to continue. An agent must be selected. Test1", SelectedAgentError{msg: "Test1"}.Error())
}

func (suite *ProjectRunnerSuite) TestRunnerHasError() {
	suite.False(runner.HasErrors(), "Runner should have no errors")
	runner.Errors = &multierror.Error{
		Errors: []error{errors.New("some error")},
	}
	suite.True(runner.HasErrors(), "Runner should have errors")
}

func (suite *ProjectRunnerSuite) TestAppendError() {
	runner.AppendError(errors.New("some error"))

	suite.True(runner.HasErrors(), "Runner should have errors")
}

func (suite *ProjectRunnerSuite) TestExec() {
	mock := util.NewMock()

	runner.Exec(mock.Updates)
	suite.Equal(1, mock.Calls)

	runner.Exec(mock.Errors)
	suite.Equal(2, mock.Calls)

	//After an error occurs, Updates will not be called again
	runner.Exec(mock.Updates)
	suite.Equal(2, mock.Calls)
}

func (suite *ProjectRunnerSuite) TestExecWith() {
	mock := util.NewMock()

	runner.ExecWith(mock.UpdatesWith, "update")
	runner.ExecWith(mock.ErrorsWith, "error")
	//After an error occurs, Updates will not be called again
	runner.ExecWith(mock.UpdatesWith, "update")
	suite.Equal(";update;error", mock.CalledWith)
}

func (suite *ProjectRunnerSuite) TestWontFailOnErrorWithoutOne() {
	runner.FailOnError()
}

func (suite *ProjectRunnerSuite) TestFailOnError() {
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

func (suite *ProjectRunnerSuite) TestSelectAgentRequiresFoundAgents() {
	runner.AppendError(SelectedAgentError{})
	runner.SelectAgent("")
	suite.Equal(1, len(runner.Errors.Errors), "Should only have one error")
	if _, ok := runner.Errors.Errors[0].(SelectedAgentError); !ok {
		suite.Fail("Should not have found any agents")
	}
}

func (suite *ProjectRunnerSuite) TestSelectAgentSkips() {
	runner.SelectAgent("")
	if _, ok := runner.Errors.Errors[0].(NoAgentsFoundError); !ok {
		suite.Fail("Should not have found any agents")
	}
}

func (suite *ProjectRunnerSuite) TestSelectsMatchingAgentName() {
	const expectedAgentName = "matchesOne"
	runner.AgentIdentifiers = &[]string{expectedAgentName, "anotherOne"}
	suite.Equal(expectedAgentName, runner.SelectAgent(expectedAgentName))

	suite.False(runner.HasErrors())
}

func (suite *ProjectRunnerSuite) TestErrorWhenSelectFindsNoMatch() {
	const expectedAgentName = "matchesNONE"
	runner.AgentIdentifiers = &[]string{"matchesOne", "anotherOne"}
	suite.Equal("", runner.SelectAgent(expectedAgentName))
	suite.True(runner.HasErrors())
	if _, ok := runner.Errors.Errors[0].(SelectedAgentError); !ok {
		suite.Fail("Should have had an issue selecting an agent that doesn't exist")
	}
}

func (suite *ProjectRunnerSuite) TestPopulateAgentsSkips() {
	calls := 0
	og := orgGetAgents
	defer func() { orgGetAgents = og }()

	orgGetAgents = MockGetAgents(&calls, mockAgents, nil)

	runner.AppendError(NoAgentsFoundError{})
	runner.PopulateAgents()
	suite.Equal(0, calls, "Expected no calls to get agents if the runner has an error")
}

func (suite *ProjectRunnerSuite) TestPopulateAgentsGetAgentsError() {
	calls := 0
	og := orgGetAgents
	defer func() { orgGetAgents = og }()

	orgGetAgents = MockGetAgents(&calls, mockAgents, errors.New("test err"))

	runner.PopulateAgents()
	suite.True(runner.HasErrors(), "error expected when returned from org.GetAgents")
	suite.Equal(1, calls, "Expected a call to orgGetAgents")
}

func (suite *ProjectRunnerSuite) TestPopulateAgentsSuccess() {
	calls := 0
	og := orgGetAgents
	defer func() { orgGetAgents = og }()
	orgGetAgents = MockGetAgents(&calls, mockAgents, nil)

	runner.PopulateAgents()
	suite.Equal(1, calls, "Expected a call to orgGetAgents")
	suite.Equal(2, len(*runner.AgentIdentifiers), "Should have two agent identifiers")
	suite.Equal("abc;xyz", strings.Join(*runner.AgentIdentifiers, ";"), "Should have two agent identifiers")
}

func (suite *ProjectRunnerSuite) TestPopulateAgentsNotFound() {
	calls := 0
	og := orgGetAgents
	defer func() { orgGetAgents = og }()
	orgGetAgents = MockGetAgents(&calls, []org.Agent{}, nil)

	runner.PopulateAgents()

	suite.Equal(1, calls, "Expected a call to orgGetAgents")
	suite.Equal(0, len(*runner.AgentIdentifiers), "Should have no agent identifiers")
	if _, ok := runner.Errors.Errors[0].(NoAgentsFoundError); !ok {
		suite.Fail("Should not have found any agents")
	}
}

func MockGetAgents(numCalls *int, agents []org.Agent, apiError error) func(ArmoryCloudAddr *url.URL, accessToken string) ([]org.Agent, error) {
	return func(ArmoryCloudAddr *url.URL, accessToken string) ([]org.Agent, error) {
		*numCalls = *numCalls + 1
		return agents, apiError
	}
}
