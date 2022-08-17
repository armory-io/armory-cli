package deploy

import (
	"errors"
	"fmt"
)

const (
	errYamlFileReadText                  = "error trying to read the YAML file: %w"
	errInvalidDeploymentObjectText       = "error invalid deployment object: %w"
	errDeploymentObjectConversionText    = "error converting deployment object: %w"
	errDeploymentStatusResponseParseText = "error trying to parse response: %w"
	errDeploymentStatusRequestText       = "request returned an error: status code(%d) %w"
)

var (
	ErrNoApplicationNameDefined = errors.New("application name must be defined in deployment file or by application opt")
)

func newYamlFileReadError(err error) error {
	return fmt.Errorf(errYamlFileReadText, err)
}

func newInvalidDeploymentObjectError(err error) error {
	return fmt.Errorf(errInvalidDeploymentObjectText, err)
}

func newDeploymentObjectConversionError(err error) error {
	return fmt.Errorf(errDeploymentObjectConversionText, err)
}

func newDeploymentStatusResponseParseError(err error) error {
	return fmt.Errorf(errDeploymentStatusResponseParseText, err)
}

func newDeploymentStatusRequestError(statusCode int, err error) error {
	return fmt.Errorf(errDeploymentStatusRequestText, statusCode, err)
}
