package create

import (
	"github.com/armory/armory-cli/pkg/cmdUtils"
	"github.com/armory/armory-cli/pkg/config"
	"github.com/spf13/cobra"
)

const (
	createShort = "Create a resource"
	createLong  = `
		Create a resource.

		`
	createExample = `

		`
)

func NewCmdCreate(configuration *config.Configuration) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create",
		Aliases: []string{},
		Short:   createShort,
		Long:    createLong,
		Example: createExample,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cmdUtils.ExecuteParentHooks(cmd, args)
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {},
		Hidden:            true,
	}

	// create subcommands
	cmd.AddCommand(NewCmdCreateAgent(configuration))

	cmdUtils.SetPersistentFlagsFromEnvVariables(cmd.Commands())

	return cmd
}
