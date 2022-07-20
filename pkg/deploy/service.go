package deploy

import (
	"errors"
	"fmt"
	de "github.com/armory-io/deploy-engine/api"
	cyclopsutils "github.com/armory-io/deploy-engine/cyclops/utils"
	"github.com/armory/armory-cli/pkg/model"
	"github.com/armory/armory-cli/pkg/util"
	"io/fs"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"os"
	"path/filepath"
	"strings"
)

var ErrorNoStrategyDeployment = errors.New("invalid deployment: strategy required for Deployment kind manifests")
var ErrorBadObject = errors.New("invalid deployment: manifest is not valid Kubernetes object")

func CreateDeploymentRequest(application string, config *model.OrchestrationConfig, contextOverrides map[string]string) (*de.StartPipelineRequest, error) {
	environments := make([]de.PipelineEnvironment, 0, len(*config.Targets))
	deployments := make([]de.PipelineDeployment, 0, len(*config.Targets))
	var analysis de.AnalysisConfig
	var webhooks []*de.WebhookRunConfig
	if config.Analysis != nil {
		analysis.DefaultAccount = config.Analysis.DefaultMetricProviderName
		queries, err := CreateAnalysisQueries(*config.Analysis, config.Analysis.DefaultMetricProviderName)
		if err != nil {
			return nil, err
		}
		analysis.Queries = queries
	}
	if config.Webhooks != nil {
		webhooksToAdd, err := buildWebhooks(*config.Webhooks)
		if err != nil {
			return nil, err
		}
		webhooks = webhooksToAdd
	}
	for key, element := range *config.Targets {

		envName := key
		target := element
		environments = append(environments, de.PipelineEnvironment{
			Name:      envName,
			Namespace: target.Namespace,
			Account:   target.Account,
		})

		deploymentToAdd := de.PipelineDeployment{
			Environment: envName,
		}

		if config.Strategies != nil {
			strategy, err := buildStrategy(*config, element.Strategy, key, contextOverrides)
			if err != nil {
				return nil, err
			}
			deploymentToAdd.Strategy = *strategy
		}

		files, err := GetManifestsFromFile(config.Manifests, envName)
		if err != nil {
			return nil, err
		}

		manifests, err := CreateDeploymentManifests(files, config.Strategies)
		if err != nil {
			return nil, err
		}
		deploymentToAdd.Manifests = manifests

		pipelineConstraint := de.ConstraintConfiguration{}
		if target.Constraints != nil {
			if target.Constraints.DependsOn != nil {
				pipelineConstraint.DependsOn = append(pipelineConstraint.DependsOn, *target.Constraints.DependsOn...)
			} else {
				pipelineConstraint.DependsOn = []string{}
			}

			beforeDeployment, err := CreateBeforeDeploymentConstraints(target.Constraints.BeforeDeployment, contextOverrides)
			if err != nil {
				return nil, err
			}
			pipelineConstraint.BeforeDeployment = beforeDeployment

			afterDeployment, err := CreateAfterDeploymentConstraints(target.Constraints.AfterDeployment, contextOverrides, config.Analysis)
			if err != nil {
				return nil, err
			}
			pipelineConstraint.AfterDeployment = afterDeployment
		}
		deploymentToAdd.Constraints = &pipelineConstraint

		if config.Analysis != nil {
			deploymentToAdd.Analysis = &analysis
		}
		if config.Webhooks != nil {
			deploymentToAdd.Webhooks = webhooks
		}
		deployments = append(deployments, deploymentToAdd)
	}
	req := de.StartPipelineRequest{
		Application:  application,
		Environments: environments,
		Deployments:  deployments,
	}
	if config.DeploymentConfig != nil && config.DeploymentConfig.Timeout != nil {
		req.DeploymentConfig = &de.DeploymentConfig{
			Timeout: &de.Timeout{
				Duration: config.DeploymentConfig.Timeout.Duration,
				Unit:     de.TimeUnit(strings.ToUpper(config.DeploymentConfig.Timeout.Unit)),
			},
		}
	}
	return &req, nil
}

