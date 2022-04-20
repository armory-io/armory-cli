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
	features []string
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
	cmd.Flags().StringArrayVarP(&options.features, "features", "f", []string {}, "features to include in the template. Available options [manual, automated, traffic]")
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

	// Weight nodes
	weight := buildWeightStepNode("33", "The percentage of pods that should be running the canary version for this step. Set it to an integer between 0 and 100, inclusive.")
	weight100 := buildWeightStepNode("100", "Setting weight to 100 is optional. Traffic automatically goes to 100 after passing the final step.")

	// Pause UntilApproved root
	pauseUA := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	pauseUANode, pauseUAValuesNode := util.BuildMapNode("pause", "A pause step type. The pipeline stops until the pause behavior is completed.")
	pauseUAValuesNode.Content = append(pauseUAValuesNode.Content, util.BuildBoolNode("untilApproved", "true",
		"The pause behavior is the deployment waits until a manual approval is given to continue. Only set this to true if there is no duration pause behavior for this step.")...)
	pauseUA.Content = append(pauseUA.Content, pauseUANode, pauseUAValuesNode)

	if len(options.features) == 0 {
		options.features = append(options.features, "manual")
	}
	for _, feature := range options.features {
		switch feature {
		case "manual":
			stepsValuesNode.Content = append(stepsValuesNode.Content, pause, weight, pauseUA)
		case "traffic":
			trafficNode, trafficValuesNode := util.BuildSequenceNode("trafficManagement", "")
			trafficItemNode := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}

			targetSliceNode, targetSliceValuesNode := util.BuildSequenceNodeWithTailComment("targets", "Specify a list of target names where the traffic management should occur. They should be defined in the top level targets block.")
			trafficItemNode.Content = append(trafficItemNode.Content, targetSliceNode, targetSliceValuesNode)

			smiNode, smiValuesNode := util.BuildSequenceNode("smi", "")
			smiItemNode := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}

			smiItemNode.Content = append(smiItemNode.Content, util.BuildStringNode("rootServiceName", "", "Name of the root service for deployment. The root service is required and must exist in your target environment at the time of deployment.")...)
			smiItemNode.Content = append(smiItemNode.Content, util.BuildStringNode("trafficSplitName", "", "Optional name of the service serving the new version. By default, \"<rootServiceName>-canary\".")...)
			smiItemNode.Content = append(smiItemNode.Content, util.BuildStringNode("canaryServiceName", "", "Optional name of the auto-generated trafficSplit custom resource. By default \"<rootServiceName>\".")...)
			smiValuesNode.Content = append(smiValuesNode.Content, smiItemNode)

			trafficItemNode.Content = append(trafficItemNode.Content, smiNode, smiValuesNode)
			trafficValuesNode.Content = append(trafficValuesNode.Content, trafficItemNode)
			root.Content = append(root.Content, trafficNode, trafficValuesNode)
		case "automated":
			// automated adds an analysis definition at the root level
			analysisNode, analysisValuesNode := util.BuildMapNode("analysis", "Define queries and thresholds used for automated analysis.")
			defaultMetricProviderNameNode := util.BuildStringNode("defaultMetricProviderName", "prometheus-prod", "Optional. Name of the default provider to use for the queries. Add providers in the Configuration UI.")
			queriesNode, queriesValuesNode := buildAnalysisQueries()
			analysisValuesNode.Content = append(analysisValuesNode.Content, defaultMetricProviderNameNode...)
			analysisValuesNode.Content = append(analysisValuesNode.Content, queriesNode, queriesValuesNode)
			root.Content = append(root.Content, analysisNode, analysisValuesNode)

			//automated uses an automated approval via canary analysis
			pauseUA.Content = buildAutomatedPauseStep()
			stepsValuesNode.Content = append(stepsValuesNode.Content, weight, pauseUA, weight100)

			webhooksNode, webhooksValuesNode := util.BuildSequenceNode("webhooks", "Define webhooks to be executed before, after or during deployment.")
			webhooksValuesNode.Content = append(webhooksValuesNode.Content,
				buildWebhookDefinitionNode("run integration test", "POST", "http://example.com/myurl/{{armory.deploymentId}}", "direct", "agent-rna", "2", "{ \"callbackUri\": \"{{armory.callbackUri}}\" }"))
			root.Content = append(root.Content, webhooksNode, webhooksValuesNode)
		default:
			return fmt.Errorf("unknown feature specified for template: %s", feature)
		}
	}

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
func buildAnalysisQueries() (*yaml.Node, *yaml.Node) {
	queriesNode, queriesValuesNode := util.BuildSequenceNode("queries", "Note that the example queries require Prometheus to have \"kube-state-metrics.metricAnnotationsAllowList[0]=pods=[*]\"\n" +
		"set and for your applications pods to have the annotation \"prometheus.io/scrape\": \"true\"")
	queryTemplate1 := "avg (avg_over_time(container_cpu_system_seconds_total{job=\"kubelet\"}[{{armory.promQlStepInterval}}]) * on (pod)  group_left (annotation_app)\n" +
		"sum(kube_pod_annotations{job=\"kube-state-metrics\",annotation_deploy_armory_io_replica_set_name=\"{{armory.replicaSetName}}\"})\n" +
		"by (annotation_app, pod)) by (annotation_app)"
	queryTemplate2 := "avg (avg_over_time(container_memory_working_set_bytes{job=\"kubelet\"}[{{armory.promQlStepInterval}}]) * on (pod)  group_left (annotation_app)\n" +
		"sum(kube_pod_annotations{job=\"kube-state-metrics\",annotation_deploy_armory_io_replica_set_name=\"{{armory.replicaSetName}}\"})\n" +
		"by (annotation_app, pod)) by (annotation_app)"
	queriesValuesNode.Content =  append(queriesValuesNode.Content,
		buildAnalysisQueryDefinitionNode("containerCPUSeconds", "my-prometheus-provider", "100", "0", queryTemplate1),
		buildAnalysisQueryDefinitionNode("avgMemoryUsage", "", "10", "0", queryTemplate2),
	)
	return queriesNode, queriesValuesNode
}

