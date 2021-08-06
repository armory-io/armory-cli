package deployCmd

import (
	"github.com/armory/armory-cli/internal/deng"
	"github.com/spf13/cobra"
)

var resumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "Resume a paused deployment",
	RunE: executeResumeCmd,
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

func executeResumeCmd(cmd *cobra.Command, args []string) error {
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
	err = deployEngine.ResumeDeployment(deploymentId, atomicDeploymentPartName, atomicDeploymentPartType)
	if err != nil {
		return err
	}
	return nil
}