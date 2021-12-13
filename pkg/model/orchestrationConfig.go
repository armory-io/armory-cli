package model

type OrchestrationConfig struct {
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


