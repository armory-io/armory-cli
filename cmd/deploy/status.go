package deploy

import (
	"github.com/armory/armory-cli/cmd"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	deployStatusShort   = "Watch deployment on Armory Cloud"
	deployStatusLong    = "Watch deployment on Armory Cloud"
	deployStatusExample  = "armory deploy status [options]"
)

func NewDeployStatusCmd(deployOptions *cmd.RootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "status",
		Aliases: []string{"status"},
		Short:   deployStatusShort,
		Long:    deployStatusLong,
		Example: deployStatusExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return status(cmd,deployOptions, args)
		},
	}
	return cmd
}

func status(cmd *cobra.Command, options *cmd.RootOptions, args []string) error {
	t, _ :=options.Auth.GetToken()
	logrus.Info(t)
	logrus.Info("Not implemented")
	return nil
}
