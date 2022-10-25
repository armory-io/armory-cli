package cmd

import (
	"bufio"
	"context"
	configCmd "github.com/armory/armory-cli/cmd/config"
	"github.com/armory/armory-cli/cmd/deploy"
	"github.com/armory/armory-cli/cmd/login"
	"github.com/armory/armory-cli/cmd/logout"
	"github.com/armory/armory-cli/cmd/quickStart"
	"github.com/armory/armory-cli/cmd/template"
	"github.com/armory/armory-cli/cmd/version"
	"github.com/armory/armory-cli/pkg/cmdUtils"
	"github.com/armory/armory-cli/pkg/config"
	"github.com/armory/armory-cli/pkg/output"
	"github.com/fatih/color"
	"github.com/google/go-github/v48/github"
	"github.com/spf13/cobra"
	log "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"net/http"
	"time"
)

func NewCmdRoot(outWriter, errWriter io.Writer) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "armory",
		Short: "A CLI for using Armory CD-as-a-Service",
	}

	addr := rootCmd.PersistentFlags().StringP("addr", "", "https://api.cloud.armory.io", "")
	err := rootCmd.PersistentFlags().MarkHidden("addr")
	test := rootCmd.PersistentFlags().BoolP("test", "", false, "")
	err = rootCmd.PersistentFlags().MarkHidden("test")

	if err != nil {
		return nil
	}

	clientId := rootCmd.PersistentFlags().StringP("clientId", "c", "", "configure oidc client credentials for Armory CD-as-a-Service API")
	clientSecret := rootCmd.PersistentFlags().StringP("clientSecret", "s", "", "configure oidc client credentials for Armory CD-as-a-Service API")
	accessToken := rootCmd.PersistentFlags().StringP("authToken", "a", "", "use an existing access token, rather than client id and secret or user login")
	verbose := rootCmd.PersistentFlags().BoolP("verbose", "v", false, "show more details")
	outFormat := rootCmd.PersistentFlags().StringP("output", "o", "text", "Set the output type. Available options: [json, yaml, text].")
	rootCmd.SetOut(outWriter)
	rootCmd.SetErr(errWriter)

	configureLogging(*verbose, *test, rootCmd)

	configuration := config.New(&config.Input{
		ApiAddr:      addr,
		ClientId:     clientId,
		ClientSecret: clientSecret,
		AccessToken:  accessToken,
		OutFormat:    outFormat,
		IsTest:       test,
	})

	CheckForUpdate(configuration)

	rootCmd.AddCommand(version.NewCmdVersion())
	rootCmd.AddCommand(deploy.NewDeployCmd(configuration))
	rootCmd.AddCommand(quickStart.NewQuickStartCmd(configuration))
	rootCmd.AddCommand(template.NewTemplateCmd())
	rootCmd.AddCommand(login.NewLoginCmd(configuration))
	rootCmd.AddCommand(logout.NewLogoutCmd())
	rootCmd.AddCommand(configCmd.NewConfigCmd(configuration))
	cmdUtils.SetPersistentFlagsFromEnvVariables(rootCmd.Commands())
	cmdUtils.SetPersistentFlagsFromEnvVariables([]*cobra.Command{rootCmd})
	return rootCmd
}

func configureLogging(verboseFlag, isTest bool, cmd *cobra.Command) {
	lvl := log.InfoLevel
	if verboseFlag {
		lvl = log.DebugLevel
	}

	loggerConfig := log.NewProductionConfig()
	encodingConfig := log.NewDevelopmentEncoderConfig()
	encodingConfig.TimeKey = ""
	encodingConfig.LevelKey = ""
	encodingConfig.NameKey = ""
	encodingConfig.CallerKey = ""

	loggerConfig.Encoding = "console"
	loggerConfig.Level = log.NewAtomicLevelAt(lvl)
	loggerConfig.EncoderConfig = encodingConfig
	logger, err := loggerConfig.Build()

	if isTest {
		encoder := zapcore.NewConsoleEncoder(encodingConfig)
		writer := bufio.NewWriter(cmd.OutOrStdout())

		logger = log.New(
			zapcore.NewCore(encoder, zapcore.AddSync(writer), zapcore.DebugLevel))
	}

	if err != nil {
		panic(err)
	}

	defer logger.Sync()
	log.ReplaceGlobals(logger)

}

func CheckForUpdate(cli *config.Configuration) {
	ctx := context.Background()
	currentVersion := version.Version
	http := &http.Client{
		Timeout: 5 * time.Second,
	}
	ghClient := github.NewClient(http)
	currentRelease, _, err := ghClient.Repositories.GetLatestRelease(ctx, "armory-io", "armory-cli")
	if err != nil {
		return
	}
	if ((*currentRelease.TagName != currentVersion) || (currentVersion == "development")) && cli.GetOutputType() == output.Text {
		color.Set(color.BgGreen)
		log.S().Infof("\nA new version of the Armory CLI, %s, is available. Please upgrade your cli by running avm install \n", *currentRelease.TagName)
		color.Unset()
	}
}
