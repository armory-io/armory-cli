package sourceControl

type SourceControlManager string
type Reference string
type Event string

const (
	gitHub      SourceControlManager = "gitHub"
	bitbucket   SourceControlManager = "bitbucket"
	branch      Reference            = "branch"
	tagRef      Reference            = "tag"
	pullRequest Event                = "pull_request"
)

type ScmContext struct {
	scm             SourceControlManager
	event           Event
	reference       Reference
	source          string
	target          string
	actor           string
	triggeringActor string
	prTitle         string
	prUrl           string
	sha             string
	repository      string
}
