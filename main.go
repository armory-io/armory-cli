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
	os.Exit(processCmdResults(err))
}

// processCmdResults process the results of a commands execution formatting and printing errors to stderr and determining the exit code to use
func processCmdResults(err error) int {
	if err == nil {
		return int(exitcodes.Success)
	}

	// if the error is an API Error deal with it
	var apiError *clierr.APIError
	if errors.As(err, &apiError) {
		console.Stderrln(apiError.DetailedError())
		return apiError.ExitCode()
	}

	var cliError *clierr.Error
	if errors.As(err, &cliError) {
		console.Stderrln(cliError.DetailedError())
		return cliError.ExitCode()
	}

	// else assume it's a plain error
	console.Stderrln(color.New(color.FgRed, color.Bold).Sprintf(err.Error()))
	return int(exitcodes.Error)
}
