package deploy

import (
	"errors"
	"fmt"
)

const yamlFileReadErrorText = "error trying to read the YAML file: %w"
const invalidDeploymentObjectErrorText = "error invalid deployment object: %w"
const deploymentObjectConversionErrorText = "error converting deployment object: %w"
const deploymentStatusResponseParseErrorText = "error trying to parse response: %w"
const deploymentStatusRequestErrorText = "request returned an error: status code(%d) %w"

var ErrNoApplicationNameDefined = errors.New("application name must be defined in deployment file or by application opt")

func newYamlFileReadError(err error) error {
	return fmt.Errorf(yamlFileReadErrorText, err)
}

func newInvalidDeploymentObjectError(err error) error {
	return fmt.Errorf(invalidDeploymentObjectErrorText, err)
}

func newDeploymentObjectConversionError(err error) error {
	return fmt.Errorf(deploymentObjectConversionErrorText, err)
}

func newDeploymentStatusResponseParseError(err error) error {
	return fmt.Errorf(deploymentStatusResponseParseErrorText, err)
}

func newDeploymentStatusRequestError(statusCode int, err error) error {
	return fmt.Errorf(deploymentStatusRequestErrorText, statusCode, err)
}
