package deployCmd

import (
	"github.com/armory/armory-cli/internal/deng"
	"github.com/armory/armory-cli/internal/status"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"os"
)

// For now we only map to a single artifact
// which could become multiple (e.g. kustomization, helm)
// Later we could read from a deployment file definition or even
// keep adding to a server-side deployment before kicking it off.
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Initiate a deployment to a target environment",
	Long: `Initiate and optionally monitor the deployment of artifacts to a target environment.
Environments are made available by agents - see https://github.com/armory-io/armory-agents`,
	RunE: executeStartCmd,
}

func init() {
	startCmd.Flags().BoolP(ParameterWait, "w", false, "wait for deployment success or failure")
	startCmd.Flags().String(ParameterEnvironmentType, "kubernetes", "deployment account type")
	startCmd.Flags().StringP(ParameterEnvironmentName, "a", "", "deployment account name")
	startCmd.Flags().StringP(ParameterEnvironmentNamespace, "n", "", "(Kubernetes only) namespace to deploy to. Defaults to manifest namespace.")
	startCmd.Flags().String(ParameterViaAccount, "", "use specified agent to retrieve artifact")
	startCmd.Flags().String(ParameterViaProvider, "", "use agent of specified provider to retrieve artifact")
	startCmd.Flags().StringSlice(ParameterVersion, nil, "(Kubernetes only) specific container versions to override")
	startCmd.Flags().BoolP(ParameterKustomize, "k", false, "(Kubernetes only) parameter is a Kustomization")
	startCmd.Flags().BoolP(ParameterLocal, "l", false, "resolve artifacts locally")
	startCmd.Flags().String(ParameterApplication, "", "application this deployment is part of")
	startCmd.Flags().StringP(ParameterStrategy, "s", "update", "Strategy one of update,bluegreen,canary")
	startCmd.Flags().StringArray(ParameterStrategySteps, nil, "wait(duration), pause, ratio(valueOrPct), traffic(percent)")
	_ = startCmd.MarkFlagRequired(ParameterEnvironmentName)
}

func executeStartCmd(cmd *cobra.Command, args []string) error {
	var deploymentConfiguration, err = extractDeploymentConfigurationFromFlags(cmd.Flags())
	if err != nil {
		return err
	}

	var deployEngine = deng.GetDeployEngineInstance()
	_, deploymentDescriptor, err := deployEngine.StartKubernetesDeployment(deploymentConfiguration)
	if err != nil {
		return err
	}

	status.PrintStatus(os.Stdout, deploymentDescriptor)
	if deploymentConfiguration.Wait {
		return status.ShowStatus(ctx, log, desc.Id, client, w, false)
	}
	return nil
}

func extractDeploymentConfigurationFromFlags(flags *pflag.FlagSet) (*deng.DeploymentConfiguration, error) {
	application, err := flags.GetString(ParameterApplication)
	if err != nil {
		return nil, err
	}
	environmentType, err := flags.GetString(ParameterEnvironmentType)
	if err != nil {
		return nil, err
	}
	environmentName, err := flags.GetString(ParameterEnvironmentName)
	if err != nil {
		return nil, err
	}
	environmentNamespace, err := flags.GetString(ParameterEnvironmentNamespace)
	if err != nil {
		return nil, err
	}
	viaAccount, err := flags.GetString(ParameterViaAccount)
	if err != nil {
		return nil, err
	}
	viaProvider, err := flags.GetString(ParameterViaProvider)
	if err != nil {
		return nil, err
	}
	version, err := flags.GetStringSlice(ParameterVersion)
	if err != nil {
		return nil, err
	}
	kustomize, err := flags.GetBool(ParameterKustomize)
	if err != nil {
		return nil, err
	}
	local, err := flags.GetBool(ParameterLocal)
	if err != nil {
		return nil, err
	}
	strategy, err := flags.GetString(ParameterStrategy)
	if err != nil {
		return nil, err
	}
	strategySteps, err := flags.GetStringArray(ParameterStrategySteps)
	if err != nil {
		return nil, err
	}
	wait, err := flags.GetBool(ParameterWait)
	if err != nil {
		return nil, err
	}


	return &deng.DeploymentConfiguration{
		Application: application,
		EnvironmentType: environmentType,
		EnvironmentName: environmentName,
		EnvironmentNamespace: environmentNamespace,
		ViaAccount: viaAccount,
		ViaProvider: viaProvider,
		Version: version,
		Kustomize: kustomize,
		Local: local,
		Strategy: strategy,
		StrategySteps: strategySteps,
		Wait: wait,
	}, nil
}