package org

import (
	"errors"
	"fmt"
)

var (
	ErrUnauthorized           = errors.New("unauthorized. Run `armory login` to ensure you're using the correct tenant")
	EnvironmentRetrievalError = errors.New("error retrieving environment to login to")
	ErrRNARetrieval           = errors.New("error retrieving Remote Network Agents to connect with")
)

func newEnvironmentRetrievalError(errorID string, errors *[]AppError) error {
	return fmt.Errorf("%w. ErrorId: %s, Desc: %v", EnvironmentRetrievalError, errorID, errors)
}

func newRNARetrievalError(errorID string, errors *[]AppError) error {
	return fmt.Errorf("%w. ErrorId: %s, Desc: %v", ErrRNARetrieval, errorID, errors)
}
