package org

import (
	"errors"
)

var (
	ErrUnauthorized         = errors.New("unauthorized. Run `armory login` to ensure you're using the correct tenant")
	ErrEnvironmentRetrieval = errors.New("error retrieving environment to login to")
	ErrRNARetrieval         = errors.New("error retrieving Remote Network Agents to connect with")
)
