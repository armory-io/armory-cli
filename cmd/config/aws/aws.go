package aws

import (
	"github.com/armory/armory-cli/pkg/cmdUtils"
	cliconfig "github.com/armory/armory-cli/pkg/config"
	"github.com/spf13/cobra"
)

const (
	configAWSShort = "Manage CD-as-a-Service AWS Access"
	configAWSLong  = "Manage CD-as-a-Service AWS Access\n\n" +
		"For usage documentation, visit TODO"
	configAWSExample = "armory aws <subcommand>"
)

func NewAWSCmd(configuration *cliconfig.Configuration) *cobra.Command {
	command := &cobra.Command{
		Use:     "aws",
		Aliases: []string{},
		Short:   configAWSShort,
		Long:    configAWSLong,
		Example: configAWSExample,
		GroupID: "admin",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cmdUtils.ExecuteParentHooks(cmd, args)
		},
	}
	// create subcommands
	command.AddCommand(NewCreateRoleCmd(configuration))

	cmdUtils.SetPersistentFlagsFromEnvVariables(command.Commands())

	return command
}
