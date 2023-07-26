package deploy

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"

	de "github.com/armory-io/deploy-engine/pkg/api"
	"github.com/armory/armory-cli/pkg/config"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func getExpectedPipelineDeployment() *de.PipelineStatusResponse {
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

	return expected
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
	expected := getExpectedPipelineDeployment()
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	responder, err := httpmock.NewJsonResponder(200, expected)
	if err != nil {
		t.Fatalf("TestDeployStatusJsonSuccess failed with: %s", err)
	}
	httpmock.RegisterResponder("GET", "https://localhost/pipelines/12345", responder)

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
	output, err := io.ReadAll(outWriter)
	if err != nil {
		t.Fatalf("TestDeployStartJsonSuccess failed with: %s", err)
	}
	var received = FormattableDeployStatus{
		Pipeline: &de.PipelineStatusResponse{},
	}
	json.Unmarshal(output, &received.Pipeline)
	assert.Equal(t, received.Pipeline.ID, expected.ID, "they should be equal")
	assert.Equal(t, received.Pipeline.Application, expected.Application, "they should be equal")
	assert.Equal(t, received.Pipeline.Status, expected.Status, "they should be equal")
	assert.Equal(t, len(received.Pipeline.Steps), len(expected.Steps), "they should be equal")
	receivedDeployment := received.Pipeline.Steps
	expectedDeployment := expected.Steps[0].Deployment
	assert.Equal(t, receivedDeployment[0].Deployment.ID, expectedDeployment.ID, "they should be equal")
}

func TestDeployStatusYAMLSuccess(t *testing.T) {
	expected := getExpectedPipelineDeployment()
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	responder, err := httpmock.NewJsonResponder(200, expected)
	if err != nil {
		t.Fatalf("TestDeployStatusYAMLSuccess failed with: %s", err)
	}
	httpmock.RegisterResponder("GET", "https://localhost/pipelines/12345", responder)

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
	output, err := io.ReadAll(outWriter)
	if err != nil {
		t.Fatalf("TestDeployStatusYAMLSuccess failed with: %s", err)
	}
	var received = FormattableDeployStatus{
		Pipeline: &de.PipelineStatusResponse{},
	}
	yaml.Unmarshal(output, &received.Pipeline)
	assert.Equal(t, received.Pipeline.ID, expected.ID, "they should be equal")
	assert.Equal(t, received.Pipeline.Application, expected.Application, "they should be equal")
	assert.Equal(t, received.Pipeline.Status, expected.Status, "they should be equal")
	receivedDeployment := received.Pipeline.Steps
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
	output, err := io.ReadAll(outWriter)
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
