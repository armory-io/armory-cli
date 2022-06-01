package quickStart

import (
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestGithubQuickStartProjectTestSuite(t *testing.T) {
	suite.Run(t, new(QuickStartSuite))
}

func tempDir() string {
	dir := os.Getenv("TMPDIR")
	if dir == "" {
		dir = "/tmp"
	}
	return dir
}

var testProject = GithubQuickStartProject{
	ProjectName:   "cdCon-cdaas-demo",
	BranchName:    "main",
	DirName:       tempDir() + "/test-project",
	IsZipFile:     true,
	DeployYmlName: "deploy.yml",
}

type QuickStartSuite struct {
	suite.Suite
}

func (suite *QuickStartSuite) SetupSuite() {

}

func (suite *QuickStartSuite) SetupTest() {

}

func (suite *QuickStartSuite) TearDownSuite() {
	if testProject.IsZipFile {
		os.Remove(testProject.DirName + ".zip")
	}
	os.RemoveAll(testProject.DirName)
}

func (suite *QuickStartSuite) TestSkipUnzip() {
	project := GithubQuickStartProject{
		IsZipFile: false,
	}

	suite.Equal(nil, project.Unzip(), "Should exit without error when project is not a zip")

}

func (suite *QuickStartSuite) TestGithubQuickStartProject() {
	log.SetLevel(log.DebugLevel)
	if err := testProject.Download(); err != nil {
		suite.Failf("Failed downloading project", "%s", err.Error())
	}

	suite.testExists(testProject.DirName + ".zip")

	if err := testProject.Unzip(); err != nil {
		suite.Failf("Failed unzipping project", "%s", err.Error())
	}

	suite.testExists(testProject.DirName)
	suite.testExists(testProject.DirName + "/manifests")
	suite.testExists(testProject.DirName + "/manifests/demo-app.yml")
	suite.testExists(testProject.DirName + "/manifests/demo-namespace.yml")
	suite.testExists(testProject.DirName + "/manifests/sample-namespace.yml")
	deployFilePath := testProject.DirName + "/" + testProject.DeployYmlName
	selectedAgentName := "test-agent-account"
	suite.testExists(deployFilePath)

	if err := testProject.UpdateAgentAccount(selectedAgentName); err != nil {
		suite.Failf("Failed updating agent account", "%s", err.Error())
	}

	yaml, err := ioutil.ReadFile(deployFilePath)
	if err != nil {
		suite.Failf("Failed reading deploy.yml", "%s", err.Error())
	}

	lines := strings.Split(string(yaml), "\n")
	foundAgentLines := 0
	for _, line := range lines {
		if strings.TrimLeft(line, " ") == ("account: " + selectedAgentName) {
			foundAgentLines++
		}
	}

	suite.Equal(2, foundAgentLines, "Expected two agent lines to be updated")
}

func (suite *QuickStartSuite) TestSuffix() {
	suite.Equal("/archive/refs/heads/main.zip", GithubQuickStartProject{IsZipFile: true}.GetUrlSuffix(), "testProject should be a zip")
	suite.Equal("", GithubQuickStartProject{IsZipFile: false}.GetUrlSuffix(), "testProject should be a zip")
}

func (suite *QuickStartSuite) testExists(fileName string) {
	if _, err := os.Stat(fileName); err != nil {
		suite.Failf("Failed checking for existence of file or dir", "%s should exist.\n %s", fileName, err.Error())
	}
}
