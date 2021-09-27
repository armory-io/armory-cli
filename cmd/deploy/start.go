package deploy

import (
	"fmt"
	de "github.com/armory-io/deploy-engine/deploy/client"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"io/ioutil"
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

type deployStartResponse struct {
	// The deployment's ID.
	DeploymentId string `json:"deploymentId,omitempty" yaml:"deploymentId,omitempty"`
}

func newDeployStartResponse(raw *de.DeploymentV2StartDeploymentResponse) deployStartResponse {
	deployment := deployStartResponse{
		raw.GetDeploymentId(),
	}
	return deployment
}

func (u deployStartResponse) String() string {
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
	payload := de.KubernetesV2StartKubernetesDeploymentRequest{}
	// read yaml file
	file, err := ioutil.ReadFile(options.deploymentFile)
	if err != nil {
		return fmt.Errorf("error trying to read the yaml file: %s", err)
	}
	// unmarshall data into struct
	err = yaml.Unmarshal(file, &payload)
	if err != nil {
		return fmt.Errorf("error invalid deployment object: %s", err)
	}
	// prepare request
	request := options.DeployClient.DeploymentServiceApi.
		DeploymentServiceStartKubernetes(options.DeployClient.Context).Body(payload)
	// execute request
	raw, response, err := request.Execute()
	if err != nil && response.StatusCode >= 300 {
		openAPIErr := err.(de.GenericOpenAPIError)
		return fmt.Errorf("deployment returns an error: status code(%d) %s",
			response.StatusCode, string(openAPIErr.Body()))
	}
	if err != nil {
		return fmt.Errorf("invalid request: %s", err)
	}
	// create response object
	deploy := newDeployStartResponse(&raw)
	// format response
	dataFormat, err := options.Output.Formatter(deploy, nil)
	if err != nil {
		return fmt.Errorf("error trying to parse respone: %s", err)
	}
	options.deploymentId = deploy.DeploymentId
	fmt.Fprintln(cmd.OutOrStdout(), dataFormat)
	return nil
}