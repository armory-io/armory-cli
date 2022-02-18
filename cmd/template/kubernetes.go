package template

import (
	"github.com/spf13/cobra"
)

const (
	kubernetesShort = "Generate a Kubernetes deployment template."
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
  - path: path/to/manifest.yaml`

func NewTemplateKubernetesCmd(rootOptions *templateOptions) *cobra.Command {
	command := &cobra.Command{
		Use:     "kubernetes",
		Aliases: []string{"kubernetes"},
		Short:   kubernetesShort,
		Long:    "",
		Example: "",
	}
	// create subcommands
	command.AddCommand(NewTemplateCanaryCmd(rootOptions))
	command.AddCommand(NewTemplateBlueGreenCmd(rootOptions))
	return command
}