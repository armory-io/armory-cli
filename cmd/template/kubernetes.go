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
		"<namespace>", "(Recommended) Set the namespace where the app gets deployed to. Overrides the namespaces that are in your manifests")...)
	devValuesNode.Content = append(devValuesNode.Content, util.BuildStringNode("strategy",
		"strategy1", "A named strategy from the strategies block. This example uses the name strategy1.")...)
	targetValuesNode.Content = append(targetValuesNode.Content, devNode, devValuesNode)
	root.Content = append(root.Content, targetNode, targetValuesNode)

	// Manifest sequence/array
	manifestsNode, manifestValuesNode := util.BuildSequenceNode("manifests", "The list of manifest sources. Can be a directory or file.")

	path := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	path2 := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	path.Content = append(path.Content, util.BuildStringNode("path", "path/to/manifests", "Read all yaml|yml files in the directory and deploy all the manifests found.")...)
	path2.Content = append(path2.Content, util.BuildStringNode("path", "path/to/manifest.yaml",
		"Deploy this specific manifest.")...)
	manifestValuesNode.Content = append(manifestValuesNode.Content, path, path2)

	root.Content = append(root.Content, manifestsNode, manifestValuesNode)

	return root
}
