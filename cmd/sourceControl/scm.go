package sourceControl

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type Manager string
type Reference string
type Event string

const (
	github           Manager   = "github"
	bitbucket        Manager   = "bitbucket"
	branch           Reference = "branch"
	tagRef           Reference = "tag"
	pullRequest      Event     = "pull_request"
	push             Event     = "push"
	workflowDispatch Event     = "workflow_dispatch"
)

type ScmContext interface {
	GetContext() (ScmContext, error)
}

type BaseScmc struct {
	Type                Manager   `json:"type,omitempty"`
	Event               Event     `json:"event,omitempty"`
	Reference           Reference `json:"reference,omitempty"`
	ReferenceName       string    `json:"referenceName,omitempty"`
	Source              string    `json:"source,omitempty"`
	Target              string    `json:"target,omitempty"`
	Principal           string    `json:"principal,omitempty"`
	TriggeringPrincipal string    `json:"triggeringPrincipal,omitempty"`
	PrTitle             string    `json:"prTitle,omitempty"`
	PrUrl               string    `json:"prUrl,omitempty"`
	Sha                 string    `json:"sha,omitempty"`
	Repository          string    `json:"repository,omitempty"`
	Server              string    `json:"server,omitempty"`
}

func RetrieveScmData(cmd *cobra.Command) ScmContext {
	var scmc ScmContext
	var err error

	scmc, err = GithubScmc{}.GetContext()

	if err != nil {
		msg := color.New(color.FgYellow, color.Bold).Sprint("scm error: ")
		fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n\n", msg, err)
	}

	return scmc
}
