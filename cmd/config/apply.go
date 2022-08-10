package config

import (
	"context"
	"fmt"
	"github.com/armory/armory-cli/pkg/cmdUtils"
	"github.com/armory/armory-cli/pkg/config"
	"github.com/armory/armory-cli/pkg/configCmd"
	"github.com/armory/armory-cli/pkg/model"
	"github.com/spf13/cobra"
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
		return fmt.Errorf("error trying to read the YAML file: %s", err)
	}
	cmd.SilenceUsage = true
	// unmarshall data into struct
	err = yaml.UnmarshalStrict(file, &payload)
	if err != nil {
		return fmt.Errorf("error invalid configuration object: %s", err)
	}
	configClient := configCmd.GetConfigClient(configuration)
	if payload.Roles != nil {
		//get existing rolls
		ctx, cancel := context.WithTimeout(configClient.ArmoryCloudClient.Context, time.Minute)
		defer cancel()
		// execute request
		roles, _, err := configClient.GetRoles(ctx)
		if err != nil {
			return fmt.Errorf("error getting existing roles: %s", err)
		}
		//check to see if roll in config file exists already, if so perform a PUT, if not perform a POST to create
		exists := false
		for _, roleInConfig := range payload.Roles {
			for _, roleInExisting := range roles {
				if roleInConfig.Name == roleInExisting.Name {
					exists = true
					//update existing role
					ctx, cancel := context.WithTimeout(configClient.ArmoryCloudClient.Context, time.Minute)
					defer cancel()
					req, err := configCmd.UpdateRolesRequest(&roleInConfig)
					_, _, err = configClient.UpdateRole(ctx, req)
					if err != nil {
						return fmt.Errorf("error trying to update role: %s", err)
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
					return fmt.Errorf("error trying to update role: %s", err)
				}
			}
			exists = false
		}
	}
	return err
}
