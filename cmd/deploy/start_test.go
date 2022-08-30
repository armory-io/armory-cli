package deploy

import (
	"bytes"
	"encoding/json"
	de "github.com/armory-io/deploy-engine/api"
	"github.com/armory/armory-cli/pkg/config"
	"github.com/armory/armory-cli/pkg/util"
	"github.com/jarcoal/httpmock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestDeployStartTestSuite(t *testing.T) {
	suite.Run(t, new(DeployStartTestSuite))
}

type DeployStartTestSuite struct {
	suite.Suite
}

func (suite *DeployStartTestSuite) SetupSuite() {
	os.Setenv("ARMORY_CLI_TEST", "true")
	httpmock.Activate()
}

func (suite *DeployStartTestSuite) SetupTest() {
	httpmock.Reset()
}

func (suite *DeployStartTestSuite) TearDownSuite() {
	os.Unsetenv("ARMORY_CLI_TEST")
	httpmock.DeactivateAndReset()
}

func (suite *DeployStartTestSuite) TestDeployStartJsonSuccess() {
	expected := &de.StartPipelineResponse{
		PipelineID: "12345",
	}
	err := registerResponder(expected, http.StatusAccepted)
	if err != nil {
		suite.T().Fatalf("TestDeployStartJsonSuccess failed with: %s", err)
	}
	tempFile := util.TempAppFile("", "app", testAppYamlStr)
	if tempFile == nil {
		suite.T().Fatal("TestDeployStartJsonSuccess failed with: Could not create temp app file.")
	}
	suite.T().Cleanup(func() { os.Remove(tempFile.Name()) })
	outWriter := bytes.NewBufferString("")
	cmd := getDeployCmdWithTmpFile(outWriter, tempFile, "json")
	err = cmd.Execute()
	if err != nil {
		suite.T().Fatalf("TestDeployStartJsonSuccess failed with: %s", err)
	}
	output, err := ioutil.ReadAll(outWriter)
	if err != nil {
		suite.T().Fatalf("TestDeployStartJsonSuccess failed with: %s", err)
	}
	var received = FormattableDeployStartResponse{}
	json.Unmarshal(output, &received)
	suite.Equal(expected.PipelineID, received.DeploymentId, "they should be equal")
}

func (suite *DeployStartTestSuite) TestDeployStartYAMLSuccess() {
	expected := &de.StartPipelineResponse{
		PipelineID: "12345",
	}
	err := registerResponder(expected, http.StatusAccepted)
	if err != nil {
		suite.T().Fatalf("TestDeployStartYAMLSuccess failed with: %s", err)
	}
	tempFile := util.TempAppFile("", "app", testAppYamlStr)
	if tempFile == nil {
		suite.T().Fatal("TestDeployStartYAMLSuccess failed with: Could not create temp app file.")
	}
	suite.T().Cleanup(func() { os.Remove(tempFile.Name()) })
	outWriter := bytes.NewBufferString("")
	cmd := getDeployCmdWithTmpFile(outWriter, tempFile, "yaml")
	err = cmd.Execute()
	if err != nil {
		suite.T().Fatalf("TestDeployStartYAMLSuccess failed with: %s", err)
	}
	output, err := ioutil.ReadAll(outWriter)
	if err != nil {
		suite.T().Fatalf("TestDeployStartYAMLSuccess failed with: %s", err)
	}
	var received = FormattableDeployStartResponse{}
	yaml.Unmarshal(output, &received)
	suite.Equal(expected.PipelineID, received.DeploymentId, "they should be equal")
}

func (suite *DeployStartTestSuite) TestDeployStartHttpError() {
	err := registerResponder(`{"code":2, "message":"invalid operation", "details":[]}`, 500)
	if err != nil {
		suite.T().Fatalf("TestDeployStartYAMLSuccess failed with: %s", err)
	}
	tempFile := util.TempAppFile("", "app", testAppYamlStr)
	if tempFile == nil {
		suite.T().Fatal("TestDeployStartHttpError failed with: Could not create temp app file.")
	}
	suite.T().Cleanup(func() { os.Remove(tempFile.Name()) })
	outWriter := bytes.NewBufferString("")
	cmd := getDeployCmdWithTmpFile(outWriter, tempFile, "yaml")

	err = cmd.Execute()
	if err != nil {
		suite.T().Fatalf("TestDeployStartHttpError failed with: %s", err)
	}
	output, err := ioutil.ReadAll(outWriter)
	if err != nil {
		suite.T().Fatalf("TestDeployStartHttpError failed with: %s", err)
	}
	suite.Equal(`error: "request returned an error: status code(500), thrown error: "{\"code\":2, \"message\":\"invalid operation\", \"details\":[]}"`,
		strings.TrimSpace(string(output)), "they should be equal")
}

func (suite *DeployStartTestSuite) TestDeployStartFlagFileRequired() {
	token := "some-token"
	configuration := config.New(&config.Input{
		AccessToken: &token,
	})
	deployCmd := NewDeployCmd(configuration)
	args := []string{"start"}
	deployCmd.SetArgs(args)
	err := deployCmd.Execute()
	if err == nil {
		suite.T().Fatal("TestDeployStartFlagRequired failed with: error should not be null")
	}
	suite.EqualError(err, "required flag(s) \"file\" not set")
}

