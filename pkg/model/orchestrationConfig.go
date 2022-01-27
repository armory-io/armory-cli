package model

type OrchestrationConfig struct {
	Version     string                       `yaml:"version,omitempty"`
	Kind        string                       `yaml:"kind,omitempty"`
	Application string                       `yaml:"application,omitempty"`
	Targets     *map[string]DeploymentTarget `yaml:"targets,omitempty"`
	Manifests   *[]ManifestPath              `yaml:"manifests,omitempty"`
	Strategies  *map[string]Strategy         `yaml:"strategies,omitempty"`
}

type Strategy struct {
	Canary    *CanaryStrategy    `yaml:"canary,omitempty"`
	BlueGreen *BlueGreenStrategy `yaml:"blue-green,omitempty"`
}

type CanaryStrategy struct {
	Steps *[]CanaryStep `yaml:"steps,omitempty"`
}

type CanaryStep struct {
	SetWeight *WeightStep `yaml:"setWeight,omitempty"`
	Pause     *PauseStep  `yaml:"pause,omitempty"`
}

type BlueGreenStrategy struct {
	RedirectTrafficAfter    *RedirectTrafficAfter    `yaml:"redirectTrafficAfter,omitempty"`
	ShutdownOldVersionAfter *ShutdownOldVersionAfter `yaml:"shutdownOldVersionAfter,omitempty"`
	ActiveService           string                   `yaml:"activeService,omitempty"`
	PreviewService          string                   `yaml:"previewService,omitempty"`
	ActiveRootUrl           string                   `yaml:"activeRootUrl,omitempty"`
	PreviewRootUrl          string                   `yaml:"previewRootUrl,omitempty"`
}

type RedirectTrafficAfter struct {
	Steps *[]BlueGreenStep `yaml:"steps,omitempty"`
}

type ShutdownOldVersionAfter struct {
	Steps *[]BlueGreenStep `yaml:"steps,omitempty"`
}

// TODO(cat): analysis step
type BlueGreenStep struct {
	Pause *PauseStep `yaml:"pause,omitempty"`
}

type WeightStep struct {
	Weight int32 `yaml:"weight,omitempty"`
}

type PauseStep struct {
	// The duration of the pause. If duration is non-zero, untilApproved should be set to false.
	Duration int32  `yaml:"duration,omitempty"`
	Unit     string `yaml:"unit,omitempty"`
	// If set to true, the progressive canary will wait until a manual judgment to continue. This field should not be set to true unless duration and unit are unset.
	UntilApproved bool `yaml:"untilApproved,omitempty"`
}

type DeploymentTarget struct {
	// The name of the Kubernetes account to be used for this deployment.
	Account string `yaml:"account,omitempty"`
	// The Kubernetes namespace where the provided manifests will be deployed.
	Namespace string `yaml:"namespace,omitempty"`
	// This is the key to a strategy under the strategies map
	Strategy    string       `yaml:"strategy,omitempty"`
	Constraints *Constraints `yaml:"constraints,omitempty"`
}

type ManifestPath struct {
	Path    string   `yaml:"path,omitempty"`
	Targets []string `yaml:"targets,omitempty"`
	Inline  string   `yaml:"inline,omitempty"`
}

type Constraints struct {
	DependsOn        *[]string           `yaml:"dependsOn,omitempty"`
	BeforeDeployment *[]BeforeDeployment `yaml:"beforeDeployment,omitempty"`
}

type BeforeDeployment struct {
	Pause *PauseStep `yaml:"pause,omitempty"`
}
