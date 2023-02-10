package template

import (
	"github.com/armory/armory-cli/pkg/cmdUtils"
	"github.com/spf13/cobra"
)

const (
	templateShort   = "Generate a customizable deployment template"
	templateLong    = "Generate a customizable deployment template"
	templateExample = "armory template"
)

func NewTemplateCmd() *cobra.Command {
	command := &cobra.Command{
		Use:     "template",
		Aliases: []string{"template"},
		Short:   templateShort,
		Long:    templateLong,
		Example: templateExample,
		GroupID: "deployment",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cmdUtils.ExecuteParentHooks(cmd, args)
		},
	}
	// create subcommands
	command.AddCommand(NewTemplateKubernetesCmd())
	cmdUtils.SetPersistentFlagsFromEnvVariables(command.Commands())
	return command
}
