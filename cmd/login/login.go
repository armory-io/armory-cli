package login

import (
	"errors"
	"fmt"
	"github.com/ahmetb/go-linq/v3"
	"github.com/armory/armory-cli/pkg/auth"
	"github.com/armory/armory-cli/pkg/cmdUtils"
	"github.com/armory/armory-cli/pkg/config"
	errorUtils "github.com/armory/armory-cli/pkg/errors"
	"github.com/armory/armory-cli/pkg/org"
	"github.com/armory/armory-cli/pkg/util"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/manifoldco/promptui"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
	"io"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	loginShort   = "Log in as User to Armory CD-as-a-Service"
	loginLong    = ""
	loginExample = ""
)

const scope = "openid profile email offline_access"

func NewLoginCmd(configuration *config.Configuration) *cobra.Command {
	envName := ""
	command := &cobra.Command{
		Use:     "login",
		Aliases: []string{"login"},
		Short:   loginShort,
		Long:    loginLong,
		Example: loginExample,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cmdUtils.ExecuteParentHooks(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return login(cmd, configuration, envName)
		},
	}
	command.Flags().StringVarP(&envName, "envName", "e", "", "")
	return command
}

func login(cmd *cobra.Command, configuration *config.Configuration, envName string) error {
	cmd.SilenceUsage = true
	armoryCloudEnvironmentConfiguration := configuration.GetArmoryCloudEnvironmentConfiguration()
	clientId := armoryCloudEnvironmentConfiguration.CliClientId
	audience := armoryCloudEnvironmentConfiguration.Audience
	TokenIssuerUrl := armoryCloudEnvironmentConfiguration.TokenIssuerUrl

	deviceTokenResponse, err := auth.GetDeviceCodeFromAuthorizationServer(clientId, scope, audience, TokenIssuerUrl)
	if err != nil {
		return errorUtils.NewWrappedError(ErrGettingDeviceCode, err)
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

	response, err := auth.PollAuthorizationServerForResponse(clientId, TokenIssuerUrl, deviceTokenResponse, authStartedAt)
	if err != nil {
		return errorUtils.NewWrappedError(ErrPollingServerResponse, err)
	}
	parsedJwt, err := auth.ParseJwtWithoutValidation(response.AccessToken)
	if err != nil {
		return errorUtils.NewWrappedError(ErrDecodingJwt, err)
	}

	selectedEnv, err := selectEnvironment(configuration.GetArmoryCloudAddr(), response.AccessToken, envName)
	if err != nil {
		return err
	}

	response, err = auth.RefreshAuthToken(clientId, TokenIssuerUrl, response.RefreshToken, selectedEnv.Id)
	if err != nil {
		return err
	}
	parsedJwt, err = auth.ParseJwtWithoutValidation(response.AccessToken)
	if err != nil {
		return errorUtils.NewWrappedError(ErrDecodingJwt, err)
	}

	err = writeCredentialToFile(err, configuration, parsedJwt, response)
	if err != nil {
		return err
	}

	claims := parsedJwt.PrivateClaims()["https://cloud.armory.io/principal"].(map[string]interface{})
	fmt.Fprintf(cmd.OutOrStdout(), "Welcome %s user: %s to tenant %s your token expires at: %s\n", claims["orgName"], claims["name"], selectedEnv.Name, parsedJwt.Expiration().Format(time.RFC1123))
	return nil
}

func createArmoryDirectoryIfNotExists(dir string) {
	_, error := os.Stat(dir)
	if error == nil {
		return
	}
	os.MkdirAll(dir, 0755)
}

func writeCredentialToFile(err error, configuration *config.Configuration, jwt jwt.Token, response *auth.SuccessfulResponse) error {
	dirname, err := os.UserHomeDir()
	if err != nil {
		return errorUtils.NewWrappedError(ErrGettingHomeDirectory, err)
	}

	armoryCloudEnvironmentConfiguration := configuration.GetArmoryCloudEnvironmentConfiguration()
	clientId := armoryCloudEnvironmentConfiguration.CliClientId
	audience := armoryCloudEnvironmentConfiguration.Audience

	credentials := auth.NewCredentials(audience, "user-login", clientId, jwt.Expiration().Format(time.RFC3339), response.AccessToken, response.RefreshToken)
	createArmoryDirectoryIfNotExists(dirname + "/.armory/")
	err = credentials.WriteCredentials(dirname + "/.armory/credentials")
	if err != nil {
		return errorUtils.NewWrappedError(ErrWritingCredentialsFile, err)
	}
	return nil
}

func selectEnvironment(armoryCloudAddr *url.URL, accessToken string, namedEnvironment ...string) (*org.Environment, error) {
	environments, err := org.GetEnvironments(armoryCloudAddr, &accessToken)
	if err != nil {
		return nil, err
	}
	var environmentNames []string
	linq.From(environments).Select(func(c interface{}) interface{} {
		return c.(org.Environment).Name
	}).ToSlice(&environmentNames)

	// If there is only 1 environment for the org, we will auto-select it
	if len(environments) == 1 {
		return &environments[0], nil
	}

	if len(namedEnvironment) > 0 && namedEnvironment[0] != "" {
		requestedEnv := getEnvForEnvName(environments, namedEnvironment[0])
		if requestedEnv != nil {
			return requestedEnv, nil
		}
		return nil, errors.New(fmt.Sprintf("Tenant %s not found, please choose a known tenant: [%s]", namedEnvironment[0], strings.Join(environmentNames[:], ",")))
	}

	prompt := promptui.Select{
		Label:  "Select tenant",
		Items:  environmentNames,
		Stdout: &util.BellSkipper{},
	}

	_, requestedEnv, err := prompt.Run()

	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to select an tenant to login to; %v\n", err))
	}
	selectedEnv := getEnvForEnvName(environments, requestedEnv)
	if selectedEnv == nil {
		return nil, errors.New("unable to select chosen tenant")
	}

	return selectedEnv, nil
}

func getEnvForEnvName(environments []org.Environment, envName string) *org.Environment {
	env := linq.
		From(environments).
		Where(func(c interface{}) bool {
			return c.(org.Environment).Name == envName
		}).
		Select(func(c interface{}) interface{} {
			return c.(org.Environment)
		}).
		First()
	sel := env.(org.Environment)
	return &sel
}
