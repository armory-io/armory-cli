package deploy

import (
	"github.com/armory/armory-cli/cmd"
	"github.com/spf13/cobra"
)

const (
	deployShort   = ""
	deployLong    = ""
	deployExample = ""
)

func NewDeployCmd(rootOptions *cmd.RootOptions) *cobra.Command {
	command := &cobra.Command{
		Use:     "deploy",
		Aliases: []string{"deploy"},
		Short:   deployShort,
		Long:    deployLong,
		Example: deployExample,
	}
	cmd.AddLoginFlags(command, rootOptions)
	// create subcommands
	command.AddCommand(NewDeployStartCmd(rootOptions))
	command.AddCommand(NewDeployStatusCmd(rootOptions))
	return command
}
