package login

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/ahmetb/go-linq/v3"
	"github.com/armory/armory-cli/pkg/auth"
	"github.com/armory/armory-cli/pkg/cmdUtils"
	"github.com/armory/armory-cli/pkg/config"
	"github.com/armory/armory-cli/pkg/configuration"
	errorUtils "github.com/armory/armory-cli/pkg/errors"
	"github.com/armory/armory-cli/pkg/model/configClient"
	"github.com/armory/armory-cli/pkg/util"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/manifoldco/promptui"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
	log "go.uber.org/zap"
)

const (
	loginShort   = "Log in to Armory CD-as-a-Service"
	loginLong    = "Log in to Armory CD-as-a-Service"
	loginExample = "armory login"
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
		GroupID: "admin",
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

func login(cmd *cobra.Command, cli *config.Configuration, envName string) error {
	cmd.SilenceUsage = true
	armoryCloudEnvironmentConfiguration := cli.GetArmoryCloudEnvironmentConfiguration()
	clientId := armoryCloudEnvironmentConfiguration.CliClientId
	audience := armoryCloudEnvironmentConfiguration.Audience
	TokenIssuerUrl := armoryCloudEnvironmentConfiguration.TokenIssuerUrl

	deviceTokenResponse, err := auth.GetDeviceCodeFromAuthorizationServer(clientId, scope, audience, TokenIssuerUrl)
	if err != nil {
		return errorUtils.NewWrappedError(ErrGettingDeviceCode, err)
	}
	log.S().Info("You are about to be prompted to verify the following code in your default browser.")
	log.S().Infof("Device Code: %s\n", deviceTokenResponse.UserCode)

	authStartedAt := time.Now()

	// Sleep for 3 seconds so the user has time to read the above message
	time.Sleep(3 * time.Second)

	// Don't pollute our beautiful terminal with garbage
	browser.Stderr = io.Discard
	browser.Stdout = io.Discard
	err = browser.OpenURL(deviceTokenResponse.VerificationUriComplete)
	if err != nil {
		log.S().Info("You are about to be prompted to verify the following code in your default browser.")
		log.S().Info(deviceTokenResponse.VerificationUriComplete)
	}

	response, err := auth.PollAuthorizationServerForResponse(clientId, TokenIssuerUrl, deviceTokenResponse, authStartedAt)
	if err != nil {
		return errorUtils.NewWrappedError(ErrPollingServerResponse, err)
	}
	parsedJwt, err := auth.ParseJwtWithoutValidation(response.AccessToken)
	if err != nil {
		return errorUtils.NewWrappedError(ErrDecodingJwt, err)
	}

	err = writeCredentialToFile(cli, parsedJwt, response)
	if err != nil {
		return err
	}

	CloudClient := configuration.NewClient(cli)
	selectedEnv, err := selectEnvironment(CloudClient, envName)
	if err != nil {
		return err
	}

	response, err = auth.RefreshAuthToken(clientId, TokenIssuerUrl, response.RefreshToken, selectedEnv.ID)
	if err != nil {
		return err
	}
	parsedJwt, err = auth.ParseJwtWithoutValidation(response.AccessToken)
	if err != nil {
		return errorUtils.NewWrappedError(ErrDecodingJwt, err)
	}

	err = writeCredentialToFile(cli, parsedJwt, response)
	if err != nil {
		return err
	}

	claims := parsedJwt.PrivateClaims()["https://cloud.armory.io/principal"].(map[string]interface{})
	log.S().Infof("Welcome %s user: %s to tenant %s your token expires at: %s\n", claims["orgName"], claims["name"], selectedEnv.Name, parsedJwt.Expiration().Format(time.RFC1123))
	return nil
}

func createArmoryDirectoryIfNotExists(dir string) {
	_, err := os.Stat(dir)
	if err == nil {
		return
	}
	os.MkdirAll(dir, 0755)
}

func writeCredentialToFile(configuration *config.Configuration, jwt jwt.Token, response *auth.SuccessfulResponse) error {
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

func selectEnvironment(cc *configuration.ConfigClient, namedEnvironment ...string) (*configClient.Environment, error) {
	ctx, cancel := context.WithTimeout(cc.ArmoryCloudClient.Context, time.Minute)
	defer cancel()
	environments, err := cc.GetEnvironments(ctx)
	if err != nil {
		return nil, err
	}
	var environmentNames []string
	linq.From(environments).Select(func(c interface{}) interface{} {
		return c.(configClient.Environment).Name
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
		//lint:ignore ST1005 errors are user facing
		return nil, fmt.Errorf("Tenant %s not found, please choose a known tenant: [%s]", namedEnvironment[0], strings.Join(environmentNames[:], ","))
	}

	prompt := promptui.Select{
		Label:  "Select tenant",
		Items:  environmentNames,
		Stdout: &util.BellSkipper{},
	}

	_, requestedEnv, err := prompt.Run()

	if err != nil {
		//lint:ignore ST1005 errors are user facing
		return nil, fmt.Errorf("Failed to select an tenant to login to; %v\n", err)
	}
	selectedEnv := getEnvForEnvName(environments, requestedEnv)
	if selectedEnv == nil {
		return nil, errors.New("unable to select chosen tenant")
	}

	return selectedEnv, nil
}

func getEnvForEnvName(environments []configClient.Environment, envName string) *configClient.Environment {
	env := linq.
		From(environments).
		Where(func(c interface{}) bool {
			return c.(configClient.Environment).Name == envName
		}).
		Select(func(c interface{}) interface{} {
			return c.(configClient.Environment)
		}).
		First()
	sel := env.(configClient.Environment)
	return &sel
}
