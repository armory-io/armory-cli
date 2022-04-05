package model

type OrchestrationConfigV2 struct {
	Version     string                `yaml:"version"`
	Kind        string                `yaml:"kind"`
	Application string                `yaml:"application" minLength:"3" maxLength:"255" doc:"The name of the application to deploy."`
	Targets     *[]DeploymentTargetV2 `yaml:"targets" minItems:"1" doc:"List of of your deployment target, Borealis supports deploying to one target cluster."'`
	Manifests   *[]ManifestPath       `yaml:"manifests" minItems:"1" doc:"The list of manifest sources. Can be a directory or file."`
	Strategies  *[]StrategyV2         `yaml:"strategies" minItems:"1" doc:"A list of named strategies that can be assigned to deployment targets in the targets block."`
}

type DeploymentTargetV2 struct {
	Name string `yaml:"name" doc:"Name for your deployment. Use a descriptive value such as the environment name."`
	DeploymentTarget
}

type StrategyV2 struct {
	Name string `yaml:"name" doc:"Name for a strategy that you use to refer to it. Used in the target block. This example uses strategy1 as the name."`
	Strategy
}
