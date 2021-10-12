package template

import (
	"github.com/armory/armory-cli/cmd"
	"github.com/spf13/cobra"
)

const (
	templateShort   = "Generate a deployment template that you can customize."
	templateLong    = ""
	templateExample = ""
)

type templateOptions struct {
	*cmd.RootOptions
}

func NewTemplateCmd(rootOptions *cmd.RootOptions) *cobra.Command {
	options := &templateOptions{
		RootOptions: rootOptions,
	}
	command := &cobra.Command{
		Use:     "template",
		Aliases: []string{"template"},
		Short:   templateShort,
		Long:    templateLong,
		Example: templateExample,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	// create subcommands
	command.AddCommand(NewTemplateKubernetesCmd(options))
	return command
}
