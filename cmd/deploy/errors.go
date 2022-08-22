package deploy

import (
	"errors"
)

var (
	ErrYamlFileRead                  = errors.New("error trying to read the YAML file")
	ErrInvalidDeploymentObject       = errors.New("error invalid deployment object")
	ErrDeploymentObjectConversion    = errors.New("error converting deployment object")
	ErrDeploymentStatusResponseParse = errors.New("error trying to parse response")
	ErrDeploymentStatusRequest       = errors.New("request returned an error: status code(%d)")
	ErrNoApplicationNameDefined      = errors.New("application name must be defined in deployment file or by application opt")
)
