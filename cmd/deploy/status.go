package deploy

import (
	"context"
	"fmt"
	deploy "github.com/armory-io/deploy-engine/deploy/client"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"time"
	_nethttp "net/http"
)

const (
	deployStatusShort   = "Watch deployment on Armory Cloud"
	deployStatusLong    = "Watch deployment on Armory Cloud"
	deployStatusExample = "armory deploy status [options]"
)

type deployStatusOptions struct {
	*deployOptions
}

type FormattableDeployStatus struct {
	DeployResp deploy.DeploymentV2DeploymentStatusResponse `json:"deployment"`
	httpResponse *_nethttp.Response
	err error
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

func newDeployStatusResponseWrapper(raw deploy.DeploymentV2DeploymentStatusResponse, response *_nethttp.Response, err error) FormattableDeployStatus {
	wrapper := FormattableDeployStatus{
		DeployResp: raw,
		httpResponse: response,
		err: err,
	}
	return wrapper
}

func (u FormattableDeployStatus) String() string {
	ret := ""
	if u.err != nil {
		logrus.Error(u.err)
		logrus.Fatalf("Error getting deployment status")
	}

	now := time.Now().Format(time.RFC3339)
	ret += fmt.Sprintf("[%v] application: %s, started: %s\n", now, u.DeployResp.GetApplication(), u.DeployResp.GetStartedAtIso8601())
	ret += fmt.Sprintf("[%v] status: ", now)
	switch status := u.DeployResp.GetStatus(); status {
	case deploy.DEPLOYMENT_PAUSED:
		end := u.DeployResp.Kubernetes.Canary.PauseInfo.GetEndTimeIso8601()
		reason := u.DeployResp.Kubernetes.Canary.PauseInfo.GetReason()
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
	ctx, cancel := context.WithTimeout(options.DeployClient.Context, time.Second * 5)
	defer cancel()
	req := options.DeployClient.DeploymentServiceApi.DeploymentServiceStatus(ctx, options.deploymentId)
	deployResp, response, err := req.Execute()
	dataFormat, err := options.Output.Formatter(newDeployStatusResponseWrapper(deployResp, response, err))
	if err != nil {
		return fmt.Errorf("error trying to parse respone: %s", err)
	}
	_, err = fmt.Fprintln(cmd.OutOrStdout(), dataFormat)
	return err
}
