package deploy

import (
	"bytes"
	"encoding/json"
	de "github.com/armory-io/deploy-engine/pkg"
	"github.com/armory/armory-cli/pkg/model"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"strings"
	"testing"
)

func TestDeployStatusJsonSuccess(t *testing.T) {
	expected := &de.PipelinePipelineStatusResponse{}
	expected.SetId("12345")
	expected.SetApplication("app")
	expected.SetStatus(de.PIPELINEPIPELINESTATUS_RUNNING)
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	responder, err := httpmock.NewJsonResponder(200, expected)
	if err != nil {
		t.Fatalf("TestDeployStatusJsonSuccess failed with: %s", err)
	}
	httpmock.RegisterResponder("GET", "https://localhost/pipelines/12345", responder)
	outWriter := bytes.NewBufferString("")
	rootCmd, options, err := getOverrideRootCmd(outWriter)
	if err != nil {
		t.Fatalf("TestDeployStatusJsonSuccess failed with: %s", err)
	}
	rootCmd.AddCommand(NewDeployCmd(options))
	args := []string{
		"deploy", "status",
		"--deploymentId=12345",
		"--output=json",
	}
	rootCmd.SetArgs(args)
	err = rootCmd.Execute()
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
}

func TestDeployStatusYAMLSuccess(t *testing.T) {
	expected := &de.PipelinePipelineStatusResponse{}
	expected.SetId("12345")
	expected.SetApplication("app")
	expected.SetStatus(de.PIPELINEPIPELINESTATUS_RUNNING)
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	responder, err := httpmock.NewJsonResponder(200, expected)
	if err != nil {
		t.Fatalf("TestDeployStatusYAMLSuccess failed with: %s", err)
	}
	httpmock.RegisterResponder("GET", "https://localhost/pipelines/12345", responder)
	outWriter := bytes.NewBufferString("")
	rootCmd, options, err := getOverrideRootCmd(outWriter)
	if err != nil {
		t.Fatalf("TestDeployStatusYAMLSuccess failed with: %s", err)
	}
	rootCmd.AddCommand(NewDeployCmd(options))
	args := []string{
		"deploy", "status",
		"--deploymentId=12345",
		"--output=yaml",
	}
	rootCmd.SetArgs(args)
	err = rootCmd.Execute()
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
}

func TestDeployStatusHttpError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	responder, err := httpmock.NewJsonResponder(500, `{"code":2, "message":"invalid operation", "details":[]}`)
	if err != nil {
		t.Fatalf("TestDeployStatusHttpError failed with: %s", err)
	}
	httpmock.RegisterResponder("GET", "https://localhost/pipelines/12345", responder)
	outWriter := bytes.NewBufferString("")
	rootCmd, options, err := getOverrideRootCmd(outWriter)
	if err != nil {
		t.Fatalf("TestDeployStatusHttpError failed with: %s", err)
	}
	rootCmd.AddCommand(NewDeployCmd(options))
	args := []string{
		"deploy", "status",
		"--deploymentId=12345",
		"--output=json",
	}
	rootCmd.SetArgs(args)
	err = rootCmd.Execute()
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
	outWriter := bytes.NewBufferString("")
	rootCmd, options, err := getOverrideRootCmd(outWriter)
	if err != nil {
		t.Fatalf("TestDeployStatusFlagDeploymentIdRequired failed with: %s", err)
	}
	rootCmd.AddCommand(NewDeployCmd(options))

	args := []string{
		"deploy", "status",
		"--output=json",
	}
	rootCmd.SetArgs(args)
	err = rootCmd.Execute()
	if err == nil {
		t.Fatal("TestDeployStatusFlagDeploymentIdRequired failed with: error should not be null")
	}
	assert.EqualError(t, err,"required flag(s) \"deploymentId\" not set")
}
