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
	root.Content = append(root.Content, util.BuildStringNode("version", "v1", "")...)
	root.Content = append(root.Content, util.BuildStringNode("kind", "kubernetes", "")...)
	root.Content = append(root.Content, util.BuildStringNode("application", "<App Name>", "The name of the deployed application.")...)

	// Target root
	targetNode, targetValuesNode := util.BuildMapNode("targets","Map of Deployment Targets, " +
		"this is set up in a way where we can do multi-target deployments (multi-region or multi-cluster)")
	devNode, devValuesNode := util.BuildMapNode("dev-west","This in the name of a deployment, underneath it are its configuration")
	devValuesNode.Content = append(devValuesNode.Content, util.BuildStringNode("account",
		"armory-cloud-hosted-services", "The name of an agent configured account")...)
	devValuesNode.Content = append(devValuesNode.Content, util.BuildStringNode("namespace",
		"hosted-services-dev", "Optionally override the namespaces that are in the manifests")...)
	devValuesNode.Content = append(devValuesNode.Content, util.BuildStringNode("strategy",
		"strategy1", "This is the key to a strategy under the strategies map")...)
	targetValuesNode.Content = append(targetValuesNode.Content, devNode, devValuesNode)
	root.Content = append(root.Content, targetNode, targetValuesNode)

	// Manifest sequence/array
	manifestsNode, manifestValuesNode := util.BuildSequenceNode("manifests", "The list of manifests sources.")

	path := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	path2 := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	path.Content = append(path.Content, util.BuildStringNode("path", "infrastructure/manifests/configmaps", "This will read all yaml|yml files in a dir and deploy all manifests in that directory to all targets.")...)
	path2.Content = append(path2.Content, util.BuildStringNode("path", "infrastructure/manifests/deployment.yaml",
		"This will read all yaml|yml files in a dir and deploy all manifests in that directory to all targets.")...)
	manifestValuesNode.Content = append(manifestValuesNode.Content, path, path2)
	root.Content = append(root.Content, manifestsNode, manifestValuesNode)


	return root
}

