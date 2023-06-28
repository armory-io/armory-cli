package deploy

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	de "github.com/armory-io/deploy-engine/pkg/api"
	"github.com/armory/armory-cli/pkg/config"
	"github.com/armory/armory-cli/pkg/util"
	"github.com/jarcoal/httpmock"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"
)

func TestDeployStartTestSuite(t *testing.T) {
	statusCheckTick = time.Second
	suite.Run(t, new(DeployStartTestSuite))
}

type DeployStartTestSuite struct {
	suite.Suite
}

func (suite *DeployStartTestSuite) SetupSuite() {
	suite.NoError(os.Setenv("ARMORY_CLI_TEST", "true"))
	httpmock.Activate()
}

func (suite *DeployStartTestSuite) SetupTest() {
	httpmock.Reset()
}

func (suite *DeployStartTestSuite) TearDownSuite() {
	suite.NoError(os.Unsetenv("ARMORY_CLI_TEST"))
	httpmock.DeactivateAndReset()
}

func (suite *DeployStartTestSuite) TestDeployStart() {
	cases := []struct {
		name        string
		kind        string
		file        string
		contentType string
		expected    de.StartPipelineResponse
		assertion   func(*testing.T, []byte, de.StartPipelineResponse)
	}{
		{
			name:        "start pipeline json success",
			kind:        "kubernetes",
			file:        testAppYamlStr,
			contentType: "json",
			expected: de.StartPipelineResponse{
				PipelineID: "12345",
			},
			assertion: func(t *testing.T, result []byte, expected de.StartPipelineResponse) {
				var received FormattableDeployStartResponse
				assert.NoError(t, json.Unmarshal(result, &received))
				assert.Equal(t, expected.PipelineID, received.DeploymentId)
			},
		},
		{
			name:        "start pipeline yaml success",
			kind:        "kubernetes",
			file:        testAppYamlStr,
			contentType: "yaml",
			expected: de.StartPipelineResponse{
				PipelineID: "12345",
			},
			assertion: func(t *testing.T, result []byte, expected de.StartPipelineResponse) {
				var received FormattableDeployStartResponse
				assert.NoError(t, yaml.Unmarshal(result, &received))
				assert.Equal(t, expected.PipelineID, received.DeploymentId)
			},
		},
		{
			name:        "start pipeline yaml fail validation",
			kind:        "kubernetes",
			file:        testAppYamlStrInvalid,
			contentType: "yaml",
			expected: de.StartPipelineResponse{
				PipelineID: "12345",
			},
			assertion: func(t *testing.T, result []byte, expected de.StartPipelineResponse) {
				assert.Equal(t, expectedValidationError, string(result))
			},
		},
		{
			name:        "start pipeline yaml success lambda",
			kind:        "lambda",
			file:        testLambdaYamlStr,
			contentType: "yaml",
			expected: de.StartPipelineResponse{
				PipelineID: "12345",
			},
			assertion: func(t *testing.T, result []byte, expected de.StartPipelineResponse) {
				// verify there is no validation errors being shown to the user
				assert.NotContains(t, string(result), "YAML is NOT valid. See the following errors:")

				// verify that response formats as expected
				var received FormattableDeployStartResponse
				assert.NoError(t, yaml.Unmarshal(result, &received))
				assert.Equal(t, expected.PipelineID, received.DeploymentId)
			},
		},
	}

	for _, c := range cases {
		suite.T().Run(c.name, func(t *testing.T) {
			assert.NoError(t, registerResponder(c.expected, http.StatusAccepted, c.kind))
			tempFile := util.TempAppFile("", "app", c.file)
			if tempFile == nil {
				t.Fatal("failed with: Could not create temp app file.")
			}
			t.Cleanup(func() {
				assert.NoError(t, os.Remove(tempFile.Name()))
			})
			outWriter := bytes.NewBufferString("")
			cmd := getDeployCmdWithFileName(outWriter, tempFile.Name(), c.contentType)
			assert.NoError(t, cmd.Execute())
			output, err := io.ReadAll(outWriter)
			assert.NoError(t, err)
			c.assertion(t, output, c.expected)
		})
	}
}

