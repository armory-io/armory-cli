package deploy

import (
	"errors"
)

var (
	ErrMissingMetricsProvider         = errors.New("metric provider must be provided either in the analysis config, as defaultMetricProviderName, or in the query as metricProviderName")
	ErrTargetsNotSpecified            = errors.New("please omit targets to include the manifests for all targets or specify the targets")
	ErrMissingQueryConfig             = errors.New("query in step does not exist in top-level analysis config")
	ErrManifestFileRead               = errors.New("error trying to read manifest file")
	ErrInvalidTrafficManagementConfig = errors.New("invalid traffic management config")
	ErrManifestFileNameRead           = errors.New("error trying to read manifest file name")
	ErrMinDeployConfigTimeout         = errors.New("invalid deployment config: timeout must be equal to or greater than 1 minute")
	ErrorNoStrategyDeployment         = errors.New("invalid deployment: strategy required for Deployment kind manifests")
	ErrorBadObject                    = errors.New("invalid deployment: manifest is not valid Kubernetes object")
)
