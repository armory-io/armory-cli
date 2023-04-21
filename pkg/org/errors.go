package org

import (
	"errors"
)

var (
	ErrUnauthorized      = errors.New("unauthorized. Run `armory login` to ensure you're using the correct tenant")
	ErrRNARetrieval      = errors.New("error retrieving Remote Network Agents to connect with")
	ErrNoConnectedAgents = errors.New("unable to retrieve Remote Network Agents to connect with; please ensure a Remote Network Agent is connected and try again")
)
