package app

import (
	"context"
	"fmt"
	"github.com/armory/armory-cli/internal/deng"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/juju/ansiterm"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"time"
)

const (
	ParameterAccount  = "account"
	ParameterProvider = "provider"
)

func Execute(ctx context.Context, log *logrus.Logger, cmd *cobra.Command, client deng.DeploymentServiceClient, args []string) error {
	if len(args) == 0 {
		// Get all applications
		return getAllApplications(ctx, log, cmd, client)
	}
	return nil
}

func getAllApplications(ctx context.Context, log *logrus.Logger, cmd *cobra.Command, client deng.DeploymentServiceClient) error {
	envName, err := cmd.Flags().GetString(ParameterAccount)
	if err != nil {
		return err
	}
	envProvider, err := cmd.Flags().GetString(ParameterProvider)
	if err != nil {
		return err
	}

	r, err := client.GetApplications(ctx, &deng.GetAppRequest{
		Env: &deng.Environment{Provider: envProvider, Account: envName},
	})
	if err != nil {
		return err
	}

	w := os.Stdout
	_, _ = fmt.Fprintf(w, "\nAPPLICATIONS\n")
	wt := ansiterm.NewTabWriter(w, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprint(wt, "Name\tDeployments\tLast successful\tLast Failure\n")
	for _, c := range r.Apps {
		_, _ = fmt.Fprintf(wt, "%s\t%d\t%s\t%s\n", c.Name, c.Deployments, timeAsString(c.LastSuccessful), timeAsString(c.LastFailure))
	}
	_ = wt.Flush()
	return nil
}

func timeAsString(t *timestamp.Timestamp) string {
	tm := t.AsTime()
	if tm.IsZero() {
		return "-"
	}
	return tm.Local().Format(time.RFC822)
}
