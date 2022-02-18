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
	if config.Analysis != nil {
		analysis.DefaultAccount = &config.Analysis.DefaultMetricProviderName
		queries, err := CreateAnalysisQueries(*config.Analysis.Queries, config.Analysis.DefaultMetricProviderName)
		if err != nil {
			return nil, err
		}
		analysis.Queries = queries
	}
	for key, element := range *config.Targets {

		envName := key
		target := element
		environments = append(environments, de.PipelinePipelineEnvironment{
			Name:      &envName,
			Namespace: &target.Namespace,
			Account:   &target.Account,
		})


		strategy, err := buildStrategy(*config.Strategies, element.Strategy)
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
		deployments = append(deployments, deploymentToAdd)
	}
	req := de.PipelineStartPipelineRequest{
		Application:  &application,
		Environments: &environments,
		Deployments:  &deployments,
	}
	return &req, nil
}

func createDeploymentCanarySteps(strategy model.Strategy) ([]de.KubernetesV2CanaryStep, error) {
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
			analysis, err := createDeploymentCanaryAnalysisStep(step.Analysis)
			if err != nil {
				return nil, err
			}

			steps = append(
				steps,
				de.KubernetesV2CanaryStep{
					Analysis: analysis,
				})
		}
	}
	return steps, nil
}

func CreateAnalysisQueries(queries []model.Query, defaultMetricProviderName string) (*[]de.AnalysisAnalysisQueries, error) {
	analysisQueries := make([]de.AnalysisAnalysisQueries, 0, len(queries))
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
	for _, obj := range *beforeDeployment {
		pause, err := createPauseConstraint(obj.Pause)
		if err != nil {
			return nil, err
		}
		constraint := de.PipelineConstraint{
			Pause: pause,
		}
		pipelineConstraints = append(pipelineConstraints, constraint)
	}
	return pipelineConstraints, nil
}

func buildStrategy(configStrategies map[string]model.Strategy, strategyName string) (*de.PipelinePipelineStrategy, error) {
	strategy := configStrategies[strategyName]
	if strategy.Canary != nil {
		steps, err := createDeploymentCanarySteps(strategy)
		if err != nil {
			return nil, err
		}
		return &de.PipelinePipelineStrategy{
			Canary: &de.KubernetesV2CanaryStrategy{
				Steps: steps,
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
				ActiveUrl:      &strategy.BlueGreen.ActiveRootUrl,
				PreviewUrl:     &strategy.BlueGreen.PreviewRootUrl,
			},
		}
		if strategy.BlueGreen.RedirectTrafficAfter != nil {
			redirectTrafficAfter, err := createBlueGreenRedirectConditions(strategy.BlueGreen.RedirectTrafficAfter)
			if err != nil {
				return nil, err
			}
			ps.BlueGreen.RedirectTrafficAfter = &redirectTrafficAfter
		}
		if strategy.BlueGreen.ShutdownOldVersionAfter != nil {
			shutdownOldVersionAfter, err := createBlueGreenShutdownConditions(strategy.BlueGreen.ShutdownOldVersionAfter)
			if err != nil {
				return nil, err
			}
			ps.BlueGreen.ShutdownOldVersionAfter = &shutdownOldVersionAfter
		}
		return ps, nil
	}

	return nil, fmt.Errorf("%s is not a valid strategy; define canary or bluegreen strategy", strategyName)
}

func createDeploymentCanaryAnalysisStep(analysis *model.AnalysisStep) (*de.AnalysisAnalysisStepInput, error) {
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

func createBlueGreenRedirectConditions(conditions []*model.BlueGreenCondition) ([]de.KubernetesV2RedirectTrafficAfter, error) {
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
		// TODO(cat): analysis condition
	}
	return redirectConditions, nil
}

func createBlueGreenShutdownConditions(conditions []*model.BlueGreenCondition) ([]de.KubernetesV2ShutdownOldVersionAfter, error) {
	var shutdownConditions []de.KubernetesV2ShutdownOldVersionAfter
	for _, condition := range conditions {
		if condition.Pause != nil {
			pause, err := createPauseStep(condition.Pause)
			if err != nil {
				return nil, err
			}
			shutdownConditions = append(
				shutdownConditions,
				de.KubernetesV2ShutdownOldVersionAfter{
					Pause: pause,
				})
		}
		// TODO(cat): analysis condition
	}
	return shutdownConditions, nil
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

func createCanaryPause(pause *model.PauseStep) (*de.KubernetesV2CanaryPauseStep, error) {
	if err := validatePauseStep(pause); err != nil {
		return nil, err
	}
	pauseStep := de.NewKubernetesV2CanaryPauseStep()
	unit, err := createTimeUnit(pause)
	if err != nil {
		return nil, err
	}
	pauseStep.SetUnit(*unit)
	pauseStep.SetUntilApproved(pause.UntilApproved)
	pauseStep.SetDuration(pause.Duration)
	return pauseStep, nil
}

func createTimeUnit(pause *model.PauseStep) (*de.TimeTimeUnit, error){
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