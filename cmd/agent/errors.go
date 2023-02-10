package agent

import "errors"

var (
	ErrDuplicateAgent        = errors.New("Sorry, there's already an agent with that name in your tenant")
	ErrRoleMissing           = errors.New("The default role Remote Network Agent role was missing, please ask your tenant admins to recreate it.")
	ErrAgentAlreadyInstalled = errors.New("Sorry, there’s already an agent installed in this namespace.")
)
