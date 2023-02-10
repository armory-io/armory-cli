package deploy

import (
	"context"
	"errors"
	"fmt"
	de "github.com/armory-io/deploy-engine/pkg/api"
	"github.com/armory/armory-cli/cmd/utils"
	"github.com/armory/armory-cli/pkg/cmdUtils"
	"github.com/armory/armory-cli/pkg/config"
	deployment "github.com/armory/armory-cli/pkg/deploy"
	errorUtils "github.com/armory/armory-cli/pkg/errors"
	"github.com/armory/armory-cli/pkg/output"
	"github.com/spf13/cobra"
	log "go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"io"
	"io/ioutil"
	_nethttp "net/http"
	"os"
	"time"
)

const (
	deployStartShort = "Start a deployment"
	deployStartLong  = "Start a deployment\n\n" +
		"For deployment configuration YAML documentation, visit https://docs.armory.io/cd-as-a-service/reference/ref-deployment-file"
	deployStartExample = "armory deploy start [options]"
)

type deployStartOptions struct {
	dryRun         bool
	deploymentFile    string
	application       string
	context           map[string]string
	waitForCompletion bool
}

type FormattableDeployStartResponse struct {
	// The deployment's ID.
	DeploymentId    string `json:"deploymentId,omitempty" yaml:"deploymentId,omitempty"`
	ExecutionStatus string `json:"status,omitempty" yaml:"status,omitempty"`
	httpResponse    *_nethttp.Response
	err             error
}

const (
	DeployResultDeploymentID = "DEPLOYMENT_ID"
	DeployResultLink         = "LINK"
	DeployResultSyncStatus   = "RUN_RESULT"
)

var statusCheckTick = time.Second * 10

func newDeployStartResponse(raw *de.StartPipelineResponse, response *_nethttp.Response, err error) FormattableDeployStartResponse {
	var pipelineID string
	if raw != nil {
		pipelineID = raw.PipelineID
	}

	deployment := FormattableDeployStartResponse{
		DeploymentId: pipelineID,
		httpResponse: response,
		err:          err,
	}
	return deployment
}

func (u FormattableDeployStartResponse) Get() interface{} {
	return u
}

func (u FormattableDeployStartResponse) GetHttpResponse() *_nethttp.Response {
	return u.httpResponse
}

func (u FormattableDeployStartResponse) GetFetchError() error {
	return u.err
}

func (u FormattableDeployStartResponse) String() string {
	return fmt.Sprintf("[%v] Deployment ID: %s", time.Now().Format(time.RFC3339), u.DeploymentId)
}

func NewDeployStartCmd(configuration *config.Configuration) *cobra.Command {
	options := &deployStartOptions{}
	cmd := &cobra.Command{
		Use:     "start --file [<path to file>]",
		Aliases: []string{"start"},
		Short:   deployStartShort,
		Long:    deployStartLong,
		Example: deployStartExample,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cmdUtils.ExecuteParentHooks(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return start(cmd, configuration, options)
		},
	}
	cmd.Flags().BoolVarP(&options.dryRun, "dryRun", "d", false, "output the json of the deployment request without submitting it")
	cmd.Flags().StringVarP(&options.deploymentFile, "file", "f", "", "path to the deployment file")
	cmd.Flags().StringVarP(&options.application, "application", "n", "", "application name for deployment")
	cmd.Flags().StringToStringVar(&options.context, "add-context", map[string]string{}, "add context values to be used in strategy steps")
	cmd.Flags().BoolVarP(&options.waitForCompletion, "watch", "w", false, "wait for deployment to complete")
	cmd.MarkFlagRequired("file")
	return cmd
}