func createDeploymentCanarySteps(strategy model.Strategy, analysisConfig *model.AnalysisConfig, context map[string]string) ([]*de.DeploymentStep, error) {
	var steps []*de.DeploymentStep
	for _, step := range *strategy.Canary.Steps {
		if step.SetWeight != nil {
			steps = append(
				steps,
				&de.DeploymentStep{
					SetWeight: &de.CanarySetWeightStepRequest{
						Weight: step.SetWeight.Weight,
					},
				})
		}

		if step.Pause != nil {
			pause, err := createPauseStep(step.Pause)
			if err != nil {
				return nil, err
			}
			steps = append(
				steps,
				&de.DeploymentStep{
					Pause: pause,
				})
		}

		if step.Analysis != nil {
			analysis, err := createDeploymentCanaryAnalysisStep(step.Analysis, analysisConfig, context)
			if err != nil {
				return nil, err
			}

			steps = append(
				steps,
				&de.DeploymentStep{
					Analysis: analysis,
				})
		}
		if step.RunWebhook != nil {
			steps = append(
				steps,
				&de.DeploymentStep{
					WebhookRun: &de.WebhookRunStepRequest{
						Name:    step.RunWebhook.Name,
						Context: util.MergeMaps(step.RunWebhook.Context, context),
					},
				})
		}
	}
	return steps, nil
}

func CreateAnalysisQueries(analysis model.AnalysisConfig, defaultMetricProviderName string) ([]*de.AnalysisQuery, error) {
	if analysis.Queries == nil {
		// we will only return a validation error if there is an analysis step being used in a canary or blue-green strategy
		return nil, nil
	}
	queries := analysis.Queries
	var analysisQueries []*de.AnalysisQuery
	for _, query := range queries {
		if query.MetricProviderName == nil {
			if defaultMetricProviderName == "" {
				return nil, fmt.Errorf("metric provider must be provided either in the analysis config, as defaultMetricProviderName, or in the query as metricProviderName")
			}
			query.MetricProviderName = &defaultMetricProviderName
		}
		analysisQueries = append(analysisQueries, &de.AnalysisQuery{
			Name:               query.Name,
			QueryTemplate:      query.QueryTemplate,
			UpperLimit:         query.UpperLimit,
			LowerLimit:         query.LowerLimit,
			MetricProviderName: *query.MetricProviderName,
		})
	}
	return analysisQueries, nil
}

func GetManifestsFromFile(manifests *[]model.ManifestPath, env string) (*[]string, error) {
	var allFileNames []string
	var files []string
	for _, manifestPath := range *manifests {
		if manifestPath.Targets != nil && len(manifestPath.Targets) == 0 {
			return nil, fmt.Errorf("please omit targets to include the manifests for all targets or specify the targets")
		}
		if util.Contains(manifestPath.Targets, env) || manifestPath.Targets == nil {
			if manifestPath.Inline != "" {
				files = append(files, manifestPath.Inline)
			}
			fileNames, err := getFileNamesFromManifestPath(manifestPath)
			if err != nil {
				return nil, err
			}
			allFileNames = append(allFileNames, fileNames...)
		}
	}

	dirFiles, err := funcName(allFileNames)
	if err != nil {
		return nil, err
	}
	files = append(files, dirFiles...)

	return &files, nil
}

func getFileNamesFromManifestPath(manifestPath model.ManifestPath) ([]string, error) {
	var allFileNames []string
	gitWorkspace, present := os.LookupEnv("GITHUB_WORKSPACE")
	_, isATest := os.LookupEnv("ARMORY_CLI_TEST")

	if manifestPath.Path != "" {
		if present && !isATest {
			manifestPath.Path = gitWorkspace + "/" + manifestPath.Path
		}
		err, fileNames := getFileNames(manifestPath)
		if err != nil {
			return nil, fmt.Errorf("unable to read manifest(s) from file: %s", err)
		}
		allFileNames = append(allFileNames, fileNames...)
	}
	return allFileNames, nil
}

