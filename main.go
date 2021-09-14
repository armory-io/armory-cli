package main

import (
	"github.com/armory/armory-cli/cmd"
	"github.com/armory/armory-cli/cmd/assembler"
	"os"
)

func main() {
	command, options := cmd.NewCmdRoot(os.Stdout, os.Stderr)
	assembler.AddSubCommands(command, options)
	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