func (suite *DeployStartTestSuite) TestDeployStartWithURLSuccess() {
	expected := &de.StartPipelineResponse{
		PipelineID: "12345",
	}
	suite.NoError(registerResponder(expected, http.StatusAccepted, "kubernetes"))

	outWriter := bytes.NewBufferString("")
	cmd := getDeployCmdWithFileName(outWriter, "https://myhostedfile.example.com/deploy.yaml", "yaml")
	suite.NoError(cmd.Execute())

	output, err := io.ReadAll(outWriter)
	suite.NoError(err)
	var received FormattableDeployStartResponse
	suite.NoError(yaml.Unmarshal(output, &received))
	suite.Equal(expected.PipelineID, received.DeploymentId, "they should be equal")
}

func (suite *DeployStartTestSuite) TestDeployWithURLUsesExpectedOptions() {
	cmd := &cobra.Command{}
	expected := &de.StartPipelineResponse{
		PipelineID: "12345",
	}
	configuration := getDefaultConfiguration("json")
	deployClient := GetMockDeployClient(configuration)
	deployClient.MockStartPipelineResponse(func() (*de.StartPipelineResponse, *http.Response, error) {
		return expected, &http.Response{Status: "200"}, nil
	})
	pipelineResp, rawResp, err := WithURL(cmd, &deployStartOptions{
		account:           "jimbob-dev",
		targetFilters:     []string{"dev-a"},
		deploymentFile:    "http://mytesturl.com/deploy.yml",
		waitForCompletion: false,
	},
		deployClient,
	)

	suite.NoError(err)
	suite.Equal("200", rawResp.Status, "rawResp should be returned by WithURL")
	suite.Equal(expected.PipelineID, pipelineResp.PipelineID, "they should be equal")
	suite.Equal(deployClient.RecordedStartPipelineOptions.UnstructuredDeployment, map[string]any{"account": "jimbob-dev", "targetFilters": []map[string]any{{"includeTarget": "dev-a"}}}, "there should be body/deployment specification for the request WithURL")
	suite.Equal(deployClient.RecordedStartPipelineOptions.Headers["Content-Type"], mediaTypePipelineV2Link, "they should be equal")
	suite.Equal(deployClient.RecordedStartPipelineOptions.Headers["Accept"], mediaTypePipelineV2, "they should be equal")
	suite.Equal(deployClient.RecordedStartPipelineOptions.Headers[armoryConfigLocationHeader], "http://mytesturl.com/deploy.yml", "they should be equal")

	pipelineResp, rawResp, err = WithURL(cmd, &deployStartOptions{
		application:       "this will cause a failure",
		deploymentFile:    "http://mytesturl.com/deploy.yml",
		waitForCompletion: false,
	},
		deployClient,
	)

	suite.Nil(pipelineResp)
	suite.Nil(rawResp)
	suite.Error(err)
	suite.ErrorIs(err, ErrApplicationNameOverrideNotSupported)
}

func (suite *DeployStartTestSuite) TestDeployWithFileUsesExpectedOptions() {
	cmd := &cobra.Command{}
	expected := &de.StartPipelineResponse{
		PipelineID: "12345",
	}
	configuration := getDefaultConfiguration("json")
	deployClient := GetMockDeployClient(configuration)
	deployClient.MockStartPipelineResponse(func() (*de.StartPipelineResponse, *http.Response, error) {
		return expected, &http.Response{Status: "200"}, nil
	})
	tempFile := util.TempAppFile("", "app", testAppYamlStr)
	if tempFile == nil {
		suite.T().Fatal("TestDeployStartYAMLSuccess failed with: Could not create temp app file.")
	}
	suite.T().Cleanup(func() { os.Remove(tempFile.Name()) })

	pipelineResp, rawResp, err := WithLocalFile(cmd, &deployStartOptions{
		targetFilters:     []string{"dev-a"},
		deploymentFile:    tempFile.Name(),
		waitForCompletion: false,
	},
		deployClient,
	)
	dr := deployClient.RecordedStartPipelineOptions.UnstructuredDeployment
	suite.NoError(err)
	suite.Equal("200", rawResp.Status, "rawResp should be returned by WithURL")
	suite.Equal(expected.PipelineID, pipelineResp.PipelineID, "they should be equal")
	suite.Equal(dr["targetFilters"], []map[string]any{{"includeTarget": "dev-a"}}, "there should be body/deployment specification for the request WithURL")
	suite.Equal("", deployClient.RecordedStartPipelineOptions.Headers[armoryConfigLocationHeader], "they should be equal")
}

