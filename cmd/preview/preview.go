package preview

import (
	"github.com/armory/armory-cli/pkg/cmdUtils"
	"github.com/armory/armory-cli/pkg/config"
	"github.com/spf13/cobra"
)

const (
	previewShort = "Manage network previews"
	previewLong  = "Manage network previews"
)

func NewCmdPreview(configuration *config.Configuration) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "preview",
		GroupID:      "admin",
		Aliases:      []string{},
		Short:        previewShort,
		Long:         previewLong,
		SilenceUsage: true,
		Hidden:       true,
	}

	cmd.AddCommand(NewCmdCreate(configuration))

	cmdUtils.SetPersistentFlagsFromEnvVariables(cmd.Commands())

	return cmd
}
