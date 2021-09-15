package deploy

import (
	"github.com/armory/armory-cli/cmd"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	deployStartShort   = "Start deployment with Armory Cloud"
	deployStartLong    = "Start deployment with Armory Cloud"
	deployStartExample  = "armory deploy start [options]"
)

func NewDeployStartCmd(deployOptions *cmd.RootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "start",
		Aliases: []string{"start"},
		Short:   deployStartShort,
		Long:    deployStartLong,
		Example: deployStartExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return start(cmd,deployOptions, args)
		},
	}
	return cmd
}

func start(cmd *cobra.Command, options *cmd.RootOptions, args []string) error {
	logrus.Fatalf("Not implemented")
	return nil
}