func funcName(dirFileNames []string) ([]string, error) {
	var files []string
	for _, fileName := range dirFileNames {
		file, err := ioutil.ReadFile(fileName)
		if err != nil {
			return nil, fmt.Errorf("error trying to read manifest file '%s': %s", fileName, err)
		}
		files = append(files, string(file))
	}
	return files, nil
}

func getFileNames(manifestPath model.ManifestPath) (error, []string) {
	var fileNames []string
	err := filepath.WalkDir(manifestPath.Path, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml" {
			fileNames = append(fileNames, path)
		}

		return nil
	})
	return err, fileNames
}

func CreateDeploymentManifests(manifests *[]string, strategy *map[string]model.Strategy) ([]de.Manifest, error) {
	deManifests := make([]de.Manifest, 0, len(*manifests))
	for _, manifest := range *manifests {
		var un unstructured.Unstructured
		if strategy == nil {
			if err := cyclopsutils.DeserializeKubernetes([]byte(manifest), &un); err != nil {
				return nil, ErrorBadObject
			}
			if un.GetKind() == "Deployment" {
				return nil, ErrorNoStrategyDeployment
			}
		}
		deManifests = append(
			deManifests,
			de.Manifest{
				Inline: de.InlineManifest{
					Value: manifest,
				},
			})
	}
	return deManifests, nil
}

func CreateBeforeDeploymentConstraints(beforeDeployment *[]model.BeforeDeployment, contextOverrides map[string]string) ([]de.Constraint, error) {
	if beforeDeployment == nil {
		return []de.Constraint{}, nil
	}
	pipelineConstraints := make([]de.Constraint, 0, len(*beforeDeployment))
	var constraint de.Constraint
	for _, obj := range *beforeDeployment {
		if obj.Pause != nil {
			pause, err := createPauseStep(obj.Pause)
			if err != nil {
				return nil, err
			}
			constraint = de.Constraint{
				Pause: pause,
			}
		} else if obj.RunWebhook != nil {
			webhook, err := createWebhookConstraint(obj.RunWebhook, contextOverrides)
			if err != nil {
				return nil, err
			}
			constraint = de.Constraint{
				Webhook: webhook,
			}
		}

		pipelineConstraints = append(pipelineConstraints, constraint)
	}
	return pipelineConstraints, nil
}

func CreateAfterDeploymentConstraints(afterDeployment *[]model.AfterDeployment, contextOverrides map[string]string, analysisConfig *model.AnalysisConfig) ([]de.Constraint, error) {
	if afterDeployment == nil {
		return []de.Constraint{}, nil
	}
	pipelineConstraints := make([]de.Constraint, 0, len(*afterDeployment))
	var constraint de.Constraint
	for _, obj := range *afterDeployment {
		if obj.Pause != nil {
			pause, err := createPauseStep(obj.Pause)
			if err != nil {
				return nil, err
			}
			constraint = de.Constraint{
				Pause: pause,
			}
		} else if obj.RunWebhook != nil {
			webhook, err := createWebhookConstraint(obj.RunWebhook, contextOverrides)
			if err != nil {
				return nil, err
			}
			constraint = de.Constraint{
				Webhook: webhook,
			}
		} else if obj.Analysis != nil {
			analysis, err := createDeploymentCanaryAnalysisStep(obj.Analysis, analysisConfig, contextOverrides)
			if err != nil {
				return nil, err
			}
			constraint = de.Constraint{
				Analysis: analysis,
			}
		}

		pipelineConstraints = append(pipelineConstraints, constraint)
	}
	return pipelineConstraints, nil
}

