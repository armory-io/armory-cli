package template

import (
	"fmt"
	"github.com/armory/armory-cli/pkg/util"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const (
	templateCanaryShort   = "Print a canary deployment template"
	templateCanaryLong    = "Print a canary deployment template in yaml format"
	templateCanaryExample = "armory template canary > canary.yaml"
)

type templateCanaryOptions struct {
	*templateOptions
}

func NewTemplateCanaryCmd(templateOptions *templateOptions) *cobra.Command {
	options := &templateCanaryOptions{
		templateOptions: templateOptions,
	}
	cmd := &cobra.Command{
		Use:     "canary",
		Aliases: []string{"canary"},
		Short:   templateCanaryShort,
		Long:    templateCanaryLong,
		Example: templateCanaryExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return canary(cmd, options, args)
		},
	}
	return cmd
}

func canary(cmd *cobra.Command, options *templateCanaryOptions, args []string) error {
	root := buildTemplateKubernetesCore()

	// Strategies root
	strategiesNode, strategyValuesNode := util.BuildMapNode("strategies","A map of strategies, each of which can be assigned to deployment targets in the targets map.")
	// Strategy1
	strategy1Node, strategy1ValuesNode := util.BuildMapNode("strategy1",
		"Specify a strategy. The identifier for a strategy is its name.")

	// Canary root
	canaryNode, canaryValuesNode := util.BuildMapNode("canary","This map key, is the deployment strategy type.")

	// Steps sequence/array
	stepsNode, stepsValuesNode := util.BuildSequenceNode("steps", "A list of progressive canary steps.")

	// Pause root
	pause := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	pauseNode, pauseValuesNode := util.BuildMapNode("pause","The map key is the step type")
	pauseValuesNode.Content = append(pauseValuesNode.Content, util.BuildIntNode("duration", "1", "The duration of the pause. If duration is non-zero, untilApproved should be set to false.")...)
	pauseValuesNode.Content = append(pauseValuesNode.Content, util.BuildStringNode("unit", "SECONDS", "")...)
	pause.Content = append(pause.Content, pauseNode, pauseValuesNode)

	// Weight root
	weight := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	weightNode, weightValuesNode := util.BuildMapNode("setweight","")
	weightValuesNode.Content = append(weightValuesNode.Content, util.BuildIntNode("weight", "33", "The percent of pods that should be running the canary version. Weight should be between 0 and 100 inclusive.")...)
	weight.Content = append(weight.Content, weightNode, weightValuesNode)

	// Pause UntilApproved root
	pauseUA := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	pauseUANode, pauseUAValuesNode := util.BuildMapNode("pause","")
	pauseUAValuesNode.Content = append(pauseUAValuesNode.Content, util.BuildBoolNode("untilapproved", "true", "If set to true, the progressive canary will wait until a manual judgment to continue. This field should not be set to true unless duration and unit are unset.")...)
	pauseUA.Content = append(pauseUA.Content, pauseUANode, pauseUAValuesNode)

	stepsValuesNode.Content = append(stepsValuesNode.Content, pause, weight, pauseUA)
	canaryValuesNode.Content = append(canaryValuesNode.Content, stepsNode, stepsValuesNode)
	strategy1ValuesNode.Content = append(strategy1ValuesNode.Content, canaryNode, canaryValuesNode)
	strategyValuesNode.Content = append(strategyValuesNode.Content, strategy1Node, strategy1ValuesNode)

	root.Content = append(root.Content, strategiesNode, strategyValuesNode)

	bytes, err := yaml.Marshal(root)
	if err != nil {
		return fmt.Errorf("error trying to build canary template: %s", err)
	}
	_, err = cmd.OutOrStdout().Write(bytes)
	if err != nil {
		return fmt.Errorf("error trying to parse canary template: %s", err)
	}
	return nil
}