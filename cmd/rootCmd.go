package cmd

import (
	"bufio"
	"context"
	"github.com/armory/armory-cli/cmd/agent"
	"github.com/armory/armory-cli/cmd/cluster"
	configCmd "github.com/armory/armory-cli/cmd/config"
	"github.com/armory/armory-cli/cmd/deploy"
	"github.com/armory/armory-cli/cmd/login"
	"github.com/armory/armory-cli/cmd/logout"
	"github.com/armory/armory-cli/cmd/preview"
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
	"os"
	"time"
)

func NewCmdRoot(outWriter, errWriter io.Writer) (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use:   "armory",
		Short: "CLI for Armory CD-as-a-Service",
	}

	addr := rootCmd.PersistentFlags().StringP("addr", "", "https://api.cloud.armory.io", "")
	if err := rootCmd.PersistentFlags().MarkHidden("addr"); err != nil {
		return nil, err
	}
	test := rootCmd.PersistentFlags().BoolP("test", "", false, "")
	if err := rootCmd.PersistentFlags().MarkHidden("test"); err != nil {
		return nil, err
	}
	accessToken := rootCmd.PersistentFlags().StringP("authToken", "a", "", "Authenticate using a raw JWT token")
	if err := rootCmd.PersistentFlags().MarkHidden("authToken"); err != nil {
		return nil, err
	}

	clientId := rootCmd.PersistentFlags().StringP("clientId", "c", "", "Authenticate using an Armory CD-as-a-Service client ID")
	clientSecret := rootCmd.PersistentFlags().StringP("clientSecret", "s", "", "Authenticate using an Armory CD-as-a-Service client secret")
	verbose := rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose logging")
	outFormat := rootCmd.PersistentFlags().StringP("output", "o", "text", "Set the output type. Available options: [json, yaml, text]")
	rootCmd.SetOut(outWriter)
	rootCmd.SetErr(errWriter)

	if err := configureLogging(verbose, test, rootCmd); err != nil {
		return nil, err
	}

	configuration := config.New(&config.Input{
		ApiAddr:      addr,
		ClientId:     clientId,
		ClientSecret: clientSecret,
		AccessToken:  accessToken,
		OutFormat:    outFormat,
		IsTest:       test,
	})

	CheckForUpdate(configuration)
	rootCmd.AddGroup(
		&cobra.Group{
			ID:    "deployment",
			Title: "Deployment Commands:",
		},
		&cobra.Group{
			ID:    "admin",
			Title: "Administrative Commands:",
		},
	)

	rootCmd.AddCommand(
		deploy.NewDeployCmd(configuration),
		quickStart.NewQuickStartCmd(configuration),
		template.NewTemplateCmd(),
		login.NewLoginCmd(configuration),
		logout.NewLogoutCmd(),
		configCmd.NewConfigCmd(configuration),
		version.NewCmdVersion(),
		agent.NewCmdAgent(configuration),
		cluster.NewClusterCmd(configuration, &cluster.SandboxClusterFileStore{}),
		preview.NewCmdPreview(configuration),
	)

	cmdUtils.SetPersistentFlagsFromEnvVariables(rootCmd.Commands())
	cmdUtils.SetPersistentFlagsFromEnvVariables([]*cobra.Command{rootCmd})
	return rootCmd, nil
}

func configureLogging(verboseFlag, isTest *bool, cmd *cobra.Command) error {
	encoderConfig := zapcore.EncoderConfig{
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// info level enabler
	infoLevel := log.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level == zapcore.InfoLevel
	})

	// error, fatal, warn, and debug level enabler
	notInfoLevel := log.LevelEnablerFunc(func(level zapcore.Level) bool {
		switch level {
		case zapcore.FatalLevel:
			return true
		case zapcore.ErrorLevel:
			return true
		case zapcore.WarnLevel:
			return true
		case zapcore.DebugLevel:
			// is verbose logging turned on?
			return *verboseFlag
		default:
			return false
		}
	})

	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	// write syncers
	stdoutSyncer := zapcore.Lock(os.Stdout)
	stderrSyncer := zapcore.Lock(os.Stderr)

	core := zapcore.NewTee(
		zapcore.NewCore(encoder, stdoutSyncer, infoLevel),
		// insures that all other output goes to stderr
		zapcore.NewCore(encoder, stderrSyncer, notInfoLevel),
	)
	logger := log.New(core)

	if *isTest {
		writer := bufio.NewWriter(cmd.OutOrStdout())

		logger = log.New(
			zapcore.NewCore(encoder, zapcore.AddSync(writer), zapcore.DebugLevel))
	}

	defer logger.Sync()
	log.ReplaceGlobals(logger)

	return nil
}

func CheckForUpdate(cli *config.Configuration) {
	ctx := context.Background()
	currentVersion := version.Version
	http := &http.Client{
		Timeout: 5 * time.Second,
	}
	ghClient := github.NewClient(http)
	log.S().Debugf("Trying to contact github to find our current release...")
	currentRelease, _, err := ghClient.Repositories.GetLatestRelease(ctx, "armory-io", "armory-cli")
	if err != nil {
		return
	}
	if ((*currentRelease.TagName != currentVersion) || (currentVersion == "development")) && cli.GetOutputType() == output.Text {
		color.Set(color.FgGreen)
		log.S().Warnf("\nA new version of the Armory CLI is available. Please upgrade to %s by running `avm install`.\n", *currentRelease.TagName)
		color.Unset()
	}
}
