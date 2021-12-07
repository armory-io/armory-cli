package deploy

import (
	"fmt"
	"github.com/armory/armory-cli/cmd"
	"github.com/spf13/cobra"
	"strings"
	"time"
)

const (
	deployShort   = ""
	deployLong    = ""
	deployExample = ""
	cloudConsoleBaseUrl = "https://console.cloud.armory.io"
	cloudConsoleStagingBaseUrl = "https://console.staging.cloud.armory.io"
)

type deployOptions struct {
	*cmd.RootOptions
	deploymentId string
}

func NewDeployCmd(rootOptions *cmd.RootOptions) *cobra.Command {
	options := &deployOptions{
		RootOptions: rootOptions,
	}
	command := &cobra.Command{
		Use:     "deploy",
		Aliases: []string{"deploy"},
		Short:   deployShort,
		Long:    deployLong,
		Example: deployExample,
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if options.O == "" {
				url := cloudConsoleBaseUrl
				if strings.Contains(options.TokenIssuerUrl, "staging") {
					url = cloudConsoleStagingBaseUrl
				}
				url += "/deployments/pipeline/" + options.deploymentId + "?environmentId=" + options.Environment
				fmt.Fprintf(cmd.OutOrStdout(), "[%v] See the deployment status UI: %s\n", time.Now().Format(time.RFC3339), url)
			}
		},
	}
	cmd.AddLoginFlags(command, options.RootOptions)
	// create subcommands
	command.AddCommand(NewDeployStartCmd(options))
	command.AddCommand(NewDeployStatusCmd(options))
	return command
}
