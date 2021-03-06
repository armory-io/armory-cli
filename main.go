package main

import (
	"github.com/armory/armory-cli/cmd"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	rootCmd := cmd.NewCmdRoot(os.Stdout, os.Stderr)
	// required so errors aren't double printed
	rootCmd.SilenceErrors = true
	// required so errors have an exit code
	cobra.CheckErr(rootCmd.Execute())
}