func buildStrategy(modelStrategy model.OrchestrationConfig, strategyName string, target string, context map[string]string) (*de.PipelineStrategy, error) {
	configStrategies := *modelStrategy.Strategies
	strategy := configStrategies[strategyName]

	tm, err := createTrafficManagement(&modelStrategy, target)
	if err != nil {
		return nil, fmt.Errorf("invalid traffic management config: %s", err)
	}

	if strategy.Canary != nil {
		steps, err := createDeploymentCanarySteps(strategy, modelStrategy.Analysis, context)
		if err != nil {
			return nil, err
		}
		canary := de.CanaryStrategy{
			Steps: steps,
		}
		if tm != nil && tm.SMI != nil {
			canary.TrafficManagement = &de.TrafficManagementRequest{
				SMI: tm.SMI,
			}
		}
		return &de.PipelineStrategy{
			Canary: &canary,
		}, nil
	} else if strategy.BlueGreen != nil {
		ps := &de.PipelineStrategy{
			BlueGreen: &de.BlueGreenStrategy{},
		}

		if strategy.BlueGreen.ActiveService != "" {
			ps.BlueGreen.ActiveService = strategy.BlueGreen.ActiveService

		}

		if strategy.BlueGreen.PreviewService != "" {
			ps.BlueGreen.PreviewService = strategy.BlueGreen.PreviewService
		}

		if tm != nil && tm.Kubernetes != nil {
			ps.BlueGreen.TrafficManagement = &de.TrafficManagementRequest{
				Kubernetes: tm.Kubernetes,
			}
		}

		if strategy.BlueGreen.RedirectTrafficAfter != nil {
			redirectTrafficAfter, err := createBlueGreenRedirectConditions(strategy.BlueGreen.RedirectTrafficAfter, modelStrategy.Analysis)
			if err != nil {
				return nil, err
			}
			ps.BlueGreen.RedirectTrafficAfter = redirectTrafficAfter
		}
		if strategy.BlueGreen.ShutDownOldVersionAfter != nil {
			shutDownOldVersionAfter, err := createBlueGreenShutdownConditions(strategy.BlueGreen.ShutDownOldVersionAfter, modelStrategy.Analysis)
			if err != nil {
				return nil, err
			}
			ps.BlueGreen.ShutDownOldVersionAfter = shutDownOldVersionAfter
		}
		return ps, nil
	}

	return nil, fmt.Errorf("%s is not a valid strategy; define canary or blueGreen strategy", strategyName)
}

func createTrafficManagement(mo *model.OrchestrationConfig, currentTarget string) (*de.TrafficManagementRequest, error) {
	if mo.TrafficManagement == nil {
		return nil, nil
	}
	var tms de.TrafficManagementRequest
	for _, tm := range *mo.TrafficManagement {
		if len(tm.SMI) > 0 {
			smis, err := createSMIs(tm)
			if err != nil {
				return nil, err
			}
			// missing targets means smi config will be applied to all targets
			if len(tm.Targets) == 0 {
				tms.SMI = smis
				break
			}
			// otherwise we apply the smi config to user-defined targets
			for _, t := range tm.Targets {
				if t == currentTarget {
					tms.SMI = smis
					break
				}
			}
		}
		if len(tm.Kubernetes) > 0 {
			kubernetesTraffic, err := createKubernetesTraffic(tm)
			if err != nil {
				return nil, err
			}
			// missing targets means kubernetes config will be applied to all targets
			if len(tm.Targets) == 0 {
				tms.Kubernetes = kubernetesTraffic
				break
			}
			for _, t := range tm.Targets {
				if t == currentTarget {
					tms.Kubernetes = kubernetesTraffic
					break
				}
			}
		}
	}
	if tms.SMI != nil || tms.Kubernetes != nil {
		return &tms, nil
	}
	return nil, nil
}

