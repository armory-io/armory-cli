package rollout

import (
	"context"
	"errors"
	"fmt"
	"github.com/armory/armory-cli/internal/deng"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

const (
	ParamName = "name"
	ParamType = "type"
)

func Resume(ctx context.Context, log *logrus.Logger, cmd *cobra.Command, client deng.DeploymentServiceClient, args []string) error {
	return performRolloutOperation(ctx, cmd, args, client.Resume)
}

func Restart(ctx context.Context, log *logrus.Logger, cmd *cobra.Command, client deng.DeploymentServiceClient, args []string) error {
	return performRolloutOperation(ctx, cmd, args, client.Restart)
}

func Abort(ctx context.Context, log *logrus.Logger, cmd *cobra.Command, client deng.DeploymentServiceClient, args []string) error {
	return performRolloutOperation(ctx, cmd, args, client.Abort)
}

func performRolloutOperation(ctx context.Context, cmd *cobra.Command, args []string, call func(ctx context.Context, in *deng.RolloutRequest, opts ...grpc.CallOption) (*deng.RolloutResponse, error)) error {
	if len(args) == 0 {
		return errors.New("please provide deployment ID")
	}

	n, err := cmd.Flags().GetString(ParamName)
	if err != nil {
		return err
	}

	t, err := cmd.Flags().GetString(ParamType)
	if err != nil {
		return err
	}

	var req *deng.RolloutRequest
	depId := args[0]
	if n == "" {
		req = &deng.RolloutRequest{All: true, DeploymentId: depId}
	} else {
		req = &deng.RolloutRequest{Name: n, Type: t, DeploymentId: depId}
	}

	res, err := call(ctx, req)
	if err != nil {
		return err
	}
	fmt.Println(res.Message)
	return nil
}
