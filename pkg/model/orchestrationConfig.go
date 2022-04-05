package model

type OrchestrationConfig struct {
	Version     string                       `yaml:"version"`
	Kind        string                       `yaml:"kind"`
	Application string                       `yaml:"application"`
	Targets     *map[string]DeploymentTarget `yaml:"targets"`
	Manifests   *[]ManifestPath              `yaml:"manifests"`
	Strategies  *map[string]Strategy         `yaml:"strategies"`
}

type Strategy struct {
	Canary *CanaryStrategy `yaml:"canary,omitempty" doc:"The deployment strategy type. Use canary."`
}

type CanaryStrategy struct {
	Steps *[]CanaryStep `yaml:"steps,omitempty" doc:"The steps for your deployment strategy."`
}

type CanaryStep struct {
	SetWeight *WeightStep `yaml:"setWeight,omitempty"`
	Pause     *PauseStep  `yaml:"pause,omitempty" doc:" A pause step type. The pipeline stops until the pause behavior is completed. The pause behavior can be duration or untilApproved."`
}

type WeightStep struct {
	Weight int32 `yaml:"weight,omitempty" doc:" The percentage of pods that should be running the canary version for this step. Set it to an integer between 0 and 100, inclusive."`
}

type PauseStep struct {
	// The duration of the pause. If duration is non-zero, untilApproved should be set to false.
	Duration int32  `yaml:"duration,omitempty" doc:"The pause behavior is time (integer) before the deployment continues. If duration is set for this step, omit untilApproved."`
	Unit     string `yaml:"unit,omitempty" doc:"# The unit of time to use for the pause. Can be seconds, minutes, or hours. Required if duration is set."`
	// If set to true, the progressive canary will wait until a manual judgment to continue. This field should not be set to true unless duration and unit are unset.
	UntilApproved bool `yaml:"untilApproved,omitempty" doc:"# The pause behavior is the deployment waits until a manual approval is given to continue. Only set this to true if there is no duration pause behavior for this step."`
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
