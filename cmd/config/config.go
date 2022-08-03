package config

import (
	"github.com/armory/armory-cli/pkg/cmdUtils"
	"github.com/spf13/cobra"
)

const (
	configShort   = ""
	configLong    = ""
	configExample = ""
)

func NewConfigCmd() *cobra.Command {
	command := &cobra.Command{
		Use:     "config",
		Aliases: []string{"config"},
		Short:   configShort,
		Long:    configLong,
		Example: configExample,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cmdUtils.ExecuteParentHooks(cmd, args)
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {},
	}
	// create subcommands
	command.AddCommand(NewConfigApplyCmd())

	cmdUtils.SetPersistentFlagsFromEnvVariables(command.Commands())

	return command
}
