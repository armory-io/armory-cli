package model

import deploy "github.com/armory-io/deploy-engine/pkg"

type Pipeline struct {
	Id                 *string                        			`json:"id,omitempty"`
	StartedAtIso8601   *string                        			`json:"startedAtIso8601,omitempty"`
	CompletedAtIso8601 *string                        			`json:"completedAtIso8601,omitempty"`
	Application        *string                        			`json:"application,omitempty"`
	Source             *deploy.PipelinePipelineSource        	`json:"source,omitempty"`
	Environments       *[]deploy.PipelinePipelineEnvironment 	`json:"environments,omitempty"`
	Status             *deploy.PipelinePipelineStatus        	`json:"status,omitempty"`
	Steps              *[]Step       							`json:"steps,omitempty"`
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
	Status     *deploy.PipelinePipelineStatus           	`json:"status,omitempty"`
	Ref        *string                          			`json:"ref,omitempty"`
	DependsOn  *[]string                        			`json:"dependsOn,omitempty"`
	Type       *string                          			`json:"type,omitempty"`
	Deployment *deploy.DeploymentV2DeploymentStatusResponse `json:"deployment,omitempty"`
	Pause      *deploy.PipelinePipelinePauseStage      		`json:"pause,omitempty"`
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