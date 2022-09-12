package deploy

import (
	"fmt"
	"github.com/armory/armory-cli/pkg/cmdUtils"
	"github.com/armory/armory-cli/pkg/config"
	"github.com/armory/armory-cli/pkg/output"
	"github.com/spf13/cobra"
	"time"
)

const (
	deployShort   = ""
	deployLong    = ""
	deployExample = ""
)

func NewDeployCmd(configuration *config.Configuration) *cobra.Command {
	command := &cobra.Command{
		Use:     "deploy",
		Aliases: []string{},
		Short:   deployShort,
		Long:    deployLong,
		Example: deployExample,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cmdUtils.ExecuteParentHooks(cmd, args)
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if configuration.GetOutputType() == output.Text {
				deploymentId := cmd.Context().Value("deploymentId").(string)
				armoryConfig := configuration.GetArmoryCloudEnvironmentConfiguration()
				url := armoryConfig.CloudConsoleBaseUrl
				url += "/deployments/pipeline/" + deploymentId + "?environmentId=" + configuration.GetCustomerEnvironmentId()
				fmt.Fprintf(cmd.OutOrStdout(), "[%v] See the deployment status UI: %s\n", time.Now().Format(time.RFC3339), url)
			}
		},
	}

	command.PersistentFlags().BoolP("test", "", false, "")
	command.PersistentFlags().MarkHidden("test")

	// create subcommands
	command.AddCommand(NewDeployStartCmd(configuration))
	command.AddCommand(NewDeployStatusCmd(configuration))

	cmdUtils.SetPersistentFlagsFromEnvVariables(command.Commands())

	return command
}
