package config

import (
	"context"
	"github.com/armory/armory-cli/pkg/cmdUtils"
	"github.com/armory/armory-cli/pkg/config"
	"github.com/armory/armory-cli/pkg/configCmd"
	errorUtils "github.com/armory/armory-cli/pkg/errors"
	"github.com/armory/armory-cli/pkg/model"
	"github.com/spf13/cobra"
	log "go.uber.org/zap"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"time"
)

const (
	configApplyShort   = "Sync configuration file with Armory CD-as-a-Service"
	configApplyLong    = "Sync configuration file with Armory CD-as-a-Service"
	configApplyExample = "armory config apply [options]"
)

type configApplyOptions struct {
	configFile string
}

func NewConfigApplyCmd(configuration *config.Configuration) *cobra.Command {
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

func apply(cmd *cobra.Command, options *configApplyOptions, configuration *config.Configuration) error {
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
	// unmarshall data into struct
	err = yaml.UnmarshalStrict(file, &payload)
	if err != nil {
		return errorUtils.NewWrappedError(ErrInvalidConfigurationObject, err)
	}
	configClient := configCmd.GetConfigClient(configuration)
	if payload.Roles != nil {
		err := processRoles(configClient, payload.Roles, cmd, payload.AllowAutoDelete)
		if err != nil {
			return err
		}
	}
	return err
}

func findDeletedRoles(rolesInConfigFile, existingRoles []model.RoleConfig) []string {
	ma := make(map[string]bool, len(rolesInConfigFile))
	var deletedRoles []string
	for _, configRole := range rolesInConfigFile {
		ma[configRole.Name] = true
	}
	for _, existingRole := range existingRoles {
		if !ma[existingRole.Name] && !existingRole.SystemDefined {
			deletedRoles = append(deletedRoles, existingRole.Name)
		}
	}
	return deletedRoles
}

func processRoles(configClient *configCmd.ConfigClient, rolesFromConfig []model.RoleConfig, cmd *cobra.Command, allowAutoDelete bool) error {
	//get existing rolls
	ctx, cancel := context.WithTimeout(configClient.ArmoryCloudClient.Context, time.Minute)
	defer cancel()
	// execute request
	existingRoles, _, err := configClient.GetRoles(ctx)
	if err != nil {
		return errorUtils.NewWrappedError(ErrGettingRoles, err)
	}
	//check to see if role in config file exists already, if so perform a PUT, if not perform a POST to create
	for _, roleInConfig := range rolesFromConfig {
		exists := false
		for _, roleInExisting := range existingRoles {
			if roleInConfig.Name == roleInExisting.Name {
				exists = true
				if !roleInExisting.SystemDefined {
					//update existing role
					ctx, cancel := context.WithTimeout(configClient.ArmoryCloudClient.Context, time.Minute)
					defer cancel()
					req, err := configCmd.UpdateRolesRequest(&roleInConfig)
					_, _, err = configClient.UpdateRole(ctx, req)
					if err != nil {
						return errorUtils.NewWrappedError(ErrUpdateRole, err)
					}
					log.S().Infof("Updated Role: '%s'", roleInConfig.Name)
				} else {
					log.S().Infof("Role %s is a system defined role and cannot be updated via CLI.", roleInConfig.Name)
				}
				break
			}
		}
		if !exists {
			//create new role
			ctx, cancel := context.WithTimeout(configClient.ArmoryCloudClient.Context, time.Minute)
			defer cancel()
			req, err := configCmd.CreateRoleRequest(&roleInConfig)
			_, _, err = configClient.CreateRole(ctx, req)
			if err != nil {
				return errorUtils.NewWrappedError(ErrCreatingRole, err)
			}
			log.S().Infof("Created role: %s", roleInConfig.Name)
		}
	}
	//Check to see if any existing roles are no longer in the config file, if so delete them
	deletedRoles := findDeletedRoles(rolesFromConfig, existingRoles)
	if len(deletedRoles) > 0 && !allowAutoDelete {
		log.S().Info("Detected the following roles that should be deleted. Doing so may be destructive.")
		log.S().Info("You can enable deletes by setting 'allowAutoDelete' to 'true' in the configuration file.")
	}
	for _, deletedRole := range deletedRoles {
		if !allowAutoDelete {
			log.S().Info(deletedRole)
		} else {
			ctx, cancel := context.WithTimeout(configClient.ArmoryCloudClient.Context, time.Minute)
			defer cancel()
			req, err := configCmd.DeleteRolesRequest(deletedRole)
			_, err = configClient.DeleteRole(ctx, req)
			if err != nil {
				return errorUtils.NewWrappedError(ErrDeletingRole, err)
			}
			log.S().Infof("Deleted role: %s", deletedRole)
		}
	}
	return nil
}
