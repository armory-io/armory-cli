package deploy

import (
	"bytes"
	"encoding/json"
	"fmt"
	de "github.com/armory-io/deploy-engine/pkg"
	"github.com/armory/armory-cli/cmd"
	"github.com/armory/armory-cli/pkg/deploy"
	"github.com/armory/armory-cli/pkg/output"
	"github.com/armory/armory-cli/pkg/util"
	"github.com/jarcoal/httpmock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"
	"io"
	"io/ioutil"
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
	expected := de.NewPipelineStartPipelineResponse()
	expected.SetPipelineId("12345")
	err := registerResponder(expected, 200)
	if err != nil {
		suite.T().Fatalf("TestDeployStartJsonSuccess failed with: %s", err)
	}
	tempFile := util.TempAppFile("", "app", testAppYamlStr)
	if tempFile == nil {
		suite.T().Fatal("TestDeployStartJsonSuccess failed with: Could not create temp app file.")
	}
	suite.T().Cleanup(func() { os.Remove(tempFile.Name()) })
	outWriter := bytes.NewBufferString("")
	rootCmd, err := getDeployCmdWithTmpFile(outWriter, tempFile, "json")
	err = rootCmd.Execute()
	if err != nil {
		suite.T().Fatalf("TestDeployStartJsonSuccess failed with: %s", err)
	}
	output, err := ioutil.ReadAll(outWriter)
	if err != nil {
		suite.T().Fatalf("TestDeployStartJsonSuccess failed with: %s", err)
	}
	var received = FormattableDeployStartResponse{}
	json.Unmarshal(output, &received)
	suite.Equal(received.DeploymentId, expected.GetPipelineId(), "they should be equal")
}

func (suite *DeployStartTestSuite) TestDeployStartYAMLSuccess() {
	expected := de.NewPipelineStartPipelineResponse()
	expected.SetPipelineId("12345")
	err := registerResponder(expected, 200)
	if err != nil {
		suite.T().Fatalf("TestDeployStartYAMLSuccess failed with: %s", err)
	}
	tempFile := util.TempAppFile("", "app", testAppYamlStr)
	if tempFile == nil {
		suite.T().Fatal("TestDeployStartYAMLSuccess failed with: Could not create temp app file.")
	}
	suite.T().Cleanup(func() { os.Remove(tempFile.Name()) })
	outWriter := bytes.NewBufferString("")
	rootCmd, err := getDeployCmdWithTmpFile(outWriter, tempFile, "yaml")
	err = rootCmd.Execute()
	if err != nil {
		suite.T().Fatalf("TestDeployStartYAMLSuccess failed with: %s", err)
	}
	output, err := ioutil.ReadAll(outWriter)
	if err != nil {
		suite.T().Fatalf("TestDeployStartYAMLSuccess failed with: %s", err)
	}
	var received = FormattableDeployStartResponse{}
	yaml.Unmarshal(output, &received)
	suite.Equal(received.DeploymentId, expected.GetPipelineId(), "they should be equal")
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
	rootCmd, err := getDeployCmdWithTmpFile(outWriter, tempFile, "yaml")
	err = rootCmd.Execute()
	if err != nil {
		suite.T().Fatalf("TestDeployStartHttpError failed with: %s", err)
	}
	output, err := ioutil.ReadAll(outWriter)
	if err != nil {
		suite.T().Fatalf("TestDeployStartHttpError failed with: %s", err)
	}
	suite.Equal(`error: "request returned an error: status code(500) "{\"code\":2, \"message\":\"invalid operation\", \"details\":[]}"`,
		strings.TrimSpace(string(output)), "they should be equal")
}

func (suite *DeployStartTestSuite) TestDeployStartFlagFileRequired() {
	outWriter := bytes.NewBufferString("")
	rootCmd, options, err := getOverrideRootCmd(outWriter)
	if err != nil {
		suite.T().Fatalf("TestDeployStartFlagRequired failed with: %s", err)
	}
	rootCmd.AddCommand(NewDeployCmd(options))

	args := []string{
		"deploy", "start",
		"--output=json",
	}
	rootCmd.SetArgs(args)
	err = rootCmd.Execute()
	if err == nil {
		suite.T().Fatal("TestDeployStartFlagRequired failed with: error should not be null")
	}
	suite.EqualError(err, "required flag(s) \"file\" not set")
}

