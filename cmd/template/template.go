package template

import (
	"github.com/armory/armory-cli/cmd"
	"github.com/armory/armory-cli/pkg/util"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const (
	templateShort   = "Generate an Armory's deployment template"
	templateLong    = ""
	templateExample = ""
)

type templateOptions struct {
	*cmd.RootOptions
	deploymentFile string
}

func NewTemplateCmd(rootOptions *cmd.RootOptions) *cobra.Command {
	options := &templateOptions{
		RootOptions: rootOptions,
	}
	command := &cobra.Command{
		Use:     "template",
		Aliases: []string{"template"},
		Short:   templateShort,
		Long:    templateLong,
		Example: templateExample,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	// create subcommands
	command.AddCommand(NewTemplateCanaryCmd(options))
	return command
}

func buildTemplateCore() *yaml.Node {
	root := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	//Core
	root.Content = append(root.Content, util.BuildStringNode("application", "<App Name>", "The name of the deployed application.")...)
	root.Content = append(root.Content, util.BuildStringNode("account", "<Account>", "The name of the Kubernetes account to be used for this deployment.")...)
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
