package template

import (
	"github.com/armory/armory-cli/pkg/util"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const (
	kubernetesShort   = "Generate an kubernetes deployment template"
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
	root.Content = append(root.Content, util.BuildStringNode("account", "<Account>", "The name of an agent configured account.")...)
	root.Content = append(root.Content, util.BuildStringNode("namespace", "<Namespace>", "The Kubernetes namespace where the provided manifests will be deployed.")...)
	// Manifest sequence/array
	manifestsNode, manifestValuesNode := util.BuildSequenceNode("manifests", "The list of manifests sources.")

	// Inline root
	inline := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	inlineNode, inlineValuesNode := util.BuildMapNode("inline","This map key, is the manifest source type.")
	inlineValuesNode.Content = append(inlineValuesNode.Content, util.BuildStringNode("value", "| apiVersion: apps/v1...", "A YAML-encoded string containing a Kubernetes resource manifest.")...)
	inline.Content = append(inline.Content, inlineNode, inlineValuesNode)

	manifestValuesNode.Content = append(manifestValuesNode.Content, inline)
	root.Content = append(root.Content, manifestsNode, manifestValuesNode)

	return root
}

