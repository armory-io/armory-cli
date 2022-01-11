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
	root.Content = append(root.Content, util.BuildStringNode("application", "<AppName>", "The name of the application to deploy.")...)

	// Target root
	targetNode, targetValuesNode := util.BuildMapNode("targets","Map of your deployment target, " +
		"Borealis supports deploying to one target cluster.")
	devNode, devValuesNode := util.BuildMapNode("<deploymentName>",
		"Name for your deployment. Use a descriptive value such as the environment name.")
	devValuesNode.Content = append(devValuesNode.Content, util.BuildStringNode("account",
		"<accountName>", "The account name that was assigned to the deployment target when you installed the RNA.")...)
	devValuesNode.Content = append(devValuesNode.Content, util.BuildStringNode("namespace",
		"<namespace>", "(Recommended) Set the namespace that the app gets deployed to. Overrides the namespaces that are in your manifests")...)
	devValuesNode.Content = append(devValuesNode.Content, util.BuildStringNode("strategy",
		"strategy1", "A named strategy from the strategies block. This example uses the name strategy1.")...)

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
	manifestsNode, manifestValuesNode := util.BuildSequenceNode("manifests", "The list of manifest sources. Can be a directory or file.")

	targetsOnNode, targetsValuesNode := util.BuildSequenceNode("targets", "")
	targetsValuesNode.Content = append(targetsValuesNode.Content, &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!str",
		Value: "dev-west",
	})

	path := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	path2 := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	path.Content = append(path.Content, util.BuildStringNode("path", "path/to/manifests", "Read all yaml|yml files in the directory and deploy all the manifests found.")...)
	path.Content = append(path.Content, targetsOnNode, targetsValuesNode)
	path2.Content = append(path2.Content, util.BuildStringNode("path", "path/to/manifest.yaml",
		"Deploy this specific manifest.")...)
	path2.Content = append(path2.Content, targetsOnNode, targetsValuesNode)
	manifestValuesNode.Content = append(manifestValuesNode.Content, path, path2)

	root.Content = append(root.Content, manifestsNode, manifestValuesNode)

	return root
}
