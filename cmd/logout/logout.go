package logout

import (
	"fmt"
	"github.com/armory/armory-cli/pkg/cmdUtils"
	"github.com/armory/armory-cli/pkg/input"
	"github.com/spf13/cobra"
	"os"
)

const (
	logoutShort   = "Log out from Armory CDaaS."
	logoutLong    = "Log out from Armory CDaaS and delete any credentials stored."
	logoutExample = "armory logout"
)

func NewLogoutCmd() *cobra.Command {
	command := &cobra.Command{
		Use:     "logout",
		Aliases: []string{"logout"},
		Short:   logoutShort,
		Long:    logoutLong,
		Example: logoutExample,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cmdUtils.ExecuteParentHooks(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return logout(cmd)
		},
	}
	return command
}

func logout(cmd *cobra.Command) error {
	promptMsg := input.PromptMsg{
		Text:     "Are you sure you want to log out? Y/N",
		ErrorMsg: "Invalid answer",
	}
	word, err := input.PromptConfirmInput(promptMsg)
	if err != nil {
		return err
	}

	if word {
		dirname, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("error at getting user home dir: %s", err)
		}
		if err = os.Remove(dirname + "/.armory/credentials"); os.IsNotExist(err) {
			fmt.Fprintln(cmd.OutOrStdout(), "You are not logged in, skipping logout")
			return nil
		}
		fmt.Fprintln(cmd.OutOrStdout(), "You have successfully been logged out")
	}
	return nil
}
