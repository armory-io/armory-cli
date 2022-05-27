package quickStart

import (
	"github.com/armory/armory-cli/pkg/cmdUtils"
	"github.com/armory/armory-cli/pkg/config"
	"github.com/armory/armory-cli/pkg/util"
	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
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
	return command
}

func quickStart(cmd *cobra.Command, configuration *config.Configuration, options *quickStartOptions) error {
	if options.verbose {
		log.Info("Increasing log level")
		log.SetLevel(log.DebugLevel)
	}
	log.Info("Welcome to Armory CLI!\nQuick Start will download a sample project from Github and tell you how to deploy it.\n")

	prompt := promptui.Prompt{
		Label:     "Ready to get started",
		IsConfirm: true,
		Stdout:    &util.BellSkipper{},
	}

	_, err := prompt.Run()

	if err != nil {
		log.Fatalf("Exiting %s\n", err)
		os.Exit(0)
	}
	//cmd.SilenceUsage = true
	demo := CdConDemo
	wasCancelled, err := demo.Download()
	if wasCancelled {
		return nil
	}
	if err != nil {
		log.Fatalf("Unable to download project from Github. Please download and unzip %s, then execute `%s`", demo.GetUrl(), demo.GetDeployCommand())
	}
	err = demo.Unzip()
	if err != nil {
		log.Fatalf("Failed to unzip\n%s", err)
	}

	selectedAgent, err := demo.SelectAgent(configuration, options.agentName)
	if err != nil {
		log.Fatalf("Failed to select agent\n%s", err)
	}

	err = demo.UpdateAgentAccount(configuration, selectedAgent)

	if err != nil {
		log.Fatalf("Failed to update agent\n %s", err)
	}

	log.Infof("Execute `cd %s && %s` to deploy your application", demo.DirName, demo.GetDeployCommand())
	log.Infof("Note: You will want to deploy the application twice. The first deployment will create a new application and the second will be a regular deployment.")
	return nil
}
