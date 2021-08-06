package deploy

import (
	"context"
	"github.com/armory/armory-cli/internal/deng/protobuff"
	"github.com/armory/armory-cli/internal/status"
	"github.com/spf13/cobra"
	"os"
)

func Execute(ctx context.Context, cmd *cobra.Command, client protobuff.DeploymentServiceClient, args []string) error {
	p := newParser(cmd.Flags(), args)
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
		return status.ShowStatus(ctx, desc.Id, client, w, false)
	}
	return nil
}
