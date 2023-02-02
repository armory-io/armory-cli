package config

import (
	"github.com/armory/armory-cli/pkg/cmdUtils"
	cliconfig "github.com/armory/armory-cli/pkg/config"
	"github.com/spf13/cobra"
)

const (
	configShort = "Manage your RBAC configuration"
	configLong  = "Manage your RBAC configuration\n\n" +
		"For usage documentation, visit https://docs.armory.io/cd-as-a-service/concepts/iam/rbac"
	configExample = ""
)

func NewConfigCmd(configuration *cliconfig.Configuration) *cobra.Command {
	command := &cobra.Command{
		Use:     "config",
		Aliases: []string{},
		Short:   configShort,
		Long:    configLong,
		Example: configExample,
		GroupID: "admin",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cmdUtils.ExecuteParentHooks(cmd, args)
		},
	}
	// create subcommands
	command.AddCommand(NewConfigApplyCmd(configuration))
	command.AddCommand(NewConfigGetCmd(configuration))

	cmdUtils.SetPersistentFlagsFromEnvVariables(command.Commands())

	return command
}
