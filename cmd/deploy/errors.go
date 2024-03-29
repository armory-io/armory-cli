package deploy

import (
	"errors"
)

var (
	ErrConfigurationRequired                = errors.New("a configuration file must be provided if not specifying a previously run pipelineId to redeploy")
	ErrTwoDeploymentConfigurationsSpecified = errors.New("when providing a pipelineId, do not provide a configuration file. The same configuration will be used to redeploy that pipeline")
	ErrYAMLFileRead                         = errors.New("error trying to read the YAML file")
	ErrInvalidDeploymentObject              = errors.New("error invalid deployment object")
	ErrDeploymentStatusResponseParse        = errors.New("error trying to parse response")
	ErrDeploymentStatusRequest              = errors.New("request returned an error")
	ErrApplicationNameOverrideNotSupported  = errors.New("application name override not supported when using a URL as your deployment configuration file")
)
