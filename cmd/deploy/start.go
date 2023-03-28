package deploy

import (
	"context"
	"errors"
	"fmt"
	de "github.com/armory-io/deploy-engine/pkg/api"
	"github.com/armory/armory-cli/cmd/utils"
	"github.com/armory/armory-cli/pkg/armoryCloud"
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
	nethttp "net/http"
	"os"
	"time"
)

const (
	deployStartShort = "Start a deployment"
	deployStartLong  = "Start a deployment\n\n" +
		"For deployment configuration YAML documentation, visit https://docs.armory.io/cd-as-a-service/reference/ref-deployment-file"
	deployStartExample         = "armory deploy start [options]"
	armoryConfigLocationHeader = "X-Armory-Config-Location"
	mediaTypePipelineV2        = "application/vnd.start.kubernetes.pipeline.v2+json"
	mediaTypePipelineV2Link    = "application/vnd.start.kubernetes.pipeline.v2.link+json"
	mediaTypePipelineRedeploy  = "application/vnd.armory.pipeline-redeploy+json"
)

type deployStartOptions struct {
	account           string
	deploymentFile    string
	pipelineId        string
	application       string
	context           map[string]string
	waitForCompletion bool
}

type WithDeployConfiguration func(cmd *cobra.Command, options *deployStartOptions, deployClient ArmoryDeployClient) (*de.StartPipelineResponse, *nethttp.Response, error)

type FormattableDeployStartResponse struct {
	// The deployment's ID.
	DeploymentId    string `json:"deploymentId,omitempty" yaml:"deploymentId,omitempty"`
	ExecutionStatus string `json:"status,omitempty" yaml:"status,omitempty"`
	httpResponse    *nethttp.Response
	err             error
}

type ArmoryDeployClient interface {
	PipelineStatus(ctx context.Context, pipelineID string) (*de.PipelineStatusResponse, *nethttp.Response, error)
	DeploymentStatus(ctx context.Context, deploymentID string) (*de.DeploymentStatusResponse, *nethttp.Response, error)
	StartPipeline(ctx context.Context, options deployment.StartPipelineOptions) (*de.StartPipelineResponse, *nethttp.Response, error)
	GetArmoryCloudClient() *armoryCloud.Client
}

const (
	DeployResultDeploymentID = "DEPLOYMENT_ID"
	DeployResultLink         = "LINK"
	DeployResultSyncStatus   = "RUN_RESULT"
)

var statusCheckTick = time.Second * 10

