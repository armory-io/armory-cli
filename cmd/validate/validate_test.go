package validate

import (
	"bytes"
	"fmt"
	"github.com/armory/armory-cli/pkg/config"
	"github.com/armory/armory-cli/pkg/util"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

func getValidateCmdWithFileName(outWriter io.Writer, fileName string, output string, additionalOpts ...string) *cobra.Command {
	configuration := getTestConfig(output)
	validateCmd := NewValidateCmd(configuration)
	validateCmd.SetOut(outWriter)
	args := []string{
		"--file=" + fileName,
	}
	args = append(args, additionalOpts...)
	validateCmd.SetArgs(args)
	return validateCmd
}

func getTestConfig(output string) *config.Configuration {
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

func TestDeployValidate(t *testing.T) {
	cases := []struct {
		testName   string
		deployYaml string
		output     string
	}{
		{
			testName:   "valid yaml should pass",
			deployYaml: validDeployYamlStr,
			output:     "YAML is valid.\n",
		},
		{
			testName:   "invalid yaml should fail",
			deployYaml: invalidDeployYamlStr,
			output: `YAML is NOT valid. See the following errors:

#PipelineRequest.targets.dev_1.strategy: 1 errors in empty disjunction:

#PipelineRequest.targets.dev_1.strategy: conflicting values "strategy1" and "strategy0"

`,
		},
		{
			testName:   "invalid lambda yaml should pass",
			deployYaml: invalidLambdaDeployYamlStr,
			output:     "YAML is valid.\n",
		},
		{
			testName:   "RNA webhook without explicit agentIdentifier should pass",
			deployYaml: rnaWebhookWithImplicitAgent,
			output:     "YAML is valid.\n",
		},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("test-%s", c.testName), func(t *testing.T) {
			tempFile := util.TempAppFile("", "deploy.yaml", c.deployYaml)
			if tempFile == nil {
				t.Fatal("TestDeployValidateSuccess failed with: Could not create temp app file.")
			}
			outWriter := bytes.NewBufferString("")
			cmd := getValidateCmdWithFileName(outWriter, tempFile.Name(), "yaml")
			err := cmd.Execute()
			if err != nil {
				t.Fatalf("TestDeployValidateSuccess failed with: %s", err)
			}
			assert.Equal(t, c.output, outWriter.String())
		})
	}

}

const validDeployYamlStr = `
version: apps/v1
kind: kubernetes
application: deployment-test
targets:
  dev_1:
    account: dev
    namespace: dev-1
    strategy: strategy1
  dev_2:
    account: dev
    namespace: dev-2
    strategy: strategy1
    constraints: 
      dependsOn: ["dev_1"]
manifests:
  - path: deployment.yaml
strategies:
  strategy1:
    canary:
      steps:
        - pause:
            duration: 1
            unit: SECONDS
`

const invalidDeployYamlStr = `
version: apps/v1
kind: kubernetes
application: deployment-test
targets:
  dev_1:
    account: dev
    namespace: dev-1
    strategy: strategy0
  dev_2:
    account: dev
    namespace: dev-2
    strategy: strategy1
    constraints: 
    dependsOn: ["dev_1"]
manifests:
  - path: deployment.yaml
strategies:
  strategy1:
    canary:
      steps:
        - pause:
            duration: 1
            unit: SECONDS
`

const invalidLambdaDeployYamlStr = `
version: v1
kind: lambda
application: first-lambda-app
context:
  foo: bar
targets:
  firstTarget:
    account: firstAccount
    deployAsIamRole: "<some-deployment-role-arn>"
    region: us-west-2
    oops: baff
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

const rnaWebhookWithImplicitAgent = `
version: apps/v1
kind: kubernetes
application: deployment-test
targets:
  dev_1:
    account: dev
    namespace: dev-1
    strategy: strategy1
    constraints:
      afterDeployment:
        - runWebhook:
            name: Sample webhook
manifests:
  - path: deployment.yaml
strategies:
  strategy1:
    canary:
      steps:
        - setWeight:
            weight: 100
webhooks:
    - name: Sample webhook
      method: GET
      uriTemplate: https://webhook.site
      networkMode: remoteNetworkAgent
      disableCallback: true
`
