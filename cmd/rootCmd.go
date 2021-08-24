package cmd

import (
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
	Short: "A CLI for using Armory Cloud",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, ParamVerbose, "v", false, "show more details")
	rootCmd.PersistentPreRunE = configureLogging
}

func configureLogging(cmd *cobra.Command, args []string) error {
	lvl := log.InfoLevel
	if verboseFlag {
		lvl = log.DebugLevel
	}
	log.SetLevel(lvl)
	log.SetFormatter(&log.TextFormatter{})
	return nil
}