package deploy

import (
	"bytes"
	"encoding/json"
	"fmt"
	de "github.com/armory-io/deploy-engine/pkg"
	"github.com/armory/armory-cli/cmd"
	"github.com/armory/armory-cli/pkg/deploy"
	"github.com/armory/armory-cli/pkg/output"
	"github.com/jarcoal/httpmock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestDeployStartJsonSuccess(t *testing.T) {
	expected := &de.DeploymentV2StartDeploymentResponse{
		DeploymentId: "12345",
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	responder, err := httpmock.NewJsonResponder(200, expected)
	if err != nil {
		t.Fatalf("TestDeployStartJsonSuccess failed with: %s", err)
	}
	httpmock.RegisterResponder("POST", "https://localhost/deployments/kubernetes", responder)

	tempFile := tempAppFile(testAppYamlStr)
	if tempFile == nil {
		t.Fatal("TestDeployStartJsonSuccess failed with: Could not create temp app file.")
	}
	defer os.Remove(tempFile.Name())

	outWriter := bytes.NewBufferString("")
	rootCmd, options, err := getOverrideRootCmd(outWriter)
	if err != nil {
		t.Fatalf("TestDeployStartJsonSuccess failed with: %s", err)
	}
	rootCmd.AddCommand(NewDeployCmd(options))

	args := []string{
		"deploy", "start",
		"--file=" + tempFile.Name(),
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
	var received = FormattableDeployStartResponse{}
	json.Unmarshal(output, &received)
	assert.Equal(t, received.DeploymentId, expected.GetDeploymentId(), "they should be equal")
}

func TestDeployStartYAMLSuccess(t *testing.T) {
	expected := &de.DeploymentV2StartDeploymentResponse{
		DeploymentId: "12345",
	}
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	responder, err := httpmock.NewJsonResponder(200, expected)
	if err != nil {
		t.Fatalf("TestDeployStartYAMLSuccess failed with: %s", err)
	}
	httpmock.RegisterResponder("POST", "https://localhost/deployments/kubernetes", responder)

	tempFile := tempAppFile(testAppYamlStr)
	if tempFile == nil {
		t.Fatal("TestDeployStartYAMLSuccess failed with: Could not create temp app file.")
	}
	defer os.Remove(tempFile.Name())

	outWriter := bytes.NewBufferString("")
	rootCmd, options, err := getOverrideRootCmd(outWriter)
	if err != nil {
		t.Fatalf("TestDeployStartYAMLSuccess failed with: %s", err)
	}
	rootCmd.AddCommand(NewDeployCmd(options))

	args := []string{
		"deploy", "start",
		"--file=" + tempFile.Name(),
		"--output=yaml",
	}
	rootCmd.SetArgs(args)
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("TestDeployStartYAMLSuccess failed with: %s", err)
	}
	output, err := ioutil.ReadAll(outWriter)
	if err != nil {
		t.Fatalf("TestDeployStartYAMLSuccess failed with: %s", err)
	}
	var received = FormattableDeployStartResponse{}
	yaml.Unmarshal(output, &received)
	assert.Equal(t, received.DeploymentId, expected.GetDeploymentId(), "they should be equal")
}

func TestDeployStartHttpError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	responder, err := httpmock.NewJsonResponder(500, `{"code":2, "message":"invalid operation", "details":[]}`)
	if err != nil {
		t.Fatalf("TestDeployStartYAMLSuccess failed with: %s", err)
	}
	httpmock.RegisterResponder("POST", "https://localhost/deployments/kubernetes", responder)

	tempFile := tempAppFile(testAppYamlStr)
	if tempFile == nil {
		t.Fatal("TestDeployStartHttpError failed with: Could not create temp app file.")
	}
	defer os.Remove(tempFile.Name())

	outWriter := bytes.NewBufferString("")
	rootCmd, options, err := getOverrideRootCmd(outWriter)
	if err != nil {
		t.Fatalf("TestDeployStartHttpError failed with: %s", err)
	}
	rootCmd.AddCommand(NewDeployCmd(options))

	args := []string{
		"deploy", "start",
		"--file=" + tempFile.Name(),
		"--output=yaml",
	}
	rootCmd.SetArgs(args)
	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("TestDeployStartHttpError failed with: %s", err)
	}
	output, err := ioutil.ReadAll(outWriter)
	if err != nil {
		t.Fatalf("TestDeployStartHttpError failed with: %s", err)
	}
	assert.Equal(t, `error: "request returned an error: status code(500) "{\"code\":2, \"message\":\"invalid operation\", \"details\":[]}"`,
		strings.TrimSpace(string(output)), "they should be equal")
}

func TestDeployStartFlagFileRequired(t *testing.T) {
	outWriter := bytes.NewBufferString("")
	rootCmd, options, err := getOverrideRootCmd(outWriter)
	if err != nil {
		t.Fatalf("TestDeployStartFlagRequired failed with: %s", err)
	}
	rootCmd.AddCommand(NewDeployCmd(options))

	args := []string{
		"deploy", "start",
		"--output=json",
	}
	rootCmd.SetArgs(args)
	err = rootCmd.Execute()
	if err == nil {
		t.Fatal("TestDeployStartFlagRequired failed with: error should not be null")
	}
	assert.EqualError(t, err,"required flag(s) \"file\" not set")
}

func TestDeployStartBadPath(t *testing.T) {
	outWriter := bytes.NewBufferString("")
	rootCmd, options, err := getOverrideRootCmd(outWriter)
	if err != nil {
		t.Fatalf("TestDeployStartBadPath failed with: %s", err)
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
		t.Fatal("TestDeployStartBadPath failed with: error should not be null")
	}
	assert.EqualError(t, err,"error trying to read the YAML file: open /badPath/test.yml: no such file or directory")
}

func tempAppFile(appContent string) *os.File {
	tempFile, _ := ioutil.TempFile("" /* /tmp dir. */, "app")
	bytes, err := tempFile.Write([]byte(appContent))
	if err != nil || bytes == 0 {
		fmt.Println("Could not write temp file.")
		return nil
	}
	return tempFile
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