package template

import (
	"fmt"
	"github.com/armory/armory-cli/pkg/util"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const (
	templateCanaryShort   = "Generate a canary deployment template"
	templateCanaryLong    = "Generate a canary deployment template in YAML format"
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
	strategiesNode, strategyValuesNode := util.BuildMapNode("strategies", "A map of named strategies that can be assigned to deployment targets in the targets block.")
	// Strategy1
	strategy1Node, strategy1ValuesNode := util.BuildMapNode("strategy1",
		"Name for a strategy that you use to refer to it. Used in the target block. This example uses strategy1 as the name.")

	// Canary root
	canaryNode, canaryValuesNode := util.BuildMapNode("canary", "The deployment strategy type. Use canary.")

	// Steps sequence/array
	stepsNode, stepsValuesNode := util.BuildSequenceNode("steps", "The steps for your deployment strategy.")

	// Pause root
	pause := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	pauseNode, pauseValuesNode := util.BuildMapNode("pause", "A pause step type. The pipeline stops until the pause behavior is completed. The pause behavior can be duration or untilApproved. ")
	pauseValuesNode.Content = append(pauseValuesNode.Content, util.BuildIntNode("duration", "1", "The pause behavior is time (integer) before the deployment continues. If duration is set for this step, omit untilApproved.")...)
	pauseValuesNode.Content = append(pauseValuesNode.Content, util.BuildStringNode("unit", "seconds", "The unit of time to use for the pause. Can be seconds, minutes, or hours. Required if duration is set.")...)
	pause.Content = append(pause.Content, pauseNode, pauseValuesNode)

	// Weight root
	weight := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	weightNode, weightValuesNode := util.BuildMapNode("setWeight", "")
	weightValuesNode.Content = append(weightValuesNode.Content, util.BuildIntNode("weight", "33", "The percentage of pods that should be running the canary version for this step. Set it to an integer between 0 and 100, inclusive.")...)
	weight.Content = append(weight.Content, weightNode, weightValuesNode)

	// Pause UntilApproved root
	pauseUA := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	pauseUANode, pauseUAValuesNode := util.BuildMapNode("pause", "A pause step type. The pipeline stops until the pause behavior is completed.")
	pauseUAValuesNode.Content = append(pauseUAValuesNode.Content, util.BuildBoolNode("untilApproved", "true",
		"The pause behavior is the deployment waits until a manual approval is given to continue. Only set this to true if there is no duration pause behavior for this step.")...)
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
