package config

import (
	"context"
	"fmt"
	"github.com/armory/armory-cli/pkg/cmdUtils"
	"github.com/armory/armory-cli/pkg/config"
	"github.com/armory/armory-cli/pkg/configCmd"
	errorUtils "github.com/armory/armory-cli/pkg/errors"
	"github.com/armory/armory-cli/pkg/model"
	"github.com/armory/armory-cli/pkg/output"
	"github.com/spf13/cobra"
	_nethttp "net/http"
	"os"
	"time"
)

const (
	configGetShort   = "Get the current configuration from Armory CD-as-a-Service"
	configGetLong    = "Get the current configuration from Armory CD-as-a-Service"
	configGetExample = "armory config get"
)

func NewConfigGetCmd(configuration *config.Configuration) *cobra.Command {
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

func get(cmd *cobra.Command, options *configApplyOptions, configuration *config.Configuration) error {
	//in case this is running on a GitHub instance
	gitWorkspace, present := os.LookupEnv("GITHUB_WORKSPACE")
	_, isATest := os.LookupEnv("ARMORY_CLI_TEST")
	if present && !isATest {
		options.configFile = gitWorkspace + options.configFile
	}
	// since we use text as the global default we need to override that for config get
	if configuration.GetOutputType() == output.Text {
		configuration.SetOutputFormatter("yaml")
	}
	configClient := configCmd.GetConfigClient(configuration)
	ctx, cancel := context.WithTimeout(configClient.ArmoryCloudClient.Context, time.Minute)
	defer cancel()
	// execute request
	roles, resp, err := configClient.GetRoles(ctx)
	dataFormat, err := configuration.GetOutputFormatter()(newGetConfigWrapper(roles, resp, err))

	cmd.SilenceUsage = true
	if err != nil {
		return errorUtils.NewWrappedError(ErrParsingGetConfigResponse, err)
	}
	_, err = fmt.Fprintln(cmd.OutOrStdout(), dataFormat)
	return nil
}

type FormattableConfiguration struct {
	Configuration model.ConfiguationOutput `json:"roles", yaml:"roles"`
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

func newGetConfigWrapper(rawRoles []model.RoleConfig, response *_nethttp.Response, err error) FormattableConfiguration {
	wrapper := FormattableConfiguration{
		Configuration: model.ConfiguationOutput{
			Roles: rawRoles,
		},
		httpResponse: response,
		err:          err,
	}
	return wrapper
}