func (suite *DeployStartTestSuite) TestDeployStartBadPath() {
	outWriter := bytes.NewBufferString("")
	rootCmd, options, err := getOverrideRootCmd(outWriter)
	if err != nil {
		suite.T().Fatalf("TestDeployStartBadPath failed with: %s", err)
	}
	rootCmd.AddCommand(NewDeployCmd(options))

	args := []string{
		"deploy", "start",
		"--file=/badPath/test.yml",
		"--output=json",
	}
	rootCmd.SetArgs(args)
	err = rootCmd.Execute()
	if err == nil {
		suite.T().Fatal("TestDeployStartBadPath failed with: error should not be null")
	}
	suite.EqualError(err, "error trying to read the YAML file: open /badPath/test.yml: no such file or directory")
}

func (suite *DeployStartTestSuite) TestWhenTheManifestAndFlagDoNotHaveAppNameAnErrorIsRaised() {
	expected := de.NewPipelineStartPipelineResponse()
	expected.SetPipelineId("12345")
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
	rootCmd, err := getDeployCmdWithTmpFile(outWriter, tempFile, "yaml")
	err = rootCmd.Execute()
	if err == nil {
		suite.T().Fatal("TestWhenTheManifestAndFlagDoNotHaveAppNameAnErrorIsRaised failed with: error should not be null")
	}
	suite.EqualError(err, "application name must be defined in deployment file or by application-name opt")
}

func (suite *DeployStartTestSuite) TestWhenTheManifestAndFlagDoNotHaveAppNameButFlagIsSuppliedAnErrorIsNotRaised() {
	expected := de.NewPipelineStartPipelineResponse()
	expected.SetPipelineId("12345")
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
	rootCmd, options, err := getOverrideRootCmd(outWriter)
	if err != nil {
		suite.T().Fatalf("TestWhenTheManifestAndFlagDoNotHaveAppNameButFlagIsSuppliedAnErrorIsNotRaised failed with: %s", err)
	}
	rootCmd.AddCommand(NewDeployCmd(options))
	args := []string{
		"deploy", "start",
		"--file=" + tempFile.Name(),
		"--application=foo",
		"--output=json",
	}
	rootCmd.SetArgs(args)

	err = rootCmd.Execute()
	if err != nil {
		suite.T().Fatalf("TestWhenTheManifestAndFlagDoNotHaveAppNameButFlagIsSuppliedAnErrorIsNotRaised failed with: %s", err)
	}
}

func getOverrideRootCmd(outWriter io.Writer) (*cobra.Command, *cmd.RootOptions, error) {
	rootCmd, options := cmd.NewCmdRoot(outWriter, ioutil.Discard)
	client, err := deploy.NewDeployClient(
		"localhost",
		"token",
	)
	if err != nil {
		return nil, nil, fmt.Errorf("could not create the deploy client")
	}
	options.DeployClient = client
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		options.Output = output.NewOutput(options.O)
		return nil
	}
	return rootCmd, options, nil
}

func registerResponder(body interface{}, status int) error {
	responder, err := httpmock.NewJsonResponder(status, body)
	if err != nil {
		return err
	}
	httpmock.RegisterResponder("POST", "https://localhost/pipelines/kubernetes", responder)
	return nil
}

func getDeployCmdWithTmpFile(outWriter io.Writer, tmpFile *os.File, output string) (*cobra.Command, error) {
	rootCmd, options, err := getOverrideRootCmd(outWriter)
	if err != nil {
		return nil, err
	}
	rootCmd.AddCommand(NewDeployCmd(options))
	args := []string{
		"deploy", "start",
		"--file=" + tmpFile.Name(),
		"--output=" + output,
	}
	rootCmd.SetArgs(args)
	return rootCmd, nil
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
