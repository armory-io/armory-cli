package applicationCmd

import (
	"github.com/spf13/cobra"
)

const (
	ParameterAccount  = "account"
	ParameterProvider = "provider"
)

var BaseCmd = &cobra.Command{
	Use:   "application",
	Short: "Manage Armory Deployment applications",
}

func init() {
	BaseCmd.AddCommand(getAllApplicationsCommand)
}