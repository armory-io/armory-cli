package login

import (
	"errors"
	"fmt"
	"github.com/ahmetb/go-linq/v3"
	"github.com/armory/armory-cli/cmd"
	"github.com/armory/armory-cli/pkg/auth"
	"github.com/armory/armory-cli/pkg/org"
	"github.com/armory/armory-cli/pkg/util"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/manifoldco/promptui"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
	"io"
	"os"
	"strings"
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
	envName string
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
	command.Flags().StringVarP(&options.envName, "envName", "e", "", "")
	command.Flags().StringVarP(&options.scope, "scope", "", "openid profile email offline_access", "")
	command.Flags().StringVarP(&options.audience, "audience", "", "https://api.cloud.armory.io", "")
	command.Flags().StringVarP(&options.TokenIssuerUrl, "tokenIssuerUrl", "", "https://auth.cloud.armory.io/oauth", "")

	command.Flags().MarkHidden("clientId")
	command.Flags().MarkHidden("scope")
	command.Flags().MarkHidden("audience")
	command.Flags().MarkHidden("tokenIssuerUrl")
	return command
}

func login(cmd *cobra.Command, options *loginOptions, args []string) error {
	if options.clientId != "" {
		UserClientId = options.clientId
	}
	cmd.SilenceUsage = true
	deviceTokenResponse, err := auth.GetDeviceCodeFromAuthorizationServer(UserClientId, options.scope, options.audience, options.TokenIssuerUrl)
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

	response, err := auth.PollAuthorizationServerForResponse(UserClientId, options.TokenIssuerUrl, deviceTokenResponse, authStartedAt)
	if err != nil {
		return fmt.Errorf("error at polling auth server for response. Err: %s", err)
	}
	jwt, err := auth.ValidateJwt(response.AccessToken)
	if err != nil {
		return fmt.Errorf("error at decoding jwt. Err: %s", err)
	}

	selectedEnv, err := selectEnvironment(options.audience, response.AccessToken, options.envName)
	if err != nil {
		return err
	}

	response, err = auth.RefreshAuthToken(UserClientId, options.TokenIssuerUrl, response.RefreshToken, selectedEnv.Id)
	if err != nil {
		return err
	}
	jwt, err = auth.ValidateJwt(response.AccessToken)
	if err != nil {
		return fmt.Errorf("error at decoding jwt. Err: %s", err)
	}

	err = writeCredentialToFile(err, options, jwt, response)
	if err != nil {
		return err
	}

	claims := jwt.PrivateClaims()["https://cloud.armory.io/principal"].(map[string]interface{})
	fmt.Fprintf(cmd.OutOrStdout(), "Welcome %s user: %s to environment %s your token expires at: %s\n", claims["orgName"], claims["name"], selectedEnv.Name, jwt.Expiration().Format(time.RFC1123))
	return nil
}

func writeCredentialToFile(err error, options *loginOptions, jwt jwt.Token, response *auth.SuccessfulResponse) error {
	dirname, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("there was an error getting the home directory. Err: %s", err)
	}

	credentials := auth.NewCredentials(options.audience, "user-login", UserClientId, jwt.Expiration().Format(time.RFC3339), response.AccessToken, response.RefreshToken)
	err = credentials.WriteCredentials(dirname + "/.armory/credentials")
	if err != nil {
		return fmt.Errorf("there was an error writing the credentials file. Err: %s", err)
	}
	return nil
}

func selectEnvironment(audience string, accessToken string, namedEnvironment ...string) (*org.Environment, error) {
	environments, err := org.GetEnvironments(audience, &accessToken)
	if err != nil {
		return nil, err
	}
	var environmentNames []string
	linq.From(environments).Select(func(c interface{}) interface {} {
		return c.(org.Environment).Name
	}).ToSlice(&environmentNames)

	if len(namedEnvironment) > 0 && namedEnvironment[0] != "" {
		requestedEnv := linq.From(environments).Where(func(c interface{}) bool {
			return c.(org.Environment).Name == namedEnvironment[0]
		}).Select(func(c interface{}) interface {} {
			return c.(org.Environment)
		}).First()
		if requestedEnv != nil {
			sel := requestedEnv.(org.Environment)
			return &sel, nil
		}
		return nil, errors.New(fmt.Sprintf("Environment %s not found, please choose a known environment: [%s]", namedEnvironment[0], strings.Join(environmentNames[:], ",")))
	}


	prompt := promptui.Select{
		Label: "Select environment",
		Items: environmentNames,
		Stdout: &util.BellSkipper{},
	}

	_, requestedEnv, err := prompt.Run()

	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to select an environment to login to; %v\n", err))
	}
	selectedEnv := linq.From(environments).Where(func(c interface{}) bool {
		return c.(org.Environment).Name == requestedEnv
	}).Select(func(c interface{}) interface{} {
		return c.(org.Environment)
	}).First()

	if selectedEnv == nil {
		return nil, errors.New("unable to select chosen environment")
	}
	sel:= selectedEnv.(org.Environment)

	return &sel, nil
}
