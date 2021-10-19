package login

import (
	"fmt"
	"github.com/armory/armory-cli/cmd"
	"github.com/armory/armory-cli/pkg/auth"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
	"io"
	"os"
	"time"
)

const (
	loginShort   = "Login as User to Armory Cloud"
	loginLong    = ""
	loginExample = ""
)

var UserClientId = ""

type loginOptions struct {
	*cmd.RootOptions
	clientId string
	scope string
	audience string
	authUrl string
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
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return login(cmd, options, args)
		},
	}
	command.Flags().StringVarP(&options.clientId, "clientId", "", "", "")
	command.Flags().StringVarP(&options.scope, "scope", "", "openid profile email", "")
	command.Flags().StringVarP(&options.audience, "audience", "", "https://api.cloud.armory.io", "")
	command.Flags().StringVarP(&options.authUrl, "authUrl", "", "https://auth.cloud.armory.io/oauth", "")

	command.Flags().MarkHidden("clientId")
	command.Flags().MarkHidden("scope")
	command.Flags().MarkHidden("audience")
	command.Flags().MarkHidden("authUrl")
	return command
}

func login(cmd *cobra.Command, options *loginOptions, args []string) error {
	if options.clientId != "" {
		UserClientId = options.clientId
	}
	deviceTokenResponse, err := auth.GetDeviceCodeFromAuthorizationServer(UserClientId, options.scope, options.audience, options.authUrl)
	if err != nil {
		return fmt.Errorf("error at getting device code: %s", err)
	}
	fmt.Fprintln(cmd.OutOrStdout(), "You are about to be prompted to verify the following code in your default browser.")
	fmt.Fprintf(cmd.OutOrStdout(), "Device Code: %s\n", deviceTokenResponse.UserCode)

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

	token, err := auth.PollAuthorizationServerForResponse(UserClientId, options.authUrl, deviceTokenResponse, authStartedAt)
	if err != nil {
		return fmt.Errorf("error at polling auth server for response. Err: %s", err)
	}
	jwt, err := auth.ValidateJwt(token)
	if err != nil {
		return fmt.Errorf("error at decoding jwt. Err: %s", err)
	}

	dirname, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("there was an error getting the home directory. Err: %s", err)
	}

	credentials := auth.NewCredentials(options.audience, "user-login", UserClientId, jwt.Expiration().Format(time.RFC3339), token)
	err = credentials.WriteCredentials(dirname + "/.armory/credentials")
	if err != nil {
		return fmt.Errorf("there was an error writing the credentials file. Err: %s", err)
	}
	claims := jwt.PrivateClaims()["https://cloud.armory.io/principal"].(map[string]interface{})
	fmt.Fprintf(cmd.OutOrStdout(), "Welcome %s user: %s, your token expires at: %s", claims["orgName"], claims["name"], jwt.Expiration().Format(time.RFC1123))
	return nil
}
