package service

import (
	"fmt"
	de "github.com/armory-io/deploy-engine/pkg"
	"github.com/armory/armory-cli/pkg/model"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func CreateDeploymentRequest(config *model.OrchestrationConfig) (*de.KubernetesV2StartKubernetesDeploymentRequest, error) {
	if len(*config.Targets) != 1 {
		return nil, fmt.Errorf("exactly one target is required for a deployment")
	}
	targetKeys := make([]string, 0, len(*config.Targets))
	for key := range *config.Targets {
		targetKeys = append(targetKeys, key)
	}
	target := (*config.Targets)[targetKeys[0]]

	strategyKeys := make([]string, 0, len(*config.Strategies))
	for key := range *config.Strategies {
		strategyKeys = append(strategyKeys, key)
	}

	strategy := (*config.Strategies)[target.Strategy]
	if &strategy.Canary == nil  {
		return nil, fmt.Errorf("error converting steps for canary deployment strategy; canary strategy not provided and is required")
	}

	steps, err := CreateDeploymentCanaryStep(strategy)
	if err != nil {
		return nil, err
	}
	files, err := GetManifestsFromFile(config.Manifests)
	if err != nil {
		return nil, err
	}
	req := de.KubernetesV2StartKubernetesDeploymentRequest{
		Application: config.Application,
		Account:     target.Account,
		Namespace:   target.Namespace,
		Manifests:   CreateDeploymentManifests(files),
		Canary: de.KubernetesV2CanaryStrategy{
			Steps: steps,
		},
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

func GetManifestsFromFile(manifests *[]model.ManifestPath) (*[]string, error) {
	var fileNames []string
	gitWorkspace, present := os.LookupEnv("GITHUB_WORKSPACE")
	for _, manifestPath := range *manifests {
		if present {
			manifestPath.Path = gitWorkspace + manifestPath.Path
		}
		err := filepath.WalkDir(manifestPath.Path, func(path string, info fs.DirEntry, err error) error {
			if err != nil {
				fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
				return err
			}
			if filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml"  {
				fileNames = append(fileNames, path)
			}

			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("unable to read manifest(s) from file: %s", err)
		}

	}
	var files []string
	for _, fileName := range fileNames {
		file, err := ioutil.ReadFile(fileName)
		if err != nil {
			return nil, fmt.Errorf("error trying to read manifest file '%s': %s", fileName, err)
		}
		files = append(files, string(file))
	}

	return &files, nil
}

func CreateDeploymentManifests(manifests *[]string) []de.KubernetesV2Manifest{
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
	return deManifests
}