func (suite *DeployStartTestSuite) TestDeployWithPipelineValidation() {
	cmd := &cobra.Command{}
	deployClient := GetMockDeployClient(getDefaultConfiguration("json"))
	cases := []struct {
		name        string
		expectedErr error
		options     *deployStartOptions
	}{
		{
			name:        "application override not allowed",
			expectedErr: ErrApplicationNameOverrideNotSupported,
			options: &deployStartOptions{
				application:       "this will cause a failure",
				pipelineID:        "12345",
				waitForCompletion: false,
			},
		},
		{
			name: "account override is allowed",
			options: &deployStartOptions{
				account:           "this will not cause a failure",
				pipelineID:        "12345",
				waitForCompletion: false,
			},
		},
	}

	for _, c := range cases {
		suite.T().Run(c.name, func(t *testing.T) {
			_, _, err := WithURL(cmd, c.options, deployClient)
			assert.ErrorIs(t, err, c.expectedErr)
		})
	}
}

func (suite *DeployStartTestSuite) TestDeployWithPipelineIdUsesExpectedOptions() {
	var filters []map[string]any
	expectedBody := map[string]any{
		"account":       "",
		"targetFilters": filters,
	}
	expectedBody["account"] = ""
	cmd := &cobra.Command{}
	expected := &de.StartPipelineResponse{
		PipelineID: "123456789",
	}
	suite.NoError(registerResponder(expected, http.StatusAccepted, "kubernetes"))
	configuration := getDefaultConfiguration("json")
	deployClient := GetMockDeployClient(configuration)
	deployClient.MockStartPipelineResponse(func() (*de.StartPipelineResponse, *http.Response, error) {
		return expected, &http.Response{Status: "200"}, nil
	})
	pipelineResp, rawResp, err := WithURL(cmd, &deployStartOptions{
		deploymentFile:    "armory::http://localhost:9099/pipelines/012345/config",
		waitForCompletion: false,
	},
		deployClient,
	)
	suite.NoError(err)
	suite.Equal("200", rawResp.Status, "rawResp should be returned by WithPipelineId")
	suite.Equal(expected.PipelineID, pipelineResp.PipelineID, "response PipelineId should match expected")
	suite.Equal(expectedBody, deployClient.RecordedStartPipelineOptions.UnstructuredDeployment, "there should not be body/deployment specification for the request WithPipelineId")
	suite.Equal(mediaTypePipelineV2Link, deployClient.RecordedStartPipelineOptions.Headers["Content-Type"], "content-type hedaer should be set to redeploy json")
	suite.Equal(mediaTypePipelineV2, deployClient.RecordedStartPipelineOptions.Headers["Accept"], "accept header should be set to pipeline V2 json")
	suite.Equal("armory::http://localhost:9099/pipelines/012345/config", deployClient.RecordedStartPipelineOptions.Headers[armoryConfigLocationHeader], "header should contain specified pipelineID for redeploy")
}

func (suite *DeployStartTestSuite) TestDeployStartHttpError() {
	err := registerResponder(`{"code":2, "message":"invalid operation", "details":[]}`, 500, "kubernetes")
	if err != nil {
		suite.T().Fatalf("TestDeployStartYAMLSuccess failed with: %s", err)
	}
	tempFile := util.TempAppFile("", "app", testAppYamlStr)
	if tempFile == nil {
		suite.T().Fatal("TestDeployStartHttpError failed with: Could not create temp app file.")
	}
	suite.T().Cleanup(func() { os.Remove(tempFile.Name()) })
	outWriter := bytes.NewBufferString("")
	cmd := getDeployCmdWithFileName(outWriter, tempFile.Name(), "yaml")

	err = cmd.Execute()
	if err != nil {
		suite.T().Fatalf("TestDeployStartHttpError failed with: %s", err)
	}
	output, err := io.ReadAll(outWriter)
	if err != nil {
		suite.T().Fatalf("TestDeployStartHttpError failed with: %s", err)
	}
	suite.Equal(`error: "request returned an error: status code(500), thrown error: "{\"code\":2, \"message\":\"invalid operation\", \"details\":[]}"`,
		strings.TrimSpace(string(output)), "they should be equal")
}

func (suite *DeployStartTestSuite) TestDeployStartFlagFileRequired() {
	token := "some-token"
	configuration := config.New(&config.Input{
		IsTest:      lo.ToPtr(true),
		AccessToken: &token,
	})
	deployCmd := NewDeployCmd(configuration)
	args := []string{"start"}
	deployCmd.SetArgs(args)
	err := deployCmd.Execute()
	if err == nil {
		suite.T().Fatal("TestDeployStartFlagRequired failed with: error should not be null")
	}
	suite.ErrorIs(err, ErrConfigurationRequired)
}

