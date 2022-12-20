package deploy

import (
	"fmt"
	"github.com/armory/armory-cli/cmd/utils"
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
			status := cmd.Context().Value("deployStatus").(*deploymentCmdStatus)

			armoryConfig := configuration.GetArmoryCloudEnvironmentConfiguration()
			url := armoryConfig.CloudConsoleBaseUrl
			env := ""
			if !*configuration.GetIsTest() {
				env = configuration.GetCustomerEnvironmentId()
			}
			url += "/deployments/pipeline/" + status.deploymentID + "?environmentId=" + env
			if configuration.GetOutputType() == output.Text {
				fmt.Fprintf(cmd.OutOrStdout(), "[%v] See the deployment status UI: %s\n", time.Now().Format(time.RFC3339), url)
			}
			utils.TryWriteGitHubStepSummary(url)
			startResult := []string{"PIPELINE_ID", status.deploymentID, "LINK", url}

			if status.executionResult != nil {
				runResult := <-status.executionResult
				if configuration.GetOutputType() == output.Text && runResult != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "[%v] Deployment %s completed with status: %s\n", time.Now().Format(time.RFC3339), status.deploymentID, runResult)
				}
				startResult = append(startResult, "RUN_RESULT", runResult)
			}
			utils.TryWriteGitHubContext(startResult...)
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
