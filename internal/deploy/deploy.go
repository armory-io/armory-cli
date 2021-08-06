package deploy

import (
	"context"
	"github.com/armory/armory-cli/internal/deng"
	"github.com/armory/armory-cli/internal/status"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

func Execute(ctx context.Context, log *logrus.Logger, cmd *cobra.Command, client deng.DeploymentServiceClient, args []string) error {
	p := newParser(cmd.Flags(), args, log)
	dep, err := p.parse()
	if err != nil {
		return err
	}
	desc, err := client.Start(ctx, dep)
	if err != nil {
		return err
	}
	w, err := cmd.Flags().GetBool(ParameterWait)
	if err != nil {
		return err
	}
	status.PrintStatus(os.Stdout, desc)
	if w {
		// Show a watch on status
		return status.ShowStatus(ctx, log, desc.Id, client, w, false)
	}
	return nil
}
