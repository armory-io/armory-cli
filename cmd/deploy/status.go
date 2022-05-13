package deploy

import (
	"context"
	"fmt"
	deploy "github.com/armory-io/deploy-engine/pkg"
	"github.com/armory/armory-cli/pkg/cmdUtils"
	"github.com/armory/armory-cli/pkg/config"
	"github.com/armory/armory-cli/pkg/model"
	"github.com/armory/armory-cli/pkg/output"
	"github.com/spf13/cobra"
	_nethttp "net/http"
	"time"
)

const (
	deployStatusShort   = "Watch deployment on Armory CD-as-a-Service"
	deployStatusLong    = "Watch deployment on Armory CD-as-a-Service"
	deployStatusExample = "armory deploy status [options]"
)

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
	case deploy.WORKFLOWWORKFLOWSTATUS_PAUSED:
		for _, stages := range *u.DeployResp.Steps {
			if *stages.Type == "pause" && *stages.Status == deploy.WORKFLOWWORKFLOWSTATUS_PAUSED {
				ret += fmt.Sprintf("[%s] msg: Paused for %d %s. You can skip the pause in the CD-as-a-Service Console or CLI\n", status, stages.Pause.GetDuration(), stages.Pause.GetUnit())
			}
		}
	case deploy.WORKFLOWWORKFLOWSTATUS_AWAITING_APPROVAL:
		ret += fmt.Sprintf("[%s] msg: Paused for Manual Judgment. You can approve the rollout and continue the deployment in the CD-as-a-Service Console or CLI.\n", status)
	default:
		ret += string(status) + "\n"
	}
	return ret
}

func NewDeployStatusCmd(configuration *config.Configuration) *cobra.Command {
	deploymentId := ""
	cmd := &cobra.Command{
		Use:     "status --deploymentId [deploymentId]",
		Aliases: []string{"status"},
		Short:   deployStatusShort,
		Long:    deployStatusLong,
		Example: deployStatusExample,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cmdUtils.ExecuteParentHooks(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return status(cmd, configuration, deploymentId)
		},
	}
	cmd.Flags().StringVarP(&deploymentId, "deploymentId", "i", "", "(Required) The ID of an existing deployment.")
	cmd.MarkFlagRequired("deploymentId")
	return cmd
}

func status(cmd *cobra.Command, configuration *config.Configuration, deploymentId string) error {
	cmd.SetContext(context.WithValue(cmd.Context(), "deploymentId", deploymentId))
	deployClient := configuration.GetDeployEngineClient()
	ctx, cancel := context.WithTimeout(deployClient.Context, time.Second*5)
	defer cancel()
	req := deployClient.DeploymentServiceApi.DeploymentServicePipelineStatus(ctx, deploymentId)
	pipelineResp, response, err := req.Execute()
	var steps []model.Step
	if response != nil && response.StatusCode == 200 && configuration.GetOutputType() != output.Text {
		for _, stages := range pipelineResp.GetSteps() {
			var step = model.Step{}
			if stages.GetType() == "deployment" && stages.GetStatus() != deploy.WORKFLOWWORKFLOWSTATUS_NOT_STARTED {
				deployment := stages.GetDeployment()
				request := deployClient.DeploymentServiceApi.DeploymentServiceStatus(ctx, deployment.GetId())
				deployRes, response, err := request.Execute()
				err = getRequestError(response, err)
				if err != nil {
					return err
				}
				step = model.NewStep(stages, &deployRes)
			} else {
				step = model.NewStep(stages, nil)
			}
			steps = append(steps, step)
		}
	}
	pipeline := model.NewPipeline(pipelineResp, &steps)
	dataFormat, err := configuration.GetOutputFormatter()(newDeployStatusResponseWrapper(*pipeline, response, err))
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
