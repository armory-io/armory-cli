package accountCmd

import (
	"fmt"
	"github.com/armory/armory-cli/internal/deng"
	"github.com/juju/ansiterm"
	"github.com/spf13/cobra"
	"os"
)

const (
	ParameterAccount  = "account"
	ParameterProvider = "provider"
)

var listAccountsCmd = &cobra.Command{
	Use:   "accounts",
	Short: "List available accounts",
	RunE: executeListAccountsCmd,
}

func init() {
	listAccountsCmd.Flags().String(
		"--provider",
		"kubernetes",
		"The deployment provider. EX: kubernetes",
	)
}

func executeListAccountsCmd(cmd *cobra.Command, args []string) error {
	provider, err := cmd.Flags().GetString(ParameterProvider)
	if err != nil {
		return err
	}

	deployEngine := deng.GetDeployEngineInstance()
	accounts, err := deployEngine.GetAccounts(provider)
	if err != nil {
		return err
	}

	printAccountsAsHumanReadableText(accounts, provider)

	return nil
}

func printAccountsAsHumanReadableText(accounts []*deng.AccountSummary, provider string) {
	if len(accounts) == 0 {
		msg := fmt.Sprintf("There where no accounts connected to Armory Deployments for provides %s", provider)
		fmt.Println(msg)
	}

	wt := ansiterm.NewTabWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprint(wt, "Provider\tName\n")
	for _, a := range accounts {
		_, _ = fmt.Fprintf(wt, "%s\t%s\n", provider, a.Name)
	}
	_ = wt.Flush()
}