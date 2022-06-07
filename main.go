package main

import (
	"github.com/armory/armory-cli/cmd"
	"os"
)

func main() {
	rootCmd := cmd.NewCmdRoot(os.Stdout, os.Stderr)
	rootCmd.Execute()
}