func newDeployStartResponse(raw *de.StartPipelineResponse, response *nethttp.Response, err error) FormattableDeployStartResponse {
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

func (u FormattableDeployStartResponse) GetHttpResponse() *nethttp.Response {
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
	cmd.Flags().StringVarP(&options.account, "account", "", "", "override the deployment YAML account field for each target when --file is a URL")
	cmd.Flags().StringVarP(&options.deploymentFile, "file", "f", "", "path to the deployment file")
	cmd.Flags().StringVarP(&options.pipelineId, "pipelineId", "i", "", "the ID of a previously deployed pipeline. Request will automatically use the original deployment configuration for that pipeline including its manifests")
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
	deployClient := deployment.GetDeployClient(configuration)
	var startResp *de.StartPipelineResponse
	var rawResp *nethttp.Response
	var err error

	//TODO - Can we use cue to easily validate that deploymentFile and pipelineId are not both provided?
	if options.deploymentFile != "" && options.pipelineId != "" {
		return ErrTwoDeploymentConfigurationsSpecified
	}

	var withConfiguration WithDeployConfiguration
	if options.pipelineId != "" {
		withConfiguration = WithPipelineId
	} else if deployment.IsURL(options.deploymentFile) {
		withConfiguration = WithURL
	} else {
		withConfiguration = WithLocalFile
	}
	startResp, rawResp, err = withConfiguration(cmd, options, deployClient)

	if err != nil && errors.Is(err, ErrYAMLFileRead) {
		return err
	}
	// create response object
	deploy := newDeployStartResponse(startResp, rawResp, err)
	storeCommandResult(cmd, DeployResultDeploymentID, deploy.DeploymentId)

	if options.waitForCompletion && err == nil {
		beginTrackingDeployment(cmd, configuration, &deploy, deployClient)
	}
	// format response
	return outputCommandResult(deploy, configuration)
}

func WithURL(cmd *cobra.Command, options *deployStartOptions, deployClient ArmoryDeployClient) (*de.StartPipelineResponse, *nethttp.Response, error) {
	if options.application != "" {
		return nil, nil, ErrApplicationNameOverrideNotSupported
	}
	cmd.SilenceUsage = true
	ctx, cancel := context.WithTimeout(deployClient.GetArmoryCloudClient().Context, time.Minute)
	defer cancel()
	// execute request
	raw, response, err := deployClient.StartPipeline(ctx, deployment.StartPipelineOptions{
		ApplicationNameOverride: options.application,
		Context:                 options.context,
		Headers: map[string]string{
			"Content-Type":             mediaTypePipelineV2Link,
			"Accept":                   mediaTypePipelineV2,
			armoryConfigLocationHeader: options.deploymentFile,
		},
		UnstructuredDeployment: map[string]any{
			"account": options.account,
		},
		IsURL: true,
	})
	return raw, response, err
}

func WithPipelineId(cmd *cobra.Command, options *deployStartOptions, deployClient ArmoryDeployClient) (*de.StartPipelineResponse, *nethttp.Response, error) {
	//TODO cue validation instead?
	if options.account != "" {
		return nil, nil, ErrAccountNameOverrideNotSupported
	}
	if options.application != "" {
		return nil, nil, ErrApplicationNameOverrideNotSupported
	}
	if options.deploymentFile != "" {
		return nil, nil, ErrTwoDeploymentConfigurationsSpecified
	}
	cmd.SilenceUsage = true
	ctx, cancel := context.WithTimeout(deployClient.GetArmoryCloudClient().Context, time.Minute)
	defer cancel()
	// execute request
	raw, response, err := deployClient.StartPipeline(ctx, deployment.StartPipelineOptions{
		ApplicationNameOverride: options.application,
		Context:                 options.context,
		Headers: map[string]string{
			"Content-Type":             mediaTypePipelineRedeploy,
			"Accept":                   mediaTypePipelineV2,
			armoryConfigLocationHeader: options.pipelineId,
		},
		//TODO - CDAAS-2230 add ability to specify specific targets
		//UnstructuredDeployment: map[string]any{
		//	"targets": [...],
		//},
		IsURL: true,
	})
	return raw, response, err
}

func WithLocalFile(cmd *cobra.Command, options *deployStartOptions, deployClient ArmoryDeployClient) (*de.StartPipelineResponse, *nethttp.Response, error) {
	//in case this is running on a github instance
	gitWorkspace, present := os.LookupEnv("GITHUB_WORKSPACE")
	_, isATest := os.LookupEnv("ARMORY_CLI_TEST")
	if present && !isATest {
		options.deploymentFile = gitWorkspace + options.deploymentFile
	}
	// read yaml file
	file, err := ioutil.ReadFile(options.deploymentFile)
	if err != nil {
		return nil, nil, errorUtils.NewWrappedError(ErrYAMLFileRead, err)
	}
	cmd.SilenceUsage = true
	// unmarshall data into struct
	var payload map[string]any
	if err = yaml.Unmarshal(file, &payload); err != nil {
		return nil, nil, errorUtils.NewWrappedError(ErrInvalidDeploymentObject, err)
	}

	ctx, cancel := context.WithTimeout(deployClient.GetArmoryCloudClient().Context, time.Minute)
	defer cancel()
	// execute request
	raw, response, err := deployClient.StartPipeline(ctx, deployment.StartPipelineOptions{
		UnstructuredDeployment:  payload,
		ApplicationNameOverride: options.application,
		Context:                 options.context,
		Headers: map[string]string{
			"Content-Type": mediaTypePipelineV2,
			"Accept":       mediaTypePipelineV2,
		},
	})
	return raw, response, err
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
