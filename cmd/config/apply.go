package config

import (
	"context"
	"io/ioutil"
	"os"
	"time"

	"github.com/armory/armory-cli/pkg/cmdUtils"
	cliconfig "github.com/armory/armory-cli/pkg/config"
	"github.com/armory/armory-cli/pkg/configuration"
	errorUtils "github.com/armory/armory-cli/pkg/errors"
	"github.com/armory/armory-cli/pkg/model"
	"github.com/armory/armory-cli/pkg/model/configClient"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	log "go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

const (
	configApplyShort   = "Sync configuration file with Armory CD-as-a-Service"
	configApplyLong    = "Sync configuration file with Armory CD-as-a-Service"
	configApplyExample = "armory config apply [options]"
)

type configApplyOptions struct {
	configFile string
}

func NewConfigApplyCmd(configuration *cliconfig.Configuration) *cobra.Command {
	options := &configApplyOptions{}
	cmd := &cobra.Command{
		Use:     "apply --file [<path to file>]",
		Aliases: []string{"apply"},
		Short:   configApplyShort,
		Long:    configApplyLong,
		Example: configApplyExample,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cmdUtils.ExecuteParentHooks(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return apply(cmd, options, configuration)
		},
	}
	cmd.Flags().StringVarP(&options.configFile, "file", "f", "", "path to the configuration file")
	err := cmd.MarkFlagRequired("file")
	if err != nil {
		return nil
	}
	return cmd
}

func apply(cmd *cobra.Command, options *configApplyOptions, cli *cliconfig.Configuration) error {
	payload := model.ConfigurationConfig{}
	//in case this is running on a GitHub instance
	gitWorkspace, present := os.LookupEnv("GITHUB_WORKSPACE")
	_, isATest := os.LookupEnv("ARMORY_CLI_TEST")
	if present && !isATest {
		options.configFile = gitWorkspace + options.configFile
	}
	// read yaml file
	file, err := ioutil.ReadFile(options.configFile)
	if err != nil {
		return errorUtils.NewWrappedError(ErrReadingYamlFile, err)
	}
	cmd.SilenceUsage = true
	// unmarshal data into struct
	if err := yaml.UnmarshalStrict(file, &payload); err != nil {
		return errorUtils.NewWrappedError(ErrInvalidConfigurationObject, err)
	}
	cc := configuration.NewClient(cli)
	if payload.Roles != nil {
		return processRoles(cc, payload.Roles, payload.AllowAutoDelete)
	}
	return err
}

func findDeletedRoles(rolesInConfigFile, apiRoles []model.RoleConfig, environments []configClient.Environment) []model.RoleConfig {
	var deletedRoles []model.RoleConfig
	for _, apiRole := range apiRoles {
		if apiRole.SystemDefined {
			continue
		}

		_, ok := lo.Find(rolesInConfigFile, func(configRole model.RoleConfig) bool {
			return configRoleMatchesAPIRole(configRole, apiRole, environments)
		})
		if !ok {
			deletedRoles = append(deletedRoles, apiRole)
		}
	}

	return deletedRoles
}

func processRoles(configClient *configuration.ConfigClient, rolesFromConfig []model.RoleConfig, allowAutoDelete bool) error {
	//get existing roles
	ctx, cancel := context.WithTimeout(configClient.ArmoryCloudClient.Context, time.Minute)
	defer cancel()
	// execute request
	existingRoles, _, err := configClient.GetRoles(ctx)
	if err != nil {
		return errorUtils.NewWrappedError(ErrGettingRoles, err)
	}

	environments, err := configClient.GetEnvironments(ctx)
	if err != nil {
		return err
	}

	//check to see if role in config file exists already, if so perform a PUT, if not perform a POST to create
	for _, roleInConfig := range rolesFromConfig {
		exists := false
		for _, roleInExisting := range existingRoles {
			if configRoleMatchesAPIRole(roleInConfig, roleInExisting, environments) {
				exists = true
				if !roleInExisting.SystemDefined {
					//update existing role
					ctx, cancel := context.WithTimeout(configClient.ArmoryCloudClient.Context, time.Minute)
					defer cancel()
					req, err := configuration.UpdateRolesRequest(roleInExisting.ID, roleInConfig.Tenant, roleInConfig.Grants)
					if err != nil {
						return errorUtils.NewWrappedError(ErrUpdateRole, err)
					}
					_, _, err = configClient.UpdateRole(ctx, req)
					if err != nil {
						return errorUtils.NewWrappedError(ErrUpdateRole, err)
					}
					log.S().Infof("Updated role: %s", roleInConfig.Name)
				} else {
					log.S().Infof("Role %s is a system role. You cannot update it via the CLI.", roleInConfig.Name)
				}
				break
			}
		}
		if !exists {
			//create new role
			ctx, cancel := context.WithTimeout(configClient.ArmoryCloudClient.Context, time.Minute)
			defer cancel()
			req, err := configuration.CreateRoleRequest(&roleInConfig)
			_, _, err = configClient.CreateRole(ctx, req)
			if err != nil {
				return errorUtils.NewWrappedError(ErrCreatingRole, err)
			}
			log.S().Infof("Created role: %s", roleInConfig.Name)
		}
	}
	//Check to see if any existing roles are no longer in the config file, if so delete them
	deletedRoles := findDeletedRoles(rolesFromConfig, existingRoles, environments)
	if len(deletedRoles) > 0 && !allowAutoDelete {
		log.S().Info("Detected the following roles that should be deleted. Doing so may be destructive.")
		log.S().Info("You can enable deletes by setting 'allowAutoDelete' to 'true' in the configuration file.")
	}
	for _, deletedRole := range deletedRoles {
		if !allowAutoDelete {
			log.S().Info(deletedRole.Name)
		} else {
			ctx, cancel := context.WithTimeout(configClient.ArmoryCloudClient.Context, time.Minute)
			defer cancel()
			req, err := configuration.DeleteRolesRequest(deletedRole.ID)
			_, err = configClient.DeleteRole(ctx, req)
			if err != nil {
				return errorUtils.NewWrappedError(ErrDeletingRole, err)
			}
			log.S().Infof("Deleted role: %s", deletedRole.Name)
		}
	}
	return nil
}

func configRoleMatchesAPIRole(configRole model.RoleConfig, apiRole model.RoleConfig, environments []configClient.Environment) bool {
	if configRole.Tenant == "" {
		return configRole.Name == apiRole.Name
	}

	environment, ok := lo.Find(environments, func(e configClient.Environment) bool {
		return e.Name == configRole.Tenant
	})
	if !ok {
		return false
	}

	return apiRole.EnvID == environment.ID && configRole.Name == apiRole.Name
}
