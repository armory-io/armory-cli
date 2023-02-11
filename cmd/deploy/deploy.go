package deploy

import (
	"fmt"
	"github.com/armory/armory-cli/cmd/utils"
	"github.com/armory/armory-cli/pkg/cmdUtils"
	"github.com/armory/armory-cli/pkg/config"
	"github.com/armory/armory-cli/pkg/output"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"time"
)

const (
	deployShort = "Manage your deployments"
	deployLong  = "Manage your deployments\n\n" +
		"For deployment configuration YAML documentation, visit https://docs.armory.io/cd-as-a-service/reference/ref-deployment-file"
	deployExample = ""
)

func NewDeployCmd(configuration *config.Configuration) *cobra.Command {
	command := &cobra.Command{
		Use:     "deploy",
		Aliases: []string{},
		Short:   deployShort,
		Long:    deployLong,
		Example: deployExample,
		GroupID: "deployment",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cmdUtils.ExecuteParentHooks(cmd, args)
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if cmd.Context().Value("dryRun") != nil {
				dryRun, ok := cmd.Context().Value("dryRun").(bool)
				if !ok || dryRun {
					return
				}
			}

			deploymentID := fetchCommandResult(cmd, DeployResultDeploymentID)
			url := buildMonitoringUrl(configuration, deploymentID)

			reportableStatus := []string{DeployResultDeploymentID, deploymentID, DeployResultLink, url}

			syncRunStatus := fetchCommandResult(cmd, DeployResultSyncStatus)
			if syncRunStatus != "" {
				reportableStatus = append(reportableStatus, DeployResultSyncStatus, syncRunStatus)
			}

			if configuration.GetOutputType() == output.Text {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "[%v] See the deployment status UI: %s\n", time.Now().Format(time.RFC3339), url)
			}

			utils.TryWriteGitHubStepSummary(url)
			utils.TryWriteGitHubContext(reportableStatus...)
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

func buildMonitoringUrl(configuration *config.Configuration, deploymentID string) (string string) {
	armoryConfig := configuration.GetArmoryCloudEnvironmentConfiguration()
	url := armoryConfig.CloudConsoleBaseUrl
	env := lo.If(lo.FromPtrOr[bool](configuration.GetIsTest(), false), "").ElseF(configuration.GetCustomerEnvironmentId)
	url += "/deployments/pipeline/" + deploymentID + "?environmentId=" + env
	return url
}
