package utils

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"testing"
)

func TestGithubIntegration(t *testing.T) {
	suite.Run(t, new(GithubIntegrationSuite))
}

type GithubIntegrationSuite struct {
	suite.Suite
	outputFile  *os.File
	summaryFile *os.File
}

func (suite *GithubIntegrationSuite) SetupSuite() {
}

func (suite *GithubIntegrationSuite) SetupTest() {
	suite.outputFile, _ = ioutil.TempFile("", "")
	suite.summaryFile, _ = ioutil.TempFile("", "")
}

func (suite *GithubIntegrationSuite) TearDownSuite() {
	os.Unsetenv(GithubOutput)
	os.Unsetenv(GithubSummary)
	os.Remove(suite.outputFile.Name())
	os.Remove(suite.summaryFile.Name())
}

func (suite *GithubIntegrationSuite) TestCanWriteOutputToFile() {
	os.Setenv(GithubOutput, suite.outputFile.Name())

	TryWriteGitHubContext("key1", "value1", "key2", "value2")

	bytes, err := ioutil.ReadFile(suite.outputFile.Name())
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "KEY1=value1\nKEY2=value2\n", string(bytes))
}

func (suite *GithubIntegrationSuite) TestWriteOutputToFileIsSkippedWhenNoEnvVariableIsSet() {
	TryWriteGitHubContext("key1", "value1", "key2", "value2")

	bytes, err := ioutil.ReadFile(suite.outputFile.Name())
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "", string(bytes))
}

func (suite *GithubIntegrationSuite) TestCanWriteOutputSummary() {
	os.Setenv(GithubSummary, suite.summaryFile.Name())

	TryWriteGitHubStepSummary("this is step summary")

	bytes, err := ioutil.ReadFile(suite.summaryFile.Name())
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "this is step summary\n", string(bytes))
}

func (suite *GithubIntegrationSuite) TestWriteOutputSummaryIsSkippedWhenNoEnvVariableIsSet() {
	TryWriteGitHubStepSummary("this is step summary")

	bytes, err := ioutil.ReadFile(suite.summaryFile.Name())
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "", string(bytes))
}
