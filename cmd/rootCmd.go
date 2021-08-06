package cmd

import (
	"flag"
	"github.com/armory/armory-cli/cmd/applicationCmd"
	"github.com/armory/armory-cli/cmd/accountCmd"
	"github.com/armory/armory-cli/cmd/deployCmd"
	"github.com/armory/armory-cli/internal/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	deployCliName = "armory"
	ParamVerbose  = "verbose"
)

var verboseFlag bool

var rootCmd = &cobra.Command{
	Use:   deployCliName,
	Short: "Trigger, monitor, and diagnose your deployments",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	// Add base commands
	rootCmd.AddCommand(applicationCmd.BaseCmd)
	rootCmd.AddCommand(accountCmd.BaseCmd)
	rootCmd.AddCommand(deployCmd.BaseCmd)

	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, ParamVerbose, "v", false, "show more details")
	rootCmd.PersistentFlags().StringP(config.ParamContext, "C", "default", "context")

	// Hidden flags
	rootCmd.PersistentFlags().String(config.ParamEndpoint, "deployCmd.cloud.armory.io:443", "deployCmd engine endpoint")

	rootCmd.PersistentFlags().Bool(config.ParamInsecure, false, "do not verify server certificate")
	rootCmd.PersistentFlags().Bool(config.ParamPlaintext, false, "use a plaintext connection (warning insecure!)")
	rootCmd.PersistentFlags().Bool(config.ParamNoProxy, false, "skip system defined proxy (HTTP_PROXY, HTTPS_PROXY)")
	rootCmd.PersistentFlags().String(config.ParamCacert, "", "path to server certificate authority")
	rootCmd.PersistentFlags().String(config.ParamCert, "", "path to client certificate (mTLS)")
	rootCmd.PersistentFlags().String(config.ParamKey, "", "path to client certificate key (mTLS)")
	rootCmd.PersistentFlags().String(config.ParamKeyPassword, "", "password to the client certificate key (mTLS)")
	rootCmd.PersistentFlags().String(config.ParamServerName, "", "override server name")
	rootCmd.PersistentFlags().String(config.ParamToken, "", "authentication token")
	rootCmd.PersistentFlags().Bool(config.ParamAnonymously, false, "connect anonymously. This will likely fail in a non test environment.")

	rootCmd.PersistentPreRunE = configureLogging
	rootCmd.SilenceUsage = true
}

func configureLogging(cmd *cobra.Command, args []string) error {
	lvl := log.InfoLevel
	if verboseFlag {
		lvl = log.DebugLevel
	}
	log.SetLevel(lvl)
	log.SetFormatter(&log.TextFormatter{})
	_ = flag.Set("logtostderr", "true")
	return nil
}