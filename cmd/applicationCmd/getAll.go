package applicationCmd

import (
	"fmt"
	"github.com/armory/armory-cli/internal/deng"
	"github.com/juju/ansiterm"
	"github.com/spf13/cobra"
	"os"
	"time"
)

var getAllApplicationsCommand = &cobra.Command{
	Use:   "get-all",
	Short: "gets a summery of applications for the given provider and account",
	RunE: executeListApplicationsCommand,
}

func init() {
	getAllApplicationsCommand.Flags().String(ParameterProvider, "kubernetes", "provider")
	getAllApplicationsCommand.Flags().String(ParameterAccount, "", "account name")
	BaseCmd.AddCommand(getAllApplicationsCommand)
}

func executeListApplicationsCommand(cmd *cobra.Command, args []string) error {
	account, err := cmd.Flags().GetString(ParameterAccount)
	if err != nil {
		return err
	}

	provider, err := cmd.Flags().GetString(ParameterProvider)
	if err != nil {
		return err
	}

	deployEngine := deng.GetDeployEngineInstance()
	apps, err := deployEngine.GetApplications(provider, account)
	if err != nil {
		return err
	}

	printApplicationsAsHumanReadableText(apps)
 	return nil
}

func printApplicationsAsHumanReadableText(apps []*deng.AppSummary) {
	w := os.Stdout
	_, _ = fmt.Fprintf(w, "\nAPPLICATIONS\n")
	wt := ansiterm.NewTabWriter(w, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprint(wt, "Name\tDeployments\tLast successful\tLast Failure\n")
	for _, c := range apps {
		_, _ = fmt.Fprintf(wt, "%s\t%d\t%s\t%s\n", c.Name, c.Deployments, c.LastSuccessful.Format(time.RFC3339), c.LastFailure.Format(time.RFC3339))
	}
	_ = wt.Flush()
}