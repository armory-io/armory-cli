package config

import (
	"github.com/armory/armory-cli/pkg/cmdUtils"
	"github.com/armory/armory-cli/pkg/config"
	"github.com/spf13/cobra"
)

const (
	configShort   = ""
	configLong    = ""
	configExample = ""
)

func NewConfigCmd(configuration *config.Configuration) *cobra.Command {
	command := &cobra.Command{
		Use:     "config",
		Aliases: []string{},
		Short:   configShort,
		Long:    configLong,
		Example: configExample,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cmdUtils.ExecuteParentHooks(cmd, args)
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {},
		Hidden:            true,
	}
	// create subcommands
	command.AddCommand(NewConfigApplyCmd(configuration))
	command.AddCommand(NewConfigGetCmd(configuration))

	cmdUtils.SetPersistentFlagsFromEnvVariables(command.Commands())

	return command
}
