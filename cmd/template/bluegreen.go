package template

import (
	"github.com/armory/armory-cli/pkg/cmdUtils"
	errorUtils "github.com/armory/armory-cli/pkg/errors"
	"github.com/spf13/cobra"
	"strings"
)

const (
	templateBlueGreenShort   = "Generate a blue/green deployment template"
	templateBlueGreenLong    = "Generate a blue/green deployment template in YAML format"
	templateBlueGreenExample = "armory template bluegreen > bluegreen.yaml"
)

const KubernetesCoreTemplate = `version: v1
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

      # The name that you assigned to the deployment target cluster when you installed the RNA.
      account: <accountName>

      # (Recommended) Set the namespace that the app gets deployed to. 
      # Overrides the namespaces that are in your manifests.
      namespace: <namespace>

      # A named strategy from the strategies block.
      strategy: <strategy>

# The list of manifest sources. Each entry can be a directory or file.
manifests:

  # Read all yaml|yml files in the directory and deploy all the manifests found.
  - path: path/to/manifests

    # The deployment targets that should use the manifest. Used for all targets if omitted.
    targets: ["<target>"]

  # Deploy this specific manifest.
  - path: path/to/manifest.yaml`

const blueGreenTemplate = `
# A map of deployment strategies, keyed by name.
strategies:

  # Strategy name. Use a descriptive name (e.g., "prod-strategy").
  # Use a strategy by assigning it to a deployment target above.
  <strategy>:

    # Define a blue/green deployment strategy.
    blueGreen:

      # The redirectTrafficAfter steps are pre-conditions for exposing the new
      # version to the activeService. The steps are executed
      # in parallel.
      redirectTrafficAfter:

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

        # A pause step type.
        # The pause type defined below is a judgment-based pause.
        - pause:

            # Pause the deployment until manual approval.
            # You can approve or rollback a deployment in the Cloud Console.
            # Do not provide a "duration" or "unit" value when defining
            # a judgment-based pause.
            untilApproved: true

      # The shutDownOldVersionAfter steps are pre-conditions for deleting the old
      # version of your software. The steps are executed in parallel.
      shutDownOldVersionAfter:
        - pause:
            untilApproved: true

trafficManagement:

   # Determines the targets that this configuration should be applied to.
   # If omitted, this configuration is applied to all targets.
 - targets: ["<target>"]

   # The Kubernetes traffic management block defines the Kubernetes Service
   # resources that should be manipulated as part of the deployment.
   #
   # You can specify more than one active / preview pair here; pairs are
   # automatically joined with your Kubernetes Deployment resources according to
   # their label selectors.
   kubernetes:

     # The name of a Kubernetes Service resource.
     # The activeService must be deployed out-of-band and should be configured
     # to direct traffic to your application.
   - activeService: active-service

     # The name of a Kubernetes Service resource. Optional.
     # The previewService must be deployed out-of-band and should be configured
     # to direct traffic to your application. You can use this service to
     # preview the new version of your application before it is exposed to users.
     previewService: preview-service
`

func NewTemplateBlueGreenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "bluegreen",
		Aliases: []string{"bluegreen"},
		Short:   templateBlueGreenShort,
		Long:    templateBlueGreenLong,
		Example: templateBlueGreenExample,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cmdUtils.ExecuteParentHooks(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return blueGreen(cmd)
		},
	}
	return cmd
}

func blueGreen(cmd *cobra.Command) error {
	template := strings.Join([]string{KubernetesCoreTemplate, blueGreenTemplate}, "\n")
	_, err := cmd.OutOrStdout().Write([]byte(template))
	if err != nil {
		return errorUtils.NewWrappedError(ErrBlueGreenTemplateParse, err)
	}
	return nil
}