func (suite *DeployStartTestSuite) TestDeployStartBadPath() {
	token := "some-token"
	addr := "https://localhost"
	clientId := ""
	clientSecret := ""
	output := "json"
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
	err := registerResponder(expected, 200, "kubernetes")
	if err != nil {
		suite.T().Fatalf("TestDeployStartYAMLSuccess failed with: %s", err)
	}
	tempFile := util.TempAppFile("", "app", testAppYamlStrWithoutApplicationName)
	if tempFile == nil {
		suite.T().Fatal("TestWhenTheManifestAndFlagDoNotHaveAppNameAnErrorIsRaised failed with: Could not create temp app file.")
	}
	suite.T().Cleanup(func() { os.Remove(tempFile.Name()) })
	outWriter := bytes.NewBufferString("")
	cmd := getDeployCmdWithFileName(outWriter, tempFile.Name(), "yaml")
	cmd.Execute()

	msg, err := io.ReadAll(outWriter)
	suite.NoError(err)
	suite.Contains(string(msg), "application name must be defined in deployment file or by application opt")
}

func (suite *DeployStartTestSuite) TestWhenTheManifestAndFlagDoNotHaveAppNameButFlagIsSuppliedAnErrorIsNotRaised() {
	expected := &de.StartPipelineResponse{
		PipelineID: "12345",
	}
	err := registerResponder(expected, 200, "kubernetes")
	if err != nil {
		suite.T().Fatalf("TestDeployStartYAMLSuccess failed with: %s", err)
	}
	tempFile := util.TempAppFile("", "app", testAppYamlStrWithoutApplicationName)
	if tempFile == nil {
		suite.T().Fatal("TestWhenTheManifestAndFlagDoNotHaveAppNameButFlagIsSuppliedAnErrorIsNotRaised failed with: Could not create temp app file.")
	}
	suite.T().Cleanup(func() { os.Remove(tempFile.Name()) })
	outWriter := bytes.NewBufferString("")
	cmd := getDeployCmdWithFileName(outWriter, tempFile.Name(), "yaml")
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

func (suite *DeployStartTestSuite) TestDeployStartJsonAndWaitForCompletionSuccess() {
	expected := &de.StartPipelineResponse{
		PipelineID: "456678",
	}
	err := registerResponder(expected, http.StatusAccepted, "kubernetes")
	if err != nil {
		suite.T().Fatalf("TestDeployAndWaitForCompletionSuccess failed with: %s", err)
	}
	err = registerStatusResponder([]de.WorkflowStatus{de.WorkflowStatusUnknown, de.WorkflowStatusRunning, de.WorkflowStatusSucceeded}, expected.PipelineID)
	if err != nil {
		suite.T().Fatalf("TestDeployAndWaitForCompletionSuccess : Failed to register status responder: %s", err)
	}
	tempFile := util.TempAppFile("", "app", testAppYamlStr)
	if tempFile == nil {
		suite.T().Fatal("TestDeployStartJsonSuccess failed with: Could not create temp app file.")
	}
	suite.T().Cleanup(func() { os.Remove(tempFile.Name()) })
	outWriter := bytes.NewBufferString("")
	cmd := getDeployCmdWithFileName(outWriter, tempFile.Name(), "json", "-w")
	err = cmd.Execute()
	if err != nil {
		suite.T().Fatalf("TestDeployStartJsonSuccess failed with: %s", err)
	}
	output, err := io.ReadAll(outWriter)
	if err != nil {
		suite.T().Fatalf("TestDeployStartJsonSuccess failed with: %s", err)
	}
	var received = FormattableDeployStartResponse{}
	json.Unmarshal(output, &received)
	suite.Equal(expected.PipelineID, received.DeploymentId, "they should be equal")
	suite.Equal(string(de.WorkflowStatusSucceeded), received.ExecutionStatus, "status should be SUCCESS")
}

func (suite *DeployStartTestSuite) TestDeployStartYAMLAndWaitForCompletionSuccess() {
	expected := &de.StartPipelineResponse{
		PipelineID: "23456",
	}
	err := registerResponder(expected, http.StatusAccepted, "kubernetes")
	if err != nil {
		suite.T().Fatalf("TestDeployStartYAMLSuccess failed with: %s", err)
	}
	err = registerStatusResponder([]de.WorkflowStatus{de.WorkflowStatusUnknown, de.WorkflowStatusRunning, de.WorkflowStatusCancelled}, expected.PipelineID)
	if err != nil {
		suite.T().Fatalf("TestDeployAndWaitForCompletionSuccess : Failed to register status responder: %s", err)
	}

	tempFile := util.TempAppFile("", "app", testAppYamlStr)
	if tempFile == nil {
		suite.T().Fatal("TestDeployStartYAMLSuccess failed with: Could not create temp app file.")
	}
	suite.T().Cleanup(func() { os.Remove(tempFile.Name()) })
	outWriter := bytes.NewBufferString("")
	cmd := getDeployCmdWithFileName(outWriter, tempFile.Name(), "yaml", "-w")
	err = cmd.Execute()
	suite.Error(err, "expected deployment failure due to cancelled status")
	output, err := io.ReadAll(outWriter)
	if err != nil {
		suite.T().Fatalf("TestDeployStartYAMLSuccess failed with: %s", err)
	}
	var received = FormattableDeployStartResponse{}
	yaml.Unmarshal(output, &received)
	suite.Equal(expected.PipelineID, received.DeploymentId, "they should be equal")
	suite.Equal(string(de.WorkflowStatusCancelled), received.ExecutionStatus, "pipeline status should be cancelled")
}

func registerResponder(body interface{}, status int, kind string) error {
	responder, err := httpmock.NewJsonResponder(status, body)
	if err != nil {
		return err
	}
	if kind == "kubernetes" {
		httpmock.RegisterResponder("POST", "https://localhost/pipelines/kubernetes", responder)
	} else {
		httpmock.RegisterResponder("POST", "https://localhost/pipelines", responder)
	}

	return nil
}

func registerStatusResponder(statuses []de.WorkflowStatus, pipelineID string) error {
	var responses []*http.Response
	for _, status := range statuses {
		r, err := httpmock.NewJsonResponse(200, map[string]string{"status": string(status)})
		if err != nil {
			return err
		}
		responses = append(responses, r)
	}
	rep := httpmock.ResponderFromMultipleResponses(responses)
	httpmock.RegisterResponder("GET", "https://localhost/pipelines/"+pipelineID, rep)
	return nil
}

func getDeployCmdWithFileName(outWriter io.Writer, fileName string, output string, additionalOpts ...string) *cobra.Command {
	configuration := getDefaultConfiguration(output)
	deployCmd := NewDeployCmd(configuration)
	deployCmd.SetOut(outWriter)
	args := []string{
		"start",
		"--file=" + fileName,
	}
	args = append(args, additionalOpts...)
	deployCmd.SetArgs(args)
	return deployCmd
}

func getDefaultConfiguration(output string) *config.Configuration {
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
	return configuration
}

const expectedValidationError = `YAML is NOT valid. See the following errors:

#PipelineRequest.targets."dev-west".strategy: 1 errors in empty disjunction:

#PipelineRequest.targets."dev-west".strategy: conflicting values "strategy1" and "strategy0"

deploymentId: "12345"

`

const testAppYamlStr = `
version: apps/v1
kind: kubernetes
application: deployment-test
targets:
    dev-west:
        account: dev
        namespace: test
        strategy: strategy1
manifests:
  - inline: ""
strategies:
    strategy1:
        canary:
            steps:
                - pause:
                    duration: 1
                    unit: SECONDS
`

const testLambdaYamlStr = `
version: v1
kind: lambda
application: first-lambda-app
description: Deploys a simple "hello world" NodeJS function
context:
  foo: bar
targets:
  firstTarget:
    account: firstAccount
    deployAsIamRole: "<some-deployment-role-arn>"
    region: us-west-2
artifacts:
  - path: "s3://<fill-me-in>/node/v0.0.1.zip"
    functionName: hello-lambda
    type: zipFile
providerOptions:
  lambda:
    - name: hello-lambda
      target: firstTarget
      runAsIamRole: "<some-lambda-role-arn>"
      handler: index.handler
      runtime: nodejs18.x
`

const testAppYamlStrInvalid = `
version: apps/v1
kind: kubernetes
application: deployment-test
targets:
    dev-west:
        account: dev
        namespace: test
        strategy: strategy0
manifests:
  - inline: ""
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
