package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
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
	suite.outputFile, _ = os.CreateTemp("", "")
	suite.summaryFile, _ = os.CreateTemp("", "")
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

	bytes, err := os.ReadFile(suite.outputFile.Name())
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "KEY1=value1\nKEY2=value2\n", string(bytes))
}

func (suite *GithubIntegrationSuite) TestWriteOutputToFileIsSkippedWhenNoEnvVariableIsSet() {
	TryWriteGitHubContext("key1", "value1", "key2", "value2")

	bytes, err := os.ReadFile(suite.outputFile.Name())
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "", string(bytes))
}

func (suite *GithubIntegrationSuite) TestCanWriteOutputSummary() {
	os.Setenv(GithubSummary, suite.summaryFile.Name())

	TryWriteGitHubStepSummary("this is step summary")

	bytes, err := os.ReadFile(suite.summaryFile.Name())
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "this is step summary\n", string(bytes))
}

func (suite *GithubIntegrationSuite) TestWriteOutputSummaryIsSkippedWhenNoEnvVariableIsSet() {
	TryWriteGitHubStepSummary("this is step summary")

	bytes, err := os.ReadFile(suite.summaryFile.Name())
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "", string(bytes))
}
