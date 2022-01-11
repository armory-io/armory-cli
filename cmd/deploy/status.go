package deploy

import (
	"context"
	"fmt"
	deploy "github.com/armory-io/deploy-engine/pkg"
	"github.com/armory/armory-cli/pkg/model"
	"github.com/spf13/cobra"
	_nethttp "net/http"
	"time"
)

const (
	deployStatusShort   = "Watch deployment on Armory Cloud"
	deployStatusLong    = "Watch deployment on Armory Cloud"
	deployStatusExample = "armory deploy status [options]"
)

type deployStatusOptions struct {
	*deployOptions
}

type FormattableDeployStatus struct {
	DeployResp   model.Pipeline `json:"deployment"`
	httpResponse *_nethttp.Response
	err          error
}

func (u FormattableDeployStatus) Get() interface{} {
	return u.DeployResp
}

func (u FormattableDeployStatus) GetHttpResponse() *_nethttp.Response {
	return u.httpResponse
}

func (u FormattableDeployStatus) GetFetchError() error {
	return u.err
}

func newDeployStatusResponseWrapper(raw model.Pipeline, response *_nethttp.Response, err error) FormattableDeployStatus {
	wrapper := FormattableDeployStatus{
		DeployResp:   raw,
		httpResponse: response,
		err:          err,
	}
	return wrapper
}

func (u FormattableDeployStatus) String() string {
	ret := ""
	now := time.Now().Format(time.RFC3339)
	ret += fmt.Sprintf("[%v] application: %s, started: %s\n", now, *u.DeployResp.Application, *u.DeployResp.StartedAtIso8601)
	ret += fmt.Sprintf("[%v] status: ", now)
	switch status := *u.DeployResp.Status; status {
	case deploy.PIPELINEPIPELINESTATUS_PAUSED:
		for _, stages := range *u.DeployResp.Steps {
			if *stages.Type == "pause" && *stages.Status == deploy.PIPELINEPIPELINESTATUS_PAUSED {
				ret += fmt.Sprintf("[%s] msg: Paused for %d %s. You can skip the pause in the cloud console or CLI\n", status, stages.Pause.GetDuration(), stages.Pause.GetUnit())
			}
		}
	case deploy.PIPELINEPIPELINESTATUS_AWAITING_APPROVAL:
		ret += fmt.Sprintf("[%s] msg: Paused for Manual Judgment. You can approve the rollout and continue the deployment in the cloud console or CLI.\n", status)
	default:
		ret += string(status) + "\n"
	}
	return ret
}

func NewDeployStatusCmd(deployOptions *deployOptions) *cobra.Command {
	options := &deployStatusOptions{
		deployOptions: deployOptions,
	}
	cmd := &cobra.Command{
		Use:     "status --deploymentId [deploymentId]",
		Aliases: []string{"status"},
		Short:   deployStatusShort,
		Long:    deployStatusLong,
		Example: deployStatusExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return status(cmd, options)
		},
	}
	cmd.Flags().StringVarP(&options.deploymentId, "deploymentId", "i", "", "(Required) The ID of an existing deployment.")
	cmd.MarkFlagRequired("deploymentId")
	return cmd
}

func status(cmd *cobra.Command, options *deployStatusOptions) error {
	ctx, cancel := context.WithTimeout(options.DeployClient.Context, time.Second*5)
	defer cancel()
	req := options.DeployClient.DeploymentServiceApi.DeploymentServicePipelineStatus(ctx, options.deploymentId)
	pipelineResp, response, err := req.Execute()
	var steps []model.Step
	if response != nil && response.StatusCode == 200 && options.O != "" {
		for _, stages := range pipelineResp.GetSteps() {
			var step = model.Step{}
			if stages.GetType() == "deployment" && stages.GetStatus() != deploy.PIPELINEPIPELINESTATUS_NOT_STARTED {
				deployment := stages.GetDeployment()
				request := options.DeployClient.DeploymentServiceApi.DeploymentServiceStatus(ctx, deployment.GetId())
				deploy, response, err := request.Execute()
				err = getRequestError(response, err)
				if err != nil {
					return err
				}
				step = model.NewStep(stages, &deploy)
			} else {
				step = model.NewStep(stages, nil)
			}
			steps = append(steps, step)
		}
	}
	pipeline := model.NewPipeline(pipelineResp, &steps)
	dataFormat, err := options.Output.Formatter(newDeployStatusResponseWrapper(*pipeline, response, err))
	// if we've made it this far, the command is valid. if an error occurs it isn't a usage error
	cmd.SilenceUsage = true
	if err != nil {
		return fmt.Errorf("error trying to parse respone: %s", err)
	}
	_, err = fmt.Fprintln(cmd.OutOrStdout(), dataFormat)
	return err
}

func getRequestError(response *_nethttp.Response, err error) error {
	if err != nil {
		// don't override the received error unless we have an unexpected http response status
		if response != nil && response.StatusCode >= 300 {
			openAPIErr := err.(deploy.GenericOpenAPIError)
			err = fmt.Errorf("request returned an error: status code(%d) %s",
				response.StatusCode, string(openAPIErr.Body()))
		}
	}
	return err
}
