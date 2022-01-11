package model

import deploy "github.com/armory-io/deploy-engine/pkg"

type Pipeline struct {
	Id                 *string                        			`json:"id,omitempty" yaml:"id,omitempty"`
	StartedAtIso8601   *string                        			`json:"startedAtIso8601,omitempty" yaml:"startedAtIso8601,omitempty"`
	CompletedAtIso8601 *string                        			`json:"completedAtIso8601,omitempty" yaml:"completedAtIso8601,omitempty"`
	Application        *string                        			`json:"application,omitempty" yaml:"application,omitempty"`
	Source             *deploy.PipelinePipelineSource        	`json:"source,omitempty" yaml:"source,omitempty"`
	Environments       *[]deploy.PipelinePipelineEnvironment 	`json:"environments,omitempty" yaml:"environments,omitempty"`
	Status             *deploy.PipelinePipelineStatus        	`json:"status,omitempty" yaml:"status,omitempty"`
	Steps              *[]Step       							`json:"steps,omitempty" yaml:"steps,omitempty"`
}

func NewPipeline(pipelineStatus deploy.PipelinePipelineStatusResponse, steps *[]Step) *Pipeline {
	return &Pipeline{
		Id:       			pipelineStatus.Id,
		StartedAtIso8601:	pipelineStatus.StartedAtIso8601,
		CompletedAtIso8601:	pipelineStatus.CompletedAtIso8601,
		Application: 		pipelineStatus.Application,
		Source:				pipelineStatus.Source,
		Environments:		pipelineStatus.Environments,
		Status:       		pipelineStatus.Status,
		Steps:         		steps,
	}
}

type Step struct {
	Status     *deploy.PipelinePipelineStatus           	`json:"status,omitempty" yaml:"status,omitempty"`
	Ref        *string                          			`json:"ref,omitempty" yaml:"ref,omitempty"`
	DependsOn  *[]string                        			`json:"dependsOn,omitempty" yaml:"dependsOn,omitempty"`
	Type       *string                          			`json:"type,omitempty" yaml:"type,omitempty"`
	Deployment *deploy.DeploymentV2DeploymentStatusResponse `json:"deployment,omitempty" yaml:"deployment,omitempty"`
	Pause      *deploy.PipelinePipelinePauseStage      		`json:"pause,omitempty" yaml:"pause,omitempty"`
}

func NewStep(pipelineStage deploy.PipelinePipelineStage, deployment *deploy.DeploymentV2DeploymentStatusResponse) Step {
	return Step{
		Status:		pipelineStage.Status,
		Ref:		pipelineStage.Ref,
		DependsOn:	pipelineStage.DependsOn,
		Type: 		pipelineStage.Type,
		Deployment:	deployment,
		Pause:		pipelineStage.Pause,
	}
}