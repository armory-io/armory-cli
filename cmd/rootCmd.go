package cmd

import (
	"github.com/armory/armory-cli/cmd/deploy"
	"github.com/armory/armory-cli/cmd/login"
	"github.com/armory/armory-cli/cmd/logout"
	"github.com/armory/armory-cli/cmd/template"
	"github.com/armory/armory-cli/cmd/version"
	"github.com/armory/armory-cli/pkg/cmdUtils"
	"github.com/armory/armory-cli/pkg/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	easy "github.com/t-tomalak/logrus-easy-formatter"
	"io"
)

func NewCmdRoot(outWriter, errWriter io.Writer) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "armory",
		Short: "A CLI for using Armory CD-as-a-Service",
	}

	addr := rootCmd.PersistentFlags().StringP("addr", "", "https://api.cloud.armory.io", "")
	rootCmd.PersistentFlags().MarkHidden("addr")

	clientId := rootCmd.PersistentFlags().StringP("clientId", "c", "", "configure oidc client credentials for Armory CD-as-a-Service API")
	clientSecret := rootCmd.PersistentFlags().StringP("clientSecret", "s", "", "configure oidc client credentials for Armory CD-as-a-Service API")
	accessToken := rootCmd.PersistentFlags().StringP("authToken", "a", "", "use an existing access token, rather than client id and secret or user login")
	verbose := rootCmd.PersistentFlags().BoolP("verbose", "v", false, "show more details")
	outFormat := rootCmd.PersistentFlags().StringP("output", "o", "text", "Set the output type. Available options: [json, yaml, text].")
	rootCmd.SetOut(outWriter)
	rootCmd.SetErr(errWriter)

	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		configureLogging(*verbose)
	}

	configuration := config.New(&config.Input{
		ApiAddr:      addr,
		ClientId:     clientId,
		ClientSecret: clientSecret,
		AccessToken:  accessToken,
		OutFormat:    outFormat,
	})

	rootCmd.AddCommand(version.NewCmdVersion())
	rootCmd.AddCommand(deploy.NewDeployCmd(configuration))
	rootCmd.AddCommand(template.NewTemplateCmd())
	rootCmd.AddCommand(login.NewLoginCmd(configuration))
	rootCmd.AddCommand(logout.NewLogoutCmd())
	cmdUtils.SetPersistentFlagsFromEnvVariables(rootCmd.Commands())
	cmdUtils.SetPersistentFlagsFromEnvVariables([]*cobra.Command{rootCmd})
	return rootCmd
}

func configureLogging(verboseFlag bool) {
	lvl := log.InfoLevel
	if verboseFlag {
		lvl = log.DebugLevel
	}
	log.SetLevel(lvl)
	log.SetFormatter(&easy.Formatter{
		LogFormat: "%msg%\n",
	})
}