func (suite *DeployStartTestSuite) TestDeployStartBadPath() {
	isTest := true
	configuration := config.New(&config.Input{IsTest: &isTest})
	deployCmd := NewDeployCmd(configuration)
	outWriter := bytes.NewBufferString("")
	deployCmd.SetOut(outWriter)

	args := []string{
		"start",
		"--file=/badPath/test.yml",
	}
	deployCmd.SetArgs(args)
	err := deployCmd.Execute()
	if err == nil {
		suite.T().Fatal("TestDeployStartBadPath failed with: error should not be null")
	}
	suite.EqualError(err, "error trying to read the YAML file, thrown error: open /badPath/test.yml: no such file or directory")
}

func (suite *DeployStartTestSuite) TestWhenTheManifestAndFlagDoNotHaveAppNameAnErrorIsRaised() {
	expected := &de.StartPipelineResponse{
		PipelineID: "12345",
	}
	err := registerResponder(expected, 200)
	if err != nil {
		suite.T().Fatalf("TestDeployStartYAMLSuccess failed with: %s", err)
	}
	tempFile := util.TempAppFile("", "app", testAppYamlStrWithoutApplicationName)
	if tempFile == nil {
		suite.T().Fatal("TestWhenTheManifestAndFlagDoNotHaveAppNameAnErrorIsRaised failed with: Could not create temp app file.")
	}
	suite.T().Cleanup(func() { os.Remove(tempFile.Name()) })
	outWriter := bytes.NewBufferString("")
	cmd := getDeployCmdWithTmpFile(outWriter, tempFile, "yaml")
	err = cmd.Execute()
	if err == nil {
		suite.T().Fatal("TestWhenTheManifestAndFlagDoNotHaveAppNameAnErrorIsRaised failed with: error should not be null")
	}
	suite.EqualError(err, "application name must be defined in deployment file or by application opt")
}

func (suite *DeployStartTestSuite) TestWhenTheManifestAndFlagDoNotHaveAppNameButFlagIsSuppliedAnErrorIsNotRaised() {
	expected := &de.StartPipelineResponse{
		PipelineID: "12345",
	}
	err := registerResponder(expected, 200)
	if err != nil {
		suite.T().Fatalf("TestDeployStartYAMLSuccess failed with: %s", err)
	}
	tempFile := util.TempAppFile("", "app", testAppYamlStrWithoutApplicationName)
	if tempFile == nil {
		suite.T().Fatal("TestWhenTheManifestAndFlagDoNotHaveAppNameButFlagIsSuppliedAnErrorIsNotRaised failed with: Could not create temp app file.")
	}
	suite.T().Cleanup(func() { os.Remove(tempFile.Name()) })
	outWriter := bytes.NewBufferString("")
	cmd := getDeployCmdWithTmpFile(outWriter, tempFile, "yaml")
	args := []string{
		"start",
		"--file=" + tempFile.Name(),
		"--application=foo",
	}
	cmd.SetArgs(args)

	err = cmd.Execute()
	if err != nil {
		suite.T().Fatalf("TestWhenTheManifestAndFlagDoNotHaveAppNameButFlagIsSuppliedAnErrorIsNotRaised failed with: %s", err)
	}
}

func registerResponder(body interface{}, status int) error {
	responder, err := httpmock.NewJsonResponder(status, body)
	if err != nil {
		return err
	}
	httpmock.RegisterResponder("POST", "https://localhost/pipelines/kubernetes", responder)
	return nil
}

func getDeployCmdWithTmpFile(outWriter io.Writer, tmpFile *os.File, output string) *cobra.Command {
	token := "some-token"
	addr := "https://localhost"
	clientId := ""
	clientSecret := ""
	isTest := true
	configuration := config.New(&config.Input{
		AccessToken:  &token,
		ApiAddr:      &addr,
		ClientId:     &clientId,
		ClientSecret: &clientSecret,
		OutFormat:    &output,
		IsTest:       &isTest,
	})
	deployCmd := NewDeployCmd(configuration)
	deployCmd.SetOut(outWriter)
	args := []string{
		"start",
		"--file=" + tmpFile.Name(),
	}
	deployCmd.SetArgs(args)
	return deployCmd
}

const testAppYamlStr = `
version: apps/v1
kind: kubernetes
application: deployment-test
targets:
    dev-west:
        account: dev
        namespace: test
        strategy: strategy1
manifests: []
strategies:
    strategy1:
        canary:
            steps:
                - pause:
                    duration: 1
                    unit: SECONDS
`

const testAppYamlStrWithoutApplicationName = `
version: apps/v1
kind: kubernetes
targets:
    dev-west:
        account: dev
        namespace: test
        strategy: strategy1
manifests: []
strategies:
    strategy1:
        canary:
            steps:
                - pause:
                    duration: 1
                    unit: SECONDS
`
