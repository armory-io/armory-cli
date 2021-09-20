package deploy

import (
	"github.com/armory/armory-cli/cmd"
	"github.com/armory/armory-cli/pkg/auth"
	"github.com/armory/armory-cli/pkg/deploy"
	"github.com/sirupsen/logrus"
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
			token, err := rootOptions.Auth.GetToken()
			if err != nil {
				logrus.Error(err)
				logrus.Fatalf("Error getting a token")
			}
			deployClient, err := deploy.NewDeployClient(
				rootOptions.DeployHostUrl,
				token,
			)
			if err != nil {
				logrus.Error(err)
				logrus.Fatalf("Error when building the api client")
			}
			rootOptions.DeployClient = deployClient
		},
	}
	cmd.AddLoginFlags(command, rootOptions)

	// create subcommands
	command.AddCommand(NewDeployStartCmd(rootOptions))
	command.AddCommand(NewDeployStatusCmd(rootOptions))
	return command
}
