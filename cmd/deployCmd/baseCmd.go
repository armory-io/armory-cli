package deployCmd

import (
	"github.com/spf13/cobra"
)

const (
	ParamName = "name"
	ParamType = "type"
	ParamDeploymentId = "deploymentId"
	ParameterEnvironmentName      = "account"
	ParameterEnvironmentType      = "account-type"
	ParameterEnvironmentNamespace = "namespace"
	ParameterKustomize            = "kustomize"
	ParameterLocal                = "local"
	ParameterViaAccount           = "via-account"
	ParameterViaProvider          = "via-provider"
	ParameterApplication          = "app"
	ParameterWait                 = "wait"
	ParameterVersion              = "version"

	// Strategy flags
	ParameterStrategy      = "strategy"
	ParameterStrategySteps = "canary-step"
)

var BaseCmd = &cobra.Command {
	Use:   "deploy",
	Short: "Initiate and manage deployments",
}

func init() {
	// Add deploy sub commands
	BaseCmd.AddCommand(abortCmd)
	BaseCmd.AddCommand(restartCmd)
	BaseCmd.AddCommand(resumeCmd)
	BaseCmd.AddCommand(startCmd)
	BaseCmd.AddCommand(statusCmd)
}