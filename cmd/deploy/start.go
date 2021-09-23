package deploy

import (
	"errors"
	de "github.com/armory-io/deploy-engine/deploy/client"
	"github.com/armory/armory-cli/cmd"
	"github.com/sirupsen/logrus"
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
	}
	cmd.Flags().StringVarP(&options.deploymentFile, "file", "f", "", "path to the deployment file")
	cmd.MarkFlagRequired("file")
	return cmd
}

func start(cmd *cobra.Command, options *deployStartOptions, args []string) error {
	deployment := de.KubernetesV2StartKubernetesDeploymentRequest{}
	yamlFile, err := ioutil.ReadFile(options.deploymentFile)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yamlFile, &deployment)
	if err != nil {
		return err
	}
	req := options.DeployClient.DeploymentServiceApi.DeploymentServiceStartKubernetes(options.DeployClient.Context)
	req = req.Body(deployment)
	data, resp, err := req.Execute()
	if err != nil && resp.StatusCode >= 300 {
		openAPIErr := err.(de.GenericOpenAPIError)
		return errors.New(string(openAPIErr.Body()))
	}
	if err != nil {
		return err
	}
	res, err := options.Output.Formatter(data)
	if err != nil {
		return err
	}
	logrus.Info(res)
	return nil
}