func createDeploymentCanaryAnalysisStep(analysis *model.AnalysisStep, analysisConfig *model.AnalysisConfig, context map[string]string) (*de.AnalysisStepRequest, error) {
	if analysisConfig == nil {
		return nil, errors.New("analysis step is present but a top-level analysis config is not defined")
	}

	if analysisConfig.Queries == nil {
		return nil, errors.New("top-level analysis config is present but no queries are defined")
	}

	for _, query := range analysis.Queries {
		queryConfig := findByName(analysisConfig.Queries, query)
		if queryConfig == nil {
			return nil, fmt.Errorf("query in step does not exist in top-level analysis config: %q", query)
		}
	}

	var rollBackMode de.RollMode
	var rollForwardMode de.RollMode
	var units de.TimeUnit
	var lookbackMethod de.LookbackMethod

	if analysis.RollBackMode != "" {
		rollBackMode = de.RollMode(strings.ToUpper(analysis.RollBackMode))
	} else {
		rollBackMode = "AUTOMATIC"
	}

	if analysis.RollForwardMode != "" {
		rollForwardMode = de.RollMode(strings.ToUpper(analysis.RollForwardMode))
	} else {
		rollForwardMode = "AUTOMATIC"
	}

	if analysis.Units != "" {
		units = de.TimeUnit(strings.ToUpper(analysis.Units))
	} else {
		units = de.TimeUnitNone
	}

	if analysis.LookbackMethod != "" {
		lookbackMethod = de.LookbackMethod(strings.ToUpper(analysis.LookbackMethod))
	} else {
		lookbackMethod = de.LookbackMethodUnset
	}

	return &de.AnalysisStepRequest{
		Context:               util.MergeMaps(analysis.Context, context),
		RollBackMode:          rollBackMode,
		RollForwardMode:       rollForwardMode,
		Interval:              analysis.Interval,
		Units:                 units,
		NumberOfJudgmentRuns:  analysis.NumberOfJudgmentRuns,
		AbortOnFailedJudgment: analysis.AbortOnFailedJudgment,
		LookbackMethod:        lookbackMethod,
		Queries:               analysis.Queries,
	}, nil
}

func createBlueGreenRedirectConditions(conditions []*model.BlueGreenCondition, analysisConfig *model.AnalysisConfig) ([]*de.DeploymentStep, error) {
	var redirectConditions []*de.DeploymentStep
	for _, condition := range conditions {
		if condition.Pause != nil {
			pause, err := createPauseStep(condition.Pause)
			if err != nil {
				return nil, err
			}
			redirectConditions = append(
				redirectConditions,
				&de.DeploymentStep{
					Pause: pause,
				})
		}
		if condition.Analysis != nil {
			analysis, err := createDeploymentCanaryAnalysisStep(condition.Analysis, analysisConfig, map[string]string{})
			if err != nil {
				return nil, err
			}

			redirectConditions = append(
				redirectConditions,
				&de.DeploymentStep{
					Analysis: analysis,
				})
		}
		if condition.RunWebhook != nil {
			redirectConditions = append(redirectConditions, &de.DeploymentStep{
				WebhookRun: &de.WebhookRunStepRequest{
					Name:    condition.RunWebhook.Name,
					Context: condition.RunWebhook.Context,
				},
			})
		}
	}
	return redirectConditions, nil
}

func createBlueGreenShutdownConditions(conditions []*model.BlueGreenCondition, analysisConfig *model.AnalysisConfig) ([]*de.DeploymentStep, error) {
	var shutDownConditions []*de.DeploymentStep
	for _, condition := range conditions {
		if condition.Pause != nil {
			pause, err := createPauseStep(condition.Pause)
			if err != nil {
				return nil, err
			}
			shutDownConditions = append(
				shutDownConditions,
				&de.DeploymentStep{
					Pause: pause,
				})
		}
		if condition.Analysis != nil {
			analysis, err := createDeploymentCanaryAnalysisStep(condition.Analysis, analysisConfig, map[string]string{})
			if err != nil {
				return nil, err
			}

			shutDownConditions = append(
				shutDownConditions,
				&de.DeploymentStep{
					Analysis: analysis,
				})
		}
		if condition.RunWebhook != nil {
			shutDownConditions = append(shutDownConditions, &de.DeploymentStep{
				WebhookRun: &de.WebhookRunStepRequest{
					Name:    condition.RunWebhook.Name,
					Context: condition.RunWebhook.Context,
				},
			})
		}
	}
	return shutDownConditions, nil
}

func createPauseStep(pause *model.PauseStep) (*de.PauseStepRequest, error) {
	if err := validatePauseStep(pause); err != nil {
		return nil, err
	}
	unit := createTimeUnit(pause)

	return &de.PauseStepRequest{
		Duration:      pause.Duration,
		Unit:          unit,
		UntilApproved: pause.UntilApproved,
	}, nil
}

