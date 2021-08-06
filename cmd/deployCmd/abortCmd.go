package deployCmd

import (
	"github.com/armory/armory-cli/internal/deng"
	"github.com/spf13/cobra"
)

var abortCmd = &cobra.Command{
	Use:   "abort",
	Short: "Abort a deployment",
	Long: "After aborting a deployment, the resource will be put in the state where it can restarted with " +
		"armory deploy restart. This is different than rollback that reverts the deployment to the last known " +
		"good deployment.",
	RunE: executeAbortCmd,
}

func init() {
	abortCmd.Flags().String(ParamDeploymentId, "", "The deployment id")
	abortCmd.Flags().String(ParamName, "", "The name of the atomic deployment part of the deployment to resume")
	abortCmd.Flags().String(ParamType, "Deployment.apps", "The type of the atomic deployment part of the deployment to resume")
	err := abortCmd.MarkFlagRequired(ParamDeploymentId)
	if err != nil {
		panic("Failed to initialize abort command err:" + err.Error())
	}
}

func executeAbortCmd(cmd *cobra.Command, args []string) error {
	deploymentId, err := cmd.Flags().GetString(ParamDeploymentId)
	if err != nil {
		return err
	}
	
	atomicDeploymentPartName, err := cmd.Flags().GetString(ParamName)
	if err != nil {
		return err
	}

	atomicDeploymentPartType, err := cmd.Flags().GetString(ParamType)
	if err != nil {
		return err
	}

	deployEngine := deng.GetDeployEngineInstance()
	err = deployEngine.AbortDeployment(deploymentId, atomicDeploymentPartName, atomicDeploymentPartType)
	if err != nil {
		return err
	}
	return nil
}