package template

import (
	"github.com/armory/armory-cli/pkg/util"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const (
	kubernetesShort = "Generate a Kubernetes deployment template."
)

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
	return command
}

func buildTemplateKubernetesCore() *yaml.Node {
	root := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	//Core
	root.Content = append(root.Content, util.BuildStringNode("version", "v1", "")...)
	root.Content = append(root.Content, util.BuildStringNode("kind", "kubernetes", "")...)
	root.Content = append(root.Content, util.BuildStringNode("application", "<App Name>", "The name of the deployed application.")...)

	// Target root
	targetNode, targetValuesNode := util.BuildMapNode("targets","Map of Deployment Targets, " +
		"Map of deployment targets. You can specify more than one target")
	devNode, devValuesNode := util.BuildMapNode("dev-west",
		"Specify a deployment target. The identifier for a deployment target is its name.")
	devValuesNode.Content = append(devValuesNode.Content, util.BuildStringNode("account",
		"account-name", "The name of an agent configured account")...)
	devValuesNode.Content = append(devValuesNode.Content, util.BuildStringNode("namespace",
		"namespace", "Optionally override the namespaces that are in the manifests")...)
	devValuesNode.Content = append(devValuesNode.Content, util.BuildStringNode("strategy",
		"strategy1", "This is the key to a strategy specified in the strategies map below")...)

	constraintNode, constraintValuesNode := util.BuildMapNode("constraints", "")
	dependsOnNode, dependsOnValuesNode := util.BuildSequenceNode("dependsOn", "Defines the deployments that must reach a successful state (defined as status == SUCCEEDED) before this deployment can start.Deployments with the same dependsOn criteria will execute in parallel.")
	beforeNode, beforeValuesNode := util.BuildSequenceNode("beforeDeployment", "A set of steps that are executed in parallel")
	pause := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	pauseNode, pauseValuesNode := util.BuildMapNode("pause", "The map key is the step type")
	pauseValuesNode.Content = append(pauseValuesNode.Content, util.BuildIntNode("duration", "1", "The duration of the pause before the deployment continues. If duration is not zero, set untilApproved to false.")...)
	pauseValuesNode.Content = append(pauseValuesNode.Content, util.BuildStringNode("unit", "SECONDS", "")...)
	pauseValuesNode.Content = append(pauseValuesNode.Content, util.BuildBoolNode("untilApproved", "false",
		"If set to true, the deployment waits until a manual approval to continue. Only set this to true if duration and unit are not set.")...)
	pause.Content = append(pause.Content, pauseNode, pauseValuesNode)
	beforeValuesNode.Content = append(beforeValuesNode.Content, pause)
	constraintValuesNode.Content = append(constraintValuesNode.Content, beforeNode, beforeValuesNode, dependsOnNode, dependsOnValuesNode)
	devValuesNode.Content = append(devValuesNode.Content, constraintNode, constraintValuesNode)

	targetValuesNode.Content = append(targetValuesNode.Content, devNode, devValuesNode)
	root.Content = append(root.Content, targetNode, targetValuesNode)

	// Manifest sequence/array
	manifestsNode, manifestValuesNode := util.BuildSequenceNode("manifests", "The list of manifest sources.")

	targetsOnNode, targetsValuesNode := util.BuildSequenceNode("targets", "")
	targetsValuesNode.Content = append(targetsValuesNode.Content, &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!str",
		Value: "dev-west",
	})

	path := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	path2 := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	path.Content = append(path.Content, util.BuildStringNode("path", "infrastructure/manifests/configmaps", "This will read all yaml|yml files in a dir and deploy all manifests in that directory to all targets.")...)
	path.Content = append(path.Content, targetsOnNode, targetsValuesNode)
	path2.Content = append(path2.Content, util.BuildStringNode("path", "infrastructure/manifests/deployment.yaml",
		"This will read all yaml|yml files in a dir and deploy all manifests in that directory to all targets.")...)
	path2.Content = append(path2.Content, targetsOnNode, targetsValuesNode)
	manifestValuesNode.Content = append(manifestValuesNode.Content, path, path2)

	root.Content = append(root.Content, manifestsNode, manifestValuesNode)

	return root
}
