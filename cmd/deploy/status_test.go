package deploy

import (
	"bytes"
	"encoding/json"
	de "github.com/armory-io/deploy-engine/pkg/api"
	"github.com/armory/armory-cli/pkg/config"
	"github.com/armory/armory-cli/pkg/model"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"strings"
	"testing"
)

func getExpectedPipelineDeployment() (*de.PipelineStatusResponse, *de.DeploymentStatusResponse) {
	expected := &de.PipelineStatusResponse{
		ID:          "12345",
		Application: "app",
		Status:      de.WorkflowStatusRunning,
		Steps: []*de.PipelineStep{
			{
				Type:   "deployment",
				Status: de.WorkflowStatusRunning,
				Deployment: &de.PipelineDeploymentStepResponse{
					ID: "5678",
				},
			},
		},
	}

	expectedDeploy := &de.DeploymentStatusResponse{
		ID:     "5678",
		Status: de.WorkflowStatusRunning,
	}

	return expected, expectedDeploy
}

func getStdTestConfig(outFmt string) *config.Configuration {
	token := "some-token"
	addr := "https://localhost"
	clientId := ""
	clientSecret := ""
	isTest := true
	return config.New(&config.Input{
		AccessToken:  &token,
		ApiAddr:      &addr,
		ClientId:     &clientId,
		ClientSecret: &clientSecret,
		OutFormat:    &outFmt,
		IsTest:       &isTest,
	})
}

func TestDeployStatusJsonSuccess(t *testing.T) {
	expected, expectedDeploy := getExpectedPipelineDeployment()
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	responder, err := httpmock.NewJsonResponder(200, expected)
	if err != nil {
		t.Fatalf("TestDeployStatusJsonSuccess failed with: %s", err)
	}
	httpmock.RegisterResponder("GET", "https://localhost/pipelines/12345", responder)
	responderDeploy, err := httpmock.NewJsonResponder(200, expectedDeploy)
	if err != nil {
		t.Fatalf("TestDeployStatusJsonSuccess failed with: %s", err)
	}
	httpmock.RegisterResponder("GET", "https://localhost/deployments/5678", responderDeploy)

	cmd := NewDeployCmd(getStdTestConfig("json"))
	outWriter := bytes.NewBufferString("")
	cmd.SetOut(outWriter)
	args := []string{
		"status",
		"--test=true",
		"--deploymentId=12345",
	}
	cmd.SetArgs(args)
	err = cmd.Execute()
	if err != nil {
		t.Fatalf("TestDeployStartJsonSuccess failed with: %s", err)
	}
	output, err := ioutil.ReadAll(outWriter)
	if err != nil {
		t.Fatalf("TestDeployStartJsonSuccess failed with: %s", err)
	}
	var received = FormattableDeployStatus{
		DeployResp: model.Pipeline{},
	}
	json.Unmarshal(output, &received.DeployResp)
	assert.Equal(t, *received.DeployResp.Id, expected.ID, "they should be equal")
	assert.Equal(t, *received.DeployResp.Application, expected.Application, "they should be equal")
	assert.Equal(t, *received.DeployResp.Status, expected.Status, "they should be equal")
	assert.Equal(t, len(*received.DeployResp.Steps), len(expected.Steps), "they should be equal")
	receivedDeployment := *received.DeployResp.Steps
	expectedDeployment := expected.Steps[0].Deployment
	assert.Equal(t, receivedDeployment[0].Deployment.ID, expectedDeployment.ID, "they should be equal")
}

func TestDeployStatusYAMLSuccess(t *testing.T) {
	expected, expectedDeploy := getExpectedPipelineDeployment()
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	responder, err := httpmock.NewJsonResponder(200, expected)
	if err != nil {
		t.Fatalf("TestDeployStatusYAMLSuccess failed with: %s", err)
	}
	httpmock.RegisterResponder("GET", "https://localhost/pipelines/12345", responder)
	responderDeploy, err := httpmock.NewJsonResponder(200, expectedDeploy)
	if err != nil {
		t.Fatalf("TestDeployStatusJsonSuccess failed with: %s", err)
	}
	httpmock.RegisterResponder("GET", "https://localhost/deployments/5678", responderDeploy)

	cmd := NewDeployCmd(getStdTestConfig("yaml"))
	outWriter := bytes.NewBufferString("")
	cmd.SetOut(outWriter)
	args := []string{
		"status",
		"--deploymentId=12345",
	}
	cmd.SetArgs(args)
	err = cmd.Execute()
	if err != nil {
		t.Fatal("TestDeployStatusYAMLSuccess failed with: error should not be null")
	}
	output, err := ioutil.ReadAll(outWriter)
	if err != nil {
		t.Fatalf("TestDeployStatusYAMLSuccess failed with: %s", err)
	}
	var received = FormattableDeployStatus{
		DeployResp: model.Pipeline{},
	}
	yaml.Unmarshal(output, &received.DeployResp)
	assert.Equal(t, *received.DeployResp.Id, expected.ID, "they should be equal")
	assert.Equal(t, *received.DeployResp.Application, expected.Application, "they should be equal")
	assert.Equal(t, *received.DeployResp.Status, expected.Status, "they should be equal")
	receivedDeployment := *received.DeployResp.Steps
	expectedDeployment := expected.Steps[0].Deployment
	assert.Equal(t, receivedDeployment[0].Deployment.ID, expectedDeployment.ID, "they should be equal")
}

func TestDeployStatusHttpError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	responder, err := httpmock.NewJsonResponder(500, `{"code":2, "message":"invalid operation", "details":[]}`)
	if err != nil {
		t.Fatalf("TestDeployStatusHttpError failed with: %s", err)
	}
	httpmock.RegisterResponder("GET", "https://localhost/pipelines/12345", responder)
	cmd := NewDeployCmd(getStdTestConfig("json"))
	outWriter := bytes.NewBufferString("")
	cmd.SetOut(outWriter)
	args := []string{
		"status",
		"--deploymentId=12345",
	}
	cmd.SetArgs(args)
	err = cmd.Execute()
	if err != nil {
		t.Fatal("TestDeployStatusHttpError failed with: error should not be null")
	}
	output, err := ioutil.ReadAll(outWriter)
	if err != nil {
		t.Fatalf("TestDeployStatusHttpError failed with: %s", err)
	}
	assert.Equal(t, `{ "error": "request returned an error: status code(500), thrown error: "{\"code\":2, \"message\":\"invalid operation\", \"details\":[]}"" }`,
		strings.TrimSpace(string(output)), "they should be equal")
}

func TestDeployStatusFlagDeploymentIdRequired(t *testing.T) {
	cmd := NewDeployCmd(getStdTestConfig("json"))
	outWriter := bytes.NewBufferString("")
	cmd.SetOut(outWriter)
	args := []string{
		"status",
	}
	cmd.SetArgs(args)
	err := cmd.Execute()
	if err == nil {
		t.Fatal("TestDeployStatusFlagDeploymentIdRequired failed with: error should not be null")
	}
	assert.EqualError(t, err, "required flag(s) \"deploymentId\" not set")
}
