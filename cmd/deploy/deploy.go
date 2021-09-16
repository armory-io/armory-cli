package deploy

import (
	"github.com/armory/armory-cli/cmd"
	"github.com/armory/armory-cli/pkg/auth"
	"github.com/spf13/cobra"
)

const (
	deployShort   = ""
	deployLong    = ""
	deployExample = ""
)

func NewDeployCmd(rootOptions *cmd.RootOptions) *cobra.Command {
	command := &cobra.Command{
		Use:     "deploy",
		Aliases: []string{"deploy"},
		Short:   deployShort,
		Long:    deployLong,
		Example: deployExample,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			rootOptions.Auth = auth.NewAuth(
				rootOptions.ClientId, rootOptions.ClientSecret, "client_credentials",
				rootOptions.TokenIssuerUrl, rootOptions.Audience)
		},
	}
	cmd.AddLoginFlags(command, rootOptions)

	// create subcommands
	command.AddCommand(NewDeployStartCmd(rootOptions))
	command.AddCommand(NewDeployStatusCmd(rootOptions))
	return command
}
