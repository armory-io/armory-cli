package quickStart

import (
	"github.com/armory/armory-cli/pkg/cmdUtils"
	"github.com/armory/armory-cli/pkg/config"
	"github.com/armory/armory-cli/pkg/util"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	log "go.uber.org/zap"
)

var CdConDemo = GithubQuickStartProject{
	ProjectName:   "cdCon-cdaas-demo",
	BranchName:    "main",
	DirName:       "armory-sample-manualCanary",
	IsZipFile:     true,
	DeployYmlName: "deploy.yml",
}

const (
	quickStartShort   = "Deploy a sample application"
	quickStartLong    = "Deploy a sample application"
	quickStartExample = "armory quick-start"
	githubZipSuffix   = "/archive/refs/heads/main.zip"
	githubBaseUrl     = "https://github.com/armory/"
)

type quickStartOptions struct {
	verbose   bool
	agentName string
}

func NewQuickStartCmd(configuration *config.Configuration) *cobra.Command {
	options := &quickStartOptions{}
	command := &cobra.Command{
		Use:     "quick-start",
		Aliases: []string{"quick-start"},
		Short:   quickStartShort,
		Long:    quickStartLong,
		Example: quickStartExample,
		GroupID: "deployment",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cmdUtils.ExecuteParentHooks(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return quickStart(cmd, configuration, options)
		},
	}
	command.Flags().StringVarP(&options.agentName, "agent", "", "", "")
	command.Flags().BoolVarP(&options.verbose, "verbose", "v", false, "")
	return command
}

func quickStart(cmd *cobra.Command, configuration *config.Configuration, options *quickStartOptions) error {

	log.S().Info("Welcome to the Armory CD-as-a-Service CLI!\nThis quick start downloads a sample application from GitHub and tells you how to deploy it.")

	prompt := promptui.Prompt{
		Label:     "Ready to get started",
		IsConfirm: true,
		Default:   "Y",
		Stdout:    &util.BellSkipper{},
	}

	if _, err := prompt.Run(); err != nil {
		log.S().Fatalf("Exiting %s\n", err)
	}

	demo := CdConDemo
	runner := NewProjectRunner(configuration)
	selectedAgent := runner.
		PopulateAgents().
		SelectAgent(options.agentName)
	runner.
		Exec(demo.OverwritePrompt).
		Exec(demo.Download).
		Exec(demo.Unzip).
		ExecWith(demo.UpdateAgentAccount, selectedAgent).
		FailOnError()

	log.S().Infof("\nExecute `cd %s && %s` to deploy the sample application", demo.DirName, demo.GetDeployCommand())
	log.S().Infof("\nNote: You should deploy the application twice. The first deployment creates a new application and the second is a regular deployment.")
	return nil
}
