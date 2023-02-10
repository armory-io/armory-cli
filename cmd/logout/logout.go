package logout

import (
	"github.com/armory/armory-cli/pkg/cmdUtils"
	errorUtils "github.com/armory/armory-cli/pkg/errors"
	"github.com/armory/armory-cli/pkg/input"
	"github.com/spf13/cobra"
	log "go.uber.org/zap"
	"os"
)

const (
	logoutShort   = "Log out of Armory CD-as-a-Service"
	logoutLong    = "Log out of Armory CD-as-a-Service and delete any stored credentials"
	logoutExample = "armory logout"
)

func NewLogoutCmd() *cobra.Command {
	command := &cobra.Command{
		Use:     "logout",
		Aliases: []string{"logout"},
		Short:   logoutShort,
		Long:    logoutLong,
		Example: logoutExample,
		GroupID: "admin",
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
			return errorUtils.NewWrappedError(ErrGettingHomeDir, err)
		}
		if err = os.Remove(dirname + "/.armory/credentials"); os.IsNotExist(err) {
			log.S().Info("You are not logged in, skipping logout")
			return nil
		}
		log.S().Info("You have successfully been logged out")
	}
	return nil
}
