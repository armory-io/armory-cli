package orchestration

import (
	"fmt"
	de "github.com/armory-io/deploy-engine/deploy/client"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"strings"
)



type Orchestration struct {
	Version string `yaml:"version,omitempty"`
	Kind string `yaml:"kind,omitempty"`
	Application string                   `yaml:"application,omitempty"`
	Targets *map[string]DeploymentTarget `yaml:"targets,omitempty"`
	Manifests *[]ManifestPath `yaml:"manifests,omitempty"`
	Strategies *map[string]Strategy `yaml:"strategies,omitempty"`
}

type Strategy struct {
	Canary *CanaryStrategy `yaml:"canary,omitempty"`
}

type CanaryStrategy struct {
	Steps *[]CanaryStep `yaml:"steps,omitempty"`
}

type CanaryStep struct {
	SetWeight *WeightStep `yaml:"setWeight,omitempty"`
	Pause *PauseStep      `yaml:"pause,omitempty"`
}

type WeightStep struct {
	Weight int32 `yaml:"weight,omitempty"`
}

type PauseStep struct {
	// The duration of the pause. If duration is non-zero, untilApproved should be set to false.
	Duration int32 `yaml:"duration,omitempty"`
	Unit string `yaml:"unit,omitempty"`
	// If set to true, the progressive canary will wait until a manual judgment to continue. This field should not be set to true unless duration and unit are unset.
	UntilApproved bool `yaml:"untilApproved,omitempty"`
}

func (o Orchestration) GetManifestsFromFile() (*[]string, error) {
	var fileNames []string
	for _, manifestPath := range *o.Manifests {
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

func (o Orchestration) ToDeployEngineManifests(manifests *[]string) *[]de.KubernetesV2Manifest{
	deManifests := make([]de.KubernetesV2Manifest, 0, len(*manifests))
	for _, manifest := range *manifests {
		deManifests = append(
			deManifests,
			de.KubernetesV2Manifest {
				Inline: &de.KubernetesV2InlineManifest{
					Value: &manifest,
				},
			})
	}
	return &deManifests
}

type DeploymentTarget struct {
	// The name of the Kubernetes account to be used for this deployment.
	Account string `yaml:"account,omitempty"`
	// The Kubernetes namespace where the provided manifests will be deployed.
	Namespace string `yaml:"namespace,omitempty"`
	// This is the key to a strategy under the strategies map
	Strategy string `yaml:"strategy,omitempty"`
}

type ManifestPath struct {
	Path string `yaml:"path,omitempty"`
}

func (o *Orchestration) ToDeployEngineDeployment() (*de.KubernetesV2StartKubernetesDeploymentRequest, error){
	if len(*o.Targets) != 1 {
		return nil, fmt.Errorf("exactly one target is required for a deployment")
	}
	targetKeys := make([]string, 0, len(*o.Targets))
	for key := range *o.Targets {
		targetKeys = append(targetKeys, key)
	}
	target := (*o.Targets)[targetKeys[0]]

	strategyKeys := make([]string, 0, len(*o.Strategies))
	for key := range *o.Strategies {
		strategyKeys = append(strategyKeys, key)
	}

	strategy := (*o.Strategies)[target.Strategy]
	if &strategy.Canary == nil  {
		return nil, fmt.Errorf("error converting steps for canary deployment strategy; canary strategy not provided and is required")
	}

	steps := make([]de.KubernetesV2CanaryStep, 0, len(*strategy.Canary.Steps))
	for _, step := range *strategy.Canary.Steps {
		if step.SetWeight != nil {
			steps = append(
				steps,
				de.KubernetesV2CanaryStep {
					SetWeight: &de.KubernetesV2CanarySetWeightStep{
						Weight: &step.SetWeight.Weight,
					},
					Pause:     nil,
				})
		}

		if step.Pause != nil {
			var unit *de.CanaryPauseStepTimeUnit
			var err error
			if step.Pause.Unit == "" {
				unit, err = de.NewCanaryPauseStepTimeUnitFromValue("NONE")
			} else {
				unit, err = de.NewCanaryPauseStepTimeUnitFromValue(strings.ToUpper(step.Pause.Unit))
			}

			if err != nil {
				return nil, err
			}
			steps = append(
				steps,
				de.KubernetesV2CanaryStep {
					SetWeight: nil,
					Pause:     &de.KubernetesV2CanaryPauseStep{
						Duration:      &step.Pause.Duration,
						Unit:          unit.Ptr(),
						UntilApproved: &step.Pause.UntilApproved,
					},
			})
		}
	}
	files, err := o.GetManifestsFromFile()
	if err != nil {
		return nil, err
	}
	req := de.KubernetesV2StartKubernetesDeploymentRequest{
		Application: &o.Application,
		Account: &target.Account,
		Namespace: &target.Namespace,
		Manifests: o.ToDeployEngineManifests(files),
		Canary: &de.KubernetesV2CanaryStrategy{
			Steps: &steps,
		},
	}
	return &req, nil
}



