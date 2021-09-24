package deploy

import (
	"fmt"
	de "github.com/armory-io/deploy-engine/deploy/client"
	"github.com/armory/armory-cli/cmd"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

const (
	deployStartShort   = "Start deployment with Armory Cloud"
	deployStartLong    = "Start deployment with Armory Cloud"
	deployStartExample = "armory deploy start [options]"
)

type deployStartOptions struct {
	*cmd.RootOptions
	deploymentFile string
	deploymentId string
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
	return fmt.Sprintf("Deployment ID: %s", u.DeploymentId)
}

func NewDeployStartCmd(deployOptions *cmd.RootOptions) *cobra.Command {
	options := &deployStartOptions{
		RootOptions: deployOptions,
	}
	cmd := &cobra.Command{
		Use:     "start",
		Aliases: []string{"start"},
		Short:   deployStartShort,
		Long:    deployStartLong,
		Example: deployStartExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return start(cmd, options, args)
		},
		PostRun: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(cmd.OutOrStdout(), "Status Available at:")
			fmt.Fprintf(cmd.OutOrStdout(), "https://console.cloud.armory.io/deploy/deployment/%s", options.deploymentId)
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
	dataFormat, err := options.Output.Formatter(deploy)
	if err != nil {
		return fmt.Errorf("error trying to parse respone: %s", err)
	}
	options.deploymentId = deploy.DeploymentId
	fmt.Fprintln(cmd.OutOrStdout(),"Deployment successfully launch.")
	fmt.Fprintln(cmd.OutOrStdout(), dataFormat)
	return nil
}