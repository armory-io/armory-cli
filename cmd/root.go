package cmd

import (
	"context"
	"flag"
	"fmt"
	"github.com/armory/armory-cli/internal/app"
	"github.com/armory/armory-cli/internal/config"
	"github.com/armory/armory-cli/internal/deng"
	"github.com/armory/armory-cli/internal/deploy"
	"github.com/armory/armory-cli/internal/rollout"
	"github.com/armory/armory-cli/internal/status"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
)

const (
	deployCliName = "armory"
	ParamVerbose  = "verbose"
)

func MainCommand() *cobra.Command {
	var rootCommand = &cobra.Command{
		Use:   deployCliName,
		Short: "Trigger, monitor, and diagnose your deployments",
	}
	rootCommand.PersistentFlags().BoolP(ParamVerbose, "v", false, "show more details")
	rootCommand.PersistentFlags().StringP(config.ParamContext, "C", "default", "context")
	rootCommand.PersistentFlags().String(config.ParamEndpoint, "deploy.cloud.armory.io:443", "deploy engine endpoint")
	rootCommand.PersistentFlags().Bool(config.ParamInsecure, false, "do not verify server certificate")
	rootCommand.PersistentFlags().Bool(config.ParamPlaintext, false, "use a plaintext connection (warning insecure!)")
	rootCommand.PersistentFlags().Bool(config.ParamNoProxy, false, "skip system defined proxy (HTTP_PROXY, HTTPS_PROXY)")
	rootCommand.PersistentFlags().String(config.ParamCacert, "", "path to server certificate authority")
	rootCommand.PersistentFlags().String(config.ParamCert, "", "path to client certificate (mTLS)")
	rootCommand.PersistentFlags().String(config.ParamKey, "", "path to client certificate key (mTLS)")
	rootCommand.PersistentFlags().String(config.ParamKeyPassword, "", "password to the client certificate key (mTLS)")
	rootCommand.PersistentFlags().String(config.ParamServerName, "", "override server name")
	rootCommand.PersistentFlags().String(config.ParamToken, "", "authentication token")
	rootCommand.PersistentFlags().Bool(config.ParamAnonymously, false, "connect anonymously. This will likely fail in a non test environment.")
	rootCommand.AddCommand(appCommand(), deployCommand(), configCommand(), tokenCommand())
	rootCommand.SilenceUsage = true
	return rootCommand
}

func appCommand() *cobra.Command {
	appCmd := &cobra.Command{
		Use:   "app",
		Short: "Get and manage applications",
	}
	appCmd.AddCommand(appGetCommand())
	return appCmd
}

func appGetCommand() *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "get",
		Short: "Get application information",
		Long:  `Retrieve known applications. If an application is provided, prints information on deployments.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeCancelable(cmd, app.Execute, args)
		},
	}
	listCmd.Flags().String(app.ParameterProvider, "kubernetes", "provider")
	listCmd.Flags().String(app.ParameterAccount, "", "account name")
	return listCmd
}

func deployCommand() *cobra.Command {
	depCmd := &cobra.Command{
		Use:   "deploy",
		Short: "Initiate and manage deployments",
		Long:  `Initiate and manage deployments - see https://docs.armory.io/docs/deploy-engine`,
	}
	depCmd.AddCommand(startDeployCmd(), statusCommand(), resumeCommand(), abortCommand(), restartCommand())
	return depCmd
}

func startDeployCmd() *cobra.Command {
	// For now we only map to a single artifact
	// which could become multiple (e.g. kustomization, helm)
	// Later we could read from a deployment file definition or even
	// keep adding to a server-side deployment before kicking it off.
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Initiate a deployment to a target environment",
		Long: `Initiate and optionally monitor the deployment of artifacts to a target environment.
Environments are made available by agents - see https://github.com/armory-io/armory-agents`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeCancelable(cmd, deploy.Execute, args)
		},
	}

	startCmd.Flags().BoolP(deploy.ParameterWait, "w", false, "wait for deployment success or failure")
	startCmd.Flags().String(deploy.ParameterEnvironmentType, "kubernetes", "deployment account type")
	startCmd.Flags().StringP(deploy.ParameterEnvironmentName, "a", "", "deployment account name")
	startCmd.Flags().StringP(deploy.ParameterEnvironmentNamespace, "n", "", "(Kubernetes only) namespace to deploy to. Defaults to manifest namespace.")
	startCmd.Flags().String(deploy.ParameterViaAccount, "", "use specified agent to retrieve artifact")
	startCmd.Flags().String(deploy.ParameterViaProvider, "", "use agent of specified provider to retrieve artifact")
	startCmd.Flags().StringSlice(deploy.ParameterVersion, nil, "(Kubernetes only) specific container versions to override")
	startCmd.Flags().BoolP(deploy.ParameterKustomize, "k", false, "(Kubernetes only) parameter is a Kustomization")
	startCmd.Flags().BoolP(deploy.ParameterLocal, "l", false, "resolve artifacts locally")
	startCmd.Flags().String(deploy.ParameterApplication, "", "application this deployment is part of")
	startCmd.Flags().StringP(deploy.ParameterStrategy, "s", "update", "Strategy one of update,bluegreen,canary")
	startCmd.Flags().StringArray(deploy.ParameterStrategySteps, nil, "wait(duration), pause, ratio(valueOrPct), traffic(percent)")
	_ = startCmd.MarkFlagRequired(deploy.ParameterEnvironmentName)

	return startCmd
}