func start(cmd *cobra.Command, configuration *config.Configuration, options *deployStartOptions) error {
	if *configuration.GetIsTest() {
		utils.ConfigureLoggingForTesting(cmd)
	}
	//in case this is running on a github instance
	gitWorkspace, present := os.LookupEnv("GITHUB_WORKSPACE")
	_, isATest := os.LookupEnv("ARMORY_CLI_TEST")
	if present && !isATest {
		options.deploymentFile = gitWorkspace + options.deploymentFile
	}
	// read yaml file
	file, err := ioutil.ReadFile(options.deploymentFile)
	if err != nil {
		return errorUtils.NewWrappedError(ErrYamlFileRead, err)
	}
	cmd.SilenceUsage = true
	// unmarshall data into struct
	var payload map[string]any
	if err = yaml.Unmarshal(file, &payload); err != nil {
		return errorUtils.NewWrappedError(ErrInvalidDeploymentObject, err)
	}

	deployClient := deployment.GetDeployClient(configuration)

	ctx, cancel := context.WithTimeout(deployClient.ArmoryCloudClient.Context, time.Minute)
	defer cancel()
	// execute request
	raw, response, err := deployClient.StartPipeline(ctx, deployment.StartPipelineOptions{
		UnstructuredDeployment:  payload,
		ApplicationNameOverride: options.application,
		ContextOverrides:        options.context,
	}, options.dryRun)

	if err != nil && errors.Is(deployment.ErrAbortOnDryRun, err) {
		cmd.SetContext(context.WithValue(ctx, "dryRun", true))
		return nil
	}
	// create response object
	deploy := newDeployStartResponse(raw, response, err)
	storeCommandResult(cmd, DeployResultDeploymentID, deploy.DeploymentId)

	if options.waitForCompletion && err == nil {
		beginTrackingDeployment(cmd, configuration, &deploy, deployClient)
	}
	// format response
	return outputCommandResult(deploy, configuration)
}

func beginTrackingDeployment(cmd *cobra.Command, configuration *config.Configuration, deploy *FormattableDeployStartResponse, deployClient *deployment.DeployClient) {
	canWriteProgress := configuration.GetOutputType() == output.Text
	if canWriteProgress {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "[%v] Waiting for deployment to complete. Status UI: %s\n", time.Now().Format(time.RFC3339), buildMonitoringUrl(configuration, deploy.DeploymentId))
	}
	var (
		status         de.WorkflowStatus
		reportedStatus string
		err            error
	)

	if status, err = waitForCompletion(deployClient, deploy.DeploymentId, canWriteProgress, cmd.OutOrStdout()); err != nil {
		reportedStatus = de.WorkflowStatusUnknown + " (error)"
	} else {
		reportedStatus = string(status)
	}
	if canWriteProgress {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "[%v] Deployment %s completed with status: %s\n", time.Now().Format(time.RFC3339), deploy.DeploymentId, reportedStatus)
	}

	deploy.ExecutionStatus = reportedStatus
	storeCommandResult(cmd, DeployResultSyncStatus, reportedStatus)
}

func waitForCompletion(deployClient *deployment.DeployClient, pipelineID string, canWriteProgress bool, out io.Writer) (de.WorkflowStatus, error) {
	var lastStatus de.WorkflowStatus
	for range time.Tick(statusCheckTick) {
		if canWriteProgress {
			_, _ = fmt.Fprintf(out, ".")
		}

		status, err := queryStatus(deployClient, pipelineID)
		if err != nil {
			return de.WorkflowStatusUnknown, err
		}

		if lastStatus != status && canWriteProgress {
			_, _ = fmt.Fprintf(out, "\n[%v] Deployment status changed: %s\n", time.Now().Format(time.RFC3339), status)
		}

		lastStatus = status
		if isDeploymentInFinalState(status) {
			break
		}
	}
	return lastStatus, nil
}

func queryStatus(deployClient *deployment.DeployClient, pipelineID string) (de.WorkflowStatus, error) {
	ctx, cancel := context.WithTimeout(deployClient.ArmoryCloudClient.Context, time.Minute)
	defer cancel()
	status, _, err := deployClient.PipelineStatus(ctx, pipelineID)
	if err != nil {
		return de.WorkflowStatusUnknown, err
	}
	return status.Status, nil
}

func isDeploymentInFinalState(status de.WorkflowStatus) bool {
	switch status {
	case de.WorkflowStatusFailed, de.WorkflowStatusSucceeded, de.WorkflowStatusCancelled:
		return true
	}
	return false
}

func outputCommandResult(deploy FormattableDeployStartResponse, configuration *config.Configuration) error {
	if dataFormat, err := configuration.GetOutputFormatter()(deploy); err == nil {
		log.S().Info(dataFormat)
		return nil
	} else {
		return err
	}
}
