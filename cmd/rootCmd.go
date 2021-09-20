package cmd

import (
	"github.com/armory/armory-cli/pkg/auth"
	"github.com/armory/armory-cli/pkg/deploy"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
)

type RootOptions struct {
	v              bool
	ClientId       string
	ClientSecret   string
	TokenIssuerUrl string
	Audience       string
	DeployHostUrl  string
	Auth           *auth.Auth
	DeployClient   *deploy.Client
}

var rootCmd = &cobra.Command{
	Use:   "armory",
	Short: "A CLI for using Armory Cloud",
}

func NewCmdRoot(outWriter, errWriter io.Writer) (*cobra.Command, *RootOptions) {
	options := &RootOptions{}
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := configureLogging(options.v); err != nil {
			return err
		}
		return nil
	}
	rootCmd.SetOut(outWriter)
	rootCmd.SetErr(errWriter)
	rootCmd.PersistentFlags().BoolVarP(&options.v, "verbose", "v", false, "show more details")
	return rootCmd, options
}

func AddLoginFlags(cmd *cobra.Command, opts *RootOptions) {
	cmd.PersistentFlags().StringVarP(&opts.ClientId, "clientId", "c", "", "configure clientId to configure Armory Cloud")
	cmd.PersistentFlags().StringVarP(&opts.ClientSecret, "clientSecret", "s", "", "configure clientSecret to configure Armory Cloud")
	cmd.PersistentFlags().StringVarP(&opts.TokenIssuerUrl, "tokenIssuerUrl", "", "https://auth.cloud.armory.io/oauth/token", "")
	cmd.PersistentFlags().StringVarP(&opts.Audience, "audience", "", "https://api.cloud.armory.io", "")
	cmd.PersistentFlags().StringVarP(&opts.DeployHostUrl, "deployHostUrl", "", "api.cloud.armory.io", "")
	cmd.PersistentFlags().MarkHidden("tokenIssuerUrl")
	cmd.PersistentFlags().MarkHidden("audience")
}

func configureLogging(verboseFlag bool) error {
	lvl := log.InfoLevel
	if verboseFlag {
		lvl = log.DebugLevel
	}
	log.SetLevel(lvl)
	log.SetFormatter(&log.TextFormatter{})
	return nil
}
