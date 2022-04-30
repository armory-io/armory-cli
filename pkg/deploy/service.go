package deploy

import (
	"errors"
	"fmt"
	de "github.com/armory-io/deploy-engine/pkg"
	"github.com/armory/armory-cli/pkg/model"
	"github.com/armory/armory-cli/pkg/util"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func CreateDeploymentRequest(application string, config *model.OrchestrationConfig) (*de.PipelineStartPipelineRequest, error) {
	environments := make([]de.PipelinePipelineEnvironment, 0, len(*config.Targets))
	deployments := make([]de.PipelinePipelineDeployment, 0, len(*config.Targets))
	var analysis de.AnalysisAnalysisConfig
	var webhooks *[]de.WebhooksWebhookRunConfig
	if config.Analysis != nil {
		analysis.DefaultAccount = &config.Analysis.DefaultMetricProviderName
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
		environments = append(environments, de.PipelinePipelineEnvironment{
			Name:      &envName,
			Namespace: &target.Namespace,
			Account:   &target.Account,
		})

		strategy, err := buildStrategy(*config, element.Strategy, key)
		if err != nil {
			return nil, err
		}

		files, err := GetManifestsFromFile(config.Manifests, envName)
		if err != nil {
			return nil, err
		}

		pipelineConstraint := de.PipelineConstraintConfiguration{}
		if target.Constraints != nil {
			beforeDeployment, err := CreateBeforeDeploymentConstraints(target.Constraints.BeforeDeployment)
			if err != nil {
				return nil, err
			}
			if target.Constraints.DependsOn != nil {
				pipelineConstraint.SetDependsOn(*target.Constraints.DependsOn)
			} else {
				pipelineConstraint.SetDependsOn([]string{})
			}
			pipelineConstraint.SetBeforeDeployment(beforeDeployment)

			afterDeployment, err := CreateAfterDeploymentConstraints(target.Constraints.AfterDeployment, config.Analysis)
			if err != nil {
				return nil, err
			}
			if target.Constraints.DependsOn != nil {
				pipelineConstraint.SetDependsOn(*target.Constraints.DependsOn)
			} else {
				pipelineConstraint.SetDependsOn([]string{})
			}
			pipelineConstraint.SetAfterDeployment(afterDeployment)
		}
		deploymentToAdd := de.PipelinePipelineDeployment{
			Environment: &envName,
			Manifests:   CreateDeploymentManifests(files),
			Strategy:    strategy,
			Constraints: &pipelineConstraint,
		}
		if config.Analysis != nil {
			deploymentToAdd.Analysis = &analysis
		}
		if config.Webhooks != nil {
			deploymentToAdd.Webhooks = webhooks
		}
		deployments = append(deployments, deploymentToAdd)
	}
	req := de.PipelineStartPipelineRequest{
		Application:  &application,
		Environments: &environments,
		Deployments:  &deployments,
	}
	return &req, nil
}

func createDeploymentCanarySteps(strategy model.Strategy, analysisConfig *model.AnalysisConfig) ([]de.KubernetesV2CanaryStep, error) {
	var steps []de.KubernetesV2CanaryStep
	for _, step := range *strategy.Canary.Steps {
		if step.SetWeight != nil {
			steps = append(
				steps,
				de.KubernetesV2CanaryStep{
					SetWeight: &de.KubernetesV2CanarySetWeightStep{
						Weight: &step.SetWeight.Weight,
					},
					Pause: nil,
				})
		}

		if step.Pause != nil {
			pause, err := createCanaryPause(step.Pause)
			if err != nil {
				return nil, err
			}
			steps = append(
				steps,
				de.KubernetesV2CanaryStep{
					SetWeight: nil,
					Pause:     pause,
				})
		}

		if step.Analysis != nil {
			analysis, err := createDeploymentCanaryAnalysisStep(step.Analysis, analysisConfig)
			if err != nil {
				return nil, err
			}

			steps = append(
				steps,
				de.KubernetesV2CanaryStep{
					Analysis: analysis,
				})
		}
		if step.RunWebhook != nil {
			steps = append(
				steps,
				de.KubernetesV2CanaryStep{
					WebhookRun: &de.WebhooksWebhookRunStepInput{
						Name:    step.RunWebhook.Name,
						Context: step.RunWebhook.Context,
					},
				})
		}
	}
	return steps, nil
}

func CreateAnalysisQueries(analysis model.AnalysisConfig, defaultMetricProviderName string) (*[]de.AnalysisAnalysisQueries, error) {
	if analysis.Queries == nil {
		// we will only return a validation error if there is an analysis step being used in a canary or blue-green strategy
		return nil, nil
	}
	queries := *analysis.Queries
	var analysisQueries []de.AnalysisAnalysisQueries
	for _, query := range queries {
		if query.MetricProviderName == nil {
			if defaultMetricProviderName == "" {
				return nil, fmt.Errorf("metric provider must be provided either in the analysis config, as defaultMetricProviderName, or in the query as metricProviderName")
			}
			query.MetricProviderName = &defaultMetricProviderName
		}
		analysisQueries = append(analysisQueries, de.AnalysisAnalysisQueries{
			Name:               query.Name,
			QueryTemplate:      query.QueryTemplate,
			UpperLimit:         query.UpperLimit,
			LowerLimit:         query.LowerLimit,
			MetricProviderName: query.MetricProviderName,
		})
	}
	return &analysisQueries, nil
}

func GetManifestsFromFile(manifests *[]model.ManifestPath, env string) (*[]string, error) {
	var fileNames []string
	var files []string
	gitWorkspace, present := os.LookupEnv("GITHUB_WORKSPACE")
	_, isATest := os.LookupEnv("ARMORY_CLI_TEST")
	for _, manifestPath := range *manifests {
		if manifestPath.Targets != nil && len(manifestPath.Targets) == 0 {
			return nil, fmt.Errorf("please omit targets to include the manifests for all targets or specify the targets")
		}

		if util.Contains(manifestPath.Targets, env) || manifestPath.Targets == nil {
			if manifestPath.Inline != "" {
				files = append(files, manifestPath.Inline)
			}
			if present && !isATest {
				manifestPath.Path = gitWorkspace + manifestPath.Path
			}
			if manifestPath.Path != "" {
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
				if err != nil {
					return nil, fmt.Errorf("unable to read manifest(s) from file: %s", err)
				}
			}
		}
	}

	for _, fileName := range fileNames {
		file, err := ioutil.ReadFile(fileName)
		if err != nil {
			return nil, fmt.Errorf("error trying to read manifest file '%s': %s", fileName, err)
		}
		files = append(files, string(file))
	}

	return &files, nil
}

func CreateDeploymentManifests(manifests *[]string) *[]de.KubernetesV2Manifest {
	deManifests := make([]de.KubernetesV2Manifest, 0, len(*manifests))
	for _, manifest := range *manifests {
		deManifests = append(
			deManifests,
			de.KubernetesV2Manifest{
				Inline: de.KubernetesV2InlineManifest{
					Value: manifest,
				},
			})
	}
	return &deManifests
}

func CreateBeforeDeploymentConstraints(beforeDeployment *[]model.BeforeDeployment) ([]de.PipelineConstraint, error) {
	if beforeDeployment == nil {
		return []de.PipelineConstraint{}, nil
	}
	pipelineConstraints := make([]de.PipelineConstraint, 0, len(*beforeDeployment))
	var constraint de.PipelineConstraint
	for _, obj := range *beforeDeployment {
		if obj.Pause != nil {
			pause, err := createPauseConstraint(obj.Pause)
			if err != nil {
				return nil, err
			}
			constraint = de.PipelineConstraint{
				Pause: pause,
			}
		} else if obj.RunWebhook != nil {
			webhook, err := createWebhookConstraint(obj.RunWebhook)
			if err != nil {
				return nil, err
			}
			constraint = de.PipelineConstraint{
				Webhook: webhook,
			}
		}

		pipelineConstraints = append(pipelineConstraints, constraint)
	}
	return pipelineConstraints, nil
}

func CreateAfterDeploymentConstraints(afterDeployment *[]model.AfterDeployment, analysisConfig *model.AnalysisConfig) ([]de.PipelineConstraint, error) {
	if afterDeployment == nil {
		return []de.PipelineConstraint{}, nil
	}
	pipelineConstraints := make([]de.PipelineConstraint, 0, len(*afterDeployment))
	var constraint de.PipelineConstraint
	for _, obj := range *afterDeployment {
		if obj.Pause != nil {
			pause, err := createPauseConstraint(obj.Pause)
			if err != nil {
				return nil, err
			}
			constraint = de.PipelineConstraint{
				Pause: pause,
			}
		} else if obj.RunWebhook != nil {
			webhook, err := createWebhookConstraint(obj.RunWebhook)
			if err != nil {
				return nil, err
			}
			constraint = de.PipelineConstraint{
				Webhook: webhook,
			}
		} else if obj.Analysis != nil {
			analysis, err := createDeploymentCanaryAnalysisStep(obj.Analysis, analysisConfig)
			if err != nil {
				return nil, err
			}
			constraint = de.PipelineConstraint{
				Analysis: analysis,
			}
		}

		pipelineConstraints = append(pipelineConstraints, constraint)
	}
	return pipelineConstraints, nil
}

func buildStrategy(modelStrategy model.OrchestrationConfig, strategyName string, target string) (*de.PipelinePipelineStrategy, error) {
	configStrategies := *modelStrategy.Strategies
	strategy := configStrategies[strategyName]
	if strategy.Canary != nil {
		steps, err := createDeploymentCanarySteps(strategy, modelStrategy.Analysis)
		if err != nil {
			return nil, err
		}
		tm, err := createTrafficManagement(&modelStrategy, target)
		if err != nil {
			return nil, fmt.Errorf("invalid traffic management config: %s", err)
		}
		return &de.PipelinePipelineStrategy{
			Canary: &de.KubernetesV2CanaryStrategy{
				Steps:             steps,
				TrafficManagement: tm,
			},
		}, nil
	} else if strategy.BlueGreen != nil {
		if strategy.BlueGreen.ActiveService == "" {
			return nil, errors.New("invalid blueGreen config: activeService is required")
		}

		ps := &de.PipelinePipelineStrategy{
			BlueGreen: &de.KubernetesV2BlueGreenStrategy{
				ActiveService:  strategy.BlueGreen.ActiveService,
				PreviewService: strategy.BlueGreen.PreviewService,
			},
		}
		if strategy.BlueGreen.RedirectTrafficAfter != nil {
			redirectTrafficAfter, err := createBlueGreenRedirectConditions(strategy.BlueGreen.RedirectTrafficAfter, modelStrategy.Analysis)
			if err != nil {
				return nil, err
			}
			ps.BlueGreen.RedirectTrafficAfter = &redirectTrafficAfter
		}
		if strategy.BlueGreen.ShutDownOldVersionAfter != nil {
			shutDownOldVersionAfter, err := createBlueGreenShutdownConditions(strategy.BlueGreen.ShutDownOldVersionAfter, modelStrategy.Analysis)
			if err != nil {
				return nil, err
			}
			ps.BlueGreen.ShutDownOldVersionAfter = &shutDownOldVersionAfter
		}
		return ps, nil
	}

	return nil, fmt.Errorf("%s is not a valid strategy; define canary or blueGreen strategy", strategyName)
}

func createTrafficManagement(mo *model.OrchestrationConfig, currentTarget string) (*de.KubernetesV2TrafficManagement, error) {
	if mo.TrafficManagement == nil {
		return nil, nil
	}
	var tms de.KubernetesV2TrafficManagement
	for _, tm := range *mo.TrafficManagement {
		if len(tm.SMI) > 0 {
			smis, err := createSMIs(tm)
			if err != nil {
				return nil, err
			}
			// missing targets means smi config will be applied to all targets
			if len(tm.Targets) == 0 {
				tms.Smi = smis
				break
			}
			// otherwise we apply the smi config to user-defined targets
			for _, t := range tm.Targets {
				if t == currentTarget {
					tms.Smi = smis
					break
				}
			}
		}
	}
	if tms.Smi != nil {
		return &tms, nil
	}
	return nil, nil
}

func createDeploymentCanaryAnalysisStep(analysis *model.AnalysisStep, analysisConfig *model.AnalysisConfig) (*de.AnalysisAnalysisStepInput, error) {
	if analysisConfig == nil {
		return nil, errors.New("analysis step is present but a top-level analysis config is not defined")
	}

	if analysisConfig.Queries == nil {
		return nil, errors.New("top-level analysis config is present but no queries are defined")
	}

	for _, query := range *analysis.Queries {
		queryConfig := findByName(*analysisConfig.Queries, query)
		if queryConfig == nil {
			return nil, fmt.Errorf("query in step does not exist in top-level analysis config: %q", query)
		}
	}

	var rollBackMode *de.AnalysisRollMode
	var rollForwardMode *de.AnalysisRollMode
	var units *de.TimeTimeUnit
	var lookbackMethod *de.AnalysisLookbackMethod
	var err error

	if analysis.RollBackMode != "" {
		rollBackMode, err = de.NewAnalysisRollModeFromValue(strings.ToUpper(analysis.RollBackMode))
	} else {
		rollBackMode, err = de.NewAnalysisRollModeFromValue("AUTOMATIC")
	}
	if err != nil {
		return nil, err
	}

	if analysis.RollForwardMode != "" {
		rollForwardMode, err = de.NewAnalysisRollModeFromValue(strings.ToUpper(analysis.RollForwardMode))
	} else {
		rollForwardMode, err = de.NewAnalysisRollModeFromValue("AUTOMATIC")
	}
	if err != nil {
		return nil, err
	}
	if analysis.Units != "" {
		units, err = de.NewTimeTimeUnitFromValue(strings.ToUpper(analysis.Units))
	} else {
		units, err = de.NewTimeTimeUnitFromValue("NONE")
	}
	if err != nil {
		return nil, err
	}
	if analysis.LookbackMethod != "" {
		lookbackMethod, err = de.NewAnalysisLookbackMethodFromValue(strings.ToUpper(analysis.LookbackMethod))
	} else {
		lookbackMethod, err = de.NewAnalysisLookbackMethodFromValue("UNSET")
	}
	if err != nil {
		return nil, err
	}

	return &de.AnalysisAnalysisStepInput{
		Context:               &analysis.Context,
		RollBackMode:          rollBackMode,
		RollForwardMode:       rollForwardMode,
		Interval:              &analysis.Interval,
		Units:                 units,
		NumberOfJudgmentRuns:  &analysis.NumberOfJudgmentRuns,
		AbortOnFailedJudgment: &analysis.AbortOnFailedJudgment,
		LookbackMethod:        lookbackMethod,
		Queries:               analysis.Queries,
	}, nil
}

func createBlueGreenRedirectConditions(conditions []*model.BlueGreenCondition, analysisConfig *model.AnalysisConfig) ([]de.KubernetesV2RedirectTrafficAfter, error) {
	var redirectConditions []de.KubernetesV2RedirectTrafficAfter
	for _, condition := range conditions {
		if condition.Pause != nil {
			pause, err := createPauseStep(condition.Pause)
			if err != nil {
				return nil, err
			}
			redirectConditions = append(
				redirectConditions,
				de.KubernetesV2RedirectTrafficAfter{
					Pause: pause,
				})
		}
		if condition.Analysis != nil {
			analysis, err := createDeploymentCanaryAnalysisStep(condition.Analysis, analysisConfig)
			if err != nil {
				return nil, err
			}

			redirectConditions = append(
				redirectConditions,
				de.KubernetesV2RedirectTrafficAfter{
					Analysis: analysis,
				})
		}
		if condition.RunWebhook != nil {
			redirectConditions = append(redirectConditions, de.KubernetesV2RedirectTrafficAfter{
				WebhookRun: &de.WebhooksWebhookRunStepInput{
					Name:    condition.RunWebhook.Name,
					Context: condition.RunWebhook.Context,
				},
			})
		}
	}
	return redirectConditions, nil
}

func createBlueGreenShutdownConditions(conditions []*model.BlueGreenCondition, analysisConfig *model.AnalysisConfig) ([]de.KubernetesV2ShutDownOldVersionAfter, error) {
	var shutDownConditions []de.KubernetesV2ShutDownOldVersionAfter
	for _, condition := range conditions {
		if condition.Pause != nil {
			pause, err := createPauseStep(condition.Pause)
			if err != nil {
				return nil, err
			}
			shutDownConditions = append(
				shutDownConditions,
				de.KubernetesV2ShutDownOldVersionAfter{
					Pause: pause,
				})
		}
		if condition.Analysis != nil {
			analysis, err := createDeploymentCanaryAnalysisStep(condition.Analysis, analysisConfig)
			if err != nil {
				return nil, err
			}

			shutDownConditions = append(
				shutDownConditions,
				de.KubernetesV2ShutDownOldVersionAfter{
					Analysis: analysis,
				})
		}
		if condition.RunWebhook != nil {
			shutDownConditions = append(shutDownConditions, de.KubernetesV2ShutDownOldVersionAfter{
				WebhookRun: &de.WebhooksWebhookRunStepInput{
					Name:    condition.RunWebhook.Name,
					Context: condition.RunWebhook.Context,
				},
			})
		}
	}
	return shutDownConditions, nil
}

func createPauseStep(pause *model.PauseStep) (*de.KubernetesV2PauseStep, error) {
	if err := validatePauseStep(pause); err != nil {
		return nil, err
	}
	unit, err := createTimeUnit(pause)
	if err != nil {
		return nil, err
	}

	pauseStep := de.NewKubernetesV2PauseStep()
	pauseStep.SetUnit(*unit)
	pauseStep.SetUntilApproved(pause.UntilApproved)
	pauseStep.SetDuration(pause.Duration)
	return pauseStep, nil
}

func createPauseConstraint(pause *model.PauseStep) (*de.PipelinePauseConstraint, error) {
	if err := validatePauseStep(pause); err != nil {
		return nil, err
	}
	pauseConstraint := de.NewPipelinePauseConstraint()
	unit, err := createTimeUnit(pause)
	if err != nil {
		return nil, err
	}
	pauseConstraint.SetUnit(*unit)
	pauseConstraint.SetUntilApproved(pause.UntilApproved)
	pauseConstraint.SetDuration(pause.Duration)
	return pauseConstraint, nil
}

func createWebhookConstraint(webhook *model.WebhookStep) (*de.WebhooksWebhookRunStepInput, error) {
	if err := validateWebhookStep(webhook); err != nil {
		return nil, err
	}
	webhookConstraint := de.NewWebhooksWebhookRunStepInput()
	webhookConstraint.SetName(*webhook.Name)
	if webhook.Context != nil {
		webhookConstraint.SetContext(*webhook.Context)
	}
	return webhookConstraint, nil
}

func createCanaryPause(pause *model.PauseStep) (*de.KubernetesV2PauseStep, error) {
	if err := validatePauseStep(pause); err != nil {
		return nil, err
	}
	pauseStep := de.NewKubernetesV2PauseStep()
	unit, err := createTimeUnit(pause)
	if err != nil {
		return nil, err
	}
	pauseStep.SetUnit(*unit)
	pauseStep.SetUntilApproved(pause.UntilApproved)
	pauseStep.SetDuration(pause.Duration)
	return pauseStep, nil
}

func createTimeUnit(pause *model.PauseStep) (*de.TimeTimeUnit, error) {
	var unit *de.TimeTimeUnit
	var err error
	if pause.Unit == "" {
		unit, err = de.NewTimeTimeUnitFromValue("NONE")
	} else {
		unit, err = de.NewTimeTimeUnitFromValue(strings.ToUpper(pause.Unit))
	}
	if err != nil {
		return nil, err
	}
	return unit, nil
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
	if *webhook.Name == "" {
		return errors.New("webhook constraint is not valid: you must provide a name for a configured webhook")
	}
	return nil
}

func findByName(queries []model.Query, name string) *model.Query {
	for _, configQuery := range queries {
		if name == *configQuery.Name {
			return &configQuery
		}
	}
	return nil
}

func buildWebhooks(webhooks []model.WebhookConfig) (*[]de.WebhooksWebhookRunConfig, error) {
	var webhooksList []de.WebhooksWebhookRunConfig
	for _, webhook := range webhooks {
		var body string
		if webhook.BodyTemplate != nil {
			var err error
			body, err = buildBody(webhook.BodyTemplate)
			if err != nil {
				return nil, err
			}
		}
		webhooksList = append(webhooksList, de.WebhooksWebhookRunConfig{
			Name:            webhook.Name,
			Method:          webhook.Method,
			UriTemplate:     webhook.UriTemplate,
			NetworkMode:     webhook.NetworkMode,
			AgentIdentifier: webhook.AgentIdentifier,
			RetryCount:      getRetryCount(webhook.RetryCount),
			Headers:         buildHeaders(webhook.Headers),
			BodyTemplate:    &body,
		})
	}
	return &webhooksList, nil
}

func buildHeaders(headers *[]model.Header) *[]de.WebhooksWebhookHeaders {
	if headers == nil {
		return nil
	}

	var headersList []de.WebhooksWebhookHeaders
	for _, header := range *headers {
		headersList = append(headersList, de.WebhooksWebhookHeaders{
			Key:   header.Key,
			Value: header.Value,
		})
	}
	return &headersList
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

func getRetryCount(retries *int32) *int32 {
	if retries == nil {
		def := int32(0)
		return &def
	}
	return retries
}

func createSMIs(tm model.TrafficManagement) (*[]de.KubernetesV2SmiTrafficManagementConfig, error) {
	var smis []de.KubernetesV2SmiTrafficManagementConfig
	for _, s := range tm.SMI {
		if s.RootServiceName == nil {
			return nil, errors.New("rootServiceName required in smi")
		}
		smis = append(smis, de.KubernetesV2SmiTrafficManagementConfig{
			RootServiceName:   s.RootServiceName,
			CanaryServiceName: s.CanaryServiceName,
			TrafficSplitName:  s.TrafficSplitName,
		})
	}
	return &smis, nil
}
