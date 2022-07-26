package model

import deploy "github.com/armory-io/deploy-engine/api"

type Pipeline struct {
	Id                 *string                       `json:"id,omitempty" yaml:"id,omitempty"`
	StartedAtIso8601   *string                       `json:"startedAtIso8601,omitempty" yaml:"startedAtIso8601,omitempty"`
	CompletedAtIso8601 *string                       `json:"completedAtIso8601,omitempty" yaml:"completedAtIso8601,omitempty"`
	Application        *string                       `json:"application,omitempty" yaml:"application,omitempty"`
	Source             *deploy.Source                `json:"source,omitempty" yaml:"source,omitempty"`
	Environments       *[]deploy.PipelineEnvironment `json:"environments,omitempty" yaml:"environments,omitempty"`
	Status             *deploy.WorkflowStatus        `json:"status,omitempty" yaml:"status,omitempty"`
	Steps              *[]Step                       `json:"steps,omitempty" yaml:"steps,omitempty"`
}

func NewPipeline(pipelineStatus *deploy.PipelineStatusResponse, steps *[]Step) *Pipeline {
	if pipelineStatus == nil {
		return &Pipeline{}
	}

	return &Pipeline{
		Id:                 &pipelineStatus.ID,
		StartedAtIso8601:   &pipelineStatus.StartedAtIso8601,
		CompletedAtIso8601: &pipelineStatus.CompletedAtIso8601,
		Application:        &pipelineStatus.Application,
		Source:             &pipelineStatus.Source,
		Environments:       &pipelineStatus.Environments,
		Status:             &pipelineStatus.Status,
		Steps:              steps,
	}
}

type Step struct {
	Status     *deploy.WorkflowStatus           `json:"status,omitempty" yaml:"status,omitempty"`
	Ref        *string                          `json:"ref,omitempty" yaml:"ref,omitempty"`
	DependsOn  *[]string                        `json:"dependsOn,omitempty" yaml:"dependsOn,omitempty"`
	Type       *string                          `json:"type,omitempty" yaml:"type,omitempty"`
	Deployment *deploy.DeploymentStatusResponse `json:"deployment,omitempty" yaml:"deployment,omitempty"`
	Pause      *deploy.PauseStepResponse        `json:"pause,omitempty" yaml:"pause,omitempty"`
}

func NewStep(pipelineStage *deploy.PipelineStep, deployment *deploy.DeploymentStatusResponse) Step {
	return Step{
		Status:     &pipelineStage.Status,
		Ref:        &pipelineStage.Ref,
		DependsOn:  &pipelineStage.DependsOn,
		Type:       &pipelineStage.Type,
		Deployment: deployment,
		Pause:      pipelineStage.Pause,
	}
}
