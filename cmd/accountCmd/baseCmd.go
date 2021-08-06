package accountCmd

import (
	"github.com/spf13/cobra"
)

var BaseCmd = &cobra.Command{
	Use:   "account",
	Short: "Manage Accounts connected to Armory Deployments",
}

func init() {
	BaseCmd.AddCommand(listAccountsCmd)
}