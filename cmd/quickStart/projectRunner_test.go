package quickStart

import (
	"errors"
	"github.com/armory/armory-cli/pkg/config"
	"github.com/hashicorp/go-multierror"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"testing"
)

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
	runner := NewProjectRunner(&config.Configuration{})
	suite.False(runner.HasErrors(), "Runner should have no errors")
	runner.Errors = &multierror.Error{
		Errors: []error{errors.New("some error")},
	}
	suite.True(runner.HasErrors(), "Runner should have errors")
}

func (suite *ProjectRunnerSuite) TestAppendError() {
	runner := NewProjectRunner(&config.Configuration{})
	runner.AppendError(errors.New("some error"))

	suite.True(runner.HasErrors(), "Runner should have errors")
}

func (suite *ProjectRunnerSuite) TestExec() {
	runner := NewProjectRunner(&config.Configuration{})
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
	runner := NewProjectRunner(&config.Configuration{})
	mock := NewMock()

	runner.ExecWith(mock.UpdatesWith, "update")
	runner.ExecWith(mock.ErrorsWith, "error")
	//After an error occurs, Updates will not be called again
	runner.ExecWith(mock.UpdatesWith, "update")
	suite.Equal(";update;error", mock.CalledWith)
}

func (suite *ProjectRunnerSuite) TestWontFailOnErrorWithoutOne() {
	runner := NewProjectRunner(&config.Configuration{})
	runner.FailOnError()
}

func (suite *ProjectRunnerSuite) TestFailOnError() {
	runner := NewProjectRunner(&config.Configuration{})
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
