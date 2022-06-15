package deploy

import (
	"context"
	"fmt"
	de "github.com/armory-io/deploy-engine/api"
	"github.com/armory/armory-cli/pkg/cmdUtils"
	"github.com/armory/armory-cli/pkg/config"
	deployment "github.com/armory/armory-cli/pkg/deploy"
	"github.com/armory/armory-cli/pkg/model"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	_nethttp "net/http"
	"os"
	"time"
)

const (
	deployStartShort   = "Start deployment with Armory CD-as-a-Service"
	deployStartLong    = "Start deployment with Armory CD-as-a-Service"
	deployStartExample = "armory deploy start [options]"
)

type deployStartOptions struct {
	deploymentFile string
	application    string
	context        map[string]string
}

type FormattableDeployStartResponse struct {
	// The deployment's ID.
	DeploymentId string `json:"deploymentId,omitempty" yaml:"deploymentId,omitempty"`
	httpResponse *_nethttp.Response
	err          error
}

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
	cmd.Flags().StringVarP(&options.deploymentFile, "file", "f", "", "path to the deployment file")
	cmd.Flags().StringVarP(&options.application, "application", "n", "", "application name for deployment")
	cmd.Flags().StringToStringVar(&options.context, "add-context", map[string]string{}, "add context values to be used in strategy steps")
	cmd.MarkFlagRequired("file")
	return cmd
}

func start(cmd *cobra.Command, configuration *config.Configuration, options *deployStartOptions) error {
	payload := model.OrchestrationConfig{}
	//in case this is running on a github instance
	gitWorkspace, present := os.LookupEnv("GITHUB_WORKSPACE")
	_, isATest := os.LookupEnv("ARMORY_CLI_TEST")
	if present && !isATest {
		options.deploymentFile = gitWorkspace + options.deploymentFile
	}
	// read yaml file
	file, err := ioutil.ReadFile(options.deploymentFile)
	if err != nil {
		return fmt.Errorf("error trying to read the YAML file: %s", err)
	}
	cmd.SilenceUsage = true
	// unmarshall data into struct
	err = yaml.UnmarshalStrict(file, &payload)
	if err != nil {
		return fmt.Errorf("error invalid deployment object: %s", err)
	}
	applicationOpt := options.application
	var application string
	if len(applicationOpt) > 0 {
		application = applicationOpt
	} else {
		application = payload.Application
	}

	if len(application) < 1 {
		return fmt.Errorf("application name must be defined in deployment file or by application opt")
	}

	dep, err := deployment.CreateDeploymentRequest(application, &payload, options.context)
	if err != nil {
		return fmt.Errorf("error converting deployment object: %s", err)
	}

	deployClient := configuration.GetDeployEngineClient()

	ctx, cancel := context.WithTimeout(deployClient.Context, time.Minute)
	defer cancel()
	// execute request
	raw, response, err := deployClient.StartPipeline(ctx, dep)
	// create response object
	deploy := newDeployStartResponse(raw, response, err)
	// format response
	dataFormat, err := configuration.GetOutputFormatter()(deploy)
	if err != nil {
		return err
	}
	cmd.SetContext(context.WithValue(ctx, "deploymentId", deploy.DeploymentId))
	_, err = fmt.Fprintln(cmd.OutOrStdout(), dataFormat)
	return err
}
