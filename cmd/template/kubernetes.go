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
	root.Content = append(root.Content, util.BuildStringNode("application", "<App Name>", "The name of the deployed application.")...)
	root.Content = append(root.Content, util.BuildStringNode("account", "<Account>", "The account of the deployment target. You set the account name when you installed the agent.")...)
	root.Content = append(root.Content, util.BuildStringNode("namespace", "<Namespace>", "The Kubernetes namespace where you want to deploy the manifest.")...)
	// Manifest sequence/array
	manifestsNode, manifestValuesNode := util.BuildSequenceNode("manifests", "The list of manifest sources.")

	// Inline root
	inline := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	inlineNode, inlineValuesNode := util.BuildMapNode("inline","This map key is the manifest source type.")
	inlineValuesNode.Content = append(inlineValuesNode.Content, util.BuildStringNode("value", "| apiVersion: apps/v1...", "A YAML-encoded string containing a Kubernetes resource manifest.")...)
	inline.Content = append(inline.Content, inlineNode, inlineValuesNode)

	manifestValuesNode.Content = append(manifestValuesNode.Content, inline)
	root.Content = append(root.Content, manifestsNode, manifestValuesNode)

	return root
}
