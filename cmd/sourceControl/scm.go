package sourceControl

type Manager string
type Reference string
type Event string

const (
	gitHub          Manager   = "gitHub"
	bitbucket       Manager   = "bitbucket"
	branch          Reference = "branch"
	tagRef          Reference = "tag"
	pullRequest     Event     = "pull_request"
	push            Event     = "push"
	workfloDispatch Event     = "workflow_dispatch"
)

type ScmContext struct {
	scm             Manager
	event           Event
	reference       Reference
	referenceName   string
	source          string
	target          string
	actor           string
	triggeringActor string
	prTitle         string
	prUrl           string
	sha             string
	repository      string
}