func buildAnalysisQueryDefinitionNode(name string, metricProviderName string, upperLimit string, lowerLimit string, queryTemplate string) *yaml.Node {
	query := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	query.Content = append(query.Content, util.BuildStringNode("name", name,
		""	)...)
	if len(metricProviderName) > 0 {
		query.Content = append(query.Content, util.BuildStringNode("metricProviderName", metricProviderName,
			"Optional. Override the defaultMetricProviderName specified in analysis.queries."	)...)
	}
	query.Content = append(query.Content, util.BuildIntNode("upperLimit", upperLimit,
		"Optional when 'lowerLimit' is specified. If the metric exceeds this value, the automated analysis fails, causing the step to fail."	)...)
	query.Content = append(query.Content, util.BuildIntNode("lowerLimit", lowerLimit,
		"Optional when 'upperLimit' is specified. If the metric goes below this value, the automated analysis fails, causing the step to fail."	)...)
	query.Content = append(query.Content, util.BuildStringNode("queryTemplate", queryTemplate,"")...)
	return query
}

func buildWebhookDefinitionNode(name string, method string, uriTemplate string, networkMode string, agentIdentifier string, retryCount string, bodyTemplate string) *yaml.Node {
	hook := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	headers, headersValuesNode := util.BuildSequenceNode("headers", "HTTP headers for the request. i.e. for Authorization or ContentType")
	auth := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	auth.Content = append(auth.Content, util.BuildStringNode("key", "Authorization", "")...)
	auth.Content = append(auth.Content, util.BuildStringNode("value", "Bearer {{secrets.oauthToken}}", "")...)
	headersValuesNode.Content = append(headersValuesNode.Content, auth)

	hook.Content = append(hook.Content, util.BuildStringNode("name", name,
		"Name of the webhook"	)...)

	hook.Content = append(hook.Content, util.BuildStringNode("method", method, "HTTP Method")...)
	hook.Content = append(hook.Content, headers, headersValuesNode)
	hook.Content = append(hook.Content, util.BuildStringNode("uriTemplate", uriTemplate, "the URI for the HTTP request")...)
	hook.Content = append(hook.Content, util.BuildStringNode("networkMode", networkMode, "Optional: direct or remoteNetworkAgent")...)
	hook.Content = append(hook.Content, util.BuildStringNode("agentIdentifier", agentIdentifier, "Optional: name of agent or account that has a remote network agent configured")...)
	hook.Content = append(hook.Content, util.BuildIntNode("retryCount", retryCount, "Optional: number of times to execute the webhook again if the trigger is unsuccessful")...)
	hook.Content = append(hook.Content, util.BuildStringNode("bodyTemplate", bodyTemplate, "Optional: a body for the HTTP request")...)
	return hook
}

func buildAutomatedPauseStep() []*yaml.Node {
	pauseUA := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	pauseUANode, pauseUAValuesNode := util.BuildMapNode("analysis", "An analysis step pauses the deployment until analysis judgement runs complete.")
	pauseUAValuesNode.Content = append(pauseUAValuesNode.Content, util.BuildIntNode("interval", "7",
		"How long each sample of the query gets summarized over"	)...)
	pauseUAValuesNode.Content = append(pauseUAValuesNode.Content, util.BuildStringNode("units", "seconds",
		"The unit for the interval: 'seconds', 'minutes' or 'hours'"	)...)
	pauseUAValuesNode.Content = append(pauseUAValuesNode.Content, util.BuildIntNode("numberOfJudgmentRuns", "1",
		"How many times the queries run.")...)
	queriesNode, queriesValuesNode := util.BuildSequenceNode("queries", "rollBackMode: manual # Optional. Defaults to 'automatic' if omitted. Uncomment to require a manual review before rolling back if automated analysis detects an issue.\n" +
		"rollForwardMode: manual # Optional. Defaults to 'automatic' if omitted. Uncomment to require a manual review before continuing deployment if automated analysis determines the app is healthy.")
	queriesValuesNode.Content = append(queriesValuesNode.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "containerCPUSeconds",
			HeadComment: "Specify a list of queries to run. Reference them by the name you assign in analysis.queries."},
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "avgMemoryUsage"})
	pauseUAValuesNode.Content = append(pauseUAValuesNode.Content, queriesNode, queriesValuesNode)
	pauseUA.Content = append(pauseUA.Content, pauseUANode, pauseUAValuesNode)

	return pauseUA.Content
}

func buildWeightStepNode(value string, comment string) *yaml.Node {
	weight := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	weightNode, weightValuesNode := util.BuildMapNode("setWeight", "")
	weightValuesNode.Content = append(weightValuesNode.Content, util.BuildIntNode("weight", value, comment)...)
	weight.Content = append(weight.Content, weightNode, weightValuesNode)
	return weight
}