func createWebhookConstraint(webhook *model.WebhookStep, contextOverrides map[string]string) (*de.WebhookRunStepRequest, error) {
	if err := validateWebhookStep(webhook); err != nil {
		return nil, err
	}
	return &de.WebhookRunStepRequest{
		Name:    webhook.Name,
		Context: util.MergeMaps(webhook.Context, contextOverrides),
	}, nil
}

func createTimeUnit(pause *model.PauseStep) de.TimeUnit {
	if pause.Unit == "" {
		return de.TimeUnitNone
	} else {
		return de.TimeUnit(strings.ToUpper(pause.Unit))
	}
}

func validatePauseStep(pause *model.PauseStep) error {
	if pause.UntilApproved {
		if pause.Duration > 0 || pause.Unit != "" {
			return errors.New("pause is not valid: untilApproved cannot be set with both a unit and duration")
		}
	} else if pause.Duration > 0 && pause.Unit == "" {
		return errors.New("pause is not valid: duration must be set with a unit")
	} else if pause.Duration < 1 && pause.Unit != "" {
		return errors.New("pause is not valid: unit must be set with a duration")
	}
	return nil
}

func validateWebhookStep(webhook *model.WebhookStep) error {
	if webhook.Name == "" {
		return errors.New("webhook constraint is not valid: you must provide a name for a configured webhook")
	}
	return nil
}

func findByName(queries []model.Query, name string) *model.Query {
	for _, configQuery := range queries {
		if name == configQuery.Name {
			return &configQuery
		}
	}
	return nil
}

func buildWebhooks(webhooks []model.WebhookConfig) ([]*de.WebhookRunConfig, error) {
	var webhooksList []*de.WebhookRunConfig
	for _, webhook := range webhooks {
		var body string
		if webhook.BodyTemplate != nil {
			var err error
			body, err = buildBody(webhook.BodyTemplate)
			if err != nil {
				return nil, err
			}
		}
		webhooksList = append(webhooksList, &de.WebhookRunConfig{
			Name:            webhook.Name,
			Method:          webhook.Method,
			URITemplate:     webhook.UriTemplate,
			NetworkMode:     webhook.NetworkMode,
			AgentIdentifier: webhook.AgentIdentifier,
			RetryCount:      webhook.RetryCount,
			Headers:         buildHeaders(webhook.Headers),
			BodyTemplate:    body,
		})
	}
	return webhooksList, nil
}

func buildHeaders(headers *[]model.Header) []de.WebhookHeader {
	if headers == nil {
		return nil
	}

	var headersList []de.WebhookHeader
	for _, header := range *headers {
		headersList = append(headersList, de.WebhookHeader{
			Key:   header.Key,
			Value: header.Value,
		})
	}
	return headersList
}

func buildBody(bodyTemplate *model.Body) (string, error) {
	if bodyTemplate.Path != nil {
		content, err := ioutil.ReadFile(*bodyTemplate.Path)
		if err != nil {
			return "", errors.New("unable to read body template file")
		}
		return string(content), nil
	}
	return *bodyTemplate.Inline, nil
}

func createSMIs(tm model.TrafficManagement) ([]*de.SMITrafficManagementConfig, error) {
	var smis []*de.SMITrafficManagementConfig
	for _, s := range tm.SMI {
		if s.RootServiceName == "" {
			return nil, errors.New("rootServiceName required in smi")
		}
		smis = append(smis, &de.SMITrafficManagementConfig{
			RootServiceName:   s.RootServiceName,
			CanaryServiceName: s.CanaryServiceName,
			TrafficSplitName:  s.TrafficSplitName,
		})
	}
	return smis, nil
}

func createKubernetesTraffic(tm model.TrafficManagement) ([]*de.KubernetesTrafficManagementConfig, error) {
	var kubernetesTraffic []*de.KubernetesTrafficManagementConfig
	for _, kc := range tm.Kubernetes {
		trafficConfig := kc
		kubernetesTraffic = append(kubernetesTraffic, &de.KubernetesTrafficManagementConfig{
			ActiveService:  trafficConfig.ActiveService,
			PreviewService: trafficConfig.PreviewService,
		})
	}
	return kubernetesTraffic, nil
}
