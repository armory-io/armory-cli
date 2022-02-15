package template

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

func TestNewTemplateBlueGreenCmd(t *testing.T) {
	cmd := NewTemplateBlueGreenCmd(&templateOptions{})
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.Execute()
	expectTemplate :=
		`version: v1
kind: kubernetes
application: <AppName> # The name of the application to deploy.
targets: # Map of your deployment target, Borealis supports deploying to one target cluster.
    <deploymentName>: # Name for your deployment. Use a descriptive value such as the environment name.
        account: <accountName> # The account name that was assigned to the deployment target when you installed the RNA.
        namespace: <namespace> # (Recommended) Set the namespace that the app gets deployed to. Overrides the namespaces that are in your manifests
        strategy: strategy1 # A named strategy from the strategies block. This example uses the name strategy1.
        constraints:
            # A set of steps that are executed in parallel
            beforeDeployment:
                - pause: # The map key is the step type
                    # The duration of the pause before the deployment continues. If duration is not zero, set untilApproved to false.
                    duration: 1
                    unit: SECONDS
                    # If set to true, the deployment waits until a manual approval to continue. Only set this to true if duration and unit are not set.
                    untilApproved: false
            # Defines the deployments that must reach a successful state (defined as status == SUCCEEDED) before this deployment can start.Deployments with the same dependsOn criteria will execute in parallel.
            dependsOn: []
# The list of manifest sources. Can be a directory or file.
manifests:
    - path: path/to/manifests # Read all yaml|yml files in the directory and deploy all the manifests found.
      targets:
        - dev-west
    - path: path/to/manifest.yaml # Deploy this specific manifest.
      targets:
        - dev-west
strategies: # A map of named strategies that can be assigned to deployment targets in the targets block.
    strategy1: # Name for a strategy that you use to refer to it. Used in the target block. This example uses strategy1 as the name.
        blue-green: # The deployment strategy type. Use blueGreen.
            # The steps that must be completed before traffic is redirected to the new version.
            redirectTrafficAfter:
                - pause: # A pause step type. The pipeline stops until the pause behavior is completed.
                    # The pause behavior is time (integer) before the deployment continues. If duration is set for this step, omit untilApproved.
                    duration: 1
                    unit: seconds # The unit of time to use for the pause. Can be seconds, minutes, or hours. Required if duration is set.
                - pause: # A pause step type. The pipeline stops until the pause behavior is completed.
                    # The pause behavior is the deployment waits until a manual approval is given to continue. Only set this to true if there is no duration pause behavior for this step.
                    untilApproved: true
            # The steps that must be completed before the old version is scaled down.
            shutdownOldVersionAfter:
                - pause: # A pause step type. The pipeline stops until the pause behavior is completed.
                    # The pause behavior is time (integer) before the deployment continues. If duration is set for this step, omit untilApproved.
                    duration: 1
                    unit: seconds # The unit of time to use for the pause. Can be seconds, minutes, or hours. Required if duration is set.
                - pause: # A pause step type. The pipeline stops until the pause behavior is completed.
                    # The pause behavior is the deployment waits until a manual approval is given to continue. Only set this to true if there is no duration pause behavior for this step.
                    untilApproved: true
            activeService: <ActiveService> # The active service that will receive traffic (required)
            previewService: <PreviewService> # The preview service that will not receive traffic until the new version is deployed (optional)
            activeRootUrl: <ActiveRootUrl> # The old version is available on the activeRootUrl before the traffic is swapped. After the redirectTrafficAfter steps activeRootUrl will point to the new version. (optional)
            previewRootUrl: <PreviewRootUrl> # The new version is available on the previewRootUrl before traffic is redirected to it. (optional)
`
	actualTemplate, err := ioutil.ReadAll(b)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, expectTemplate, string(actualTemplate))
}