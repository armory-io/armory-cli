package login

import (
	"fmt"
	"github.com/armory/armory-cli/cmd"
	"github.com/armory/armory-cli/pkg/auth"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
	"io"
	"time"
)

const (
	loginShort   = ""
	loginLong    = ""
	loginExample = ""
)

type loginOptions struct {
	*cmd.RootOptions
	clientId string
	scope string
	audience string
}

func NewLoginCmd(rootOptions *cmd.RootOptions) *cobra.Command {
	options := &loginOptions{
		RootOptions: rootOptions,
	}
	command := &cobra.Command{
		Use:     "login",
		Aliases: []string{"login"},
		Short:   loginShort,
		Long:    loginLong,
		Example: loginExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return login(cmd, options, args)
		},
	}
	command.Flags().StringVarP(&options.clientId, "clientId", "c", "", "")
	command.Flags().StringVarP(&options.scope, "scope", "s", "openid profile email", "")
	command.Flags().StringVarP(&options.audience, "audience", "a", "https://api.cloud.armory.io", "")
	command.Flags().MarkHidden("clientId")
	command.Flags().MarkHidden("scope")
	command.Flags().MarkHidden("audience")
	return command
}

func login(cmd *cobra.Command, options *loginOptions, args []string) error {
	deviceTokenResponse, err := auth.GetDeviceCodeFromAuthorizationServer()
	if err != nil {
		return fmt.Errorf("error at getting device code: %s", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "You are about to be prompted to verify the following code in your default browser.")
	fmt.Fprintf(cmd.OutOrStdout(), "Device Code: %s", deviceTokenResponse.UserCode)

	authStartedAt := time.Now()

	// Sleep for 3 seconds so the user has time to read the above message
	time.Sleep(3 * time.Second)

	// Don't pollute our beautiful terminal with garbage
	browser.Stderr = io.Discard
	browser.Stdout = io.Discard
	err = browser.OpenURL(deviceTokenResponse.VerificationUriComplete)
	if err != nil {
		fmt.Fprintf(cmd.OutOrStdout(), "You are about to be prompted to verify the following code in your default browser.")
		fmt.Fprintf(cmd.OutOrStdout(), deviceTokenResponse.VerificationUriComplete)
	}

	token, err := auth.PollAuthorizationServerForResponse(deviceTokenResponse, authStartedAt)
	if err != nil {
		return fmt.Errorf("error at polling auth server for response: %s", err)
	}
	jwt, err := auth.DecodeJwtMetadata(token)
	if err != nil {
		return fmt.Errorf("error at decoding jwt: %s", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Welcome %s user: %s, your token expires at: %s", jwt.PrincipalMetadata.OrgName, jwt.PrincipalMetadata.Name, time.Unix(jwt.ExpiresAt, 0).Local().String())
	return nil
}
