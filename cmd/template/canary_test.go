package template

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

func TestNewTemplateCanaryCmd(t *testing.T) {
	cmd := NewTemplateCanaryCmd(&templateOptions{})
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.Execute()
	expectTemplate := `version: v1
kind: kubernetes

# The name of the application to deploy.
application: <application>

# Targets is a map of deployment environments, keyed by name.
# A named target is an (account, namespace) tuple; the tuple must be unique
# within a deployment. You can deploy to one or more targets within
# a deployment pipeline.
targets:

  # Target name. Use a descriptive value (e.g., prod or staging).
  <target>:

      # An account corresponds to a Kubernetes cluster.
      # You can create and configure accounts inside Cloud Console
      # or by installing Armory RNA inside a cluster.
      account: <accountName>

      # If provided, namespace overrides the "namespace" value
      # in all manifests deployed to this target. Recommended.
      namespace: <namespace>

      # A named strategy from the "strategies" block, configured below.
      # This strategy is used when deploying manifests to this target.
      strategy: <strategy>

# The list of manifest sources. Each entry can be a directory or file.
manifests:

  # Read all yaml|yml files in the directory and deploy all the manifests found.
  - path: path/to/manifests

    # The deployment targets that should use the manifest. Used for all targets if omitted.
    targets: ["<target>"]

  # Deploy this specific manifest.
  - path: path/to/manifest.yaml

# A map of deployment strategies, keyed by name.
strategies:

  # Strategy name. Use a descriptive name (e.g., "prod-strategy").
  # Use a strategy by assigning it to a deployment target above.
  <strategy>:

    # Define a progressive canary deployment strategy.
    # Use this strategy to incrementally shift traffic to a new version of
    # your software.
    canary:

      # A set of deployment steps. The steps are executed in order.
      steps:

        # A pause step type.
        # The deployment stops until the pause behavior is complete.
        # The pause type defined below is a duration-based pause.
        - pause:

            # Pause the deployment for <duration> <unit> (e.g., pause for 5 minutes).
            # A duration-based pause should omit the "untilApproved" flag.
            duration: 1

            # The pause's time unit. One of seconds, minutes, or hours.
            # Required if duration is set.
            unit: seconds

        # The percentage of pods that should be running the new version
        # of your software.
        # Use an integer between 0 and 100, inclusive.
        - setWeight:
            weight: 33

        # A pause step type.
        # The pause type defined below is a judgment-based pause.
        - pause:

            # Pause the deployment until manual approval.
            # You can approve or rollback a deployment in the Cloud Console.
            # Do not provide a "duration" or "unit" value when defining
            # a judgment-based pause.
            untilApproved: true
`

	actualTemplate, err := ioutil.ReadAll(b)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, expectTemplate, string(actualTemplate))
}