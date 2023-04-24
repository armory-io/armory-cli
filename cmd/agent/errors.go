package agent

import "errors"

var (
	ErrDuplicateAgent        = errors.New("sorry, there's already an agent with that name in your tenant")
	ErrRoleMissing           = errors.New("the default role Remote Network Agent role was missing, please ask your tenant admins to recreate it")
	ErrAgentAlreadyInstalled = errors.New("sorry, there’s already an agent installed in this namespace")
)
