package cluster

import (
	"github.com/armory/armory-cli/pkg/cmdUtils"
	"github.com/armory/armory-cli/pkg/config"
	"github.com/spf13/cobra"
)

const (
	createShort = "Manage a temporary Kubernetes cluster provisioned by Armory"
	createLong  = "Manage a temporary kubernetes cluster Armory provisions for you"
)

func NewClusterCmd(configuration *config.Configuration, store SandboxStorage) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "cluster",
		GroupID:      "admin",
		Aliases:      []string{},
		Short:        createShort,
		Long:         createLong,
		SilenceUsage: true,
	}

	cmd.AddCommand(NewCreateClusterCmd(configuration, store))

	cmdUtils.SetPersistentFlagsFromEnvVariables(cmd.Commands())

	return cmd
}
