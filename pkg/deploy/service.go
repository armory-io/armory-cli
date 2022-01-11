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

func CreateDeploymentRequest(config *model.OrchestrationConfig) (*de.PipelineStartPipelineRequest, error) {
	environments := make([]de.PipelinePipelineEnvironment, 0, len(*config.Targets))
	deployments := make([]de.PipelinePipelineDeployment, 0, len(*config.Targets))
	for key, element  := range *config.Targets {
		envName := key
		target := element
		environments = append(environments, de.PipelinePipelineEnvironment{
			Name: &envName,
			Namespace: &target.Namespace,
			Account: &target.Account,
		})

		strategy := (*config.Strategies)[element.Strategy]
		if &strategy.Canary == nil  {
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
			Manifests: CreateDeploymentManifests(files),
			Strategy: &de.PipelinePipelineStrategy{
				Canary: &de.KubernetesV2CanaryStrategy{
					Steps: steps,
				},
			},
			Constraints: &pipelineConstraint,
		})
	}
	req := de.PipelineStartPipelineRequest{
		Application: &config.Application,
		Environments: &environments,
		Deployments: &deployments,
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
			var unit *de.KubernetesV2CanaryPauseStepTimeUnit
			var err error
			if step.Pause.Unit == "" {
				unit, err = de.NewKubernetesV2CanaryPauseStepTimeUnitFromValue("NONE")
			} else {
				unit, err = de.NewKubernetesV2CanaryPauseStepTimeUnitFromValue(strings.ToUpper(step.Pause.Unit))
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
	}
	return steps, nil
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

func CreateDeploymentManifests(manifests *[]string) *[]de.KubernetesV2Manifest{
	deManifests := make([]de.KubernetesV2Manifest, 0, len(*manifests))
	for _, manifest := range *manifests {
		deManifests = append(
			deManifests,
			de.KubernetesV2Manifest {
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