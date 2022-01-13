package deploy

import (
	"fmt"
	de "github.com/armory-io/deploy-engine/pkg"
	"github.com/armory/armory-cli/pkg/model"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func CreateDeploymentRequest(application string, config *model.OrchestrationConfig) (*de.PipelineStartPipelineRequest, error) {
	environments := make([]de.PipelinePipelineEnvironment, 0, len(*config.Targets))
	deployments := make([]de.PipelinePipelineDeployment, 0, len(*config.Targets))
	var analysis de.AnalysisAnalysisConfig
	if &config.Analysis != nil {
		analysis.DefaultAccount = &config.Analysis.DefaultAccount
		analysis.DefaultType = &config.Analysis.DefaultType
		analysis.Queries = CreateAnalysisQueries(*config.Analysis.Queries, config.Analysis.DefaultAccount)
	}
	for key, element := range *config.Targets {
		envName := key
		target := element
		environments = append(environments, de.PipelinePipelineEnvironment{
			Name:      &envName,
			Namespace: &target.Namespace,
			Account:   &target.Account,
		})

		strategy := (*config.Strategies)[element.Strategy]
		if &strategy.Canary == nil {
			return nil, fmt.Errorf("error converting steps for canary deployment strategy; canary strategy not provided and is required")
		}

		steps, err := CreateDeploymentCanaryStep(strategy)
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

			pipelineConstraint.SetDependsOn(*target.Constraints.DependsOn)
			pipelineConstraint.SetBeforeDeployment(beforeDeployment)
		}

		deployments = append(deployments, de.PipelinePipelineDeployment{
			Environment: &envName,
			Manifests:   CreateDeploymentManifests(files),
			Strategy: &de.PipelinePipelineStrategy{
				Canary: &de.KubernetesV2CanaryStrategy{
					Steps: steps,
				},
			},
			Constraints: &pipelineConstraint,
			Analysis:    &analysis,
		})
	}
	req := de.PipelineStartPipelineRequest{
		Application:  &application,
		Environments: &environments,
		Deployments:  &deployments,
	}
	return &req, nil
}

func CreateDeploymentCanaryStep(strategy model.Strategy) ([]de.KubernetesV2CanaryStep, error) {
	steps := make([]de.KubernetesV2CanaryStep, 0, len(*strategy.Canary.Steps))
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
			var unit *de.TimeTimeUnit
			var err error
			if step.Pause.Unit == "" {
				unit, err = de.NewTimeTimeUnitFromValue("NONE")
			} else {
				unit, err = de.NewTimeTimeUnitFromValue(strings.ToUpper(step.Pause.Unit))
			}

			if err != nil {
				return nil, err
			}
			steps = append(
				steps,
				de.KubernetesV2CanaryStep{
					SetWeight: nil,
					Pause: &de.KubernetesV2CanaryPauseStep{
						Duration:      &step.Pause.Duration,
						Unit:          unit,
						UntilApproved: &step.Pause.UntilApproved,
					},
				})
		}

		if step.Analysis != nil {
			var rollBackMode *de.AnalysisRollMode
			var rollForwardMode *de.AnalysisRollMode
			var units *de.TimeTimeUnit
			var lookbackMethod *de.AnalysisLookbackMethod
			var err error

			if step.Analysis.RollBackMode != "" {
				rollBackMode, err = de.NewAnalysisRollModeFromValue(strings.ToUpper(step.Analysis.RollBackMode))
			} else {
				rollBackMode, err = de.NewAnalysisRollModeFromValue("AUTOMATIC")
			}
			if err != nil {
				return nil, err
			}

			if step.Analysis.RollForwardMode != "" {
				rollForwardMode, err = de.NewAnalysisRollModeFromValue(strings.ToUpper(step.Analysis.RollForwardMode))
			} else {
				rollForwardMode, err = de.NewAnalysisRollModeFromValue("AUTOMATIC")
			}
			if err != nil {
				return nil, err
			}
			if step.Analysis.Units != "" {
				units, err = de.NewTimeTimeUnitFromValue(strings.ToUpper(step.Analysis.Units))
			} else {
				units, err = de.NewTimeTimeUnitFromValue("NONE")
			}
			if err != nil {
				return nil, err
			}
			if step.Analysis.LookbackMethod != "" {
				lookbackMethod, err = de.NewAnalysisLookbackMethodFromValue(strings.ToUpper(step.Analysis.LookbackMethod))
			} else {
				lookbackMethod, err = de.NewAnalysisLookbackMethodFromValue("UNSET")
			}

			steps = append(
				steps,
				de.KubernetesV2CanaryStep{
					Analysis: &de.AnalysisAnalysisStepInput{
						Context:               &step.Analysis.Context,
						RollBackMode:          rollBackMode,
						RollForwardMode:       rollForwardMode,
						Interval:              &step.Analysis.Interval,
						Units:                 units,
						NumberOfJudgmentRuns:  &step.Analysis.Count,
						AbortOnFailedJudgment: &step.Analysis.AbortOnFailedJudgment,
						LookbackMethod:        lookbackMethod,
						Queries:               step.Analysis.Queries,
					},
				})
		}
	}
	return steps, nil
}

func CreateAnalysisQueries(queries []model.Query, defaultAccount string) *[]de.AnalysisAnalysisQueries {
	analysisQueries := make([]de.AnalysisAnalysisQueries, 0, len(queries))
	for _, query := range queries {

		analysisQueries = append(analysisQueries, de.AnalysisAnalysisQueries{
			Name:               &query.Name,
			QueryTemplate:      &query.QueryTemplate,
			UpperLimit:         &query.UpperLimit,
			LowerLimit:         &query.LowerLimit,
			MetricProviderName: &defaultAccount,
		})
	}
	return &analysisQueries
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

		if contains(manifestPath.Targets, env) || manifestPath.Targets == nil {
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
		var unit *de.TimeTimeUnit
		var err error
		if obj.Pause.Unit == "" {
			unit, err = de.NewTimeTimeUnitFromValue("NONE")
		} else {
			unit, err = de.NewTimeTimeUnitFromValue(strings.ToUpper(obj.Pause.Unit))
		}
		if err != nil {
			return nil, err
		}
		pause := de.NewPipelinePauseConstraint()
		pause.SetUnit(*unit)
		pause.SetUntilApproved(obj.Pause.UntilApproved)
		pause.SetDuration(obj.Pause.Duration)
		constraint := de.PipelineConstraint{
			Pause: pause,
		}
		pipelineConstraints = append(pipelineConstraints, constraint)
	}
	return pipelineConstraints, nil
}

func contains(s []string, searchterm string) bool {
	i := sort.SearchStrings(s, searchterm)
	return i < len(s) && s[i] == searchterm
}
