package main

import (
	"github.com/armory/armory-cli/cmd"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	rootCmd := cmd.NewCmdRoot(os.Stdout, os.Stderr)
	cobra.CheckErr(rootCmd.Execute())
}
