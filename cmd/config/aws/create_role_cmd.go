package aws

import (
	"fmt"
	"github.com/armory/armory-cli/pkg/auth"
	"github.com/armory/armory-cli/pkg/cmdUtils"
	cliconfig "github.com/armory/armory-cli/pkg/config"
	"github.com/armory/armory-cli/pkg/util"
	"github.com/manifoldco/promptui"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
	log "go.uber.org/zap"
	"io"
)

const (
	configCreateRoleShort = "AWS Create Role via CloudFormation Stack"
	//TODO For usage documentation, visit...
	configCreateRoleLong    = "Use AWS CloudFormation quick create to run a template that allows Armory to assume a role to manage deployments\n"
	configCreateRoleExample = "armory config aws create-role"
	templateUrl             = "https://us-east-1.console.aws.amazon.com/cloudformation/home?region=us-east-1#/stacks/quickcreate?templateURL=https://armory-cdaas-cloudformation.s3.us-west-2.amazonaws.com/iam-role-cfn.template&stackName=Armory-CDAAS-Role-Stack&param_AccountId=%s&param_ExternalId=%s"
)

func NewCreateRoleCmd(configuration *cliconfig.Configuration, reader io.ReadCloser) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create-role",
		Aliases: []string{"create-role"},
		Short:   configCreateRoleShort,
		Long:    configCreateRoleLong,
		Example: configCreateRoleExample,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cmdUtils.ExecuteParentHooks(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return createRole(cmd, configuration, reader)
		},
	}
	return cmd
}

func createRole(cmd *cobra.Command, cli *cliconfig.Configuration, reader io.ReadCloser) error {
	orgID, err := cli.GetAuth().GetOrganizationId()
	if err != nil {
		return auth.ErrNotLoggedIn
	}

	envID, err := cli.GetAuth().GetEnvironmentId()
	if err != nil {
		return auth.ErrNotLoggedIn
	}

	externalID := fmt.Sprintf("%s:%s", orgID, envID)

	cmd.SilenceUsage = true
	browser.Stderr = io.Discard
	browser.Stdout = io.Discard
	log.S().Info("In order to deploy AWS resources, we need to create a Trust Relationship in your AWS account by adding a role that Armory can assume to execute deployments on your behalf.")
	log.S().Info("1. Log in to the AWS Management Console you want to connect to in your default browser. The logged in user requires access to configure IAM roles.")
	log.S().Info("2. Click \"Create\" in the AWS CloudFormation screen and wait for the stack creation to complete.")
	log.S().Info("3. Once the CloudFormation is finished, locate the ArmoryRole by navigating to IAM > Roles in the AWS Management console. Use the `arn` value as the `account` in your CD-as-a-Service deployment targets.")
	prompt := promptui.Prompt{
		Label:     "Continue",
		IsConfirm: true,
		Default:   "Y",
		Stdout:    &util.BellSkipper{},
		Stdin:     reader,
	}

	c, err := prompt.Run()
	if err != nil {
		if c != "Y" && c != "y" {
			return nil
		}
		return err
	}
	url := fmt.Sprintf(templateUrl, cli.GetArmoryCloudEnvironmentConfiguration().AWSAccountID, externalID)
	err = browser.OpenURL(url)
	if err != nil {
		log.S().Info("Unable to open browser. Please copy and paste the following URL into your browser.")
		log.S().Info(url)
	}

	return err
}
