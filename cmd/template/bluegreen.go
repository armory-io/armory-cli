package template

import (
	"fmt"
	"github.com/armory/armory-cli/pkg/util"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const (
	templateBlueGreenShort   = "Generate a blue-green deployment template"
	templateBlueGreenLong    = "Generate a blue-green deployment template in YAML format"
	templateBlueGreenExample = "armory template blue-green > blue-green.yaml"
)

type templateBlueGreenOptions struct {
	*templateOptions
}

func NewTemplateBlueGreenCmd(templateOptions *templateOptions) *cobra.Command {
	options := &templateBlueGreenOptions{
		templateOptions: templateOptions,
	}
	cmd := &cobra.Command{
		Use:     "blue-green",
		Aliases: []string{"blue-green"},
		Short:   templateBlueGreenShort,
		Long:    templateBlueGreenLong,
		Example: templateBlueGreenExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return blueGreen(cmd, options, args)
		},
	}
	return cmd
}

func blueGreen(cmd *cobra.Command, options *templateBlueGreenOptions, args []string) error {
	root := buildTemplateKubernetesCore()

	// Strategies root
	strategiesNode, strategyValuesNode := util.BuildMapNode("strategies", "A map of named strategies that can be assigned to deployment targets in the targets block.")
	// Strategy1
	strategy1Node, strategy1ValuesNode := util.BuildMapNode("strategy1",
		"Name for a strategy that you use to refer to it. Used in the target block. This example uses strategy1 as the name.")

	// BlueGreen root
	blueGreenNode, blueGreenValuesNode := util.BuildMapNode("blue-green", "The deployment strategy type. Use blueGreen.")

	// redirectTrafficAfter Root
	redirectTrafficAfterNode, redirectTrafficAfterValuesNode := util.BuildSequenceNode("redirectTrafficAfter", "The steps that must be completed before traffic is redirected to the new version.")

	// shutdownOldVersionAfter Root
	shutdownOldVersionAfterNode, shutdownOldVersionAfterValuesNode := util.BuildSequenceNode("shutdownOldVersionAfter", "The steps that must be completed before the old version is scaled down.")

	// Pause root
	pause := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	pauseNode, pauseValuesNode := util.BuildMapNode("pause", "A pause step type. The pipeline stops until the pause behavior is completed.")
	pauseValuesNode.Content = append(pauseValuesNode.Content, util.BuildIntNode("duration", "1", "The pause behavior is time (integer) before the deployment continues. If duration is set for this step, omit untilApproved.")...)
	pauseValuesNode.Content = append(pauseValuesNode.Content, util.BuildStringNode("unit", "seconds", "The unit of time to use for the pause. Can be seconds, minutes, or hours. Required if duration is set.")...)
	pause.Content = append(pause.Content, pauseNode, pauseValuesNode)

	// Pause UntilApproved root
	pauseUA := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	pauseUANode, pauseUAValuesNode := util.BuildMapNode("pause", "A pause step type. The pipeline stops until the pause behavior is completed.")
	pauseUAValuesNode.Content = append(pauseUAValuesNode.Content, util.BuildBoolNode("untilApproved", "true",
		"The pause behavior is the deployment waits until a manual approval is given to continue. Only set this to true if there is no duration pause behavior for this step.")...)
	pauseUA.Content = append(pauseUA.Content, pauseUANode, pauseUAValuesNode)

	redirectTrafficAfterValuesNode.Content = append(redirectTrafficAfterValuesNode.Content, pause, pauseUA)
	shutdownOldVersionAfterNode.Content = append(redirectTrafficAfterValuesNode.Content, pause, pauseUA)

	redirectTrafficAfterNode.Content = append(redirectTrafficAfterNode.Content, pause, pauseUA)
	shutdownOldVersionAfterNode.Content = append(shutdownOldVersionAfterNode.Content, pause, pauseUA)

	blueGreenValuesNode.Content = append(blueGreenValuesNode.Content, redirectTrafficAfterNode, redirectTrafficAfterValuesNode)
	blueGreenValuesNode.Content = append(blueGreenValuesNode.Content, shutdownOldVersionAfterNode, shutdownOldVersionAfterValuesNode)

	activeServiceNode := util.BuildStringNode("activeService", "<ActiveService>", "The active service that will receive traffic (required)")
	previewServiceNode := util.BuildStringNode("previewService", "<PreviewService>", "The preview service that will not receive traffic until the new version is deployed (optional)")
	activeRootUrlNode := util.BuildStringNode("activeRootUrl", "<ActiveRootUrl>", "The old version is available on the activeRootUrl before the traffic is swapped. After the redirectTrafficAfter steps activeRootUrl will point to the new version. (optional)")
	previewRootUrlNode := util.BuildStringNode("previewRootUrl", "<PreviewRootUrl>", "The new version is available on the previewRootUrl before traffic is redirected to it. (optional)")

	blueGreenValuesNode.Content = append(blueGreenValuesNode.Content, activeServiceNode...)
	blueGreenValuesNode.Content = append(blueGreenValuesNode.Content, previewServiceNode...)
	blueGreenValuesNode.Content = append(blueGreenValuesNode.Content, activeRootUrlNode...)
	blueGreenValuesNode.Content = append(blueGreenValuesNode.Content, previewRootUrlNode...)

	strategy1ValuesNode.Content = append(strategy1ValuesNode.Content, blueGreenNode, blueGreenValuesNode)
	strategyValuesNode.Content = append(strategyValuesNode.Content, strategy1Node, strategy1ValuesNode)

	root.Content = append(root.Content, strategiesNode, strategyValuesNode)

	bytes, err := yaml.Marshal(root)
	if err != nil {
		return fmt.Errorf("error trying to build blueGreen template: %s", err)
	}
	_, err = cmd.OutOrStdout().Write(bytes)
	if err != nil {
		return fmt.Errorf("error trying to parse blueGreen template: %s", err)
	}
	return nil
}
