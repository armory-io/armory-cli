package main

import (
	"errors"
	"github.com/armory/armory-cli/cmd"
	cmdVersion "github.com/armory/armory-cli/cmd/version"
	"github.com/armory/armory-cli/internal/clierr"
	"github.com/armory/armory-cli/internal/clierr/exitcodes"
	"github.com/armory/armory-cli/pkg/console"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"os"
)

var version string = "development"

func main() {
	cmdVersion.Version = version
	// Disabling EnableCommandSorting allows us to set our own command sort order.
	cobra.EnableCommandSorting = false

	rootCmd, err := cmd.NewCmdRoot(os.Stdout, os.Stderr)
	if err != nil {
		console.Stderrln(err.Error())
		os.Exit(int(exitcodes.Error))
	}

	// required so errors aren't double printed
	rootCmd.SilenceErrors = true
	// execute the command
	err = rootCmd.Execute()
	if err == nil {
		os.Exit(int(exitcodes.Success))
	}

	// if the error is an API Error deal with it
	var apiError *clierr.APIError
	if errors.As(err, &apiError) {
		console.Stderrln(apiError.DetailedError())
		os.Exit(apiError.ExitCode())
	}

	// else assume it's a plain error
	color.New(color.FgRed, color.Bold).Sprintln(err.Error())
	os.Exit(int(exitcodes.Error))
}
