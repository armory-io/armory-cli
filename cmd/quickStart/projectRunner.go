package quickStart

import (
	"errors"
	"fmt"
	"github.com/ahmetb/go-linq/v3"
	"github.com/armory/armory-cli/pkg/config"
	"github.com/armory/armory-cli/pkg/org"
	"github.com/armory/armory-cli/pkg/util"
	"github.com/hashicorp/go-multierror"
	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"
	"strings"
)

type ProjectRunner struct {
	Errors        *multierror.Error
	Configuration *config.Configuration
}

func NewProjectRunner(configuration *config.Configuration) *ProjectRunner {
	return &ProjectRunner{
		Configuration: configuration,
	}
}

func (r *ProjectRunner) HasErrors() bool {
	return r.Errors.ErrorOrNil() != nil
}

func (r *ProjectRunner) FailOnError() {
	if r.HasErrors() {
		log.Fatal(r.Errors)
	}
}

func (r *ProjectRunner) AppendError(err error) *ProjectRunner {
	if err != nil {
		r.Errors = multierror.Append(r.Errors.ErrorOrNil(), err)
	}
	return r
}

func (r *ProjectRunner) Exec(f func() error) *ProjectRunner {
	if r.HasErrors() {
		return r
	}
	return r.AppendError(f())
}

func (r *ProjectRunner) ExecWith(f func(string) error, x string) *ProjectRunner {
	if r.HasErrors() {
		return r
	}
	return r.AppendError(f(x))
}

func (r *ProjectRunner) SelectAgent(namedAgent string) string {
	if r.HasErrors() {
		return ""
	}

	log.Info("Fetching armory agents that are connected to your k8s cluster...")
	agents, err := org.GetAgents(r.Configuration.GetArmoryCloudAddr(), r.Configuration.GetAuthToken())

	if err != nil {
		r.AppendError(err)
		return ""
	}
	var agentIdentifiers []string
	linq.From(agents).Select(func(c interface{}) interface{} {
		log.Debugln(fmt.Sprintf("Found agent %s", c.(org.Agent).AgentIdentifier))
		return c.(org.Agent).AgentIdentifier
	}).ToSlice(&agentIdentifiers)

	if len(agentIdentifiers) < 1 {
		r.AppendError(errors.New(fmt.Sprintf("No agents were found. Please ensure you have a connected agent: %s%s", r.Configuration.GetArmoryCloudEnvironmentConfiguration().CloudConsoleBaseUrl, "/configuration/agents")))
		return ""
	}

	if len(namedAgent) > 0 && namedAgent != "" {
		requestedAgent := ""
		for _, agentName := range agentIdentifiers {
			if agentName == namedAgent {
				requestedAgent = agentName
			}
		}
		if requestedAgent == "" {
			r.AppendError(errors.New(fmt.Sprintf("Specified agent %s not found, please choose a known agent: [%s]", namedAgent, strings.Join(agentIdentifiers[:], ","))))
			return ""
		}
	}

	selectedAgent := ""
	if len(agentIdentifiers) == 1 {
		selectedAgent = agentIdentifiers[0]
	} else {
		prompt := promptui.Select{
			Label:  "Select one of your connected agents",
			Items:  agentIdentifiers,
			Stdout: &util.BellSkipper{},
		}

		_, selectedAgent, err = prompt.Run()

		if err != nil || selectedAgent == "" {
			r.AppendError(errors.New(fmt.Sprintf("Failed to select an agent to deploy to; %v\n", err)))
			return ""
		}
	}

	log.Debugln(fmt.Sprintf("Selected agent %s", selectedAgent))
	return selectedAgent
}
