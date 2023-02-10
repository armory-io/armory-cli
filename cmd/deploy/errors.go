package deploy

import (
	"errors"
)

var (
	ErrYAMLFileRead                        = errors.New("error trying to read the YAML file")
	ErrInvalidDeploymentObject             = errors.New("error invalid deployment object")
	ErrDeploymentStatusResponseParse       = errors.New("error trying to parse response")
	ErrDeploymentStatusRequest             = errors.New("request returned an error")
	ErrApplicationNameOverrideNotSupported = errors.New("application name override not supported when using a URL as your deployment configuration file")
)
