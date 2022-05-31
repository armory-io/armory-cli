package quickStart

import (
	"github.com/armory/armory-cli/pkg/cmdUtils"
	"github.com/armory/armory-cli/pkg/config"
	"github.com/armory/armory-cli/pkg/util"
	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var CdConDemo = GithubQuickStartProject{
	ProjectName:   "cdCon-cdaas-demo",
	BranchName:    "main",
	DirName:       "armory-sample-manualCanary",
	IsZipFile:     true,
	DeployYmlName: "deploy.yml",
}

const (
	quickStartShort = "Download and run a sample project"
	quickStartLong  = "Download and run a small sample project from Armory's Github"
	loginExample    = "armory quick-start -i"
	githubZipSuffix = "/archive/refs/heads/main.zip"
	githubBaseUrl   = "https://github.com/armory/"
)

type quickStartOptions struct {
	verbose     bool
	skipPrompts bool
	agentName   string
}

func NewQuickStartCmd(configuration *config.Configuration) *cobra.Command {
	options := &quickStartOptions{}
	command := &cobra.Command{
		Use:     "quick-start",
		Aliases: []string{"quick-start"},
		Short:   quickStartShort,
		Long:    quickStartLong,
		Example: loginExample,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cmdUtils.ExecuteParentHooks(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return quickStart(cmd, configuration, options)
		},
	}
	command.Flags().StringVarP(&options.agentName, "agent", "", "", "")
	command.Flags().BoolVarP(&options.verbose, "verbose", "v", false, "")
	command.Flags().BoolVarP(&options.skipPrompts, "yes", "y", false, "")
	return command
}

func quickStart(cmd *cobra.Command, configuration *config.Configuration, options *quickStartOptions) error {
	if options.verbose {
		log.SetLevel(log.DebugLevel)
	}

	log.Info("Welcome to Armory CLI!\nQuick Start will download a sample project from Github and tell you how to deploy it.\n")
	if !options.skipPrompts {
		prompt := promptui.Prompt{
			Label:     "Ready to get started",
			IsConfirm: true,
			Stdout:    &util.BellSkipper{},
		}

		if _, err := prompt.Run(); err != nil {
			log.Fatalf("Exiting %s\n", err)
		}
	}

	demo := CdConDemo
	runner := NewProjectRunner(configuration)
	selectedAgent := runner.SelectAgent(options.agentName)
	runner.
		Exec(demo.OverwritePrompt).
		Exec(demo.Download).
		Exec(demo.Unzip).
		ExecWith(demo.UpdateAgentAccount, selectedAgent).
		FailOnError()

	log.Infof("Execute `cd %s && %s` to deploy your application", demo.DirName, demo.GetDeployCommand())
	log.Infof("Note: You will want to deploy the application twice. The first deployment will create a new application and the second will be a regular deployment.")
	return nil
}
