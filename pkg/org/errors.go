package org

import (
	"errors"
	"fmt"
)

var ErrUnauthorized = errors.New("unauthorized. Run `armory login` to ensure you're using the correct tenant")
var EnvironmentRetrievalError = errors.New("error retrieving environment to login to")
var RNARetrievalError = errors.New("Error retrieving Remote Network Agents to connect with")

func newEnvironmentRetrievalError(errorId string, errors *[]AppError) error {
	return fmt.Errorf("%w. ErrorId: %s, Desc: %v", EnvironmentRetrievalError, errorId, errors)
}

func newRNARetrievalError(errorId string, errors *[]AppError) error {
	return fmt.Errorf("%w. ErrorId: %s, Desc: %v", RNARetrievalError, errorId, errors)
}
