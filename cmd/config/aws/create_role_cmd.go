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
	"time"
)

var (
	useDevTemplate bool
)

const (
	configCreateRoleShort   = "AWS Create Role via CloudFormation Stack"
	configCreateRoleLong    = "Use AWS CloudFormation quick create to run a template that allows Armory to assume a role to manage deployments\n"
	configCreateRoleExample = "armory config aws create-role"
	templateUrl             = "https://us-east-1.console.aws.amazon.com/cloudformation/home?region=us-east-1#/stacks/quickcreate?templateURL=https://s3.us-west-2.amazonaws.com/static.cloud.armory.io/templates/aws/cloudformation/%siam-role-cfn.template&stackName=Armory-CDAAS-Role-Stack-%s&param_AccountId=%s&param_ExternalId=%s&param_ArmoryRoleName=%s"
	devTemplatePrefix       = "dev-"

	installPromptIntro  = "In order to deploy AWS resources, we need to create a Trust Relationship in your AWS account by adding a role that Armory can assume to execute deployments on your behalf."
	installPromptPrereq = "Prerequisite: Log in to the AWS Management Console you want to connect to in your default browser. The logged in user requires access to configure IAM roles."
	installPromptStep1  = "1. Type \"Y\" to begin the installation of resources, this will open your browser window with your AWS Console. You will complete the rest of this process in your browser."
	installPromptStep2  = "2. Review the resources to be created in your AWS account. Customize the name of the role that armory will use to deploy your lambda function."
	installPromptStep3  = "3. Click \"Create\" in the AWS CloudFormation screen and wait for the stack creation to complete."
	installPromptStep4  = "4. Once the CloudFormation stack creation is finished, locate the created role arn in the `Outputs` section. You will find it under the key `RoleArnOutput`. Use the `arn` value as the `account` in your CD-as-a-Service deployment targets."
	installPromptErr    = "Unable to open browser. Please copy and paste the following URL into your browser."
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

	// Note: when wanting to develop the cloudformation template being used by this command you can pass the `--dev-template` flag and then the command will use
	// `dev-iam-role-cfn.template` instead of `iam-role-cfn.template`. This allows you to make changes to the dev template and not worry about effecting users.
	cmd.Flags().BoolVar(&useDevTemplate, "dev-template", false, "")
	if err := cmd.Flags().MarkHidden("dev-template"); err != nil {
		return nil
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
	roleName := "ArmoryRole"
	epochTs := fmt.Sprintf("%d", time.Now().Unix())
	templatePrefix := ""

	if useDevTemplate {
		fmt.Println("** Using Dev Template: https://static.cloud.armory.io/templates/aws/cloudformation/dev-iam-role-cfn.template")
		templatePrefix = devTemplatePrefix
	}

	cmd.SilenceUsage = true
	browser.Stderr = io.Discard
	browser.Stdout = io.Discard
	log.S().Info(installPromptIntro)
	log.S().Info(installPromptPrereq)
	log.S().Info(installPromptStep1)
	log.S().Info(installPromptStep2)
	log.S().Info(installPromptStep3)
	log.S().Info(installPromptStep4)
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
	url := fmt.Sprintf(templateUrl, templatePrefix, epochTs, cli.GetArmoryCloudEnvironmentConfiguration().AWSAccountID, externalID, roleName)
	err = browser.OpenURL(url)
	if err != nil {
		log.S().Info(installPromptErr)
		log.S().Info(url)
	}

	return err
}
