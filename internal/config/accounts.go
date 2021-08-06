package config

import (
	"context"
	"fmt"
	"github.com/armory/armory-cli/internal/deng"
	"github.com/juju/ansiterm"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

const ParamProvider = "provider"

func GetAccounts(ctx context.Context, log *logrus.Logger, cmd *cobra.Command, client deng.DeploymentServiceClient, args []string) error {
	p, err := cmd.Flags().GetString(ParamProvider)
	if err != nil {
		return err
	}

	res, err := client.GetAccounts(ctx, &deng.GetAccountRequest{Provider: p})
	if err != nil {
		return err
	}

	fmt.Printf("Found %d accounts\n", len(res.Accounts))
	if len(res.Accounts) == 0 {
		return nil
	}

	wt := ansiterm.NewTabWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprint(wt, "Provider\tName\n")
	for _, a := range res.Accounts {
		_, _ = fmt.Fprintf(wt, "%s\t%s\n", p, a.Account)
	}
	_ = wt.Flush()
	return nil
}
