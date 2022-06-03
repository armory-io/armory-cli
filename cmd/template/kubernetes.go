package template

import (
	"fmt"
	"github.com/armory/armory-cli/pkg/cmdUtils"
	"github.com/armory/armory-cli/pkg/util"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const (
	kubernetesShort = "Generate a Kubernetes deployment template."
)

func NewTemplateKubernetesCmd() *cobra.Command {
	command := &cobra.Command{
		Use:     "kubernetes",
		Aliases: []string{"kubernetes"},
		Short:   kubernetesShort,
		Long:    "",
		Example: "",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cmdUtils.ExecuteParentHooks(cmd, args)
		},
	}
	// create subcommands
	command.AddCommand(NewTemplateCanaryCmd())
	command.AddCommand(NewTemplateBlueGreenCmd())
	cmdUtils.SetPersistentFlagsFromEnvVariables(command.Commands())
	return command
}

func buildTemplateKubernetesCore(options *templateCanaryOptions) (*yaml.Node, error) {
	root := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	//Core
	root.Content = append(root.Content, util.BuildStringNode("version", "v1", "")...)
	root.Content = append(root.Content, util.BuildStringNode("kind", "kubernetes", "")...)
	root.Content = append(root.Content, util.BuildStringNode("application", "<AppName>", "The name of the application to deploy.")...)

	// Target root
	targetNode, targetValuesNode := util.BuildMapNode("targets", "Map of your deployment target, "+
		"Armory CD-as-a-Service supports deploying to one target cluster.")
	devNode, devValuesNode := util.BuildMapNode("<deploymentName>",
		"Name for your deployment. Use a descriptive value such as the environment name.")
	devValuesNode.Content = append(devValuesNode.Content, util.BuildStringNode("account",
		"<accountName>", "The name that you assigned to the deployment target cluster when you installed the RNA.")...)
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
	pauseValuesNode.Content = append(pauseValuesNode.Content, util.BuildBoolNode("untilApproved", "true",
		"If set to true, the deployment waits until a manual approval to continue. Only set this to true if duration and unit are not set.")...)
	pause.Content = append(pause.Content, pauseNode, pauseValuesNode)
	afterNode, afterValuesNode := util.BuildSequenceNode("afterDeployment", "A set of steps that are executed in parallel, after the deployment is run")

	for _, feature := range options.features {
		switch feature {
		case "automated":
			// automated adds an analysis definition at the root level
			analysisNode, analysisValuesNode := util.BuildMapNode("analysis", "Define queries and thresholds used for automated analysis.")
			defaultMetricProviderNameNode := util.BuildStringNode("defaultMetricProviderName", "prometheus-prod", "Optional. Name of the default provider to use for the queries. Add providers in the Configuration UI.")
			queriesNode, queriesValuesNode := buildAnalysisQueries()
			analysisValuesNode.Content = append(analysisValuesNode.Content, defaultMetricProviderNameNode...)
			analysisValuesNode.Content = append(analysisValuesNode.Content, queriesNode, queriesValuesNode)
			root.Content = append(root.Content, analysisNode, analysisValuesNode)

			analysisStepNode := buildAutomatedAnalysisStep()
			analysis := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
			analysis.Content = append(analysis.Content, analysisStepNode...)
			afterValuesNode.Content = append(afterValuesNode.Content, analysis)
		case "webhook":
			hook := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
			hookNode, hookValuesNode := util.BuildMapNode("runWebhook", "The map key is the step type")
			hookValuesNode.Content = append(hookValuesNode.Content, util.BuildStringNode("name", "run integration test", "The name of a defined webhook")...)
			hookValuesNode.Content = append(hookValuesNode.Content, util.BuildStringNode("unit", "SECONDS", "")...)
			context := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
			contextNode, contextValuesNode := util.BuildMapNode("context", "A context of configured values for use as variable replacement")
			context.Content = append(context.Content, contextNode, contextValuesNode)
			hook.Content = append(hook.Content, hookNode, hookValuesNode)
			afterValuesNode.Content = append(afterValuesNode.Content, hook)

			webhooksNode, webhooksValuesNode := util.BuildSequenceNode("webhooks", "Define webhooks to be executed before, after or during deployment.")
			webhooksValuesNode.Content = append(webhooksValuesNode.Content,
				buildWebhookDefinitionNode("run integration test", "POST", "http://example.com/myurl/{{armory.deploymentId}}", "direct", "agent-rna", "2", "{ \"callbackUri\": \"{{armory.callbackUri}}\" }"))
			root.Content = append(root.Content, webhooksNode, webhooksValuesNode)
		default:
			return nil, fmt.Errorf("unknown feature specified for template: %s", feature)
		}
	}

	beforeValuesNode.Content = append(beforeValuesNode.Content, pause)
	constraintValuesNode.Content = append(constraintValuesNode.Content, beforeNode, beforeValuesNode, afterNode, afterValuesNode, dependsOnNode, dependsOnValuesNode)
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

	return root, nil
}
