package cluster

import (
	"github.com/armory/armory-cli/pkg/cmdUtils"
	"github.com/armory/armory-cli/pkg/config"
	"github.com/spf13/cobra"
)

const (
	createShort = "Manage a temporary cluster Armory provisions for you"
	createLong  = "Manage a temporary cluster Armory provisions for you"
)

func NewClusterCmd(configuration *config.Configuration) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "cluster",
		GroupID:      "admin",
		Aliases:      []string{},
		Short:        createShort,
		Long:         createLong,
		SilenceUsage: true,
	}

	// agent subcommands
	cmd.AddCommand(NewCreateClusterCmd(configuration))

	cmdUtils.SetPersistentFlagsFromEnvVariables(cmd.Commands())

	return cmd
}