func statusCommand() *cobra.Command {
	getCmd := &cobra.Command{
		Use:   "status",
		Short: "Get deployment information",
		Long:  `Get deployment information [deployment ID]`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeCancelable(cmd, status.Execute, args)
		},
	}
	getCmd.Flags().BoolP(status.ParameterWatch, "w", false, "watch changes")
	getCmd.Flags().Bool(status.ParameterShowEvents, false, "show events")
	return getCmd
}

func resumeCommand() *cobra.Command {
	getCmd := &cobra.Command{
		Use:   "resume",
		Short: "Resume a paused deployment",
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeCancelable(cmd, rollout.Resume, args)
		},
	}
	getCmd.Flags().String(rollout.ParamName, "", "name of deployment to resume")
	getCmd.Flags().String(rollout.ParamType, "Deployment.apps", "type of deployment to resume")
	return getCmd
}

func abortCommand() *cobra.Command {
	getCmd := &cobra.Command{
		Use:   "abort",
		Short: "Abort a deployment",
		Long: "After aborting a deployment, the resource will be put in the state where it can restarted with " +
			"armory deploy restart. This is different than rollback that reverts the deployment to the last known " +
			"good deployment.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeCancelable(cmd, rollout.Abort, args)
		},
	}
	getCmd.Flags().String(rollout.ParamName, "", "name of deployment to abort")
	getCmd.Flags().String(rollout.ParamType, "Deployment.apps", "type of deployment to abort")
	return getCmd
}

func restartCommand() *cobra.Command {
	getCmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart a previously aborted deployment",
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeCancelable(cmd, rollout.Restart, args)
		},
	}
	getCmd.Flags().String(rollout.ParamName, "", "name of deployment to restart")
	getCmd.Flags().String(rollout.ParamType, "Deployment.apps", "type of deployment to restart")
	return getCmd
}

func configCommand() *cobra.Command {
	cfgCmd := &cobra.Command{
		Use:   "config",
		Short: fmt.Sprintf("View and get %s configuration", deployCliName),
	}
	viewCmd := &cobra.Command{
		Use:   "view",
		Short: fmt.Sprintf("View %s configuration", deployCliName),
		Run: func(cmd *cobra.Command, args []string) {
			config.View()
		},
	}
	cfgCmd.AddCommand(viewCmd, addConfigCommand(), accountsCommand())
	return cfgCmd
}

func addConfigCommand() *cobra.Command {
	addCmd := &cobra.Command{
		Use:   "add",
		Short: "Add a context to the configuration",
		Run: func(cmd *cobra.Command, args []string) {
			if err := config.Add(cmd); err != nil {
				fmt.Printf("Unable to add context: %s\n", err.Error())
			}
		},
	}
	addCmd.Flags().String(config.ParamAddClientId, "", "Armory cloud client-id")
	addCmd.Flags().String(config.ParamAddSecret, "", "Armory cloud secret")
	addCmd.Flags().String(config.ParamAddAudience, "", "Override Armory Cloud audience")
	addCmd.Flags().String(config.ParamAddTokenIssuerUrl, "", "Override Armory Cloud token issuer endpoint")

	_ = addCmd.MarkFlagRequired(config.ParamAddSecret)
	_ = addCmd.MarkFlagRequired(config.ParamAddClientId)
	return addCmd
}

func tokenCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "token",
		Short: "Obtain a token from configured provider",
		RunE: func(cmd *cobra.Command, args []string) error {
			log := setupLog(cmd)
			return config.GetToken(context.TODO(), log, cmd)
		},
	}
}

func accountsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "accounts",
		Short: "List available accounts",
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeCancelable(cmd, config.GetAccounts, args)
		},
	}
	cmd.Flags().String(config.ParamProvider, "kubernetes", "Provider")
	return cmd
}

func setupLog(cmd *cobra.Command) *logrus.Logger {
	verbose, _ := cmd.Flags().GetBool(ParamVerbose)
	logger := logrus.New()
	lvl := logrus.FatalLevel
	if verbose {
		lvl = logrus.DebugLevel
	}
	logger.SetLevel(lvl)
	logger.SetFormatter(&logrus.TextFormatter{})
	_ = flag.Set("logtostderr", "true")
	return logger
}

func makeDeploymentClient(log *logrus.Logger, ctx context.Context, cmd *cobra.Command) (deng.DeploymentServiceClient, error) {
	conn, err := config.GetClientConnection(log, cmd)
	if err != nil {
		return nil, err
	}
	conn.Connect(ctx)
	return deng.NewDeploymentServiceClient(conn.GetConn()), nil
}

type executor func(ctx context.Context, log *logrus.Logger, cmd *cobra.Command, client deng.DeploymentServiceClient, args []string) error

func executeCancelable(cmd *cobra.Command, exe executor, args []string) error {
	ctx, cancel := context.WithCancel(context.TODO())
	logger := setupLog(cmd)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	// Wait for signal to stop server
	go func() {
		<-signalCh
		logger.Debug("signal received, stopping command...")
		cancel()
	}()

	dc, err := makeDeploymentClient(logger, ctx, cmd)
	if err != nil {
		return err
	}
	return exe(ctx, logger, cmd, dc, args)
}