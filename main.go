package main

import (
	"fmt"
	"github.com/armory/armory-cli/cmd"
	"github.com/armory/armory-cli/cmd/assembler"
	"os"
)

func main() {
	command, options := cmd.NewCmdRoot(os.Stdout, os.Stderr)
	assembler.AddSubCommands(command, options)
	if err := command.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
