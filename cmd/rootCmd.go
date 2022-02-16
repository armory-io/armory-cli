package cmd

import (
	"errors"
	"fmt"
	"github.com/armory/armory-cli/pkg/auth"
	"github.com/armory/armory-cli/pkg/deploy"
	"github.com/armory/armory-cli/pkg/output"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	easy "github.com/t-tomalak/logrus-easy-formatter"
	"io"
)

type RootOptions struct {
	dev            bool
	v              bool
	O              string
	clientId       string
	clientSecret   string
	audience       string
	deployHostUrl  string
	TokenIssuerUrl string
	Environment    string
	Token          string
	DeployClient   *deploy.Client
	Output         *output.Output
}

func NewCmdRoot(outWriter, errWriter io.Writer) (*cobra.Command, *RootOptions) {
	rootCmd := &cobra.Command{
		Use:   "armory",
		Short: "A CLI for using Armory Cloud",
	}
	options := &RootOptions{}
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := configureLogging(options.v); err != nil {
			return fmt.Errorf("error at configuring logging: %s", err)
		}
		if options.O != "" && options.O != "json" && options.O != "yaml" {
			return errors.New("the output type is invalid. Do not specify parameter to get plain output. Available options: [json]")
		}
		options.Output = output.NewOutput(options.O)
		auth := auth.NewAuth(
			options.clientId, options.clientSecret, "client_credentials",
			options.TokenIssuerUrl, options.audience, options.Token)
		token, err := auth.GetToken()
		if err != nil {
			return fmt.Errorf("error at retrieving a token: %s", err)
		}
		options.Environment, err = auth.GetEnvironment()
		if err != nil {
			return fmt.Errorf("error at retrieving the environment: %s", err)
		}
		deployClient, err := deploy.NewDeployClient(
			options.deployHostUrl,
			token,
			options.dev,
		)
		if err != nil {
			return fmt.Errorf("error at creating the http client: %s", err)
		}
		options.DeployClient = deployClient
		return nil
	}
	rootCmd.SetOut(outWriter)
	rootCmd.SetErr(errWriter)
	rootCmd.PersistentFlags().BoolVarP(&options.v, "verbose", "v", false, "show more details")
	rootCmd.PersistentFlags().StringVarP(&options.O, "output", "o", "", "Set the output type. Available options: [json, yaml]. Default plain text.")
	rootCmd.PersistentFlags().BoolVarP(&options.dev, "dev", "d", false, "local deploy engine development")
	rootCmd.PersistentFlags().MarkHidden("dev")
	return rootCmd, options
}

func AddLoginFlags(cmd *cobra.Command, opts *RootOptions) {
	cmd.PersistentFlags().StringVarP(&opts.clientId, "clientId", "c", "", "configure clientId to configure Armory Cloud")
	cmd.PersistentFlags().StringVarP(&opts.clientSecret, "clientSecret", "s", "", "configure clientSecret to configure Armory Cloud")
	cmd.PersistentFlags().StringVarP(&opts.TokenIssuerUrl, "tokenIssuerUrl", "", "https://auth.cloud.armory.io/oauth", "")
	cmd.PersistentFlags().StringVarP(&opts.Token, "authToken", "a", "", "use an existing token, rather than client id and secret or user login")
	cmd.PersistentFlags().StringVarP(&opts.audience, "audience", "", "https://api.cloud.armory.io", "")
	cmd.PersistentFlags().StringVarP(&opts.deployHostUrl, "deployHostUrl", "", "api.cloud.armory.io", "")
	cmd.PersistentFlags().MarkHidden("tokenIssuerUrl")
	cmd.PersistentFlags().MarkHidden("audience")
	cmd.PersistentFlags().MarkHidden("deployHostUrl")
}

func configureLogging(verboseFlag bool) error {
	lvl := log.InfoLevel
	if verboseFlag {
		lvl = log.DebugLevel
	}
	log.SetLevel(lvl)
	log.SetFormatter(&easy.Formatter{
		LogFormat: "%msg%\n",
	})
	return nil
}
