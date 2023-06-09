package deploy

import (
	"errors"
)

var (
	ErrManifestFileRead         = errors.New("error trying to read manifest file")
	ErrManifestFileNameRead     = errors.New("error trying to read manifest file name")
	ErrNoApplicationNameDefined = errors.New("application name must be defined in deployment file or by application opt")
	ErrNoKind                   = errors.New("deployment YAML must contain 'kind' property")
)
