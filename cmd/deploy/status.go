package deploy

import (
	"context"
	"fmt"
	deploy "github.com/armory-io/deploy-engine/pkg/api"
	"github.com/armory/armory-cli/cmd/utils"
	"github.com/armory/armory-cli/pkg/cmdUtils"
	"github.com/armory/armory-cli/pkg/config"
	deployment "github.com/armory/armory-cli/pkg/deploy"
	errorUtils "github.com/armory/armory-cli/pkg/errors"
	"github.com/spf13/cobra"
	log "go.uber.org/zap"
	"net/http"
	"time"
)

const (
	deployStatusShort   = "Get a deployment's status"
	deployStatusLong    = "Get a deployment's status"
	deployStatusExample = "armory deploy status [options]"
)

type formattableDeployStatus struct {
	currentTimestamp time.Time
	pipeline         *deploy.PipelineStatusResponse
	httpResponse     *http.Response
	err              error
}

func (ds formattableDeployStatus) Get() interface{} {
	return ds.pipeline
}

func (ds formattableDeployStatus) GetHttpResponse() *http.Response {
	return ds.httpResponse
}

func (ds formattableDeployStatus) GetFetchError() error {
	return ds.err
}

func (ds formattableDeployStatus) String() string {
	ret := ""
	currentTimestampIso8601 := ds.currentTimestamp.Format(time.RFC3339)

	ret += fmt.Sprintf("[%v] application: %s, started: %s\n", currentTimestampIso8601, ds.pipeline.Application, ds.pipeline.StartedAtIso8601)
	ret += fmt.Sprintf("[%v] status: ", currentTimestampIso8601)
	switch status := ds.pipeline.Status; status {
	case deploy.WorkflowStatusPaused:
		for _, stages := range ds.pipeline.Steps {
			if stages.Type == "pause" && stages.Status == deploy.WorkflowStatusPaused {
				ret += fmt.Sprintf("[%s] msg: Paused for %d %s. You can skip the pause in the CD-as-a-Service Console or CLI\n", status, stages.Pause.Duration, stages.Pause.Unit)
			}
		}
	case deploy.WorkflowStatusAwaitingApproval:
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
	cmd.Flags().StringVarP(&deploymentId, "deploymentId", "i", "", "(Required) The ID of an existing deployment.\n"+
		"You can find the deploymentId by navigating to the deployment status page and looking in the URL: \n"+
		"https://console.cloud.armory.io/deployments/pipeline/<deploymentId>")
	cmd.MarkFlagRequired("deploymentId")
	return cmd
}

func status(cmd *cobra.Command, configuration *config.Configuration, deploymentId string) error {
	if *configuration.GetIsTest() {
		utils.ConfigureLoggingForTesting(cmd)
	}

	storeCommandResult(cmd, DeployResultDeploymentID, deploymentId)

	deployClient := deployment.NewClient(configuration)
	ctx, cancel := context.WithTimeout(deployClient.ArmoryCloudClient.Context, time.Second*5)
	defer cancel()

	pipelineResp, response, err := deployClient.PipelineStatus(ctx, deploymentId)
	dataFormat, err := configuration.GetOutputFormatter()(formattableDeployStatus{
		currentTimestamp: configuration.Now(),
		pipeline:         pipelineResp,
		httpResponse:     response,
		err:              err,
	})

	// if we've made it this far, the command is valid. if an error occurs it isn't a usage error
	cmd.SilenceUsage = true
	if err != nil {
		return errorUtils.NewWrappedError(ErrDeploymentStatusResponseParse, err)
	}
	log.S().Info(dataFormat)
	return err
}
