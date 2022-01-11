package logout

import (
	"fmt"
	"github.com/armory/armory-cli/cmd"
	"github.com/armory/armory-cli/pkg/input"
	"github.com/spf13/cobra"
	"os"
)

const (
	logoutShort   = "Logout from Armory Cloud."
	logoutLong    = "Logout from Armory Cloud and delete any credentials stored."
	logoutExample = "armory logout"
)

type logoutOptions struct {
	*cmd.RootOptions
}

func NewLogoutCmd(rootOptions *cmd.RootOptions) *cobra.Command {
	options := &logoutOptions{
		RootOptions: rootOptions,
	}
	command := &cobra.Command{
		Use:     "logout",
		Aliases: []string{"logout"},
		Short:   logoutShort,
		Long:    logoutLong,
		Example: logoutExample,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return logout(cmd, options, args)
		},
	}
	return command
}

func logout(cmd *cobra.Command, options *logoutOptions, args []string) error {
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
