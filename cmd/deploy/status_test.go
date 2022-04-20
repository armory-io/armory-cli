package deploy

import (
	"bytes"
	"encoding/json"
	de "github.com/armory-io/deploy-engine/pkg"
	"github.com/armory/armory-cli/pkg/config"
	"github.com/armory/armory-cli/pkg/model"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"strings"
	"testing"
)

func getExpectedPipelineDeployment() (*de.PipelinePipelineStatusResponse, *de.DeploymentV2DeploymentStatusResponse) {
	expected := &de.PipelinePipelineStatusResponse{}
	expected.SetId("12345")
	expected.SetApplication("app")
	expected.SetStatus(de.WORKFLOWWORKFLOWSTATUS_RUNNING)
	stage := &de.PipelinePipelineStage{}
	stage.SetType("deployment")
	stage.SetStatus(de.WORKFLOWWORKFLOWSTATUS_RUNNING)
	deploy := &de.PipelinePipelineDeploymentStage{}
	deploy.SetId("5678")
	stage.SetDeployment(*deploy)
	expected.SetSteps([]de.PipelinePipelineStage{*stage})

	expectedDeploy := &de.DeploymentV2DeploymentStatusResponse{}
	expectedDeploy.SetId("5678")
	expectedDeploy.SetStatus(de.WORKFLOWWORKFLOWSTATUS_RUNNING)

	return expected, expectedDeploy
}

func getStdTestConfig(outFmt string) *config.Configuration {
	token := "some-token"
	addr := "https://localhost"
	clientId := ""
	clientSecret := ""
	return config.New(&config.Input{
		AccessToken:  &token,
		ApiAddr:      &addr,
		ClientId:     &clientId,
		ClientSecret: &clientSecret,
		OutFormat:    &outFmt,
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
	assert.Equal(t, *received.DeployResp.Id, expected.GetId(), "they should be equal")
	assert.Equal(t, *received.DeployResp.Application, expected.GetApplication(), "they should be equal")
	assert.Equal(t, *received.DeployResp.Status, expected.GetStatus(), "they should be equal")
	assert.Equal(t, len(*received.DeployResp.Steps), len(expected.GetSteps()), "they should be equal")
	receivedDeployment := *received.DeployResp.Steps
	expectedDeployment := expected.GetSteps()[0].GetDeployment()
	assert.Equal(t, receivedDeployment[0].Deployment.Id, expectedDeployment.GetId(), "they should be equal")
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
	assert.Equal(t, *received.DeployResp.Id, expected.GetId(), "they should be equal")
	assert.Equal(t, *received.DeployResp.Application, expected.GetApplication(), "they should be equal")
	assert.Equal(t, *received.DeployResp.Status, expected.GetStatus(), "they should be equal")
	receivedDeployment := *received.DeployResp.Steps
	expectedDeployment := expected.GetSteps()[0].GetDeployment()
	assert.Equal(t, receivedDeployment[0].Deployment.Id, expectedDeployment.GetId(), "they should be equal")
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
	assert.Equal(t, `{ "error": "request returned an error: status code(500) "{\"code\":2, \"message\":\"invalid operation\", \"details\":[]}"" }`,
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
