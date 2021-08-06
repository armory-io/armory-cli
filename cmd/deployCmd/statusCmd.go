package deployCmd

import (
	"github.com/armory/armory-cli/internal/helpers"
	"github.com/armory/armory-cli/internal/status"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get deployment information",
	Long:  `Get deployment information [deployment ID]`,
	RunE: func(c *cobra.Command, args []string) error {
		return helpers.ExecuteCancelable(c, status.Execute, args)
	},
}

func init() {
	statusCmd.Flags().BoolP(status.ParameterWatch, "w", false, "watch changes")
	statusCmd.Flags().Bool(status.ParameterShowEvents, false, "show events")
}