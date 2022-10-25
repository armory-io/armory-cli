package config

import (
	"context"
	"fmt"
	_nethttp "net/http"
	"os"
	"time"

	"github.com/armory/armory-cli/pkg/cmdUtils"
	cliconfig "github.com/armory/armory-cli/pkg/config"
	"github.com/armory/armory-cli/pkg/configuration"
	errorUtils "github.com/armory/armory-cli/pkg/errors"
	"github.com/armory/armory-cli/pkg/model"
	configClientModel "github.com/armory/armory-cli/pkg/model/configClient"
	"github.com/armory/armory-cli/pkg/output"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

const (
	configGetShort   = "Get the current configuration from Armory CD-as-a-Service"
	configGetLong    = "Get the current configuration from Armory CD-as-a-Service"
	configGetExample = "armory config get"
)

func NewConfigGetCmd(configuration *cliconfig.Configuration) *cobra.Command {
	options := &configApplyOptions{}
	cmd := &cobra.Command{
		Use:     "get",
		Aliases: []string{"get"},
		Short:   configGetShort,
		Long:    configGetLong,
		Example: configGetExample,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cmdUtils.ExecuteParentHooks(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return get(cmd, options, configuration)
		},
	}
	return cmd
}

func get(cmd *cobra.Command, options *configApplyOptions, cli *cliconfig.Configuration) error {
	//in case this is running on a GitHub instance
	gitWorkspace, present := os.LookupEnv("GITHUB_WORKSPACE")
	_, isATest := os.LookupEnv("ARMORY_CLI_TEST")
	if present && !isATest {
		options.configFile = gitWorkspace + options.configFile
	}
	// since we use text as the global default we need to override that for config get
	if cli.GetOutputType() == output.Text {
		cli.SetOutputFormatter("yaml")
	}
	configClient := configuration.NewClient(cli)
	ctx, cancel := context.WithTimeout(configClient.ArmoryCloudClient.Context, time.Minute)
	defer cancel()
	// execute request
	environments, err := configClient.GetEnvironments(ctx)
	if err != nil {
		return errorUtils.NewWrappedError(ErrGettingEnvironments, err)
	}

	roles, resp, err := configClient.GetRoles(ctx)
	dataFormat, err := cli.GetOutputFormatter()(newGetConfigWrapper(environments, roles, resp, err))

	cmd.SilenceUsage = true
	if err != nil {
		return errorUtils.NewWrappedError(ErrParsingGetConfigResponse, err)
	}
	_, err = fmt.Fprintln(cmd.OutOrStdout(), dataFormat)
	return nil
}

type FormattableConfiguration struct {
	Configuration model.ConfigurationOutput `json:"roles" yaml:"roles"`
	httpResponse  *_nethttp.Response
	err           error
}

func (u FormattableConfiguration) Get() interface{} {
	return u.Configuration
}

func (u FormattableConfiguration) GetHttpResponse() *_nethttp.Response {
	return u.httpResponse
}

func (u FormattableConfiguration) GetFetchError() error {
	return u.err
}

func newGetConfigWrapper(rawEnvironments []configClientModel.Environment, rawRoles []model.RoleConfig, response *_nethttp.Response, err error) FormattableConfiguration {
	userOnlyRoles := make([]model.RoleConfig, 0)
	for _, role := range rawRoles {
		if !role.SystemDefined {
			userOnlyRoles = append(userOnlyRoles, role)
		}
	}

	tenants := lo.Map(rawEnvironments, func(environment configClientModel.Environment, _ int) string {
		return environment.Name
	})

	wrapper := FormattableConfiguration{
		Configuration: model.ConfigurationOutput{
			Tenants: tenants,
			Roles:   userOnlyRoles,
		},
		httpResponse: response,
		err:          err,
	}
	return wrapper
}
