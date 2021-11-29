package deploy

import (
	"context"
	"fmt"
	de "github.com/armory-io/deploy-engine/pkg"
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
	deployStartShort   = "Start deployment with Armory Cloud"
	deployStartLong    = "Start deployment with Armory Cloud"
	deployStartExample = "armory deploy start [options]"
)

type deployStartOptions struct {
	*deployOptions
	deploymentFile string
}

type FormattableDeployStartResponse struct {
	// The deployment's ID.
	DeploymentId string `json:"deploymentId,omitempty" yaml:"deploymentId,omitempty"`
	httpResponse *_nethttp.Response
	err error
}

func newDeployStartResponse(raw *de.PipelineStartPipelineResponse, response *_nethttp.Response, err error) FormattableDeployStartResponse {
	deployment := FormattableDeployStartResponse{
		DeploymentId: raw.GetPipelineId(),
		httpResponse: response,
		err: err,
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

func NewDeployStartCmd(deployOptions *deployOptions) *cobra.Command {
	options := &deployStartOptions{
		deployOptions: deployOptions,
	}
	cmd := &cobra.Command{
		Use:     "start --file [<path to file>]",
		Aliases: []string{"start"},
		Short:   deployStartShort,
		Long:    deployStartLong,
		Example: deployStartExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return start(cmd, options, args)
		},
	}
	cmd.Flags().StringVarP(&options.deploymentFile, "file", "f", "", "path to the deployment file")
	cmd.MarkFlagRequired("file")
	return cmd
}

func start(cmd *cobra.Command, options *deployStartOptions, args []string) error {
	payload := model.OrchestrationConfig{}
	//in case this is running on a github instance
	gitWorkspace, present := os.LookupEnv("GITHUB_WORKSPACE")
	_, isATest := os.LookupEnv("ARMORY_CLI_TEST")
	if present && !isATest{
		options.deploymentFile = gitWorkspace + options.deploymentFile
	}
	// read yaml file
	file, err := ioutil.ReadFile(options.deploymentFile)
	if err != nil {
		return fmt.Errorf("error trying to read the YAML file: %s", err)
	}
	cmd.SilenceUsage = true
	// unmarshall data into struct
	err = yaml.Unmarshal(file, &payload)
	if err != nil {
		return fmt.Errorf("error invalid deployment object: %s", err)
	}
	dep, err := deployment.CreateDeploymentRequest(&payload)
	if err != nil {
		return fmt.Errorf("error converting deployment object: %s", err)
	}

	ctx, cancel := context.WithTimeout(options.DeployClient.Context, time.Second * 5)
	defer cancel()
	// prepare request
	request := options.DeployClient.DeploymentServiceApi.
		DeploymentServiceStartKubernetesPipeline(ctx).Body(*dep)
	// execute request
	raw, response, err := request.Execute()
	// create response object
	deploy := newDeployStartResponse(&raw, response, err)
	// format response
	dataFormat, err := options.Output.Formatter(deploy)
	if err != nil {
		return fmt.Errorf("error trying to parse response: %s", err)
	}
	options.deploymentId = deploy.DeploymentId
	_, err = fmt.Fprintln(cmd.OutOrStdout(), dataFormat)
	return err
}