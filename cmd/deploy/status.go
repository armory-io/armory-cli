package deploy

import (
	"fmt"
	deploy "github.com/armory-io/deploy-engine/deploy/client"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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
	cmd.Flags().StringVarP(&options.deploymentId, "deploymentId", "i", "", "The id of an existing deployment (required)")
	cmd.MarkFlagRequired("deploymentId")
	return cmd
}

func status(cmd *cobra.Command, options *deployStatusOptions ) error {
	req := options.DeployClient.DeploymentServiceApi.DeploymentServiceStatus(options.DeployClient.Context, options.deploymentId)
	deployResp, response, err := req.Execute()
	if err != nil && response.StatusCode >= 300 {
		openAPIErr := err.(deploy.GenericOpenAPIError)
		return fmt.Errorf("deployment returns an error: status code(%d) %s",
			response.StatusCode, string(openAPIErr.Body()))
	}
	var ret string
	if options.O != "" {
		ret, err = options.Output.Formatter(deployResp, err)
	} else {
		ret = printPlain(deployResp, err)
	}
	fmt.Fprintln(cmd.OutOrStdout(), ret)
	return nil
}

func printPlain(deployResp deploy.DeploymentV2DeploymentStatusResponse, err error) string {
	ret := ""
	if err != nil {
		logrus.Error(err)
		logrus.Fatalf("Error getting deployment status")
	}

	now := time.Now().Format(time.RFC3339)
	ret += fmt.Sprintf("[%v] application: %s, started: %s\n", now, deployResp.GetApplication(), deployResp.GetStartedAtIso8601())
	ret += fmt.Sprintf("[%v] status: ", now)
	switch status := deployResp.GetStatus(); status {
	case deploy.DEPLOYMENT_PAUSED:
		end := deployResp.Kubernetes.Canary.PauseInfo.GetEndTimeIso8601()
		reason := deployResp.Kubernetes.Canary.PauseInfo.GetReason()
		if reason == "" {
			reason = "unspecified"
		}
		ret += fmt.Sprintf("[%s] msg: Paused until %s for reason: %s. You may resume immediately in the cloud console or CLI\n", status, end, reason)
	case deploy.DEPLOYMENT_AWAITING_APPROVAL:
		ret += fmt.Sprintf("[%s] msg: Paused for Manual Judgment. You may resume immediately in the cloud console or CLI.\n", status)
	default:
		ret += string(status) + "\n"
	}
	return ret
}
