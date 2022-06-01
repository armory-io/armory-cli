package quickStart

import (
	"fmt"
	"github.com/ahmetb/go-linq/v3"
	"github.com/armory/armory-cli/pkg/config"
	"github.com/armory/armory-cli/pkg/org"
	"github.com/armory/armory-cli/pkg/util"
	"github.com/hashicorp/go-multierror"
	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"
	"net/url"
	"strings"
)

var orgGetAgents = org.GetAgents

type ProjectRunner struct {
	Errors              *multierror.Error
	ArmoryCloudAddr     *url.URL
	AuthToken           string
	CloudConsoleBaseUrl string
	AgentIdentifiers    *[]string
}

type NoAgentsFoundError struct {
	msg string
}

func (n NoAgentsFoundError) Error() string {
	return fmt.Sprintf("No armory agents found. %s", n.msg)
}

type SelectedAgentError struct {
	msg string
}

func (n SelectedAgentError) Error() string {
	return fmt.Sprintf("Unable to continue. An agent must be selected. %s", n.msg)
}

func NewProjectRunner(configuration config.CliConfiguration) *ProjectRunner {
	return &ProjectRunner{
		ArmoryCloudAddr:     configuration.GetArmoryCloudAddr(),
		AuthToken:           configuration.GetAuthToken(),
		CloudConsoleBaseUrl: configuration.GetArmoryCloudEnvironmentConfiguration().CloudConsoleBaseUrl,
		AgentIdentifiers:    &[]string{},
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
	if r.AgentIdentifiers == nil || len(*r.AgentIdentifiers) < 1 {
		r.AppendError(NoAgentsFoundError{msg: "Ensure the running process has populated the agent list"})
		return ""
	}
	if len(namedAgent) > 0 && namedAgent != "" {
		for _, agentName := range *r.AgentIdentifiers {
			if agentName == namedAgent {
				return namedAgent
			}
		}

		r.AppendError(SelectedAgentError{msg: fmt.Sprintf("Specified Remote Network Agent %s not found. Choose a connected Remote Network Agent: [%s]", namedAgent, strings.Join(*r.AgentIdentifiers, ","))})
		return ""
	}

	selectedAgent := ""
	if len(*r.AgentIdentifiers) == 1 {
		selectedAgent = (*r.AgentIdentifiers)[0]
	} else {
		prompt := promptui.Select{
			Label:  "Select one of your connected agents",
			Items:  *r.AgentIdentifiers,
			Stdout: &util.BellSkipper{},
		}

		_, selectedAgent, err := prompt.Run()

		if err != nil || selectedAgent == "" {
			r.AppendError(SelectedAgentError{msg: fmt.Sprintf("Failed to select a Remote Network Agent to deploy to; %s", err.Error())})
			return ""
		}
	}

	log.Debugln(fmt.Sprintf("Selected agent %s", selectedAgent))
	return selectedAgent
}

func (r *ProjectRunner) PopulateAgents() *ProjectRunner {
	if r.HasErrors() {
		return r
	}

	log.Info("Fetching Remote Network Agents that are connected to your Kubernetes cluster...")
	agents, err := orgGetAgents(r.ArmoryCloudAddr, r.AuthToken)

	if err != nil {
		r.AppendError(err)
		return r
	}
	foundIdentifiers := []string{}
	linq.From(agents).Select(func(c interface{}) interface{} {
		log.Debugln(fmt.Sprintf("Found Remote Network Agent '%s'", c.(org.Agent).AgentIdentifier))
		return c.(org.Agent).AgentIdentifier
	}).Distinct().ToSlice(&foundIdentifiers)
	r.AgentIdentifiers = &foundIdentifiers
	if len(*r.AgentIdentifiers) < 1 {
		r.AppendError(NoAgentsFoundError{msg: fmt.Sprintf("No Remote Network Agents were found. Please ensure you have a connected Remote Network Agent: %s%s", r.CloudConsoleBaseUrl, "/configuration/agents")})
	}

	return r
}
