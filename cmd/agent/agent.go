package agent

import (
	"github.com/armory/armory-cli/pkg/cmdUtils"
	"github.com/armory/armory-cli/pkg/config"
	"github.com/spf13/cobra"
)

const (
	createShort = "Create a resource"
	createLong  = "Create a resource."
)

func NewCmdAgent(configuration *config.Configuration) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "agent",
		Aliases:      []string{},
		Short:        createShort,
		Long:         createLong,
		SilenceUsage: true,
	}

	// agent subcommands
	cmd.AddCommand(NewCmdCreateAgent(configuration))

	cmdUtils.SetPersistentFlagsFromEnvVariables(cmd.Commands())

	return cmd
}
